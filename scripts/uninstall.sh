#!/bin/bash

# Ark CLI Uninstallation Script
# This script removes Ark CLI and cleans up all associated files

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

# Function to check if Ark is installed
is_installed() {
    command -v "$BINARY_NAME" &> /dev/null
}

# Function to remove binary
remove_binary() {
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    
    if [ -f "$binary_path" ]; then
        print_status "Removing binary from $binary_path"
        rm -f "$binary_path"
        print_success "Binary removed"
    else
        print_warning "Binary not found at $binary_path"
    fi
}

# Function to stop running Ark processes
stop_running_processes() {
    print_status "Checking for running Ark processes..."
    
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    local pids=""
    
    # Use pgrep if available (more reliable)
    if command -v pgrep &> /dev/null; then
        # Only match processes where the executable name is exactly "ark"
        pids=$(pgrep -x "ark" 2>/dev/null || true)
    else
        # Fallback: Use ps and filter more carefully
        # Match only processes where:
        # 1. The command is exactly "ark" (not containing "ark" as a substring)
        # 2. Or the executable path ends with "/ark"
        # 3. Exclude shell scripts, grep, ps, uninstall script, etc.
        pids=$(ps aux | awk '
            $11 ~ /^[\/].*\/ark$/ || ($11 == "ark" && $0 !~ /uninstall|grep|ps|awk|sh|bash|zsh/) {
                print $2
            }' | grep -v "^$$")
    fi
    
    # Also check for caffeinate processes that might be managed by ark
    # Look for caffeinate processes that have ark-related PIDs or are in ark config dir
    local caffeinate_pids=""
    if [ -d "$CONFIG_DIR/data" ]; then
        # Check for PID files that ark creates for caffeinate
        local pid_file="$CONFIG_DIR/data/caffeinate.pid"
        if [ -f "$pid_file" ]; then
            local cached_pid=$(cat "$pid_file" 2>/dev/null | tr -d '[:space:]')
            if [ -n "$cached_pid" ] && kill -0 "$cached_pid" 2>/dev/null; then
                # Verify it's actually a caffeinate process
                if ps -p "$cached_pid" -o comm= 2>/dev/null | grep -q "caffeinate"; then
                    caffeinate_pids="$cached_pid"
                fi
            fi
        fi
    fi
    
    # Combine PIDs
    if [ -n "$caffeinate_pids" ]; then
        if [ -z "$pids" ]; then
            pids="$caffeinate_pids"
        else
            pids="$pids $caffeinate_pids"
        fi
    fi
    
    if [ -z "$pids" ]; then
        print_status "No running Ark processes found"
        return 0
    fi
    
    print_warning "Found running Ark processes. Stopping them..."
    for pid in $pids; do
        # Double-check it's not the current script or its parent
        if [ "$pid" = "$$" ] || [ "$pid" = "$PPID" ]; then
            continue
        fi
        
        # Verify the process still exists and is actually ark-related
        if kill -0 "$pid" 2>/dev/null; then
            local proc_cmd=$(ps -p "$pid" -o comm= 2>/dev/null)
            if [ "$proc_cmd" = "ark" ] || [ "$proc_cmd" = "caffeinate" ]; then
                print_status "Stopping process $pid ($proc_cmd)"
                kill "$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null
            fi
        fi
    done
    
    # Wait a bit for processes to stop
    sleep 1
    
    print_success "Ark processes cleanup completed"
}

# Function to unlock all locked directories
unlock_directories() {
    print_status "Unlocking all locked directories..."
    
    if [ -d "$CONFIG_DIR/data" ]; then
        # This would need to be implemented in the actual Ark CLI
        # For now, we'll just show a message
        print_warning "Please manually unlock any locked directories using: $BINARY_NAME lock list"
        print_warning "Then unlock each directory using: $BINARY_NAME unlock <directory>"
    fi
}

# Function to remove PATH configuration
remove_from_path() {
    local rc_file="$1"
    
    if [ -f "$rc_file" ]; then
        print_status "Removing PATH configuration from $rc_file"
        
        # Create a backup of the original file
        cp "$rc_file" "$rc_file.backup.$(date +%Y%m%d_%H%M%S)"
        
        # Remove Ark CLI PATH configuration
        # Escape the INSTALL_DIR for use in sed
        local escaped_dir=$(echo "$INSTALL_DIR" | sed 's/[[\.*^$()+?{|]/\\&/g')
        
        # Create temporary file for editing (macOS sed requires this)
        local temp_file="${rc_file}.ark_uninstall"
        
        # Remove any line containing Ark CLI PATH configuration comment
        grep -v "# Ark CLI PATH configuration" "$rc_file" > "$temp_file"
        # Remove PATH export for the install directory
        # Match: export PATH="/path/to/install:$PATH"
        sed "s|export PATH=\"${escaped_dir}:\$PATH\"||g" "$temp_file" > "${temp_file}.2"
        # Also match with $HOME expansion if used
        sed "s|export PATH=\"\$HOME/.local/bin:\$PATH\"||g" "${temp_file}.2" > "${temp_file}.3"
        # Remove empty lines
        sed '/^$/d' "${temp_file}.3" > "$temp_file"
        
        # Replace original file
        mv "$temp_file" "$rc_file"
        rm -f "${temp_file}.2" "${temp_file}.3"
        
        print_success "PATH configuration removed from $rc_file"
    else
        print_warning "Configuration file $rc_file not found"
    fi
}

# Function to remove configuration directory
remove_config_dir() {
    if [ -d "$CONFIG_DIR" ]; then
        print_status "Configuration directory found: $CONFIG_DIR"
        echo -n "Do you want to remove the configuration directory and all data? [y/N]: "
        read -r response
        
        case "$response" in
            [yY]|[yY][eE][sS])
                print_status "Removing configuration directory..."
                rm -rf "$CONFIG_DIR"
                print_success "Configuration directory removed"
                ;;
            *)
                print_warning "Configuration directory preserved at $CONFIG_DIR"
                ;;
        esac
    else
        print_warning "Configuration directory not found"
    fi
}

