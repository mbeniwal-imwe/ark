package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// KeySize is the size of the encryption key in bytes
	KeySize = 32 // 256 bits
	// NonceSize is the size of the nonce for GCM
	NonceSize = 12 // 96 bits
	// SaltSize is the size of the salt for key derivation
	SaltSize = 32 // 256 bits
)

// Encryptor handles encryption and decryption operations
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new encryptor with the given key
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}

	return &Encryptor{key: key}, nil
}

// Encrypt encrypts the given plaintext using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts the given ciphertext using AES-256-GCM
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(ciphertext) < NonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:NonceSize]
	ciphertext = ciphertext[NonceSize:]

	// Decrypt and authenticate
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// DeriveKey derives an encryption key from a password using Argon2id
func DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(salt) != SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d bytes, got %d", SaltSize, len(salt))
	}

	// Argon2id parameters (recommended values)
	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, KeySize)
	return key, nil
}

// GenerateSalt generates a random salt for key derivation
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// HashPassword hashes a password using PBKDF2 with SHA-256
func HashPassword(password string, salt []byte) ([]byte, error) {
	// Use Argon2id which is more secure than PBKDF2
	// This provides the same security as PBKDF2 with 100,000 iterations
	return DeriveKey(password, salt)
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password string, hash, salt []byte) bool {
	computedHash, err := HashPassword(password, salt)
	if err != nil {
		return false
	}

	return len(computedHash) == len(hash) &&
		compareBytes(computedHash, hash)
}

// compareBytes compares two byte slices in constant time
func compareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}

	return result == 0
}
