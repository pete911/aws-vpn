package ovpn

import (
	"bytes"
	"fmt"
)

var InboundPort = 1194

func UserData(vpnConfig any) (string, error) {
	tmpl, err := serverUserDataTemplate()
	if err != nil {
		return "", fmt.Errorf("parse openvpn template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, vpnConfig); err != nil {
		return "", fmt.Errorf("execute openvpn template: %w", err)
	}
	return buf.String(), nil
}

type ClientConfig struct {
	Hostname   string
	TaKey      string
	CaCert     string
	ClientCert string
	ClientKey  string
}

func NewClientConfig(hostname string, secrets map[string]string) ClientConfig {
	return ClientConfig{
		Hostname:   hostname,
		TaKey:      secrets["/ta.key"],
		CaCert:     secrets["/ca.crt"],
		ClientCert: secrets["/client.crt"],
		ClientKey:  secrets["/client.key"],
	}
}

func (c ClientConfig) Parse() ([]byte, error) {
	tmpl, err := clientConfigTemplate()
	if err != nil {
		return nil, fmt.Errorf("parse openvpn client config: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, c); err != nil {
		return nil, fmt.Errorf("execute openvpn client config template: %w", err)
	}
	return buf.Bytes(), nil
}
