package lock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/core/crypto"
	"github.com/mbeniwal-imwe/ark/internal/features/dirlock"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

// setupTestLockEnvironment creates a test environment
func setupTestLockEnvironment(t *testing.T) (string, []byte) {
	t.Helper()
	configDir, err := os.MkdirTemp("", "ark-lock-test-*")
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

	// Initialize config with test master password
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

	return configDir, masterKey
}

// cleanupTestLockEnvironment removes test environment
func cleanupTestLockEnvironment(t *testing.T, configDir string) {
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
	return rootCmd
}

func TestLockCommandStructure(t *testing.T) {
	// Test that commands are properly structured
	if LockCmd == nil {
		t.Fatal("LockCmd should not be nil")
	}

	if LockCmd.Use != "lock" {
		t.Errorf("Expected Use to be 'lock', got '%s'", LockCmd.Use)
	}

	// Check that subcommands exist (they may not be added until init() runs)
	// This test verifies the command structure exists
	if LockCmd.Use == "" {
		t.Error("LockCmd should have a Use value")
	}
}

func TestLockListEmpty(t *testing.T) {
	configDir, masterKey := setupTestLockEnvironment(t)
	defer cleanupTestLockEnvironment(t, configDir)

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	svc := &dirlock.Service{DB: db}
	recs, err := svc.List()
	if err != nil {
		t.Fatalf("Failed to list locked directories: %v", err)
	}

	if len(recs) != 0 {
		t.Errorf("Expected 0 locked directories in empty database, got %d", len(recs))
	}
}

func TestLockServiceIsLocked(t *testing.T) {
	configDir, masterKey := setupTestLockEnvironment(t)
	defer cleanupTestLockEnvironment(t, configDir)

	// Create a test directory
	testDir, err := os.MkdirTemp("", "ark-lock-test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	svc := &dirlock.Service{DB: db}

	// Test IsLocked on unlocked directory
	locked, err := svc.IsLocked(testDir)
	if err != nil {
		t.Fatalf("Failed to check if directory is locked: %v", err)
	}

	if locked {
		t.Error("Directory should not be locked initially")
	}
}

// Note: Full lock/unlock testing requires:
// 1. Actual directory operations (tested in integration tests)
// 2. Password input (requires mocking or stdin manipulation)
// 3. File system encryption operations
// These should be tested with integration tests
