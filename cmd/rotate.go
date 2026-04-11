package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/m-hammad-faisal/envguard/crypto"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Re-encrypt secrets.enc with a new team passphrase",
	Long: `Rotate the team passphrase by decrypting with the old passphrase
and re-encrypting with a new one. Run this when a team member leaves.`,
	RunE: runRotate,
}

func init() {
	rootCmd.AddCommand(rotateCmd)
}

func runRotate(cmd *cobra.Command, args []string) error {
	// Step 1: verify secrets.enc exists before asking for anything
	encryptedPath := ".envguard/secrets.enc"
	encrypted, err := os.ReadFile(encryptedPath)
	if err != nil {
		return fmt.Errorf(
			"failed to read .envguard/secrets.enc: %w\nHint: run 'envguard push' first",
			err,
		)
	}

	// Step 2: decrypt with old passphrase
	oldPassphrase, err := promptPassphrase("Enter current team passphrase: ")
	if err != nil {
		return fmt.Errorf("failed to read passphrase: %w", err)
	}

	plaintext, err := crypto.Decrypt(encrypted, oldPassphrase)
	if err != nil {
		// crypto.ErrDecryptionFailed has a clear message already
		return err
	}
	fmt.Println("✓ Decrypted with current passphrase")

	// Step 3: prompt for new passphrase with confirmation
	newPassphrase, err := promptPassphraseWithConfirmation()
	if err != nil {
		return err
	}

	// Step 4: re-encrypt with new passphrase
	reEncrypted, err := crypto.Encrypt(plaintext, newPassphrase)
	if err != nil {
		return fmt.Errorf("re-encryption failed: %w", err)
	}

	// Step 5: write back atomically — write to temp file first, then rename
	// This prevents a partial write from corrupting the only copy of secrets.enc
	tmpPath := encryptedPath + ".tmp"
	if err := os.WriteFile(tmpPath, reEncrypted, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	if err := os.Rename(tmpPath, encryptedPath); err != nil {
		// Clean up the temp file if rename fails
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace secrets.enc: %w", err)
	}

	fmt.Println("✓ Re-encrypted with new passphrase")
	fmt.Println("✓ Written to .envguard/secrets.enc")

	// Step 6: print explicit team instructions — rotation is useless without these
	fmt.Println(`
Next steps:
  1. Commit and push .envguard/secrets.enc
        git add .envguard/secrets.enc
        git commit -m "security: rotate team passphrase"
        git push

  2. Share the new passphrase with remaining team members out-of-band
     (Signal, 1Password, in person — never Slack or email)

  3. Each team member runs:
        envguard pull

  4. Confirm the old team member no longer has repository access.`)

	return nil
}

// promptPassphraseWithConfirmation prompts for a new passphrase twice and
// verifies they match. Returns an error if they don't match or either is empty.
func promptPassphraseWithConfirmation() (string, error) {
	newPassphrase, err := promptPassphrase("Enter new team passphrase: ")
	if err != nil {
		return "", fmt.Errorf("failed to read new passphrase: %w", err)
	}

	confirm, err := promptPassphrase("Confirm new team passphrase: ")
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase confirmation: %w", err)
	}

	if newPassphrase != confirm {
		return "", fmt.Errorf("passphrases do not match — rotation aborted, secrets.enc is unchanged")
	}

	if len(newPassphrase) < 12 {
		return "", fmt.Errorf("new passphrase is too short — use at least 12 characters")
	}

	return newPassphrase, nil
}
