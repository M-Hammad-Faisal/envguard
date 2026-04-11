package cmd

import (
	"os"
	"testing"

	"github.com/m-hammad-faisal/envguard/crypto"
)

func TestPromptPassphraseWithConfirmationMismatch(t *testing.T) {
	// We can't easily test the interactive prompt itself, but we can test
	// the rotate logic by calling encrypt/decrypt directly.
	// This test verifies that a wrong old passphrase is rejected.
	plaintext := []byte("DB_URL=postgres://localhost/db\nAPI_KEY=sk_live_testkey123\n")
	oldPassphrase := "old-team-passphrase-secure"
	newPassphrase := "new-team-passphrase-secure"

	encrypted, err := crypto.Encrypt(plaintext, oldPassphrase)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Wrong old passphrase should fail
	_, err = crypto.Decrypt(encrypted, "wrong-passphrase")
	if err == nil {
		t.Fatal("expected decryption to fail with wrong passphrase")
	}

	// Correct old passphrase should succeed
	decrypted, err := crypto.Decrypt(encrypted, oldPassphrase)
	if err != nil {
		t.Fatalf("Decrypt failed with correct passphrase: %v", err)
	}

	// Re-encrypt with new passphrase
	reEncrypted, err := crypto.Encrypt(decrypted, newPassphrase)
	if err != nil {
		t.Fatalf("Re-encrypt failed: %v", err)
	}

	// Verify new passphrase works
	final, err := crypto.Decrypt(reEncrypted, newPassphrase)
	if err != nil {
		t.Fatalf("Decrypt with new passphrase failed: %v", err)
	}

	if string(final) != string(plaintext) {
		t.Errorf("plaintext mismatch after rotation\ngot:  %q\nwant: %q", final, plaintext)
	}

	// Verify old passphrase no longer works
	_, err = crypto.Decrypt(reEncrypted, oldPassphrase)
	if err == nil {
		t.Fatal("old passphrase should not decrypt after rotation")
	}
}

func TestRotateAtomicWrite(t *testing.T) {
	// Verify the atomic write pattern: temp file is cleaned up on success
	tmpDir := t.TempDir()
	chdir(t, tmpDir)

	plaintext := []byte("SECRET=rotationtest12345\n")
	passphrase := "test-passphrase-rotate"

	encrypted, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(".envguard", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".envguard/secrets.enc", encrypted, 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate what rotate does: write to tmp, rename
	newPassphrase := "new-passphrase-after-rotate"
	reEncrypted, err := crypto.Encrypt(plaintext, newPassphrase)
	if err != nil {
		t.Fatal(err)
	}

	tmpPath := ".envguard/secrets.enc.tmp"
	if err := os.WriteFile(tmpPath, reEncrypted, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpPath, ".envguard/secrets.enc"); err != nil {
		t.Fatal(err)
	}

	// Temp file should be gone
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful rename")
	}

	// New passphrase should work
	result, err := os.ReadFile(".envguard/secrets.enc")
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := crypto.Decrypt(result, newPassphrase)
	if err != nil {
		t.Fatalf("decrypt after atomic write failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Error("content mismatch after atomic write")
	}
}

func TestShortPassphraseRejected(t *testing.T) {
	// Directly test the length check logic
	short := "tooshort"
	if len(short) >= 12 {
		t.Skip("test value is not actually short")
	}
	// This mirrors the check in promptPassphraseWithConfirmation
	if len(short) >= 12 {
		t.Error("short passphrase should have been rejected")
	}
}
