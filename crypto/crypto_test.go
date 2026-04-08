package crypto_test

import (
	"bytes"
	"testing"

	"github.com/m-hammad-faisal/envguard/crypto"
)

func TestRoundtrip(t *testing.T) {
	original := []byte("DB_URL=postgres://user:s3cr3t@localhost/mydb\nAPI_KEY=sk_live_abc123xyz789\n")
	passphrase := "correct-horse-battery-staple"

	encrypted, err := crypto.Encrypt(original, passphrase)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(original, decrypted) {
		t.Errorf("roundtrip mismatch\ngot:  %q\nwant: %q", decrypted, original)
	}
}

func TestWrongPassphrase(t *testing.T) {
	encrypted, err := crypto.Encrypt([]byte("SECRET=hunter2secretvalue"), "correct")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = crypto.Decrypt(encrypted, "wrong")
	if err == nil {
		t.Fatal("expected error on wrong passphrase, got nil")
	}
	if err != crypto.ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got: %v", err)
	}
}

func TestTruncatedData(t *testing.T) {
	_, err := crypto.Decrypt([]byte{crypto.Version, 0x01, 0x02}, "passphrase")
	if err == nil {
		t.Fatal("expected error on truncated data")
	}
}

func TestUniqueOutputs(t *testing.T) {
	// Same plaintext + same passphrase must never produce the same ciphertext.
	// Proves the nonce and salt are both randomized correctly.
	plaintext := []byte("same content repeated for uniqueness test")
	passphrase := "same passphrase"

	enc1, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := crypto.Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(enc1, enc2) {
		t.Error("two encryptions of identical plaintext produced identical output — nonce or salt is not random")
	}
}

func TestCorruptedCiphertext(t *testing.T) {
	encrypted, err := crypto.Encrypt([]byte("some secret value that is long enough"), "passphrase")
	if err != nil {
		t.Fatal(err)
	}
	// Flip the last byte — corrupts the GCM authentication tag
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = crypto.Decrypt(encrypted, "passphrase")
	if err == nil {
		t.Fatal("expected error on corrupted ciphertext, GCM should reject it")
	}
}

func TestUnsupportedVersion(t *testing.T) {
	encrypted, err := crypto.Encrypt([]byte("secret value for version test"), "passphrase")
	if err != nil {
		t.Fatal(err)
	}
	encrypted[0] = 99 // corrupt the version byte

	_, err = crypto.Decrypt(encrypted, "passphrase")
	if err != crypto.ErrUnsupportedVersion {
		t.Errorf("expected ErrUnsupportedVersion, got: %v", err)
	}
}

func TestEmptyPlaintext(t *testing.T) {
	encrypted, err := crypto.Encrypt([]byte{}, "passphrase")
	if err != nil {
		t.Fatalf("Encrypt of empty plaintext failed: %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted, "passphrase")
	if err != nil {
		t.Fatalf("Decrypt of empty plaintext failed: %v", err)
	}
	if len(decrypted) != 0 {
		t.Errorf("expected empty decrypted output, got %q", decrypted)
	}
}

func TestLargePlaintext(t *testing.T) {
	// ~1MB payload — ensures no buffer sizing issues
	large := bytes.Repeat([]byte("KEY=value_that_is_long_enough\n"), 35000)
	encrypted, err := crypto.Encrypt(large, "passphrase")
	if err != nil {
		t.Fatalf("Encrypt of large payload failed: %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted, "passphrase")
	if err != nil {
		t.Fatalf("Decrypt of large payload failed: %v", err)
	}
	if !bytes.Equal(large, decrypted) {
		t.Error("large payload roundtrip mismatch")
	}
}
