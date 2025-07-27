#!/bin/bash

# Neovim Update Script
# Intelligently updates Neovim using Flatpak first, GitHub releases as fallback
# Author: DarO
# Usage: ./nvim-update.sh [--force-github] [--dry-run]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVIM_DIR="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$HOME/.nvim-backups"
CURRENT_DATE=$(date +%Y%m%d_%H%M%S)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Flags
DRY_RUN=false
FORCE_GITHUB=false

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

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN=true
                info "Dry run mode enabled - no changes will be made"
                shift
                ;;
            --force-github)
                FORCE_GITHUB=true
                info "Forcing GitHub installation method"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

show_help() {
    echo "Neovim Update Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --dry-run        Show what would be done without making changes"
    echo "  --force-github   Skip Flatpak and use GitHub releases directly"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Update Methods (tried in order):"
    echo "  1. Flatpak (io.neovim.nvim from Flathub)"
    echo "  2. GitHub releases (official binaries to ~/.local/)"
    echo ""
    echo "Safety Features:"
    echo "  - Automatic backup of current installation"
    echo "  - Configuration validation"
    echo "  - Easy rollback if issues detected"
}

# Get current Neovim version
get_current_version() {
    if command -v nvim &> /dev/null; then
        nvim --version | head -1 | grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+'
    else
        echo "none"
    fi
}

# Get latest Neovim version from GitHub
get_latest_version() {
    curl -s https://api.github.com/repos/neovim/neovim/releases/latest | \
        grep '"tag_name"' | \
        cut -d'"' -f4 | \
        sed 's/v//'
}

# Create backup of current installation
create_backup() {
    step "Creating backup of current Neovim installation..."
    
    if [[ "$DRY_RUN" == true ]]; then
        info "Would create backup in: $BACKUP_DIR/nvim-$CURRENT_DATE"
        return 0
    fi
    
    mkdir -p "$BACKUP_DIR/nvim-$CURRENT_DATE"
    
    # Backup data directories
    if [[ -d "$HOME/.local/share/nvim" ]]; then
        cp -r "$HOME/.local/share/nvim" "$BACKUP_DIR/nvim-$CURRENT_DATE/share"
        success "Backed up ~/.local/share/nvim"
    fi
    
    if [[ -d "$HOME/.cache/nvim" ]]; then
        cp -r "$HOME/.cache/nvim" "$BACKUP_DIR/nvim-$CURRENT_DATE/cache"
        success "Backed up ~/.cache/nvim"
    fi
    
    # Backup current binary location
    if command -v nvim &> /dev/null; then
        NVIM_PATH=$(which nvim)
        echo "$NVIM_PATH" > "$BACKUP_DIR/nvim-$CURRENT_DATE/binary_path.txt"
        success "Recorded current binary path: $NVIM_PATH"
    fi
    
    success "Backup created in: $BACKUP_DIR/nvim-$CURRENT_DATE"
}

# Check if Flatpak is available and working
check_flatpak() {
    if ! command -v flatpak &> /dev/null; then
        return 1
    fi
    
    if ! flatpak --version &> /dev/null; then
        return 1
    fi
    
    return 0
}

# Install/update Neovim via Flatpak
install_via_flatpak() {
    step "Attempting to install/update Neovim via Flatpak..."
    
    if ! check_flatpak; then
        warning "Flatpak not available or not working"
        return 1
    fi
    
    if [[ "$DRY_RUN" == true ]]; then
        info "Would install/update io.neovim.nvim via Flatpak"
        return 0
    fi
    
    # Add Flathub if not already added
    if ! flatpak remotes | grep -q flathub; then
        info "Adding Flathub repository..."
        flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
    fi
    
    # Install or update Neovim
    if flatpak list | grep -q io.neovim.nvim; then
        info "Updating existing Flatpak Neovim installation..."
        flatpak update -y io.neovim.nvim
    else
        info "Installing Neovim via Flatpak..."
        flatpak install -y flathub io.neovim.nvim
    fi
    
    # Create wrapper script for seamless integration
    create_flatpak_wrapper
    
    # Setup configuration symlinks for sandbox access
    setup_flatpak_config_links
    
    success "Neovim installed/updated via Flatpak"
    return 0
}

# Create wrapper script for Flatpak Neovim
create_flatpak_wrapper() {
    local wrapper_path="$HOME/.local/bin/nvim"
    
    mkdir -p "$HOME/.local/bin"
    
    cat > "$wrapper_path" << 'EOF'
#!/bin/bash
# Neovim Flatpak Wrapper
# This script runs the Flatpak version of Neovim with proper integration

exec flatpak run io.neovim.nvim "$@"
EOF
    
    chmod +x "$wrapper_path"
    success "Created Flatpak wrapper at $wrapper_path"
}

# Setup Flatpak sandbox symlinks for seamless config access
setup_flatpak_config_links() {
    step "Setting up Flatpak configuration symlinks..."
    
    if [[ "$DRY_RUN" == true ]]; then
        info "Would create symlinks for Flatpak sandbox access"
        return 0
    fi
    
    local flatpak_config_dir="$HOME/.var/app/io.neovim.nvim"
    
    # Create necessary directories
    mkdir -p "$flatpak_config_dir/config"
    mkdir -p "$flatpak_config_dir/data"
    mkdir -p "$flatpak_config_dir/cache"
    
    # Create config symlink
    if [[ ! -L "$flatpak_config_dir/config/nvim" ]]; then
        if [[ -e "$HOME/.config/nvim" ]]; then
            ln -sf "$HOME/.config/nvim" "$flatpak_config_dir/config/nvim"
            success "Created config symlink"
        else
            warning "~/.config/nvim not found - you may need to create your config"
        fi
    else
        success "Config symlink already exists"
    fi
    
    # Create data symlink
    if [[ ! -L "$flatpak_config_dir/data/nvim" ]]; then
        mkdir -p "$HOME/.local/share/nvim"
        ln -sf "$HOME/.local/share/nvim" "$flatpak_config_dir/data/nvim"
        success "Created data symlink"
    else
        success "Data symlink already exists"
    fi
    
    # Create cache symlink
    if [[ ! -L "$flatpak_config_dir/cache/nvim" ]]; then
        mkdir -p "$HOME/.cache/nvim"
        ln -sf "$HOME/.cache/nvim" "$flatpak_config_dir/cache/nvim"
        success "Created cache symlink"
    else
        success "Cache symlink already exists"
    fi
}

