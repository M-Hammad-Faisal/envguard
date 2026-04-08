package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

// Binary layout of secrets.enc:
// [1 byte version][16 bytes salt][4 bytes time uint32 LE][4 bytes memory uint32 LE][1 byte threads][12 bytes nonce][ciphertext...]
//
// Storing Argon2id params alongside the salt means future param changes
// never break decryption of existing files.

const (
	// Version is the current binary format version. Increment when the layout changes.
	Version byte = 1

	saltSize  = 16
	nonceSize = 12
	keySize   = 32

	// OWASP minimum recommended Argon2id parameters (2023)
	defaultTime    uint32 = 1
	defaultMemory  uint32 = 64 * 1024
	defaultThreads uint8  = 4

	// headerSize = version(1) + salt(16) + time(4) + memory(4) + threads(1) + nonce(12)
	headerSize = 1 + saltSize + 4 + 4 + 1 + nonceSize
)

var (
	ErrInvalidCiphertext  = errors.New("ciphertext is too short or corrupted")
	ErrDecryptionFailed   = errors.New("decryption failed: incorrect passphrase or corrupted data")
	ErrUnsupportedVersion = errors.New("unsupported secrets.enc version: re-run 'envguard push' to re-encrypt")
)

type argon2Params struct {
	time    uint32
	memory  uint32
	threads uint8
}

// Encrypt encrypts plaintext using AES-256-GCM with Argon2id key derivation.
// Returns the full binary payload including version header, salt, params, nonce, and ciphertext.
func Encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	p := argon2Params{time: defaultTime, memory: defaultMemory, threads: defaultThreads}
	key := argon2.IDKey([]byte(passphrase), salt, p.time, p.memory, p.threads, keySize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	header := make([]byte, 0, headerSize)
	header = append(header, Version)
	header = append(header, salt...)
	header = binary.LittleEndian.AppendUint32(header, p.time)
	header = binary.LittleEndian.AppendUint32(header, p.memory)
	header = append(header, p.threads)
	header = append(header, nonce...)

	return append(header, ciphertext...), nil
}

// Decrypt decrypts a payload produced by Encrypt.
// Returns ErrDecryptionFailed on wrong passphrase or corrupted data.
func Decrypt(data []byte, passphrase string) ([]byte, error) {
	if len(data) < headerSize {
		return nil, ErrInvalidCiphertext
	}

	if data[0] != Version {
		return nil, ErrUnsupportedVersion
	}

	offset := 1
	salt := data[offset : offset+saltSize]
	offset += saltSize

	var p argon2Params
	p.time = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	p.memory = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	p.threads = data[offset]
	offset++

	nonce := data[offset : offset+nonceSize]
	offset += nonceSize

	ciphertext := data[offset:]
	if len(ciphertext) == 0 {
		return nil, ErrInvalidCiphertext
	}

	key := argon2.IDKey([]byte(passphrase), salt, p.time, p.memory, p.threads, keySize)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}
