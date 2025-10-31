package vault

import (
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
)

func TestListCommand(t *testing.T) {
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

	// Add multiple credentials
	credentials := []struct {
		key         string
		value       string
		format      string
		description string
		tags        []string
	}{
		{"key1", "value1", "text", "First key", []string{"test", "one"}},
		{"key2", "value2", "text", "Second key", []string{"test", "two"}},
		{"key3", "value3", "json", "Third key", []string{"test"}},
	}

	for _, cred := range credentials {
		err := vaultManager.Set(cred.key, cred.value, cred.format, cred.description, cred.tags)
		if err != nil {
			t.Fatalf("Failed to set credential %s: %v", cred.key, err)
		}
	}

	// List all credentials
	entries, err := vaultManager.List()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 credentials, got %d", len(entries))
	}
}

func TestListCommandEmptyVault(t *testing.T) {
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

	entries, err := vaultManager.List()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 credentials in empty vault, got %d", len(entries))
	}
}

func TestDisplayAsTable(t *testing.T) {
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

	err = vaultManager.Set("table-key", "table-value", "text", "Table test", []string{"table"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	entries, err := vaultManager.List()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	// Test displayAsTable doesn't panic
	if err := displayAsTable(entries); err != nil {
		t.Errorf("displayAsTable failed: %v", err)
	}

	// Test with empty list
	if err := displayAsTable([]*VaultEntry{}); err != nil {
		t.Errorf("displayAsTable failed with empty list: %v", err)
	}
}

func TestDisplayAsJSON(t *testing.T) {
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

	err = vaultManager.Set("json-key", "json-value", "text", "JSON test", []string{"json"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	entries, err := vaultManager.List()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	// Test that displayAsJSON doesn't panic and produces output
	// We can't easily capture stdout in unit tests, so we just verify it doesn't error
	if err := displayAsJSON(entries); err != nil {
		t.Errorf("displayAsJSON failed: %v", err)
	}
}

func TestDisplayAsYAML(t *testing.T) {
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

	err = vaultManager.Set("yaml-key", "yaml-value", "text", "YAML test", []string{"yaml"})
	if err != nil {
		t.Fatalf("Failed to set credential: %v", err)
	}

	entries, err := vaultManager.List()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	// Test that displayAsYAML doesn't panic and produces output
	// We can't easily capture stdout in unit tests, so we just verify it doesn't error
	if err := displayAsYAML(entries); err != nil {
		t.Errorf("displayAsYAML failed: %v", err)
	}
}

func TestFilterByTags(t *testing.T) {
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

	// Add credentials with different tags
	vaultManager.Set("key1", "value1", "text", "", []string{"aws", "prod"})
	vaultManager.Set("key2", "value2", "text", "", []string{"aws", "dev"})
	vaultManager.Set("key3", "value3", "text", "", []string{"gcp", "prod"})

	allEntries, err := vaultManager.List()
	if err != nil {
		t.Fatalf("Failed to list credentials: %v", err)
	}

	// Filter by single tag
	filtered := filterByTags(allEntries, []string{"aws"})
	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries with 'aws' tag, got %d", len(filtered))
	}

	// Filter by multiple tags (AND logic)
	filtered = filterByTags(allEntries, []string{"aws", "prod"})
	if len(filtered) != 1 {
		t.Errorf("Expected 1 entry with both 'aws' and 'prod' tags, got %d", len(filtered))
	}
	if filtered[0].Key != "key1" {
		t.Errorf("Expected filtered entry to be 'key1', got '%s'", filtered[0].Key)
	}

	// Filter by non-existent tag
	filtered = filterByTags(allEntries, []string{"nonexistent"})
	if len(filtered) != 0 {
		t.Errorf("Expected 0 entries with 'nonexistent' tag, got %d", len(filtered))
	}
}

// Helper functions removed - display functions output to stdout which is hard to test in unit tests
// These should be tested in integration tests or with proper output capturing
