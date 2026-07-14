// Package cryptoutil provides AES-256-GCM encryption for secrets that must
// be stored at rest, not just passed through -- currently just
// postmark_servers.api_token_encrypted (see internal/provisioning). Not for
// passwords: those are never listnun's to store, since Zitadel and the
// listmonk fork each own their own credential storage.
package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidKeySize is returned by ParseKey when the decoded key isn't
// exactly 32 bytes (AES-256).
var ErrInvalidKeySize = errors.New("cryptoutil: key must decode to 32 bytes")

// ParseKey decodes a base64-encoded 32-byte AES-256 key, e.g. one generated
// with `openssl rand -base64 32`.
func ParseKey(b64Key string) ([32]byte, error) {
	var key [32]byte
	decoded, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return key, fmt.Errorf("cryptoutil: decode key: %w", err)
	}
	if len(decoded) != 32 {
		return key, ErrInvalidKeySize
	}
	copy(key[:], decoded)
	return key, nil
}

// Encrypt returns a base64-encoded (nonce || ciphertext), ready to store as
// a single text column.
func Encrypt(key [32]byte, plaintext string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("cryptoutil: generate nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt.
func Decrypt(key [32]byte, encoded string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}

	sealed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("cryptoutil: decode ciphertext: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(sealed) < nonceSize {
		return "", errors.New("cryptoutil: ciphertext too short")
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("cryptoutil: decrypt: %w", err)
	}
	return string(plaintext), nil
}

func newGCM(key [32]byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cryptoutil: init cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil: init GCM: %w", err)
	}
	return gcm, nil
}