# Function to verify uninstallation
verify_uninstallation() {
    print_status "Verifying uninstallation..."
    
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    
    # Check if binary file still exists
    if [ -f "$binary_path" ]; then
        print_error "Uninstallation verification failed - binary still exists at $binary_path"
        return 1
    fi
    
    # Check if binary is still in PATH (might be cached in current shell)
    if command -v "$BINARY_NAME" &> /dev/null; then
        local found_path=$(which "$BINARY_NAME" 2>/dev/null || command -v "$BINARY_NAME")
        if [ "$found_path" != "$binary_path" ]; then
            print_warning "Binary found in PATH at $found_path (may be from another location or cached)"
        else
            print_warning "Binary may still be cached in current shell - restart your terminal"
        fi
    fi
    
    print_success "Ark CLI successfully uninstalled"
    return 0
}

# Function to show post-uninstallation instructions
show_post_uninstall_instructions() {
    echo
    print_success "Uninstallation completed successfully!"
    echo
    echo "What was removed:"
    echo "- Binary: $INSTALL_DIR/$BINARY_NAME"
    echo "- PATH configuration from shell rc file"
    echo "- Configuration directory (if you chose to remove it)"
    echo
    echo "Note: You may need to restart your terminal for PATH changes to take effect."
    echo
}

# Function to confirm uninstallation
confirm_uninstallation() {
    echo "This will uninstall Ark CLI and remove all associated files."
    echo
    echo "The following will be removed:"
    echo "- Binary: $INSTALL_DIR/$BINARY_NAME"
    echo "- PATH configuration from shell rc file"
    echo "- Configuration directory: $CONFIG_DIR (optional)"
    echo
    echo -n "Are you sure you want to continue? [y/N]: "
    read -r response
    
    case "$response" in
        [yY]|[yY][eE][sS])
            return 0
            ;;
        *)
            print_status "Uninstallation cancelled"
            exit 0
            ;;
    esac
}

# Main uninstallation function
main() {
    echo "Ark CLI Uninstallation Script"
    echo "============================="
    echo
    
    # Check if Ark is installed
    if ! is_installed; then
        print_warning "Ark CLI is not installed or not in PATH"
        exit 0
    fi
    
    # Confirm uninstallation
    confirm_uninstallation
    
    # Detect shell
    local shell_type=$(detect_shell)
    print_status "Detected shell: $shell_type"
    
    # Get rc file
    local rc_file=$(get_rc_file "$shell_type")
    print_status "Using configuration file: $rc_file"
    
    # Stop running processes
    stop_running_processes
    
    # Unlock directories
    unlock_directories
    
    # Remove binary
    remove_binary
    
    # Remove PATH configuration
    remove_from_path "$rc_file"
    
    # Remove configuration directory
    remove_config_dir
    
    # Verify uninstallation
    if verify_uninstallation; then
        show_post_uninstall_instructions
    else
        print_error "Uninstallation failed"
        exit 1
    fi
}

# Run main function
main "$@"

