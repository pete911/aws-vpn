package main

import (
	"github.com/pete911/aws-vpn/internal/cmd"
	"os"
)

var Version = "dev"

func main() {
	cmd.Version = Version
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
