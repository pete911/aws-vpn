package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "version",
		Long:  "",
		Run:   runVersion,
	}
)

func init() {
	Root.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) {
	fmt.Println(Version)
}
