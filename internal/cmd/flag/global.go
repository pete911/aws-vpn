package flag

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Region   string
	LogLevel string
)

func InitPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&Region,
		"region",
		GetStringEnv("REGION", ""),
		"aws region",
	)
	cmd.PersistentFlags().StringVar(
		&LogLevel,
		"log-level",
		GetStringEnv("LOG", "debug"),
		"log level - debug, info, warn, error",
	)
}

func GetStringEnv(envName string, defaultValue string) string {
	env, ok := os.LookupEnv(fmt.Sprintf("AWS_VPN_%s", envName))
	if !ok {
		return defaultValue
	}
	return env
}
