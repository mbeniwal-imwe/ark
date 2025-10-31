package vault

import (
	"fmt"
	"strings"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/models"
)

// VaultManager manages vault operations
type VaultManager struct {
	db *storage.Database
}

// NewVaultManager creates a new vault manager
func NewVaultManager(db *storage.Database) *VaultManager {
	return &VaultManager{db: db}
}

// Set stores a value in the vault
func (vm *VaultManager) Set(key, value, format, description string, tags []string) error {
	// Validate format
	if !isValidFormat(format) {
		return fmt.Errorf("invalid format: %s. Supported formats: json, yaml, text", format)
	}

	// Create vault entry
	entry := models.NewVaultEntry(key, value, format)
	entry.SetDescription(description)

	// Add tags
	for _, tag := range tags {
		entry.AddTag(tag)
	}

	// Store in database
	return vm.db.Set("vault", key, entry)
}

// Get retrieves a value from the vault
func (vm *VaultManager) Get(key string) (*models.VaultEntry, error) {
	var entry models.VaultEntry
	err := vm.db.Get("vault", key, &entry)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault entry: %w", err)
	}

	// Update last accessed time
	entry.UpdatedAt = time.Now()
	vm.db.Set("vault", key, entry)

	return &entry, nil
}

// List returns all vault entries
func (vm *VaultManager) List() ([]*models.VaultEntry, error) {
	keys, err := vm.db.List("vault")
	if err != nil {
		return nil, fmt.Errorf("failed to list vault keys: %w", err)
	}

	var entries []*models.VaultEntry
	for _, key := range keys {
		entry, err := vm.Get(key)
		if err != nil {
			continue // Skip invalid entries
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Search searches for vault entries matching the query
func (vm *VaultManager) Search(query string) ([]*models.VaultEntry, error) {
	keys, err := vm.db.Search("vault", query)
	if err != nil {
		return nil, fmt.Errorf("failed to search vault: %w", err)
	}

	var entries []*models.VaultEntry
	for _, key := range keys {
		entry, err := vm.Get(key)
		if err != nil {
			continue // Skip invalid entries
		}

		// Additional client-side filtering
		if entry.MatchesSearch(query) {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Delete removes a vault entry
func (vm *VaultManager) Delete(key string) error {
	// Check if entry exists
	exists, err := vm.db.Exists("vault", key)
	if err != nil {
		return fmt.Errorf("failed to check if entry exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("vault entry '%s' not found", key)
	}

	return vm.db.Delete("vault", key)
}

// Update updates an existing vault entry
func (vm *VaultManager) Update(key, value, format, description string, tags []string) error {
	// Check if entry exists
	exists, err := vm.db.Exists("vault", key)
	if err != nil {
		return fmt.Errorf("failed to check if entry exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("vault entry '%s' not found", key)
	}

	// Get existing entry
	entry, err := vm.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get existing entry: %w", err)
	}

	// Update fields
	entry.Value = value
	entry.Format = format
	entry.SetDescription(description)
	entry.UpdatedAt = time.Now()

	// Update tags
	entry.Tags = []string{}
	for _, tag := range tags {
		entry.AddTag(tag)
	}

	// Store updated entry
	return vm.db.Set("vault", key, entry)
}

// Exists checks if a vault entry exists
func (vm *VaultManager) Exists(key string) (bool, error) {
	return vm.db.Exists("vault", key)
}

// GetByTag returns all vault entries with a specific tag
func (vm *VaultManager) GetByTag(tag string) ([]*models.VaultEntry, error) {
	entries, err := vm.List()
	if err != nil {
		return nil, err
	}

	var filtered []*models.VaultEntry
	for _, entry := range entries {
		if entry.HasTag(tag) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetByFormat returns all vault entries with a specific format
func (vm *VaultManager) GetByFormat(format string) ([]*models.VaultEntry, error) {
	entries, err := vm.List()
	if err != nil {
		return nil, err
	}

	var filtered []*models.VaultEntry
	for _, entry := range entries {
		if entry.Format == format {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// AddTag adds a tag to an existing vault entry
func (vm *VaultManager) AddTag(key, tag string) error {
	entry, err := vm.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get vault entry: %w", err)
	}

	entry.AddTag(tag)
	return vm.db.Set("vault", key, entry)
}

// RemoveTag removes a tag from an existing vault entry
func (vm *VaultManager) RemoveTag(key, tag string) error {
	entry, err := vm.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get vault entry: %w", err)
	}

	entry.RemoveTag(tag)
	return vm.db.Set("vault", key, entry)
}

// SetMetadata sets metadata for a vault entry
func (vm *VaultManager) SetMetadata(key, metaKey string, value interface{}) error {
	entry, err := vm.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get vault entry: %w", err)
	}

	entry.SetMetadata(metaKey, value)
	return vm.db.Set("vault", key, entry)
}

// GetMetadata retrieves metadata from a vault entry
func (vm *VaultManager) GetMetadata(key, metaKey string) (interface{}, bool, error) {
	entry, err := vm.Get(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get vault entry: %w", err)
	}

	value, exists := entry.GetMetadata(metaKey)
	return value, exists, nil
}

// Clear removes all vault entries
func (vm *VaultManager) Clear() error {
	keys, err := vm.db.List("vault")
	if err != nil {
		return fmt.Errorf("failed to list vault keys: %w", err)
	}

	for _, key := range keys {
		if err := vm.db.Delete("vault", key); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", key, err)
		}
	}

	return nil
}

// isValidFormat checks if the format is valid
func isValidFormat(format string) bool {
	validFormats := []string{"json", "yaml", "text"}
	format = strings.ToLower(format)

	for _, valid := range validFormats {
		if format == valid {
			return true
		}
	}

	return false
}
