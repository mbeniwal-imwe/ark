package vault

import (
	"github.com/spf13/cobra"
)

// VaultCmd represents the vault command
var VaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage encrypted credentials and secrets",
	Long: `Vault commands allow you to securely store, retrieve, and manage your credentials and secrets.

The vault uses AES-256-GCM encryption to protect your sensitive data. All data is encrypted
using your master password and stored locally in an encrypted database.

Examples:
  ark vault set my-api-key "sk-1234567890abcdef"
  ark vault get my-api-key
  ark vault list
  ark vault search aws
  ark vault delete old-key`,
}

func init() {
	// Add vault subcommands
	VaultCmd.AddCommand(setCmd)
	VaultCmd.AddCommand(getCmd)
	VaultCmd.AddCommand(listCmd)
	VaultCmd.AddCommand(searchCmd)
	VaultCmd.AddCommand(deleteCmd)
	VaultCmd.AddCommand(updateCmd)
}

// Execute adds all child commands to the root command
func Execute() error {
	return VaultCmd.Execute()
}
