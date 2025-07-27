#!/bin/bash

# Neovim Rollback Script
# Restores previous Neovim installation from backup
# Author: DarO
# Usage: ./rollback-nvim.sh [backup-date]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$HOME/.nvim-backups"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

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

show_help() {
    echo "Neovim Rollback Script"
    echo ""
    echo "Usage: $0 [BACKUP_DATE]"
    echo ""
    echo "Arguments:"
    echo "  BACKUP_DATE   Specific backup to restore (format: YYYYMMDD_HHMMSS)"
    echo "                If not provided, will show available backups"
    echo ""
    echo "Examples:"
    echo "  $0                    # List available backups"
    echo "  $0 20241127_143022    # Restore specific backup"
    echo ""
    echo "What this script does:"
    echo "  1. Lists available backups"
    echo "  2. Removes current Neovim installation"
    echo "  3. Restores Neovim data from backup"
    echo "  4. Reinstalls system Neovim if needed"
}

# List available backups
list_backups() {
    echo "Available Neovim backups:"
    echo ""
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        warning "No backup directory found at $BACKUP_DIR"
        return 1
    fi
    
    local backups=($(ls -1 "$BACKUP_DIR" | grep "nvim-" | sort -r))
    
    if [[ ${#backups[@]} -eq 0 ]]; then
        warning "No Neovim backups found"
        return 1
    fi
    
    for backup in "${backups[@]}"; do
        local backup_date=${backup#nvim-}
        local formatted_date=$(echo $backup_date | sed 's/_/ /')
        local backup_path="$BACKUP_DIR/$backup"
        
        echo "  📁 $backup_date ($formatted_date)"
        
        if [[ -f "$backup_path/binary_path.txt" ]]; then
            local binary_path=$(cat "$backup_path/binary_path.txt")
            echo "     Binary was at: $binary_path"
        fi
        
        if [[ -d "$backup_path/share" ]]; then
            echo "     ✓ Contains data backup"
        fi
        
        if [[ -d "$backup_path/cache" ]]; then
            echo "     ✓ Contains cache backup"
        fi
        
        echo ""
    done
    
    return 0
}

# Validate backup exists and has required content
validate_backup() {
    local backup_date="$1"
    local backup_path="$BACKUP_DIR/nvim-$backup_date"
    
    if [[ ! -d "$backup_path" ]]; then
        error "Backup not found: $backup_path"
        return 1
    fi
    
    info "Validating backup: nvim-$backup_date"
    
    if [[ ! -d "$backup_path/share" ]] && [[ ! -d "$backup_path/cache" ]]; then
        error "Backup appears to be incomplete (no data directories)"
        return 1
    fi
    
    success "Backup validation passed"
    return 0
}

# Remove current Neovim installation
remove_current_nvim() {
    step "Removing current Neovim installation..."
    
    # Remove Flatpak version if installed
    if command -v flatpak &> /dev/null && flatpak list | grep -q io.neovim.nvim; then
        info "Removing Flatpak Neovim..."
        flatpak uninstall -y io.neovim.nvim || true
    fi
    
    # Remove manual installation from ~/.local/
    if [[ -d "$HOME/.local/nvim-linux-x86_64" ]]; then
        info "Removing manual installation from ~/.local/"
        rm -rf "$HOME/.local/nvim-linux-x86_64"
    fi
    
    # Remove symlink
    if [[ -L "$HOME/.local/bin/nvim" ]]; then
        info "Removing symlink from ~/.local/bin/nvim"
        rm -f "$HOME/.local/bin/nvim"
    fi
    
    # Remove desktop entry
    if [[ -f "$HOME/.local/share/applications/nvim.desktop" ]]; then
        info "Removing desktop entry"
        rm -f "$HOME/.local/share/applications/nvim.desktop"
    fi
    
    success "Current installation removed"
}

# Restore data from backup
restore_data() {
    local backup_date="$1"
    local backup_path="$BACKUP_DIR/nvim-$backup_date"
    
    step "Restoring Neovim data from backup..."
    
    # Create timestamp for current data (if any exists)
    local current_timestamp=$(date +%Y%m%d_%H%M%S)
    
    # Backup current data before restore (just in case)
    if [[ -d "$HOME/.local/share/nvim" ]]; then
        warning "Moving current data to ~/.local/share/nvim.pre-rollback.$current_timestamp"
        mv "$HOME/.local/share/nvim" "$HOME/.local/share/nvim.pre-rollback.$current_timestamp"
    fi
    
    if [[ -d "$HOME/.cache/nvim" ]]; then
        warning "Moving current cache to ~/.cache/nvim.pre-rollback.$current_timestamp"
        mv "$HOME/.cache/nvim" "$HOME/.cache/nvim.pre-rollback.$current_timestamp"
    fi
    
    # Restore from backup
    if [[ -d "$backup_path/share" ]]; then
        info "Restoring ~/.local/share/nvim"
        cp -r "$backup_path/share" "$HOME/.local/share/nvim"
        success "Data restored"
    fi
    
    if [[ -d "$backup_path/cache" ]]; then
        info "Restoring ~/.cache/nvim"
        cp -r "$backup_path/cache" "$HOME/.cache/nvim"
        success "Cache restored"
    fi
}

# Reinstall system Neovim
reinstall_system_nvim() {
    step "Reinstalling system Neovim..."
    
    if command -v dnf &> /dev/null; then
        info "Reinstalling Neovim via dnf..."
        sudo dnf reinstall -y neovim
        success "System Neovim reinstalled"
    elif command -v apt &> /dev/null; then
        info "Reinstalling Neovim via apt..."
        sudo apt install --reinstall neovim
        success "System Neovim reinstalled"
    else
        warning "Could not determine package manager to reinstall system Neovim"
        warning "You may need to manually reinstall Neovim via your package manager"
    fi
}

# Confirm rollback action
confirm_rollback() {
    local backup_date="$1"
    
    echo ""
    warning "This will:"
    echo "  1. Remove current Neovim installation"
    echo "  2. Restore data from backup: nvim-$backup_date"
    echo "  3. Reinstall system Neovim"
    echo ""
    echo "Current data will be backed up before rollback."
    echo ""
    
    read -p "Do you want to proceed with the rollback? [y/N]: " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "Rollback cancelled"
        exit 0
    fi
}

# Main rollback function
main() {
    echo "=== Neovim Rollback Script ==="
    echo ""
    
    # Parse command line arguments
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        "")
            # No arguments - list backups
            list_backups
            echo "To rollback to a specific backup, run:"
            echo "  $0 <backup_date>"
            exit 0
            ;;
        *)
            # Backup date provided
            local backup_date="$1"
            ;;
    esac
    
    # Validate backup exists
    if ! validate_backup "$backup_date"; then
        error "Invalid backup: $backup_date"
        echo ""
        list_backups
        exit 1
    fi
    
    # Confirm rollback
    confirm_rollback "$backup_date"
    
    # Perform rollback
    remove_current_nvim
    restore_data "$backup_date"
    reinstall_system_nvim
    
    echo ""
    success "Rollback completed successfully!"
    echo ""
    info "Neovim has been restored from backup: nvim-$backup_date"
    info "Please restart your terminal and test your Neovim"
    echo ""
    
    # Show current version
    if command -v nvim &> /dev/null; then
        local current_version=$(nvim --version | head -1)
        info "Current Neovim: $current_version"
    fi
}

# Run main function with all arguments
main "$@"