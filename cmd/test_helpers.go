package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// setupTestConfigDir creates a temporary config directory for testing
func setupTestConfigDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "ark-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create subdirectories
	subdirs := []string{"data", "logs", "config", "backup"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(dir, subdir), 0700); err != nil {
			t.Fatalf("Failed to create subdirectory %s: %v", subdir, err)
		}
	}

	return dir
}

// cleanupTestConfigDir removes the temporary config directory
func cleanupTestConfigDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("Failed to cleanup test dir %s: %v", dir, err)
	}
}

// setupTestRootCmd creates a root command with test config directory
func setupTestRootCmd(t *testing.T, configDir string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{
		Use: "ark",
	}
	cmd.PersistentFlags().String("config-dir", configDir, "Configuration directory")
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	cmd.PersistentFlags().String("config", "", "Config file")
	return cmd
}
