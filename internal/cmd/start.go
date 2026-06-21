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
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "start OpenVPN EC2 instances",
		Long:  "",
		Run:   runStart,
	}
)

func init() {
	Root.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}

	logger := NewLogger()
	client := NewClient(logger)
	instance := SelectInstance(cmd.Context(), client, name, "stopped")
	if !prompt.Prompt(fmt.Sprintf("start %s VPN instance in %s region", instance.Name, client.Region)) {
		return
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*300)
	defer cancel()

	if err := client.Start(ctx, instance); err != nil {
		fmt.Printf("start %s VPN: %v\n", instance.Name, err)
		os.Exit(1)
	}
}
