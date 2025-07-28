#!/bin/bash

# Simple Neovim Installation Script
# Cross-platform installer for Ubuntu and Fedora
# Author: DarO
# Usage: ./install-neovim.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Global variables
OS_TYPE=""
PACKAGE_MANAGER=""
TEMP_DIR=""

info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

step() {
    echo -e "${PURPLE}▶${NC} $1"
}

# Detect operating system and package manager
detect_os() {
    step "Detecting operating system..."
    
    if [[ -f /etc/fedora-release ]]; then
        OS_TYPE="fedora"
        PACKAGE_MANAGER="dnf"
        success "Detected Fedora"
    elif [[ -f /etc/ubuntu-release ]] || [[ -f /etc/debian_version ]] || command -v apt-get &> /dev/null; then
        OS_TYPE="ubuntu"
        PACKAGE_MANAGER="apt"
        success "Detected Ubuntu/Debian"
    else
        error "Unsupported operating system. This script supports Ubuntu/Debian and Fedora only."
        exit 1
    fi
}

# Get current Neovim version if installed
get_current_version() {
    if command -v nvim &> /dev/null; then
        nvim --version | head -1 | grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+'
    else
        echo "none"
    fi
}

# Get latest Neovim version from GitHub API
get_latest_version() {
    local latest_version
    latest_version=$(curl -s https://api.github.com/repos/neovim/neovim/releases/latest | \
        grep '"tag_name"' | \
        cut -d'"' -f4 | \
        sed 's/v//')
    
    if [[ -z "$latest_version" ]]; then
        error "Failed to fetch latest version from GitHub"
        exit 1
    fi
    
    echo "$latest_version"
}

# Check for existing Neovim installation
check_existing_installation() {
    step "Checking for existing Neovim installation..."
    
    local current_version
    current_version=$(get_current_version)
    
    if [[ "$current_version" != "none" ]]; then
        info "Found existing Neovim installation: v$current_version"
        
        # Check installation method
        if command -v nvim &> /dev/null; then
            local nvim_path=$(which nvim)
            info "Neovim binary location: $nvim_path"
        fi
        
        return 0
    else
        info "No existing Neovim installation found"
        return 1
    fi
}

# Remove existing Neovim installation
remove_existing_neovim() {
    step "Removing existing Neovim installation..."
    
    # Remove system package
    if [[ "$OS_TYPE" == "fedora" ]]; then
        if dnf list installed neovim &> /dev/null; then
            info "Removing Neovim system package (Fedora)..."
            sudo dnf remove -y neovim || true
        fi
    elif [[ "$OS_TYPE" == "ubuntu" ]]; then
        if dpkg -l | grep -q neovim; then
            info "Removing Neovim system package (Ubuntu)..."
            sudo apt remove -y neovim || true
        fi
    fi
    
    # Remove compiled version from /usr/local/bin
    if [[ -f /usr/local/bin/nvim ]]; then
        info "Removing compiled Neovim from /usr/local/bin..."
        sudo rm -f /usr/local/bin/nvim
    fi
    
    # Remove manual installation from ~/.local/
    if [[ -d "$HOME/.local/nvim-linux-x86_64" ]]; then
        info "Removing manual installation from ~/.local/"
        rm -rf "$HOME/.local/nvim-linux-x86_64"
    fi
    
    if [[ -L "$HOME/.local/bin/nvim" ]]; then
        info "Removing symlink from ~/.local/bin/nvim"
        rm -f "$HOME/.local/bin/nvim"
    fi
    
    # Clean cache but preserve configuration
    if [[ -d "$HOME/.cache/nvim" ]]; then
        info "Cleaning Neovim cache..."
        rm -rf "$HOME/.cache/nvim"
    fi
    
    success "Existing Neovim installation removed"
    success "Configuration preserved in ~/.config/nvim and ~/.local/share/nvim"
}

