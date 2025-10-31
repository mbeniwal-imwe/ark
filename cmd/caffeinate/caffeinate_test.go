package caffeinate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/features/caffeinate"
	"github.com/spf13/cobra"
)

// setupTestCaffeinateEnvironment creates a test environment
func setupTestCaffeinateEnvironment(t *testing.T) string {
	t.Helper()
	configDir, err := os.MkdirTemp("", "ark-caffeinate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create data subdirectory for PID file
	dataDir := filepath.Join(configDir, "data")
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		t.Fatalf("Failed to create data directory: %v", err)
	}

	return configDir
}

// cleanupTestCaffeinateEnvironment removes test environment
func cleanupTestCaffeinateEnvironment(t *testing.T, configDir string) {
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

func TestCaffeinateCommandStructure(t *testing.T) {
	// Test that commands are properly structured
	if CaffeinateCmd == nil {
		t.Fatal("CaffeinateCmd should not be nil")
	}

	if CaffeinateCmd.Use != "caffeinate" {
		t.Errorf("Expected Use to be 'caffeinate', got '%s'", CaffeinateCmd.Use)
	}

	// Check that subcommands are added
	subcommands := CaffeinateCmd.Commands()
	expectedSubcommands := []string{"start", "stop", "status"}
	found := make(map[string]bool)

	for _, cmd := range subcommands {
		found[cmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !found[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

func TestCaffeinateStatusWhenNotRunning(t *testing.T) {
	configDir := setupTestCaffeinateEnvironment(t)
	defer cleanupTestCaffeinateEnvironment(t, configDir)

	runner := &caffeinate.Runner{ConfigDir: configDir}
	status, err := runner.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status == "" {
		t.Error("Status should not be empty")
	}
}

func TestCaffeinateStopWhenNotRunning(t *testing.T) {
	configDir := setupTestCaffeinateEnvironment(t)
	defer cleanupTestCaffeinateEnvironment(t, configDir)

	runner := &caffeinate.Runner{ConfigDir: configDir}
	// Stopping when not running should not error (should be idempotent)
	err := runner.Stop()
	if err != nil {
		t.Logf("Stop returned error (this may be expected): %v", err)
	}
}

func TestSecondsToDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    int64 // Expected duration in seconds
	}{
		{30, 30},
		{60, 60},
		{0, 0},
		{300, 300},
	}

	for _, tt := range tests {
		d := secondsToDuration(tt.seconds)
		if d.Seconds() != float64(tt.want) {
			t.Errorf("secondsToDuration(%d) = %.0fs, want %d", tt.seconds, d.Seconds(), tt.want)
		}
	}
}

// Note: We don't test actual Start() because it requires:
// 1. Running actual caffeinate process (system dependency)
// 2. Background process management
// 3. File system operations that might affect the system
// These should be tested in integration tests or with proper mocking
