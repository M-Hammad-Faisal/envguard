package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "envguard",
	Version: version,
	Short:   "EnvGuard — secure environment variable management for teams",
	Long:    "EnvGuard encrypts your .env file for team sharing and prevents hardcoded secrets from reaching your Git history.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
