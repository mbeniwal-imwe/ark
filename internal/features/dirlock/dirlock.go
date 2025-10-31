package dirlock

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/mbeniwal-imwe/ark/internal/storage/models"
)

type Service struct {
	DB *storage.Database
}

// getMasterKey retrieves the master key from config
func (s *Service) getMasterKey() ([]byte, error) {
	// This is a simplified approach - in production, get from config
	// For now, return a placeholder
	return make([]byte, 32), nil
}

func (s *Service) Lock(path string, useMaster bool, password string, hide bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Basic safeguard: directory must exist
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", abs)
	}

	// Derive encryption key
	var key []byte
	if useMaster {
		// Use master key from config
		key, err = s.getMasterKey()
		if err != nil {
			return err
		}
	} else {
		// Derive key from custom password
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return err
		}
		key, err = deriveKeyFromPassword(password, salt)
		if err != nil {
			return err
		}
	}

	// Encrypt directory content
	if err := EncryptDirectory(abs, key); err != nil {
		return fmt.Errorf("failed to encrypt directory: %w", err)
	}

	// Restrict permissions
	_ = os.Chmod(abs, 0000)

	// Optionally hide directory (macOS)
	if hide {
		_ = exec.Command("chflags", "hidden", abs).Run()
	}

	rec := models.NewLockedDirectory(abs, useMaster, hide)
	if !useMaster {
		rec.SetPassword(password)
	}
	rec.SetMetadata("mode", "encrypted")
	return s.DB.Set("locked_dirs", abs, rec)
}

func (s *Service) Unlock(path string, masterPassword string, provided string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	var rec models.LockedDirectory
	if err := s.DB.Get("locked_dirs", abs, &rec); err != nil {
		return fmt.Errorf("not locked: %s", abs)
	}

	// Password check
	pass := rec.GetPassword(masterPassword)
	if pass != provided && !rec.UseMaster {
		return fmt.Errorf("invalid password for %s", abs)
	}

	// Derive decryption key
	var key []byte
	if rec.UseMaster {
		key, err = s.getMasterKey()
		if err != nil {
			return err
		}
	} else {
		// For custom passwords, we need to derive the same key
		// This is a simplified approach - in production, store salt
		salt := make([]byte, 16) // Use stored salt in production
		key, err = deriveKeyFromPassword(provided, salt)
		if err != nil {
			return err
		}
	}

	// Decrypt directory content
	if err := DecryptDirectory(abs, key); err != nil {
		return fmt.Errorf("failed to decrypt directory: %w", err)
	}

	// Unhide and restore permissions
	_ = exec.Command("chflags", "nohidden", abs).Run()
	_ = os.Chmod(abs, 0700)

	rec.Encrypted = false
	rec.UpdateLastAccessed()
	_ = s.DB.Delete("locked_dirs", abs)
	return nil
}

func (s *Service) List() ([]models.LockedDirectory, error) {
	keys, err := s.DB.List("locked_dirs")
	if err != nil {
		return nil, err
	}
	var out []models.LockedDirectory
	for _, k := range keys {
		var rec models.LockedDirectory
		if err := s.DB.Get("locked_dirs", k, &rec); err == nil {
			out = append(out, rec)
		}
	}
	return out, nil
}

func (s *Service) IsLocked(path string) (bool, error) {
	abs := path
	if !strings.HasPrefix(path, "/") {
		a, err := filepath.Abs(path)
		if err != nil {
			return false, err
		}
		abs = a
	}
	return s.DB.Exists("locked_dirs", abs)
}

// Stamp updates last accessed time safe
func (s *Service) Stamp(path string) {
	var rec models.LockedDirectory
	if err := s.DB.Get("locked_dirs", path, &rec); err == nil {
		rec.LastAccessed = time.Now()
		_ = s.DB.Set("locked_dirs", path, rec)
	}
}
