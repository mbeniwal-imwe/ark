package models

import (
	"time"
)

// LockedDirectory represents a locked directory
type LockedDirectory struct {
	Path         string            `json:"path"`
	Password     string            `json:"-"` // Not serialized
	UseMaster    bool              `json:"use_master"`
	Hidden       bool              `json:"hidden"`
	Encrypted    bool              `json:"encrypted"`
	LockedAt     time.Time         `json:"locked_at"`
	LastAccessed time.Time         `json:"last_accessed,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NewLockedDirectory creates a new locked directory record
func NewLockedDirectory(path string, useMaster bool, hidden bool) *LockedDirectory {
	now := time.Now()
	return &LockedDirectory{
		Path:      path,
		UseMaster: useMaster,
		Hidden:    hidden,
		Encrypted: true,
		LockedAt:  now,
		Metadata:  make(map[string]string),
	}
}

// SetPassword sets the password for the directory
func (d *LockedDirectory) SetPassword(password string) {
	d.Password = password
	d.UseMaster = false
}

// SetMasterPassword indicates the directory uses the master password
func (d *LockedDirectory) SetMasterPassword() {
	d.Password = ""
	d.UseMaster = true
}

// UpdateLastAccessed updates the last accessed time
func (d *LockedDirectory) UpdateLastAccessed() {
	d.LastAccessed = time.Now()
}

// SetMetadata sets metadata for the directory
func (d *LockedDirectory) SetMetadata(key, value string) {
	if d.Metadata == nil {
		d.Metadata = make(map[string]string)
	}
	d.Metadata[key] = value
}

// IsLocked checks if the directory is currently locked
func (d *LockedDirectory) IsLocked() bool {
	return d.Encrypted
}

// GetPassword returns the password to use for this directory
func (d *LockedDirectory) GetPassword(masterPassword string) string {
	if d.UseMaster {
		return masterPassword
	}
	return d.Password
}
