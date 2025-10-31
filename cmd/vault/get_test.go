package vault

import (
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
)

func TestGetCommand(t *testing.T) {
	configDir, masterKey := setupTestVaultEnvironment(t)
	defer cleanupTestVaultEnvironment(t, configDir)

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	vaultManager := vault.NewVaultManager(db)

	// Set a credential first
	testValue := "secret-api-key-12345"
	err = vaultManager.Set("api-key", testValue, "text", "API key for testing", []string{"api", "test"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	// Get the credential
	entry, err := vaultManager.Get("api-key")
	if err != nil {
		t.Fatalf("Failed to get credential: %v", err)
	}

	if entry.Value != testValue {
		t.Errorf("Expected value '%s', got '%s'", testValue, entry.Value)
	}
	if entry.Key != "api-key" {
		t.Errorf("Expected key 'api-key', got '%s'", entry.Key)
	}
	if entry.Format != "text" {
		t.Errorf("Expected format 'text', got '%s'", entry.Format)
	}
	if entry.Description != "API key for testing" {
		t.Errorf("Expected description 'API key for testing', got '%s'", entry.Description)
	}
}

func TestGetCommandNonExistent(t *testing.T) {
	configDir, masterKey := setupTestVaultEnvironment(t)
	defer cleanupTestVaultEnvironment(t, configDir)

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	vaultManager := vault.NewVaultManager(db)

	// Try to get non-existent credential
	_, err = vaultManager.Get("non-existent-key")
	if err == nil {
		t.Error("Expected error when getting non-existent credential, but got none")
	}
}

func TestDisplayValueTextFormat(t *testing.T) {
	value := "plain text value"
	displayValue(value, "text")
	// This function just prints, so we can't easily test output
	// But we can verify it doesn't panic
}

func TestDisplayValueJSONFormat(t *testing.T) {
	value := `{"key":"value","number":123}`
	displayValue(value, "json")
	// Function prints formatted JSON, verify no panic
}

func TestDisplayValueYAMLFormat(t *testing.T) {
	value := "key: value\nnumber: 123"
	displayValue(value, "yaml")
	// Function prints formatted YAML, verify no panic
}

func TestDisplayWithMetadata(t *testing.T) {
	configDir, masterKey := setupTestVaultEnvironment(t)
	defer cleanupTestVaultEnvironment(t, configDir)

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	vaultManager := vault.NewVaultManager(db)

	err = vaultManager.Set("meta-key", "meta-value", "text", "Metadata test", []string{"meta"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	entry, err := vaultManager.Get("meta-key")
	if err != nil {
		t.Fatalf("Failed to get credential: %v", err)
	}

	// Test displayWithMetadata doesn't panic
	displayWithMetadata(entry, "text")
}
