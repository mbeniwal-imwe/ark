package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a credential from the vault",
	Long: `Delete a credential from the encrypted vault.

This action cannot be undone. The credential will be permanently removed.

Examples:
  ark vault delete my-api-key
  ark vault delete old-credential --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

var (
	forceDelete bool
)

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Load configuration
	configDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
	cfg, err := config.Load(configDir)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Open database (GetMasterKey will handle password verification)
	masterKey, err := cfg.GetMasterKey()
	if err != nil {
		return err
	}
	db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create vault manager
	vaultManager := vault.NewVaultManager(db)

	// Check if credential exists
	exists, err := vaultManager.Exists(key)
	if err != nil {
		return fmt.Errorf("failed to check if credential exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("credential '%s' not found", key)
	}

	// Confirm deletion unless --force is used
	if !forceDelete {
		if !confirmDeletion(key) {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the credential
	if err := vaultManager.Delete(key); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	fmt.Printf("✅ Successfully deleted credential '%s' from vault\n", key)
	return nil
}

func confirmDeletion(key string) bool {
	fmt.Printf("Are you sure you want to delete credential '%s'? This action cannot be undone. [y/N]: ", key)

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}
