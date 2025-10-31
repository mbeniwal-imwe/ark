package vault

import (
	"strings"
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
)

func TestSearchCommand(t *testing.T) {
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

	// Add test credentials
	vaultManager.Set("aws-access-key", "AKIAIOSFODNN7EXAMPLE", "text", "AWS access key", []string{"aws", "access"})
	vaultManager.Set("aws-secret-key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "text", "AWS secret key", []string{"aws", "secret"})
	vaultManager.Set("github-token", "ghp_1234567890abcdef", "text", "GitHub personal access token", []string{"github", "token"})

	// Search for "aws"
	entries, err := vaultManager.Search("aws")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(entries) < 2 {
		t.Errorf("Expected at least 2 entries matching 'aws', got %d", len(entries))
	}

	// Verify results contain "aws" in key, description, or tags
	for _, entry := range entries {
		found := strings.Contains(strings.ToLower(entry.Key), "aws") ||
			strings.Contains(strings.ToLower(entry.Description), "aws")
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), "aws") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Entry %s doesn't seem to match 'aws' search", entry.Key)
		}
	}
}

func TestSearchCommandNoResults(t *testing.T) {
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

	// Search for something that doesn't exist
	entries, err := vaultManager.Search("nonexistent-key-12345")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 results for non-existent search, got %d", len(entries))
	}
}

func TestDisplaySearchResults(t *testing.T) {
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

	vaultManager.Set("search-key", "search-value", "text", "Search test", []string{"search"})

	entries, err := vaultManager.Search("search")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// Test displaySearchResults doesn't panic
	if err := displaySearchResults(entries, "search"); err != nil {
		t.Errorf("displaySearchResults failed: %v", err)
	}

	// Test with empty results
	if err := displaySearchResults([]*VaultEntry{}, "empty"); err != nil {
		t.Errorf("displaySearchResults failed with empty results: %v", err)
	}
}

func TestSearchCommandWithTagFilter(t *testing.T) {
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
	vaultManager.Set("prod-key", "prod-value", "text", "Production key", []string{"prod", "aws"})
	vaultManager.Set("dev-key", "dev-value", "text", "Development key", []string{"dev", "aws"})
	vaultManager.Set("test-key", "test-value", "text", "Test key", []string{"test", "gcp"})

	// Search and filter by tags
	allEntries, err := vaultManager.Search("key")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// Filter by "aws" tag
	filtered := filterByTags(allEntries, []string{"aws"})
	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries with 'aws' tag after search, got %d", len(filtered))
	}
}
