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
	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop OpenVPN EC2 instances",
		Long:  "",
		Run:   runStop,
	}
)

func init() {
	Root.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	logger := NewLogger()
	client := NewClient(logger)
	instance := SelectInstance(cmd.Context(), client, name, "running")
	if !prompt.Prompt(fmt.Sprintf("stop %s VPN instance in %s region", instance.Name, client.Region)) {
		return
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*300)
	defer cancel()

	if err := client.Stop(ctx, instance); err != nil {
		fmt.Printf("stop %s VPN: %v\n", instance.Name, err)
		os.Exit(1)
	}
}
