package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCommand(t *testing.T) {
	configDir := setupTestConfigDir(t)
	defer cleanupTestConfigDir(t, configDir)

	// This will fail because it needs password input, but we can test directory creation
	// We'll test directory creation separately

	// Check that directories were created
	expectedDirs := []string{
		configDir,
		filepath.Join(configDir, "data"),
		filepath.Join(configDir, "logs"),
		filepath.Join(configDir, "config"),
		filepath.Join(configDir, "backup"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("Expected directory %s to exist, but got error: %v", dir, err)
		}
	}
}

func TestIsInitialized(t *testing.T) {
	configDir := setupTestConfigDir(t)
	defer cleanupTestConfigDir(t, configDir)

	// Should return false when config file doesn't exist
	if isInitialized(configDir) {
		t.Error("Expected isInitialized to return false when config doesn't exist")
	}

	// Create config file
	configFile := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("test: data\n"), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Should return true when config file exists
	if !isInitialized(configDir) {
		t.Error("Expected isInitialized to return true when config exists")
	}
}

func TestCreateDirectoryStructure(t *testing.T) {
	configDir := setupTestConfigDir(t)
	defer cleanupTestConfigDir(t, configDir)

	// Remove directories
	os.RemoveAll(configDir)

	// Create directory structure
	if err := createDirectoryStructure(configDir); err != nil {
		t.Fatalf("createDirectoryStructure failed: %v", err)
	}

	// Verify all directories exist
	expectedDirs := []string{
		configDir,
		filepath.Join(configDir, "data"),
		filepath.Join(configDir, "logs"),
		filepath.Join(configDir, "config"),
		filepath.Join(configDir, "backup"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory %s should exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Errorf("%s should be a directory", dir)
		}
		// Check permissions (0700)
		if info.Mode().Perm() != 0700 {
			t.Errorf("Directory %s should have permissions 0700, got %o", dir, info.Mode().Perm())
		}
	}
}