# Install build dependencies
install_dependencies() {
    step "Installing build dependencies..."
    
    if [[ "$OS_TYPE" == "fedora" ]]; then
        info "Installing dependencies for Fedora..."
        sudo dnf install -y git cmake ninja-build libtool autoconf automake pkgconfig unzip gettext-devel
    elif [[ "$OS_TYPE" == "ubuntu" ]]; then
        info "Updating package lists and installing dependencies for Ubuntu..."
        sudo apt-get update
        sudo apt-get install -y git cmake ninja-build libtool libtool-bin autoconf automake pkg-config unzip gettext
    fi
    
    success "Build dependencies installed"
}

# Build and install Neovim from source
build_and_install_neovim() {
    local latest_version="$1"
    
    step "Building and installing Neovim v$latest_version from source..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    info "Using temporary directory: $TEMP_DIR"
    
    # Clone Neovim repository
    info "Cloning Neovim repository..."
    if ! git clone --branch "v$latest_version" https://github.com/neovim/neovim "$TEMP_DIR/neovim"; then
        error "Failed to clone Neovim repository for version v$latest_version"
        exit 1
    fi
    
    # Build Neovim
    info "Building Neovim (this may take a few minutes)..."
    cd "$TEMP_DIR/neovim"
    if ! make -j4 CMAKE_BUILD_TYPE=RelWithDebInfo; then
        error "Failed to build Neovim"
        exit 1
    fi
    
    # Install Neovim
    info "Installing Neovim..."
    if ! sudo make -j4 install; then
        error "Failed to install Neovim"
        exit 1
    fi
    
    success "Neovim v$latest_version built and installed successfully"
}

# Cleanup temporary files
cleanup() {
    if [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        step "Cleaning up temporary files..."
        rm -rf "$TEMP_DIR"
        success "Cleanup completed"
    fi
}

# Verify installation
verify_installation() {
    step "Verifying installation..."
    
    if ! command -v nvim &> /dev/null; then
        error "Neovim not found in PATH after installation"
        return 1
    fi
    
    local installed_version
    installed_version=$(get_current_version)
    
    if [[ -z "$installed_version" ]]; then
        error "Could not determine installed Neovim version"
        return 1
    fi
    
    success "Neovim v$installed_version installed successfully"
    
    # Test basic functionality
    info "Testing basic functionality..."
    if timeout 10s nvim --headless -c "lua print('Installation test successful')" +qa; then
        success "Basic functionality test passed"
    else
        warning "Basic functionality test failed, but installation appears complete"
    fi
    
    return 0
}

# Main function
main() {
    echo "=== Simple Neovim Installation Script ==="
    echo ""
    
    # Detect OS
    detect_os
    
    # Check for existing installation
    local has_existing=false
    if check_existing_installation; then
        has_existing=true
        echo ""
        warning "An existing Neovim installation was found."
        echo "This script will:"
        echo "  ✓ Remove existing binaries and cache"
        echo "  ✓ Preserve your configuration (~/.config/nvim)"
        echo "  ✓ Preserve your data (~/.local/share/nvim)"
        echo ""
        
        read -p "Do you want to remove the existing installation and continue? [y/N]: " -n 1 -r
        echo
        
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Installation cancelled"
            exit 0
        fi
        
        remove_existing_neovim
    fi
    
    # Get latest version
    step "Fetching latest Neovim version from GitHub..."
    local latest_version
    latest_version=$(get_latest_version)
    success "Latest version: v$latest_version"
    
    # Install dependencies
    install_dependencies
    
    # Build and install
    build_and_install_neovim "$latest_version"
    
    # Cleanup
    cleanup
    
    # Verify installation
    if verify_installation; then
        echo ""
        success "Neovim installation completed successfully!"
        echo ""
        info "Neovim v$latest_version is now installed"
        info "You may need to restart your terminal or run 'hash -r' to update PATH"
        
        if [[ "$has_existing" == true ]]; then
            echo ""
            info "Your configuration and data have been preserved:"
            info "  • Configuration: ~/.config/nvim"
            info "  • Data: ~/.local/share/nvim"
        fi
    else
        error "Installation verification failed"
        exit 1
    fi
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
