package logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mbeniwal-imwe/ark/internal/core/logger"
	"github.com/spf13/cobra"
)

// setupTestLogsEnvironment creates a test environment
func setupTestLogsEnvironment(t *testing.T) string {
	t.Helper()
	configDir, err := os.MkdirTemp("", "ark-logs-test-*")
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

	return configDir
}

// cleanupTestLogsEnvironment removes test environment
func cleanupTestLogsEnvironment(t *testing.T, configDir string) {
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

func TestLogsCommandStructure(t *testing.T) {
	// Test that commands are properly structured
	if LogsCmd == nil {
		t.Fatal("LogsCmd should not be nil")
	}

	if LogsCmd.Use != "logs" {
		t.Errorf("Expected Use to be 'logs', got '%s'", LogsCmd.Use)
	}

	// Verify the command structure exists
	// Subcommands may not be added until init() runs in the actual package
	if LogsCmd.Use == "" {
		t.Error("LogsCmd should have a Use value")
	}
}

func TestLogsViewWithNoLogs(t *testing.T) {
	configDir := setupTestLogsEnvironment(t)
	defer cleanupTestLogsEnvironment(t, configDir)

	// Initialize logger
	logConfig := logger.LogConfig{
		Enabled:  true,
		MaxDays:  30,
		MaxSize:  100,
		Compress: true,
		LogDir:   filepath.Join(configDir, "logs"),
	}

	loggerInstance, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	// Get logs for non-existent feature
	logs, err := loggerInstance.GetLogs("nonexistent", 10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected 0 logs for non-existent feature, got %d", len(logs))
	}
}

func TestLogsViewWithLogs(t *testing.T) {
	configDir := setupTestLogsEnvironment(t)
	defer cleanupTestLogsEnvironment(t, configDir)

	// Initialize logger
	logConfig := logger.LogConfig{
		Enabled:  true,
		MaxDays:  30,
		MaxSize:  100,
		Compress: true,
		LogDir:   filepath.Join(configDir, "logs"),
	}

	loggerInstance, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	// Write a test log
	loggerInstance.Info("test-feature", "Test log message")

	// Get logs - GetLogs might return structured log entries or raw strings
	// The actual format depends on the logger implementation
	logs, err := loggerInstance.GetLogs("test-feature", 10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) == 0 {
		t.Error("Expected at least 1 log entry, got 0")
	} else {
		// The logger might return structured entries - check if it contains our message
		messageFound := false
		for _, logEntry := range logs {
			if strings.Contains(logEntry.Message, "Test log message") {
				messageFound = true
				break
			}
		}
		if !messageFound {
			t.Logf("Log entries received: %+v", logs)
			t.Error("Expected log message to contain 'Test log message'")
		}
	}
}

func TestLogsGetLogsWithLimit(t *testing.T) {
	configDir := setupTestLogsEnvironment(t)
	defer cleanupTestLogsEnvironment(t, configDir)

	// Initialize logger
	logConfig := logger.LogConfig{
		Enabled:  true,
		MaxDays:  30,
		MaxSize:  100,
		Compress: true,
		LogDir:   filepath.Join(configDir, "logs"),
	}

	loggerInstance, err := logger.NewLogger(logConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer loggerInstance.Close()

	// Write multiple log entries
	for i := 0; i < 5; i++ {
		loggerInstance.Info("test-feature", "Test log message "+string(rune('0'+i)))
	}

	// Get logs with limit
	logs, err := loggerInstance.GetLogs("test-feature", 3)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) > 3 {
		t.Errorf("Expected at most 3 log entries, got %d", len(logs))
	}
}

// Note: tail and clear commands require:
// 1. Real-time file watching (tail)
// 2. Interactive confirmation (clear)
// These are better tested in integration tests
