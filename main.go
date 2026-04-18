package main

import (
	"os"

	"github.com/pete911/aws-vpn/internal/cmd"
)

var Version = "dev"

func main() {
	cmd.Version = Version
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
