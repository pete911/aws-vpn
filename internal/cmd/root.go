package cmd

import (
	"context"
	"fmt"
	"github.com/pete911/aws-vpn/internal/aws"
	"github.com/pete911/aws-vpn/internal/cmd/flag"
	"github.com/pete911/aws-vpn/internal/cmd/prompt"
	"github.com/pete911/aws-vpn/internal/vpn"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"strings"
	"time"
)

var (
	Root      = &cobra.Command{}
	logLevels = map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}
	Version   string
)

func init() {
	flag.InitPersistentFlags(Root)
}

func NewLogger() *slog.Logger {
	if level, ok := logLevels[strings.ToLower(flag.LogLevel)]; ok {
		opts := &slog.HandlerOptions{Level: level}
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	}

	fmt.Printf("invalid log level %s", flag.LogLevel)
	os.Exit(1)
	return nil
}

func NewClient(logger *slog.Logger) vpn.Client {
	// prompt region if user did not select any and set it on client
	if flag.Region == "" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		regions, err := aws.ListOptedInRegions(ctx, logger)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		i, _ := prompt.Select("region", regions.Names())
		selectedRegionCode := regions[i].Code
		flag.Region = selectedRegionCode
	}

	awsClient, err := aws.NewClient(logger, flag.Region)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// TODO - test and finish wireguard
	return vpn.NewClient(logger, awsClient, vpn.OpenVpn)
}

// SelectInstance either verifies if supplied instance name exists, or prompts user to select instance if argument is empty
func SelectInstance(client vpn.Client, instanceName string) aws.Instance {
	instances, err := client.List()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !strings.HasPrefix(instanceName, vpn.NamePrefix) {
		instanceName = vpn.NamePrefix + instanceName
	}

	// name has not been provided, we only have prefix
	if instanceName != vpn.NamePrefix {
		for _, i := range instances {
			if i.Name == instanceName {
				return i
			}
		}
		fmt.Printf("instance %s not found\n", instanceName)
		os.Exit(1)
	}

	i, _ := prompt.Select("instance", instances.Names())
	return instances[i]
}
