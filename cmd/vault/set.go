package vault

import (
	"fmt"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Store a credential or secret in the vault",
	Long: `Store a credential or secret in the encrypted vault.

If no value is provided, you will be prompted to enter it interactively.
The value can be in JSON, YAML, or plain text format.

Examples:
  ark vault set my-api-key "sk-1234567890abcdef"
  ark vault set aws-credentials --format json
  ark vault set database-config --format yaml --description "Production DB config"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSet,
}

var (
	format      string
	description string
	tags        []string
	interactive bool
)

func init() {
	setCmd.Flags().StringVarP(&format, "format", "f", "text", "Format of the value (json, yaml, text)")
	setCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the credential")
	setCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to associate with the credential")
	setCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enter value interactively")
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	var value string

	// Get value from arguments or stdin
	if len(args) > 1 {
		value = args[1]
	} else if interactive {
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

	// Store the credential
	if err := vaultManager.Set(key, value, format, description, tags); err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	fmt.Printf("✅ Successfully stored credential '%s' in vault\n", key)
	return nil
}
