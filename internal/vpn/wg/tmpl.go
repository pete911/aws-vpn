package wg

import "text/template"

// TODO - https://dev.to/gabrieltetzner/setting-up-a-vpn-with-wireguard-server-on-aws-ec2-4a49
// follow these instructions but modify or amazon linux
func serverUserDataTemplate() (*template.Template, error) {
	txt := `MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="==MYBOUNDARY=="

--==MYBOUNDARY==
Content-Type: text/x-shellscript; charset="us-ascii"

#!/bin/bash -xe
echo "--- user data start ---"

yum update
yum install -y wireguard-tools
yum install -y iptables

touch /etc/wireguard/private.key
chmod go= /etc/wireguard/private.key
wg genkey > /etc/wireguard/private.key
cat /etc/wireguard/private.key | wg pubkey > /etc/wireguard/public.key

default_ni=$(ip route list default | cut -d ' ' -f5)

cat > /etc/wireguard/wg0.conf << EOF
[Interface]
Address = 192.168.2.1/24
MTU = 1420
PrivateKey = $(cat /etc/wireguard/private.key)
ListenPort = 51820
DNS = $(cat /etc/resolv.conf | grep 'nameserver ' | cut -d ' ' -f2)

PostUp = iptables -t nat -I POSTROUTING -o $default_ni -j MASQUERADE
PostUp = ip6tables -t nat -I POSTROUTING -o $default_ni -j MASQUERADE
PreDown = iptables -t nat -D POSTROUTING -o $default_ni -j MASQUERADE
PreDown = ip6tables -t nat -D POSTROUTING -o $default_ni -j MASQUERADE

[Peer] 
AllowedIPs = 192.168.2.2/32
PublicKey = $(cat /etc/wireguard/public.key)
EOF

aws secretsmanager create-secret --name {{ .SecretsPath }}/public.key --secret-string file:///etc/wireguard/public.key

echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf
sysctl -p

systemctl enable wg-quick@wg0.service
systemctl start wg-quick@wg0.service
systemctl status wg-quick@wg0.service

echo "--- user data end ---"
--==MYBOUNDARY==--\
`
	return template.New("wg-user-data").Parse(txt)
}
