package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/core/password"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Ark CLI with master password",
	Long: `Initialize Ark CLI by setting up the master password and creating necessary directories.

This command will:
- Create the ~/.ark directory structure
- Set up the master password for encryption
- Initialize the encrypted database
- Create default configuration files`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configDir := GetConfigDir()

	// Check if already initialized
	if isInitialized(configDir) {
		fmt.Println("Ark is already initialized.")
		return nil
	}

	// Create directory structure
	if err := createDirectoryStructure(configDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Initialize master password
	masterPassword, err := password.SetupMasterPassword()
	if err != nil {
		return fmt.Errorf("failed to setup master password: %w", err)
	}

	// Initialize configuration
	cfg, err := config.Initialize(configDir, masterPassword)
	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Save configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Ark CLI initialized successfully!")
	fmt.Printf("Configuration directory: %s\n", configDir)
	fmt.Println("You can now use Ark commands to manage your credentials and automate tasks.")

	return nil
}

func isInitialized(configDir string) bool {
	configFile := filepath.Join(configDir, "config.yaml")
	_, err := os.Stat(configFile)
	return err == nil
}

func createDirectoryStructure(configDir string) error {
	dirs := []string{
		configDir,
		filepath.Join(configDir, "data"),
		filepath.Join(configDir, "logs"),
		filepath.Join(configDir, "config"),
		filepath.Join(configDir, "backup"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
