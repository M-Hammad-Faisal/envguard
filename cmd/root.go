package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:     "envguard",
	Version: "0.1.4",
	Short:   "EnvGuard — secure environment variable management for teams",
	Long:    "EnvGuard encrypts your .env file for team sharing and prevents hardcoded secrets from reaching your Git history.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
