package vault

import (
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
)

func TestDeleteCommand(t *testing.T) {
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
	err = vaultManager.Set("delete-key", "delete-value", "text", "To be deleted", []string{"delete"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	// Verify it exists
	exists, err := vaultManager.Exists("delete-key")
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Credential should exist before deletion")
	}

	// Delete the credential
	err = vaultManager.Delete("delete-key")
	if err != nil {
		t.Fatalf("Failed to delete credential: %v", err)
	}

	// Verify it no longer exists
	exists, err = vaultManager.Exists("delete-key")
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Credential should not exist after deletion")
	}
}

func TestDeleteCommandNonExistent(t *testing.T) {
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

	// Try to delete non-existent credential
	err = vaultManager.Delete("non-existent-key")
	if err == nil {
		t.Error("Expected error when deleting non-existent credential, but got none")
	}
}

func TestConfirmDeletion(t *testing.T) {
	// Test confirmDeletion function
	// This function reads from stdin, so we can't easily test it in unit tests
	// But we can verify the logic with various inputs if we mock stdin
	t.Skip("confirmDeletion requires interactive input, skipping unit test")
}
