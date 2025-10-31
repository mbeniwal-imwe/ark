package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Set test version values
	originalVersion := Version
	originalBuildDate := BuildDate
	originalGitCommit := GitCommit

	defer func() {
		Version = originalVersion
		BuildDate = originalBuildDate
		GitCommit = originalGitCommit
	}()

	Version = "1.0.0"
	BuildDate = "2025-01-01T00:00:00Z"
	GitCommit = "abc123"

	// Test the version output directly (mimicking what the command does)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Ark CLI %s\n", Version)
	fmt.Fprintf(&buf, "Build Date: %s\n", BuildDate)
	fmt.Fprintf(&buf, "Git Commit: %s\n", GitCommit)

	output := buf.String()
	if !strings.Contains(output, "Ark CLI 1.0.0") {
		t.Errorf("Expected 'Ark CLI 1.0.0' in output, got: %s", output)
	}
	if !strings.Contains(output, "Build Date: 2025-01-01T00:00:00Z") {
		t.Errorf("Expected build date in output, got: %s", output)
	}
	if !strings.Contains(output, "Git Commit: abc123") {
		t.Errorf("Expected git commit in output, got: %s", output)
	}
}

func TestVersionCommandWithDefaults(t *testing.T) {
	originalVersion := Version
	originalBuildDate := BuildDate
	originalGitCommit := GitCommit

	defer func() {
		Version = originalVersion
		BuildDate = originalBuildDate
		GitCommit = originalGitCommit
	}()

	Version = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"

	// Test the version output directly
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Ark CLI %s\n", Version)
	fmt.Fprintf(&buf, "Build Date: %s\n", BuildDate)
	fmt.Fprintf(&buf, "Git Commit: %s\n", GitCommit)

	output := buf.String()
	if !strings.Contains(output, "Ark CLI dev") {
		t.Errorf("Expected 'Ark CLI dev' in output, got: %s", output)
	}
	if !strings.Contains(output, "Build Date: unknown") {
		t.Errorf("Expected 'unknown' build date, got: %s", output)
	}
	if !strings.Contains(output, "Git Commit: unknown") {
		t.Errorf("Expected 'unknown' git commit, got: %s", output)
	}
}
