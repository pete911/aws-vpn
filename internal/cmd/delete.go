package cmd

import (
	"fmt"
	"github.com/pete911/aws-vpn/internal/cmd/prompt"
	"github.com/spf13/cobra"
	"os"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete OpenVPN EC2 instance",
		Long:  "",
		Run:   runDelete,
	}
)

func init() {
	Root.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	logger := NewLogger()
	client := NewClient(logger)
	instance := SelectInstance(client, name)
	if !prompt.Prompt(fmt.Sprintf("delete %s VPN instance in %s region", instance, client.Region)) {
		return
	}

	if err := client.Delete(instance); err != nil {
		fmt.Printf("delete %s VPN: %v\n", instance.Name, err)
		os.Exit(1)
	}
}
