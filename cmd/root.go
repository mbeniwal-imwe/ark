package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	awsCmd "github.com/mbeniwal-imwe/ark/cmd/aws"
	"github.com/mbeniwal-imwe/ark/cmd/backup"
	"github.com/mbeniwal-imwe/ark/cmd/caffeinate"
	ec2Cmd "github.com/mbeniwal-imwe/ark/cmd/ec2"
	"github.com/mbeniwal-imwe/ark/cmd/lock"
	"github.com/mbeniwal-imwe/ark/cmd/logs"
	s3cmd "github.com/mbeniwal-imwe/ark/cmd/s3"
	"github.com/mbeniwal-imwe/ark/cmd/vault"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version = "dev"
	// BuildDate is set during build
	BuildDate = "unknown"
	// GitCommit is set during build
	GitCommit = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ark",
	Short: "Ark CLI - Developer automation tool",
	Long: `Ark is a comprehensive CLI tool for developers that provides:

- Encrypted credential storage and management
- AWS cloud service integration
- Directory locking and security
- Device automation (caffeinate)
- Backup and restore capabilities
- And much more...

Built with security and user experience as top priorities.`,
	// Version is handled by the version command
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is $HOME/.ark/config.yaml)")

	// Set config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get home directory: %v", err))
	}

	configDir := filepath.Join(homeDir, ".ark")
	rootCmd.PersistentFlags().String("config-dir", configDir, "Configuration directory")

	// Add subcommands
	rootCmd.AddCommand(vault.VaultCmd)
	rootCmd.AddCommand(caffeinate.CaffeinateCmd)
	rootCmd.AddCommand(lock.LockCmd)
	rootCmd.AddCommand(awsCmd.Cmd)
	rootCmd.AddCommand(ec2Cmd.EC2Cmd)
	rootCmd.AddCommand(s3cmd.S3Cmd)
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(logs.LogsCmd)
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	configDir, _ := rootCmd.PersistentFlags().GetString("config-dir")
	return configDir
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	verbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
	return verbose
}
