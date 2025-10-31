package models

import (
	"strings"
	"time"
)

// VaultEntry represents a vault entry
type VaultEntry struct {
	Key         string                 `json:"key"`
	Value       string                 `json:"value"`
	Format      string                 `json:"format"` // json, yaml, text
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NewVaultEntry creates a new vault entry
func NewVaultEntry(key, value, format string) *VaultEntry {
	now := time.Now()
	return &VaultEntry{
		Key:       key,
		Value:     value,
		Format:    format,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]interface{}),
	}
}

// SetDescription sets the description for the entry
func (e *VaultEntry) SetDescription(desc string) {
	e.Description = desc
	e.UpdatedAt = time.Now()
}

// AddTag adds a tag to the entry
func (e *VaultEntry) AddTag(tag string) {
	for _, existingTag := range e.Tags {
		if existingTag == tag {
			return // Tag already exists
		}
	}
	e.Tags = append(e.Tags, tag)
	e.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the entry
func (e *VaultEntry) RemoveTag(tag string) {
	for i, existingTag := range e.Tags {
		if existingTag == tag {
			e.Tags = append(e.Tags[:i], e.Tags[i+1:]...)
			e.UpdatedAt = time.Now()
			break
		}
	}
}

// SetMetadata sets metadata for the entry
func (e *VaultEntry) SetMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	e.UpdatedAt = time.Now()
}

// GetMetadata retrieves metadata for the entry
func (e *VaultEntry) GetMetadata(key string) (interface{}, bool) {
	if e.Metadata == nil {
		return nil, false
	}
	value, exists := e.Metadata[key]
	return value, exists
}

// HasTag checks if the entry has a specific tag
func (e *VaultEntry) HasTag(tag string) bool {
	for _, existingTag := range e.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// MatchesSearch checks if the entry matches a search query
func (e *VaultEntry) MatchesSearch(query string) bool {
	query = strings.ToLower(query)

	// Check key
	if strings.Contains(strings.ToLower(e.Key), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(e.Description), query) {
		return true
	}

	// Check tags
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Check value (for text format)
	if e.Format == "text" && strings.Contains(strings.ToLower(e.Value), query) {
		return true
	}

	return false
}
