package vault

import (
	"encoding/json"
	"fmt"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve a credential or secret from the vault",
	Long: `Retrieve a credential or secret from the encrypted vault.

The value will be displayed in its original format (JSON, YAML, or plain text).
You can copy the output to clipboard or pipe it to other commands.

Examples:
  ark vault get my-api-key
  ark vault get aws-credentials | jq .
  ark vault get database-config --format yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

var (
	outputFormat string
	showMetadata bool
)

func init() {
	getCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Override output format (json, yaml, text)")
	getCmd.Flags().BoolVarP(&showMetadata, "metadata", "m", false, "Show metadata and tags")
}

func runGet(cmd *cobra.Command, args []string) error {
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

	// Get the credential
	entry, err := vaultManager.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	// Determine output format
	format := outputFormat
	if format == "" {
		format = entry.Format
	}

	// Display the value
	if showMetadata {
		displayWithMetadata(entry, format)
	} else {
		displayValue(entry.Value, format)
	}

	return nil
}

func displayValue(value, format string) {
	switch format {
	case "json":
		var data interface{}
		if err := json.Unmarshal([]byte(value), &data); err != nil {
			// If not valid JSON, display as text
			fmt.Print(value)
		} else {
			// Pretty print JSON
			prettyJSON, _ := json.MarshalIndent(data, "", "  ")
			fmt.Print(string(prettyJSON))
		}
	case "yaml":
		var data interface{}
		if err := yaml.Unmarshal([]byte(value), &data); err != nil {
			// If not valid YAML, display as text
			fmt.Print(value)
		} else {
			// Pretty print YAML
			prettyYAML, _ := yaml.Marshal(data)
			fmt.Print(string(prettyYAML))
		}
	default:
		fmt.Print(value)
	}
}

func displayWithMetadata(entry *VaultEntry, format string) {
	fmt.Printf("Key: %s\n", entry.Key)
	fmt.Printf("Format: %s\n", entry.Format)
	fmt.Printf("Description: %s\n", entry.Description)
	fmt.Printf("Tags: %v\n", entry.Tags)
	fmt.Printf("Created: %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", entry.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("---")
	displayValue(entry.Value, format)
}
