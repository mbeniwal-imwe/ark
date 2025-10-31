package awsfeat

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CheckPrerequisites checks if AWS CLI is installed and configured
func CheckPrerequisites() error {
	// Check if AWS CLI is installed
	if err := checkAWSCLI(); err != nil {
		return fmt.Errorf("AWS CLI not found: %w", err)
	}

	// Check if credentials are configured
	if err := checkCredentials(); err != nil {
		return fmt.Errorf("AWS credentials not configured: %w", err)
	}

	return nil
}

func checkAWSCLI() error {
	_, err := exec.LookPath("aws")
	if err != nil {
		return fmt.Errorf("AWS CLI not found in PATH. Please install AWS CLI first: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
	}
	return nil
}

func checkCredentials() error {
	// Check for AWS credentials in standard locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Check for credentials file
	credsPath := filepath.Join(homeDir, ".aws", "credentials")
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		return fmt.Errorf("AWS credentials file not found at %s", credsPath)
	}

	// Check for config file
	configPath := filepath.Join(homeDir, ".aws", "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("AWS config file not found at %s", configPath)
	}

	return nil
}

// TestAWSCLI tests AWS CLI connectivity
func TestAWSCLI() error {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("AWS CLI test failed: %w", err)
	}

	fmt.Printf("✅ AWS CLI is working correctly:\n%s", string(output))
	return nil
}
