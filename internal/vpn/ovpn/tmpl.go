package ovpn

import (
	"text/template"
)

func clientConfigTemplate() (*template.Template, error) {
	txt := `client
dev tun
proto udp
remote {{ .Hostname }} 1194
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
cipher AES-256-GCM
auth SHA256
key-direction 1
verb 3

<ca>
{{ .CaCert }}
</ca>

<cert>
{{ .ClientCert }}
</cert>

<key>
{{ .ClientKey }}
</key>

<tls-crypt>
{{ .TaKey }}
</tls-crypt>`
	return template.New("vpn-client-config").Parse(txt)
}

func serverUserDataTemplate() (*template.Template, error) {
	txt := `MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="==MYBOUNDARY=="

--==MYBOUNDARY==
Content-Type: text/x-shellscript; charset="us-ascii"

#!/bin/bash -xe
echo "--- user data start ---"

yum update
yum install -y openvpn
yum install -y iptables

# setup routing

default_ni=$(ip route list default | cut -d ' ' -f5)
iptables -t nat -A POSTROUTING -s 10.4.0.1/2 -o $default_ni -j MASQUERADE
iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -o $default_ni -j MASQUERADE

echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf
sysctl -p

# setup easy rsa
mkdir /etc/openvpn/easy-rsa
cd /etc/openvpn/easy-rsa
wget -O easy-rsa.tgz https://github.com/OpenVPN/easy-rsa/releases/download/v3.2.2/EasyRSA-3.2.2.tgz
tar -xvf easy-rsa.tgz
mv EasyRSA-3.2.2/x509-types .
mv EasyRSA-3.2.2/easyrsa .
rm easy-rsa.tgz
rm -r EasyRSA-3.2.2

echo 'set_var EASYRSA_ALGO "ec"' >> vars
echo 'set_var EASYRSA_DIGEST "sha512"' >> vars
./easyrsa init-pki

./easyrsa --batch build-ca nopass
./easyrsa --batch gen-req server nopass
./easyrsa --batch sign-req server server
./easyrsa --batch gen-req client nopass
./easyrsa --batch sign-req client client

mv /etc/openvpn/easy-rsa/pki/ca.crt /etc/openvpn/ca.crt
mv /etc/openvpn/easy-rsa/pki/issued/server.crt /etc/openvpn/server/server.crt
mv /etc/openvpn/easy-rsa/pki/private/server.key /etc/openvpn/server/server.key
mv /etc/openvpn/easy-rsa/pki/issued/client.crt /etc/openvpn/client/client.crt
mv /etc/openvpn/easy-rsa/pki/private/client.key /etc/openvpn/client/client.key

cd /etc/openvpn
openvpn --genkey secret ta.key

chown -R openvpn:openvpn /etc/openvpn

# create openvpn config, sample is located at /usr/share/doc/openvpn/sample/sample-config-files/server.conf
cat > /etc/openvpn/server/server.conf << 'EOF'
port 1194
proto udp
dev tun
ca /etc/openvpn/ca.crt
cert /etc/openvpn/server/server.crt
key /etc/openvpn/server/server.key
dh none
topology subnet
server 10.8.0.0 255.255.255.0
push "redirect-gateway def1 bypass-dhcp"
push "dhcp-option DNS 8.8.8.8"
push "dhcp-option DNS 8.8.4.4"
ifconfig-pool-persist ipp.txt
keepalive 10 120
persist-key
persist-tun
status openvpn-status.log
verb 3

tls-crypt /etc/openvpn/ta.key
cipher AES-256-GCM
auth SHA256
user openvpn
group openvpn
EOF

systemctl enable openvpn-server@server.service
systemctl start openvpn-server@server.service

aws secretsmanager create-secret --name {{ .SecretsPath }}/ta.key --secret-string file:///etc/openvpn/ta.key
aws secretsmanager create-secret --name {{ .SecretsPath }}/ca.crt --secret-string file:///etc/openvpn/ca.crt
aws secretsmanager create-secret --name {{ .SecretsPath }}/client.key --secret-string file:///etc/openvpn/client/client.key
aws secretsmanager create-secret --name {{ .SecretsPath }}/client.crt --secret-string file:///etc/openvpn/client/client.crt

echo "--- user data end ---"
--==MYBOUNDARY==--\
`
	return template.New("openvpn-user-data").Parse(txt)
}
