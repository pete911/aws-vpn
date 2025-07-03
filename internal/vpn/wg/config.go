package wg

import (
	"bytes"
	"fmt"
)

var InboundPort = 51820

func UserData(vpnConfig any) (string, error) {
	tmpl, err := serverUserDataTemplate()
	if err != nil {
		return "", fmt.Errorf("parse vpn template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, vpnConfig); err != nil {
		return "", fmt.Errorf("execute vpn template: %w", err)
	}
	return buf.String(), nil
}
