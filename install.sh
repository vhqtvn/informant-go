#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Repository information
REPO="vhqtvn/informant-go"
BINARY_NAME="informant"

# Installation settings
INSTALL_DIR="/usr/local/bin"
HOOK_INSTALL=${HOOK_INSTALL:-true}

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            log_error "Supported architectures: x86_64 (amd64), aarch64 (arm64)"
            exit 1
            ;;
    esac
}

# Detect OS
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $os in
        linux)
            echo "linux"
            ;;
        *)
            log_error "Unsupported OS: $os"
            log_error "This installer currently supports Linux only"
            exit 1
            ;;
    esac
}

# Check if running as root for system installation
check_root() {
    if [[ $EUID -eq 0 ]]; then
        return 0
    else
        return 1
    fi
}

# Get the latest release version
get_latest_version() {
    log_info "Fetching latest release information..."
    local version
    version=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "$version" ]]; then
        log_error "Failed to fetch latest release version"
        exit 1
    fi
    echo "$version"
}

# Download file with progress
download_file() {
    local url="$1"
    local output="$2"
    log_info "Downloading: $(basename "$output")"
    if command -v curl >/dev/null 2>&1; then
        curl -L --progress-bar "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget --progress=bar:force "$url" -O "$output"
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local expected_checksum="$2"
    
    if ! command -v sha256sum >/dev/null 2>&1; then
        log_warning "sha256sum not found, skipping checksum verification"
        return 0
    fi
    
    log_info "Verifying checksum..."
    local actual_checksum
    actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
    
    if [[ "$actual_checksum" == "$expected_checksum" ]]; then
        log_success "Checksum verification passed"
        return 0
    else
        log_error "Checksum verification failed!"
        log_error "Expected: $expected_checksum"
        log_error "Actual:   $actual_checksum"
        exit 1
    fi
}

# Install binary
install_binary() {
    local binary_path="$1"
    local install_path="$INSTALL_DIR/$BINARY_NAME"
    
    if check_root; then
        log_info "Installing to $install_path"
        chmod +x "$binary_path"
        cp "$binary_path" "$install_path"
        log_success "Binary installed to $install_path"
    else
        log_warning "Not running as root. Installing to ~/.local/bin/"
        local user_bin="$HOME/.local/bin"
        mkdir -p "$user_bin"
        chmod +x "$binary_path"
        cp "$binary_path" "$user_bin/$BINARY_NAME"
        log_success "Binary installed to $user_bin/$BINARY_NAME"
        
        # Check if ~/.local/bin is in PATH
        if [[ ":$PATH:" != *":$user_bin:"* ]]; then
            log_warning "~/.local/bin is not in your PATH"
            log_warning "Add the following to your shell profile (.bashrc, .zshrc, etc.):"
            echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
        fi
        
        # Update INSTALL_DIR for hook installation
        INSTALL_DIR="$user_bin"
    fi
}

# Install pacman hook
install_hook() {
    if [[ "$HOOK_INSTALL" != "true" ]]; then
        log_info "Skipping pacman hook installation (HOOK_INSTALL=false)"
        return 0
    fi
    
    if ! check_root; then
        log_warning "Pacman hook installation requires root privileges"
        log_info "Run the following command after installation to install the hook:"
        echo "    sudo $INSTALL_DIR/$BINARY_NAME install"
        return 0
    fi
    
    # Check if this is an Arch-based system
    if [[ ! -f /etc/pacman.conf ]]; then
        log_info "Not an Arch Linux system, skipping pacman hook installation"
        return 0
    fi
    
    log_info "Installing pacman hook..."
    if "$INSTALL_DIR/$BINARY_NAME" install; then
        log_success "Pacman hook installed successfully"
    else
        log_warning "Failed to install pacman hook. You can install it manually later with:"
        echo "    sudo $BINARY_NAME install"
    fi
}

# Cleanup function
cleanup() {
    local temp_dir="$1"
    if [[ -n "$temp_dir" && -d "$temp_dir" ]]; then
        rm -rf "$temp_dir"
    fi
}

# Main installation function
main() {
    echo -e "${BLUE}"
    echo "========================================="
    echo "         InformantGo Installer"
    echo "   Arch Linux News Reader & Hook"
    echo "========================================="
    echo -e "${NC}"
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        log_error "Either curl or wget is required for installation"
        exit 1
    fi
    
    # Detect system
    local os arch version
    os=$(detect_os)
    arch=$(detect_arch)
    version=$(get_latest_version)
    
    log_info "System: $os/$arch"
    log_info "Latest version: $version"
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    trap "cleanup '$temp_dir'" EXIT
    
    # Download binary and checksums
    local binary_name="informant-$os-$arch"
    local binary_url="https://github.com/$REPO/releases/download/$version/$binary_name"
    local checksums_url="https://github.com/$REPO/releases/download/$version/checksums.txt"
    
    local binary_path="$temp_dir/$binary_name"
    local checksums_path="$temp_dir/checksums.txt"
    
    download_file "$binary_url" "$binary_path"
    download_file "$checksums_url" "$checksums_path"
    
    # Verify checksum
    local expected_checksum
    expected_checksum=$(grep "$binary_name" "$checksums_path" | cut -d' ' -f1)
    if [[ -n "$expected_checksum" ]]; then
        verify_checksum "$binary_path" "$expected_checksum"
    else
        log_warning "Could not find checksum for $binary_name, skipping verification"
    fi
    
    # Install binary
    install_binary "$binary_path"
    
    # Install pacman hook
    install_hook
    
    # Show success message and usage
    echo
    log_success "InformantGo installed successfully!"
    echo
    echo -e "${BLUE}Usage:${NC}"
    echo "  $BINARY_NAME check           # Check for unread news"
    echo "  $BINARY_NAME list            # List all news items"
    echo "  $BINARY_NAME list --unread   # List only unread items"
    echo "  $BINARY_NAME read            # Interactively read items"
    echo "  $BINARY_NAME tui             # Launch interactive TUI"
    echo
    if check_root; then
        echo "  $BINARY_NAME install        # Install pacman hook"
        echo "  $BINARY_NAME uninstall      # Remove pacman hook"
        echo
    fi
    echo -e "${BLUE}Documentation:${NC} https://github.com/$REPO"
}

# Handle command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-hook)
            HOOK_INSTALL=false
            shift
            ;;
        --help|-h)
            echo "InformantGo Installer"
            echo
            echo "Usage: $0 [OPTIONS]"
            echo
            echo "Options:"
            echo "  --no-hook    Skip pacman hook installation"
            echo "  --help, -h   Show this help message"
            echo
            echo "Environment Variables:"
            echo "  HOOK_INSTALL=false   Skip pacman hook installation"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main 