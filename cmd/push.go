package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/m-hammad-faisal/envguard/crypto"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Encrypt your .env and stage it for team sharing",
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(".envguard"); os.IsNotExist(err) {
		return fmt.Errorf("EnvGuard is not initialized. Run 'envguard init' first")
	}

	plaintext, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("failed to read .env: %w\nHint: create a .env file or run 'envguard init'", err)
	}
	if len(plaintext) == 0 {
		return fmt.Errorf(".env file is empty — nothing to encrypt")
	}

	passphrase, err := promptPassphrase("Enter team passphrase: ")
	if err != nil {
		return fmt.Errorf("failed to read passphrase: %w", err)
	}

	encrypted, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	if err := os.WriteFile(".envguard/secrets.enc", encrypted, 0644); err != nil {
		return fmt.Errorf("failed to write secrets.enc: %w", err)
	}

	fmt.Println("✓ Encrypted and written to .envguard/secrets.enc")
	fmt.Println("  Commit .envguard/secrets.enc and .envguard/config.json to share with your team.")
	return nil
}

// promptPassphrase reads a hidden passphrase from stdin.
// Used by push and pull — both are interactive commands where stdin is a real TTY.
// For git hook context (scan), use promptPassphraseViaTTY instead.
func promptPassphrase(prompt string) (string, error) {
	fmt.Print(prompt)
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	if len(passBytes) == 0 {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	return string(passBytes), nil
}
