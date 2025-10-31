package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/core/crypto"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

// setupTestVaultEnvironment creates a test environment with initialized config and database
func setupTestVaultEnvironment(t *testing.T) (string, []byte) {
	t.Helper()
	configDir, err := os.MkdirTemp("", "ark-vault-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create subdirectories
	subdirs := []string{"data", "logs", "config", "backup"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(configDir, subdir), 0700); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}
	}

	// Initialize config with a test master password
	masterPassword := "TestPassword123!"
	salt, _ := crypto.GenerateSalt()
	masterKey, err := crypto.DeriveKey(masterPassword, salt)
	if err != nil {
		t.Fatalf("Failed to derive master key: %v", err)
	}

	cfg := &config.Config{
		ConfigDir:    configDir,
		DatabasePath: filepath.Join(configDir, "data", "ark.db"),
		Salt:         salt,
		Security: config.SecurityConfig{
			PasswordCacheTimeout: 300,
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create and initialize database
	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	db.Close()

	return configDir, masterKey
}

// cleanupTestVaultEnvironment removes test environment
func cleanupTestVaultEnvironment(t *testing.T, configDir string) {
	t.Helper()
	if err := os.RemoveAll(configDir); err != nil {
		t.Logf("Failed to cleanup test dir: %v", err)
	}
}

// setupTestRootCmd creates a root command for testing
func setupTestRootCmd(t *testing.T, configDir string) *cobra.Command {
	t.Helper()
	rootCmd := &cobra.Command{
		Use: "ark",
	}
	rootCmd.PersistentFlags().String("config-dir", configDir, "Configuration directory")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("config", "", "Config file")
	return rootCmd
}

// mockGetMasterKey patches config.GetMasterKey for testing
// Note: This is a workaround since we can't easily mock password input
// In a real scenario, you'd want to refactor password.GetMasterPassword() to be injectable
func getTestMasterKey(t *testing.T, configDir string, masterKey []byte) []byte {
	t.Helper()
	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	// Directly set the master key in config (bypassing password prompt)
	cfg.MasterKey = masterKey
	return masterKey
}
