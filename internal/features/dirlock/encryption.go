package dirlock

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mbeniwal-imwe/ark/internal/core/crypto"
)

// EncryptDirectory encrypts all files in a directory
func EncryptDirectory(dirPath string, key []byte) error {
	// Create encrypted archive
	archivePath := dirPath + ".ark_encrypted"
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Walk directory and encrypt files
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Encrypt content
		encrypted, err := encryptContent(content, key)
		if err != nil {
			return err
		}

		// Add to zip
		relPath, _ := filepath.Rel(dirPath, path)
		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = zipFile.Write(encrypted)
		return err
	})

	if err != nil {
		return err
	}

	// Remove original files
	err = os.RemoveAll(dirPath)
	if err != nil {
		return err
	}

	// Rename archive to original directory name
	return os.Rename(archivePath, dirPath)
}

// DecryptDirectory decrypts all files in a directory
func DecryptDirectory(dirPath string, key []byte) error {
	// Check if directory is encrypted
	if !isEncryptedDirectory(dirPath) {
		return fmt.Errorf("directory is not encrypted")
	}

	// Create temporary directory
	tempDir := dirPath + ".ark_temp"
	err := os.MkdirAll(tempDir, 0700)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Open encrypted archive
	file, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return err
	}

	// Extract and decrypt files
	for _, zipFile := range zipReader.File {
		// Read encrypted content
		rc, err := zipFile.Open()
		if err != nil {
			return err
		}

		encrypted, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}

		// Decrypt content
		decrypted, err := decryptContent(encrypted, key)
		if err != nil {
			return err
		}

		// Write decrypted file
		filePath := filepath.Join(tempDir, zipFile.Name)
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return err
		}

		err = os.WriteFile(filePath, decrypted, 0644)
		if err != nil {
			return err
		}
	}

	// Remove encrypted directory
	err = os.RemoveAll(dirPath)
	if err != nil {
		return err
	}

	// Rename temp directory to original name
	return os.Rename(tempDir, dirPath)
}

// encryptContent encrypts file content using AES-256-GCM
func encryptContent(content []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, content, nil)
	return ciphertext, nil
}

// decryptContent decrypts file content using AES-256-GCM
func decryptContent(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// isEncryptedDirectory checks if a directory is encrypted
func isEncryptedDirectory(dirPath string) bool {
	// Check if it's a zip file (encrypted archive)
	file, err := os.Open(dirPath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}

	// Try to read as zip
	_, err = zip.NewReader(file, fileInfo.Size())
	return err == nil
}

// deriveKeyFromPassword derives encryption key from password
func deriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	return crypto.DeriveKey(password, salt)
}
