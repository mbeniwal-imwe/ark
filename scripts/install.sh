#!/bin/bash

# Ark CLI Installation Script
# This script installs Ark CLI and updates the shell configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="ark"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.ark"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect the shell
detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    else
        # Try to detect from $SHELL
        case "$SHELL" in
            */zsh) echo "zsh" ;;
            */bash) echo "bash" ;;
            *) echo "unknown" ;;
        esac
    fi
}

# Function to get the appropriate rc file
get_rc_file() {
    local shell_type="$1"
    case "$shell_type" in
        "zsh")
            if [ -f "$HOME/.zshrc" ]; then
                echo "$HOME/.zshrc"
            else
                echo "$HOME/.zprofile"
            fi
            ;;
        "bash")
            if [ -f "$HOME/.bashrc" ]; then
                echo "$HOME/.bashrc"
            else
                echo "$HOME/.bash_profile"
            fi
            ;;
        *)
            echo "$HOME/.profile"
            ;;
    esac
}

# Function to check if PATH already contains the install directory
is_path_configured() {
    local rc_file="$1"
    if [ -f "$rc_file" ]; then
        grep -q "export PATH.*$INSTALL_DIR" "$rc_file" 2>/dev/null
    else
        return 1
    fi
}

# Function to add PATH to shell configuration
add_to_path() {
    local rc_file="$1"
    local shell_type="$2"
    
    print_status "Adding $INSTALL_DIR to PATH in $rc_file"
    
    # Create rc file if it doesn't exist
    if [ ! -f "$rc_file" ]; then
        touch "$rc_file"
    fi
    
    # Add PATH export if not already present
    if ! is_path_configured "$rc_file"; then
        echo "" >> "$rc_file"
        echo "# Ark CLI PATH configuration" >> "$rc_file"
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$rc_file"
        print_success "Added PATH configuration to $rc_file"
    else
        print_warning "PATH already configured in $rc_file"
    fi
}

# Function to create configuration directory
create_config_dir() {
    print_status "Creating configuration directory: $CONFIG_DIR"
    
    mkdir -p "$CONFIG_DIR"/{data,logs,config,backup}
    
    # Set appropriate permissions
    chmod 700 "$CONFIG_DIR"
    chmod 700 "$CONFIG_DIR"/data
    chmod 700 "$CONFIG_DIR"/logs
    chmod 700 "$CONFIG_DIR"/config
    chmod 700 "$CONFIG_DIR"/backup
    
    print_success "Configuration directory created"
}

# Function to check if running on macOS
check_macos() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        print_warning "This script is optimized for macOS. Some features may not work on other platforms."
    fi
}

# Function to check for required dependencies
check_dependencies() {
    print_status "Checking dependencies..."
    
    # Check for Go (optional, for building from source)
    if ! command -v go &> /dev/null; then
        print_warning "Go is not installed. You'll need to install a pre-built binary."
    fi
    
    # Check for make (optional, for building from source)
    if ! command -v make &> /dev/null; then
        print_warning "Make is not installed. You'll need to install a pre-built binary."
    fi
    
    print_success "Dependency check complete"
}

# Function to install the binary
install_binary() {
    local binary_path="$1"
    
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found at $binary_path"
        print_status "Please build the binary first using 'make build' or download a pre-built binary"
        exit 1
    fi
    
    print_status "Installing binary to $INSTALL_DIR"
    
    # Copy binary to install directory
    cp "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    print_success "Binary installed successfully"
}

# Function to verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    
    if [ -f "$binary_path" ]; then
        local version=$("$binary_path" version 2>/dev/null || echo "unknown")
        print_success "Ark CLI installed successfully"
        print_status "Version: $version"
        print_status "Location: $binary_path"
    else
        print_error "Installation verification failed - binary not found at $binary_path"
        return 1
    fi
}

# Function to show post-installation instructions
show_post_install_instructions() {
    echo
    print_success "Installation completed successfully!"
    echo
    echo "Next steps:"
    echo "1. Restart your terminal or run: source $rc_file"
    echo "2. Initialize Ark CLI: $BINARY_NAME init"
    echo "3. Start using Ark: $BINARY_NAME --help"
    echo
    echo "Configuration directory: $CONFIG_DIR"
    echo "Binary location: $INSTALL_DIR/$BINARY_NAME"
    echo
}

# Main installation function
main() {
    echo "Ark CLI Installation Script"
    echo "=========================="
    echo
    
    # Check if running on macOS
    check_macos
    
    # Check dependencies
    check_dependencies
    
    # Detect shell
    local shell_type=$(detect_shell)
    print_status "Detected shell: $shell_type"
    
    # Get rc file
    local rc_file=$(get_rc_file "$shell_type")
    print_status "Using configuration file: $rc_file"
    
    # Check if binary exists in build directory
    local build_binary="./build/$BINARY_NAME"
    if [ -f "$build_binary" ]; then
        install_binary "$build_binary"
    else
        print_error "Binary not found. Please build it first using 'make build'"
        exit 1
    fi
    
    # Create configuration directory
    create_config_dir
    
    # Add to PATH
    add_to_path "$rc_file" "$shell_type"
    
    # Verify installation
    if verify_installation; then
        show_post_install_instructions
    else
        print_error "Installation failed"
        exit 1
    fi
}

# Run main function
main "$@"
