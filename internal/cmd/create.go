package cmd

import (
	"fmt"
	"github.com/pete911/aws-vpn/internal/cmd/flag"
	"github.com/pete911/aws-vpn/internal/ip"
	"github.com/spf13/cobra"
	"os"
)

var (
	createInboundCidrFlag string

	createCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "create OpenVPN EC2 instance",
		Long:  "",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run:   runCreate,
	}
)

func init() {
	Root.AddCommand(createCmd)
	createCmd.Flags().StringVar(
		&createInboundCidrFlag,
		"inbound-cidr",
		flag.GetStringEnv("INBOUND_CIDR", ""),
		"inbound cidr for security group",
	)
}

func runCreate(cmd *cobra.Command, args []string) {
	name := args[0]

	if createInboundCidrFlag == "" {
		myip, err := ip.MyIp()
		if err != nil {
			fmt.Printf("get my ip: %v\n", err)
			os.Exit(1)
		}
		createInboundCidrFlag = fmt.Sprintf("%s/32", myip)
	}

	logger := NewLogger()
	client := NewClient(logger)
	instance, err := client.Create(name, createInboundCidrFlag)
	if err != nil {
		fmt.Printf("create %s VPN: %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("VPN instance %s created\n", instance.Id)
	fmt.Printf("    public dns %s\n", instance.PublicDnsName)
	fmt.Printf("    public IP  %s\n", instance.PublicIp)
}
