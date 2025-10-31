package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/core/crypto"
	"github.com/mbeniwal-imwe/ark/internal/core/password"
	"gopkg.in/yaml.v3"
)

// Config represents the Ark configuration
type Config struct {
	Version      string         `yaml:"version" json:"version"`
	CreatedAt    time.Time      `yaml:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `yaml:"updated_at" json:"updated_at"`
	MasterKey    []byte         `yaml:"-" json:"-"` // Not serialized
	Salt         []byte         `yaml:"salt" json:"-"`
	ConfigDir    string         `yaml:"-" json:"-"`
	DatabasePath string         `yaml:"database_path" json:"database_path"`
	LogLevel     string         `yaml:"log_level" json:"log_level"`
	LogRotation  LogConfig      `yaml:"log_rotation" json:"log_rotation"`
	AWS          AWSConfig      `yaml:"aws" json:"aws"`
	Backup       BackupConfig   `yaml:"backup" json:"backup"`
	Security     SecurityConfig `yaml:"security" json:"security"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Enabled  bool `yaml:"enabled" json:"enabled"`
	MaxDays  int  `yaml:"max_days" json:"max_days"`
	MaxSize  int  `yaml:"max_size_mb" json:"max_size_mb"`
	Compress bool `yaml:"compress" json:"compress"`
}

