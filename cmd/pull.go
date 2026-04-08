package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/m-hammad-faisal/envguard/crypto"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Decrypt the shared secrets.enc and hydrate your local .env",
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	encrypted, err := os.ReadFile(".envguard/secrets.enc")
	if err != nil {
		return fmt.Errorf(
			"failed to read .envguard/secrets.enc: %w\nHint: has the team run 'envguard push' and committed the file?",
			err,
		)
	}

	passphrase, err := promptPassphrase("Enter team passphrase: ")
	if err != nil {
		return fmt.Errorf("failed to read passphrase: %w", err)
	}

	plaintext, err := crypto.Decrypt(encrypted, passphrase)
	if err != nil {
		return err
	}

	if err := os.WriteFile(".env", plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write .env: %w", err)
	}
	fmt.Println("✓ .env hydrated successfully")

	// Silently ensure the pre-commit hook is present for devs who just cloned
	if err := installHook(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not install pre-commit hook: %v\n", err)
	}

	return nil
}
