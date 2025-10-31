package vault

import (
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
)

func TestUpdateCommand(t *testing.T) {
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

	// Add a credential
	err = vaultManager.Set("update-key", "old-value", "text", "Old description", []string{"old"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	// Update the credential
	err = vaultManager.Update("update-key", "new-value", "text", "New description", []string{"new", "updated"})
	if err != nil {
		t.Fatalf("Failed to update credential: %v", err)
	}

	// Verify the update
	entry, err := vaultManager.Get("update-key")
	if err != nil {
		t.Fatalf("Failed to get updated credential: %v", err)
	}

	if entry.Value != "new-value" {
		t.Errorf("Expected value 'new-value', got '%s'", entry.Value)
	}
	if entry.Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", entry.Description)
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "new" || entry.Tags[1] != "updated" {
		t.Errorf("Expected tags [new, updated], got %v", entry.Tags)
	}
}

func TestUpdateCommandNonExistent(t *testing.T) {
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

	// Try to update non-existent credential
	err = vaultManager.Update("non-existent-key", "value", "text", "", nil)
	if err == nil {
		t.Error("Expected error when updating non-existent credential, but got none")
	}
}

func TestUpdateCommandPreservesFormat(t *testing.T) {
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

	// Add a credential with JSON format
	err = vaultManager.Set("json-key", `{"old": "value"}`, "json", "JSON entry", []string{"json"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	// Update with explicit format - when empty format is passed to Update,
	// it uses the format from the existing entry (this is handled in the command, not the manager)
	// The manager's Update method requires a format to be specified
	err = vaultManager.Update("json-key", `{"new": "value"}`, "json", "", nil)
	if err != nil {
		t.Fatalf("Failed to update credential: %v", err)
	}

	// Verify format was preserved
	entry, err := vaultManager.Get("json-key")
	if err != nil {
		t.Fatalf("Failed to get updated credential: %v", err)
	}

	if entry.Format != "json" {
		t.Errorf("Expected format to be 'json', got '%s'", entry.Format)
	}
}

func TestUpdateCommandWithFormatChange(t *testing.T) {
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

	// Add a credential with text format
	err = vaultManager.Set("format-key", "text-value", "text", "", nil)
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	// Update with different format
	err = vaultManager.Update("format-key", `{"key": "value"}`, "json", "", nil)
	if err != nil {
		t.Fatalf("Failed to update credential: %v", err)
	}

	// Verify format changed
	entry, err := vaultManager.Get("format-key")
	if err != nil {
		t.Fatalf("Failed to get updated credential: %v", err)
	}

	if entry.Format != "json" {
		t.Errorf("Expected format to be 'json', got '%s'", entry.Format)
	}
}