// AWSConfig represents AWS configuration
type AWSConfig struct {
	DefaultProfile string            `yaml:"default_profile" json:"default_profile"`
	Profiles       map[string]string `yaml:"profiles" json:"profiles"`
	Region         string            `yaml:"region" json:"region"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	S3Bucket      string `yaml:"s3_bucket" json:"s3_bucket"`
	S3Prefix      string `yaml:"s3_prefix" json:"s3_prefix"`
	EncryptionKey []byte `yaml:"-" json:"-"` // Not serialized
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	PasswordCacheTimeout int `yaml:"password_cache_timeout_seconds" json:"password_cache_timeout_seconds"` // Timeout in seconds
}

// cacheEntry represents a cached master key entry
type cacheEntry struct {
	Key       []byte    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

var (
	// Global mutex for file-based cache operations
	cacheMutex sync.RWMutex
)

// DefaultConfig returns a default configuration
func DefaultConfig(configDir string) *Config {
	return &Config{
		Version:      "1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ConfigDir:    configDir,
		DatabasePath: filepath.Join(configDir, "data", "ark.db"),
		LogLevel:     "info",
		LogRotation: LogConfig{
			Enabled:  true,
			MaxDays:  30,
			MaxSize:  100, // 100MB
			Compress: true,
		},
		AWS: AWSConfig{
			DefaultProfile: "default",
			Profiles:       make(map[string]string),
			Region:         "us-east-1",
		},
		Backup: BackupConfig{
			Enabled:  false,
			S3Prefix: "ark-backups/",
		},
		Security: SecurityConfig{
			PasswordCacheTimeout: 300, // Default 5 minutes
		},
	}
}

// Initialize creates a new configuration with master password
func Initialize(configDir, masterPassword string) (*Config, error) {
	config := DefaultConfig(configDir)

	// Generate salt for key derivation
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	config.Salt = salt

	// Derive master key from password
	masterKey, err := crypto.DeriveKey(masterPassword, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive master key: %w", err)
	}
	config.MasterKey = masterKey

	// Generate backup encryption key
	backupKey, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup key: %w", err)
	}
	config.Backup.EncryptionKey = backupKey

	return config, nil
}

// Load loads configuration from file
func Load(configDir string) (*Config, error) {
	configFile := filepath.Join(configDir, "config.yaml")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.ConfigDir = configDir
	config.DatabasePath = filepath.Join(configDir, "data", "ark.db")

	// Ensure Security config has default timeout if not set
	if config.Security.PasswordCacheTimeout <= 0 {
		config.Security.PasswordCacheTimeout = 300 // Default 5 minutes
	}

	return &config, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	configFile := filepath.Join(c.ConfigDir, "config.yaml")

	// Update timestamp
	c.UpdatedAt = time.Now()

	// Create a copy without sensitive data for serialization
	safeConfig := *c
	safeConfig.MasterKey = nil
	safeConfig.Backup.EncryptionKey = nil

	data, err := yaml.Marshal(&safeConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getCacheFilePath returns the path to the password cache file
func (c *Config) getCacheFilePath() string {
	return filepath.Join(c.ConfigDir, "data", ".master_key_cache")
}

// getCacheEncryptionKey derives an encryption key from config directory and salt
func (c *Config) getCacheEncryptionKey() ([]byte, error) {
	// Use config directory path + salt to derive a stable encryption key
	// This ensures the cache is tied to this specific Ark installation
	data := []byte(c.ConfigDir + string(c.Salt))
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// loadCachedMasterKey loads and decrypts the cached master key if valid
func (c *Config) loadCachedMasterKey() ([]byte, error) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	cachePath := c.getCacheFilePath()

	// Check if cache file exists
	if _, err := os.Stat(cachePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cache not found")
		}
		return nil, err
	}

	// Read encrypted cache file
	encryptedData, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	// Derive encryption key for cache
	encKey, err := c.getCacheEncryptionKey()
	if err != nil {
		return nil, err
	}

	// Create encryptor with cache encryption key
	encryptor, err := crypto.NewEncryptor(encKey)
	if err != nil {
		os.Remove(cachePath)
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Decrypt cache data (Encryptor.Decrypt expects nonce prepended)
	plaintext, err := encryptor.Decrypt(encryptedData)
	if err != nil {
		// If decryption fails, cache might be corrupted - delete it
		os.Remove(cachePath)
		return nil, fmt.Errorf("failed to decrypt cache: %w", err)
	}

	// Unmarshal cache entry
	var entry cacheEntry
	if err := json.Unmarshal(plaintext, &entry); err != nil {
		os.Remove(cachePath)
		return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	// Check if cache is expired
	if time.Now().After(entry.ExpiresAt) {
		os.Remove(cachePath)
		return nil, fmt.Errorf("cache expired")
	}

	return entry.Key, nil
}

// saveCachedMasterKey saves and encrypts the master key to cache file
func (c *Config) saveCachedMasterKey(masterKey []byte, timeoutSeconds int) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Create cache entry
	entry := cacheEntry{
		Key:       masterKey,
		ExpiresAt: time.Now().Add(time.Duration(timeoutSeconds) * time.Second),
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Derive encryption key
	encKey, err := c.getCacheEncryptionKey()
	if err != nil {
		return err
	}

	// Create encryptor with cache encryption key
	encryptor, err := crypto.NewEncryptor(encKey)
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}

	// Encrypt cache data (Encryptor.Encrypt prepends nonce)
	encryptedData, err := encryptor.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt cache: %w", err)
	}

	// Ensure cache directory exists
	cachePath := c.getCacheFilePath()
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write encrypted cache file with restrictive permissions
	if err := os.WriteFile(cachePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// GetMasterKey returns the master encryption key
func (c *Config) GetMasterKey() ([]byte, error) {
	// Check if master key is already loaded in config
	if len(c.MasterKey) > 0 {
		return c.MasterKey, nil
	}

	// Check file-based cache
	cachedKey, err := c.loadCachedMasterKey()
	if err == nil && len(cachedKey) > 0 {
		// Cache hit - use cached key
		c.MasterKey = cachedKey
		return cachedKey, nil
	}

	// If no master key is loaded, we need to prompt for the master password
	// and derive the key from the stored salt
	if len(c.Salt) == 0 {
		return nil, fmt.Errorf("no salt found in config - Ark may not be initialized. Run 'ark init' first")
	}

	// Prompt for master password
	masterPassword, err := password.GetMasterPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to get master password: %w", err)
	}

	// Derive master key from password and salt
	masterKey, err := crypto.DeriveKey(masterPassword, c.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive master key: %w", err)
	}

	// Cache the master key with expiration (file-based)
	timeout := c.Security.PasswordCacheTimeout
	if timeout <= 0 {
		timeout = 300 // Default 5 minutes if not set
	}

	// Save to cache file (ignore errors - cache is optional)
	if err := c.saveCachedMasterKey(masterKey, timeout); err != nil {
		// Log but don't fail - caching is a convenience feature
		// In production, you might want to log this
	}

	// Also cache in the config instance
	c.MasterKey = masterKey
	return masterKey, nil
}

// GetMasterKeySilent returns the master key without prompting (for internal use)
func (c *Config) GetMasterKeySilent() []byte {
	return c.MasterKey
}

// ClearPasswordCache clears the cached master key
func ClearPasswordCache(configDir string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cachePath := filepath.Join(configDir, "data", ".master_key_cache")
	os.Remove(cachePath)
}

// SetPasswordCacheTimeout updates the password cache timeout in the config
func (c *Config) SetPasswordCacheTimeout(timeoutSeconds int) {
	if timeoutSeconds < 0 {
		timeoutSeconds = 0 // 0 means no caching
	}
	c.Security.PasswordCacheTimeout = timeoutSeconds
}

// SetMasterPassword updates the master password and regenerates keys
func (c *Config) SetMasterPassword(password string) error {
	// Clear the cache since we're changing the password
	ClearPasswordCache(c.ConfigDir)

	// Generate new salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	c.Salt = salt

	// Derive new master key
	masterKey, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return fmt.Errorf("failed to derive master key: %w", err)
	}
	c.MasterKey = masterKey

	// Generate new backup key
	backupKey, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate backup key: %w", err)
	}
	c.Backup.EncryptionKey = backupKey

	return nil
}

// ToJSON returns the configuration as JSON (excluding sensitive data)
func (c *Config) ToJSON() ([]byte, error) {
	safeConfig := *c
	safeConfig.MasterKey = nil
	safeConfig.Backup.EncryptionKey = nil

	return json.MarshalIndent(&safeConfig, "", "  ")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	if len(c.Salt) != crypto.SaltSize {
		return fmt.Errorf("invalid salt size")
	}

	if c.DatabasePath == "" {
		return fmt.Errorf("database path is required")
	}

	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	return nil
}
