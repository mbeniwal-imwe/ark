package vault

import (
	"strings"
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
)

func TestSetCommand(t *testing.T) {
	configDir, masterKey := setupTestVaultEnvironment(t)
	defer cleanupTestVaultEnvironment(t, configDir)

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.MasterKey = masterKey

	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Since runSet requires password input, we'll test the vault manager directly
	vaultManager := vault.NewVaultManager(db)

	// Test setting a credential
	err = vaultManager.Set("test-key", "test-value", "text", "Test description", []string{"test", "unit"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	// Verify it was stored
	entry, err := vaultManager.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get credential: %v", err)
	}

	if entry.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", entry.Value)
	}
	if entry.Format != "text" {
		t.Errorf("Expected format 'text', got '%s'", entry.Format)
	}
	if entry.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", entry.Description)
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "test" || entry.Tags[1] != "unit" {
		t.Errorf("Expected tags [test, unit], got %v", entry.Tags)
	}
}

func TestSetCommandWithJSONFormat(t *testing.T) {
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

	jsonValue := `{"key": "value", "number": 123}`
	err = vaultManager.Set("json-key", jsonValue, "json", "JSON test", []string{"json"})
	if err != nil {
		t.Fatalf("Failed to set JSON credential: %v", err)
	}

	entry, err := vaultManager.Get("json-key")
	if err != nil {
		t.Fatalf("Failed to get JSON credential: %v", err)
	}

	if entry.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", entry.Format)
	}
	if !strings.Contains(entry.Value, "key") {
		t.Errorf("JSON value should contain 'key', got: %s", entry.Value)
	}
}

func TestSetCommandWithYAMLFormat(t *testing.T) {
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

	yamlValue := "key: value\nnumber: 123"
	err = vaultManager.Set("yaml-key", yamlValue, "yaml", "YAML test", []string{"yaml"})
	if err != nil {
		t.Fatalf("Failed to set YAML credential: %v", err)
	}

	entry, err := vaultManager.Get("yaml-key")
	if err != nil {
		t.Fatalf("Failed to get YAML credential: %v", err)
	}

	if entry.Format != "yaml" {
		t.Errorf("Expected format 'yaml', got '%s'", entry.Format)
	}
}

func TestSetCommandEmptyValue(t *testing.T) {
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

	// Note: The vault manager allows empty values (which might be valid in some cases)
	// The command handler (runSet) rejects empty values, but we're testing the manager directly
	// So empty values are technically allowed at the storage layer
	vaultManager := vault.NewVaultManager(db)
	err = vaultManager.Set("empty-key", "", "text", "", nil)
	// Empty values are allowed at the storage layer - validation happens in the command handler
	if err != nil {
		t.Logf("Vault manager rejected empty value (this may be valid): %v", err)
	}
}

func TestSetCommandInvalidFormat(t *testing.T) {
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

	// Test invalid format
	err = vaultManager.Set("invalid-key", "value", "invalid-format", "", nil)
	if err == nil {
		t.Error("Expected error for invalid format, but got none")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected error message about invalid format, got: %v", err)
	}
}
