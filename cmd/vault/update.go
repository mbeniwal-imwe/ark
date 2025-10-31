package vault

import (
	"fmt"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <key> [value]",
	Short: "Update an existing credential in the vault",
	Long: `Update an existing credential in the encrypted vault.

If no value is provided, you will be prompted to enter it interactively.
The value can be in JSON, YAML, or plain text format.

Examples:
  ark vault update my-api-key "new-api-key-value"
  ark vault update aws-credentials --format json
  ark vault update database-config --description "Updated DB config"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUpdate,
}

var (
	updateFormat      string
	updateDescription string
	updateTags        []string
	updateInteractive bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateFormat, "format", "f", "", "Format of the value (json, yaml, text)")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "Description of the credential")
	updateCmd.Flags().StringSliceVarP(&updateTags, "tags", "t", []string{}, "Tags to associate with the credential")
	updateCmd.Flags().BoolVarP(&updateInteractive, "interactive", "i", false, "Enter value interactively")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	key := args[0]
	var value string

	// Get value from arguments or stdin
	if len(args) > 1 {
		value = args[1]
	} else if updateInteractive {
		value = getValueInteractively()
	} else {
		value = getValueFromStdin()
	}

	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}

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

	// Get existing entry to preserve format if not specified
	if updateFormat == "" {
		entry, err := vaultManager.Get(key)
		if err != nil {
			return fmt.Errorf("failed to get existing credential: %w", err)
		}
		updateFormat = entry.Format
	}

	// Update the credential
	if err := vaultManager.Update(key, value, updateFormat, updateDescription, updateTags); err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	fmt.Printf("✅ Successfully updated credential '%s' in vault\n", key)
	return nil
}
