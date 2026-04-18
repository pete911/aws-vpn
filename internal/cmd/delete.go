package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pete911/aws-vpn/internal/cmd/prompt"
	"github.com/spf13/cobra"
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
	instance := SelectInstance(cmd.Context(), client, name)
	if !prompt.Prompt(fmt.Sprintf("delete %s VPN instance in %s region", instance, client.Region)) {
		return
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*300)
	defer cancel()

	if err := client.Delete(ctx, instance); err != nil {
		fmt.Printf("delete %s VPN: %v\n", instance.Name, err)
		os.Exit(1)
	}
}
