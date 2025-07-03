package cmd

import (
	"fmt"
	"github.com/pete911/aws-vpn/internal/cmd/out"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "list OpenVPN EC2 instances",
		Long:  "",
		Run:   runList,
	}
)

func init() {
	Root.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) {
	logger := NewLogger()
	client := NewClient(logger)

	instances, err := client.List()
	if err != nil {
		fmt.Printf("list instances: %v\n", err)
		os.Exit(1)
	}

	table := out.NewTable(logger, os.Stdout)
	table.AddRow("ID", "NAME", "HOST", "PUBLIC IP", "PRIVATE IP", "TYPE", "LAUNCH TIME")
	for _, instance := range instances {
		table.AddRow(
			instance.Id,
			instance.Name,
			instance.PublicDnsName,
			instance.PublicIp,
			instance.PrivateIp,
			instance.InstanceType,
			instance.LaunchTime.Format(time.RFC822),
		)
	}
	table.Print()
}