# Install Neovim via GitHub releases
install_via_github() {
    step "Installing Neovim from GitHub releases..."
    
    local latest_version
    latest_version=$(get_latest_version)
    
    if [[ -z "$latest_version" ]]; then
        error "Failed to get latest version from GitHub"
        return 1
    fi
    
    info "Latest version: v$latest_version"
    
    if [[ "$DRY_RUN" == true ]]; then
        info "Would download and install Neovim v$latest_version to ~/.local/"
        return 0
    fi
    
    local download_url="https://github.com/neovim/neovim/releases/download/v$latest_version/nvim-linux-x86_64.tar.gz"
    local temp_dir=$(mktemp -d)
    local install_dir="$HOME/.local"
    
    # Download latest release
    info "Downloading Neovim v$latest_version..."
    curl -L "$download_url" -o "$temp_dir/nvim.tar.gz"
    
    # Extract to temporary location
    info "Extracting archive..."
    tar -xzf "$temp_dir/nvim.tar.gz" -C "$temp_dir"
    
    # Remove old installation if it exists
    if [[ -d "$install_dir/nvim-linux-x86_64" ]]; then
        rm -rf "$install_dir/nvim-linux-x86_64"
    fi
    
    # Move to installation directory
    mv "$temp_dir/nvim-linux-x86_64" "$install_dir/"
    
    # Create/update symlink
    mkdir -p "$HOME/.local/bin"
    ln -sf "$install_dir/nvim-linux-x86_64/bin/nvim" "$HOME/.local/bin/nvim"
    
    # Create desktop entry
    create_desktop_entry "$install_dir/nvim-linux-x86_64"
    
    # Cleanup
    rm -rf "$temp_dir"
    
    success "Neovim v$latest_version installed to ~/.local/"
    return 0
}

# Create desktop entry for manual installation
create_desktop_entry() {
    local nvim_path="$1"
    local desktop_dir="$HOME/.local/share/applications"
    local desktop_file="$desktop_dir/nvim.desktop"
    
    mkdir -p "$desktop_dir"
    
    cat > "$desktop_file" << EOF
[Desktop Entry]
Name=Neovim
GenericName=Text Editor
Comment=Edit text files
Exec=$nvim_path/bin/nvim %F
Icon=$nvim_path/share/icons/hicolor/128x128/apps/nvim.png
Type=Application
Terminal=true
Categories=Utility;TextEditor;Development;
MimeType=text/plain;text/x-makefile;text/x-c++hdr;text/x-c++src;text/x-chdr;text/x-csrc;text/x-java;text/x-moc;text/x-pascal;text/x-tcl;text/x-tex;application/x-shellscript;text/x-c;text/x-c++;
EOF
    
    success "Created desktop entry at $desktop_file"
}

# Test the new Neovim installation
test_installation() {
    step "Testing new Neovim installation..."
    
    if ! command -v nvim &> /dev/null; then
        error "Neovim not found in PATH after installation"
        return 1
    fi
    
    local new_version
    new_version=$(get_current_version)
    
    if [[ -z "$new_version" ]]; then
        error "Could not determine Neovim version"
        return 1
    fi
    
    info "Installed version: v$new_version"
    
    # Test basic functionality
    if [[ "$DRY_RUN" == false ]]; then
        if ! timeout 10s nvim --headless -c "lua print('Test successful')" +qa; then
            error "Neovim failed basic functionality test"
            return 1
        fi
    fi
    
    success "Neovim installation test passed"
    return 0
}

# Main update function
main() {
    echo "=== Neovim Update Script ==="
    echo ""
    
    parse_args "$@"
    
    local current_version
    current_version=$(get_current_version)
    
    local latest_version
    latest_version=$(get_latest_version)
    
    if [[ -z "$latest_version" ]]; then
        error "Failed to get latest version from GitHub"
        exit 1
    fi
    
    info "Current version: v$current_version"
    info "Latest version: v$latest_version"
    
    if [[ "$current_version" == "$latest_version" ]]; then
        success "Neovim is already up to date!"
        exit 0
    fi
    
    # Create backup
    create_backup
    
    # Try installation methods
    local install_success=false
    
    if [[ "$FORCE_GITHUB" == false ]]; then
        info "Trying Flatpak installation first..."
        if install_via_flatpak; then
            install_success=true
        else
            warning "Flatpak installation failed, trying GitHub releases..."
        fi
    fi
    
    if [[ "$install_success" == false ]]; then
        if install_via_github; then
            install_success=true
        else
            error "GitHub installation failed"
            exit 1
        fi
    fi
    
    # Test installation
    if ! test_installation; then
        error "Installation test failed"
        warning "You may need to run the rollback script"
        exit 1
    fi
    
    echo ""
    success "Neovim successfully updated to v$latest_version!"
    echo ""
    info "Backup created in: $BACKUP_DIR/nvim-$CURRENT_DATE"
    info "To rollback if needed: $SCRIPT_DIR/rollback-nvim.sh"
    echo ""
    info "Please restart your terminal and test your Neovim configuration"
}

# Run main function with all arguments
main "$@"