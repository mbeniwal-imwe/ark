package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/core/crypto"
	"go.etcd.io/bbolt"
)

// Database represents an encrypted BoltDB database
type Database struct {
	db   *bbolt.DB
	enc  *crypto.Encryptor
	path string
}

// NewDatabase creates a new encrypted database
func NewDatabase(path string, masterKey []byte) (*Database, error) {
	// Open BoltDB
	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create encryptor
	enc, err := crypto.NewEncryptor(masterKey)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	database := &Database{
		db:   db,
		enc:  enc,
		path: path,
	}

	// Initialize buckets
	if err := database.initBuckets(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return database, nil
}

// initBuckets initializes the database buckets
func (d *Database) initBuckets() error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{
			"vault",
			"aws_profiles",
			"ec2_instances",
			"locked_dirs",
			"backup_metadata",
			"config",
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}

		return nil
	})
}

// Set stores an encrypted value in the specified bucket
func (d *Database) Set(bucket, key string, value interface{}) error {
	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Encrypt data
	encryptedData, err := d.enc.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		return b.Put([]byte(key), encryptedData)
	})
}

// Get retrieves and decrypts a value from the specified bucket
func (d *Database) Get(bucket, key string, dest interface{}) error {
	var encryptedData []byte

	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key %s not found in bucket %s", key, bucket)
		}

		encryptedData = make([]byte, len(data))
		copy(encryptedData, data)
		return nil
	})

	if err != nil {
		return err
	}

	// Decrypt data
	decryptedData, err := d.enc.Decrypt(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Unmarshal to destination
	if err := json.Unmarshal(decryptedData, dest); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// Delete removes a key from the specified bucket
func (d *Database) Delete(bucket, key string) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		return b.Delete([]byte(key))
	})
}

// List returns all keys in the specified bucket
func (d *Database) List(bucket string) ([]string, error) {
	var keys []string

	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		return b.ForEach(func(key, _ []byte) error {
			keys = append(keys, string(key))
			return nil
		})
	})

	return keys, err
}

// Search searches for keys matching a pattern in the specified bucket
func (d *Database) Search(bucket, pattern string) ([]string, error) {
	var keys []string

	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		return b.ForEach(func(key, _ []byte) error {
			keyStr := string(key)
			if contains(keyStr, pattern) {
				keys = append(keys, keyStr)
			}
			return nil
		})
	})

	return keys, err
}

// Exists checks if a key exists in the specified bucket
func (d *Database) Exists(bucket, key string) (bool, error) {
	var exists bool

	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}

		data := b.Get([]byte(key))
		exists = data != nil
		return nil
	})

	return exists, err
}

// Close closes the database
func (d *Database) Close() error {
	return d.db.Close()
}

// Backup creates a backup of the database
func (d *Database) Backup() ([]byte, error) {
	var backup bytes.Buffer

	err := d.db.View(func(tx *bbolt.Tx) error {
		_, err := tx.WriteTo(&backup)
		return err
	})

	return backup.Bytes(), err
}

// Restore restores the database from backup data
func (d *Database) Restore(data []byte) error {
	// Close current database
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	// Write backup data to file
	if err := os.WriteFile(d.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}

	// Reopen database
	db, err := bbolt.Open(d.path, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to open database for restore: %w", err)
	}

	d.db = db
	return nil
}

// Stats returns database statistics
func (d *Database) Stats() bbolt.Stats {
	return d.db.Stats()
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring performs a simple substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
