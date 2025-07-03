package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "download client config for OpenVPN EC2 instance",
		Long:  "",
		Run:   runConfig,
	}
)

func init() {
	Root.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	logger := NewLogger()
	client := NewClient(logger)

	instance := SelectInstance(client, name)
	b, err := client.GetClientConfig(instance)
	if err != nil {
		fmt.Printf("clent config for %s VPN: %v\n", instance.Name, err)
		os.Exit(1)
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("get home dir: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(dirname, fmt.Sprintf("%s-%s.ovpn", client.Region, instance.Name))
	if err := os.WriteFile(configPath, b, 0644); err != nil {
		fmt.Printf("write VPN config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("VPN config has been saved in %s\n", configPath)
}
