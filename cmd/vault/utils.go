package vault

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/storage/models"
	"gopkg.in/yaml.v3"
)

// VaultEntry represents a vault entry (alias for models.VaultEntry)
type VaultEntry = models.VaultEntry

// filterByTags filters vault entries by tags
func filterByTags(entries []*VaultEntry, tags []string) []*VaultEntry {
	var filtered []*VaultEntry

	for _, entry := range entries {
		hasAllTags := true
		for _, tag := range tags {
			if !entry.HasTag(tag) {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// displayAsTable displays vault entries in table format
func displayAsTable(entries []*VaultEntry) error {
	if len(entries) == 0 {
		fmt.Println("No credentials found in vault.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tFORMAT\tDESCRIPTION\tTAGS\tCREATED")
	fmt.Fprintln(w, "---\t------\t-----------\t----\t-------")

	for _, entry := range entries {
		tags := strings.Join(entry.Tags, ", ")
		if tags == "" {
			tags = "-"
		}

		description := entry.Description
		if description == "" {
			description = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			entry.Key,
			entry.Format,
			description,
			tags,
			entry.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	return w.Flush()
}

// displayAsJSON displays vault entries in JSON format
func displayAsJSON(entries []*VaultEntry) error {
	// Create a simplified structure for JSON output
	type VaultEntrySummary struct {
		Key         string   `json:"key"`
		Format      string   `json:"format"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		CreatedAt   string   `json:"created_at"`
		UpdatedAt   string   `json:"updated_at"`
	}

	var summaries []VaultEntrySummary
	for _, entry := range entries {
		summaries = append(summaries, VaultEntrySummary{
			Key:         entry.Key,
			Format:      entry.Format,
			Description: entry.Description,
			Tags:        entry.Tags,
			CreatedAt:   entry.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   entry.UpdatedAt.Format(time.RFC3339),
		})
	}

	jsonData, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// displayAsYAML displays vault entries in YAML format
func displayAsYAML(entries []*VaultEntry) error {
	// Create a simplified structure for YAML output
	type VaultEntrySummary struct {
		Key         string   `yaml:"key"`
		Format      string   `yaml:"format"`
		Description string   `yaml:"description"`
		Tags        []string `yaml:"tags"`
		CreatedAt   string   `yaml:"created_at"`
		UpdatedAt   string   `yaml:"updated_at"`
	}

	var summaries []VaultEntrySummary
	for _, entry := range entries {
		summaries = append(summaries, VaultEntrySummary{
			Key:         entry.Key,
			Format:      entry.Format,
			Description: entry.Description,
			Tags:        entry.Tags,
			CreatedAt:   entry.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   entry.UpdatedAt.Format(time.RFC3339),
		})
	}

	yamlData, err := yaml.Marshal(summaries)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	fmt.Println(string(yamlData))
	return nil
}

// getValueInteractively prompts user for input
func getValueInteractively() string {
	fmt.Print("Enter value: ")
	reader := bufio.NewReader(os.Stdin)
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}

// getValueFromStdin reads value from stdin
func getValueFromStdin() string {
	var value string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		value += scanner.Text() + "\n"
	}
	return strings.TrimSpace(value)
}
