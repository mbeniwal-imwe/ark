package vault

import (
	"fmt"
	"sort"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/vault"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for credentials in the vault",
	Long: `Search for credentials in the encrypted vault.

The search will match against key names, descriptions, tags, and content (for text format).
Use --format json or --format yaml for machine-readable output.

Examples:
  ark vault search aws
  ark vault search "api key"
  ark vault search database --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

var (
	searchFormat string
	searchTags   []string
)

func init() {
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, json, yaml)")
	searchCmd.Flags().StringSliceVarP(&searchTags, "tags", "t", []string{}, "Filter by tags")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

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

	// Search for credentials
	entries, err := vaultManager.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search credentials: %w", err)
	}

	// Filter by tags if specified
	if len(searchTags) > 0 {
		entries = filterByTags(entries, searchTags)
	}

	// Sort by key name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	// Display results
	switch searchFormat {
	case "json":
		return displayAsJSON(entries)
	case "yaml":
		return displayAsYAML(entries)
	default:
		return displaySearchResults(entries, query)
	}
}

func displaySearchResults(entries []*VaultEntry, query string) error {
	if len(entries) == 0 {
		fmt.Printf("No credentials found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d credential(s) matching '%s':\n\n", len(entries), query)
	return displayAsTable(entries)
}
