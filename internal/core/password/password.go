package password

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// SetupMasterPassword prompts the user to set up a master password
func SetupMasterPassword() (string, error) {
	fmt.Println("Setting up master password for Ark CLI...")
	fmt.Println("This password will be used to encrypt all your sensitive data.")
	fmt.Println()

	// Get password
	password, err := getPassword("Enter master password: ")
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters long")
	}

	// Confirm password
	confirmPassword, err := getPassword("Confirm master password: ")
	if err != nil {
		return "", fmt.Errorf("failed to read confirmation password: %w", err)
	}

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	fmt.Println("✅ Master password set successfully!")
	return password, nil
}

// GetMasterPassword prompts the user for their master password
func GetMasterPassword() (string, error) {
	password, err := getPassword("Enter master password: ")
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	return password, nil
}

// getPassword securely reads a password from stdin
func getPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Check if stdin is a terminal
	if terminal.IsTerminal(int(syscall.Stdin)) {
		// Read password without echoing
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", err
		}
		fmt.Println() // Add newline after password input
		return string(password), nil
	}

	// Fallback for non-terminal input (e.g., pipes)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(password), nil
}

// GetPasswordWithConfirmation prompts for a password with confirmation
func GetPasswordWithConfirmation(prompt, confirmPrompt string) (string, error) {
	password, err := getPassword(prompt)
	if err != nil {
		return "", err
	}

	confirmPassword, err := getPassword(confirmPrompt)
	if err != nil {
		return "", err
	}

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

// ValidatePasswordStrength checks if a password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 126:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
