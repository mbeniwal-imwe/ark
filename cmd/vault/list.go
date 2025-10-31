package vault

import (
	"fmt"
	"sort"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credentials in the vault",
	Long: `List all credentials stored in the encrypted vault.

The output shows key names, formats, descriptions, and creation dates.
Use --format json or --format yaml for machine-readable output.

Examples:
  ark vault list
  ark vault list --format json
  ark vault list --tags aws,database`,
	RunE: runList,
}

var (
	listFormat string
	listTags   []string
	listFilter string
)

func init() {
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "Output format (table, json, yaml)")
	listCmd.Flags().StringSliceVarP(&listTags, "tags", "t", []string{}, "Filter by tags")
	listCmd.Flags().StringVarP(&listFilter, "filter", "", "", "Filter by key name or description")
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Get entries
	var entries []*VaultEntry
	if listFilter != "" {
		entries, err = vaultManager.Search(listFilter)
	} else {
		entries, err = vaultManager.List()
	}
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	// Filter by tags if specified
	if len(listTags) > 0 {
		entries = filterByTags(entries, listTags)
	}

	// Sort by key name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	// Display results
	switch listFormat {
	case "json":
		return displayAsJSON(entries)
	case "yaml":
		return displayAsYAML(entries)
	default:
		return displayAsTable(entries)
	}
}
