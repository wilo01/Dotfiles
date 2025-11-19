#!/bin/bash

# Neovim Update Script
# Intelligently updates Neovim with multiple installation methods
# Author: DarO
# Usage: ./nvim-update.sh [--build-source|--dnf|--flatpak] [--dry-run]

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
INSTALL_METHOD="source"  # Default to building from source
# Valid methods: source, dnf, flatpak

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
            --build-source)
                INSTALL_METHOD="source"
                info "Using build-from-source installation method"
                shift
                ;;
            --dnf)
                INSTALL_METHOD="dnf"
                info "Using Fedora DNF package manager"
                shift
                ;;
            --flatpak)
                INSTALL_METHOD="flatpak"
                info "Using Flatpak installation method"
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
    echo "  --build-source   Build from source (DEFAULT)"
    echo "  --dnf            Install via Fedora DNF package manager"
    echo "  --flatpak        Install via Flatpak (sandboxing limitations apply)"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Installation Methods:"
    echo "  1. Build from source (DEFAULT)"
    echo "     - Latest features and optimizations"
    echo "     - CMAKE_BUILD_TYPE=Release (optimized)"
    echo "     - Uses existing host tools (rg, fd, etc.)"
    echo "     - Most stable for Fedora installations"
    echo ""
    echo "  2. Fedora DNF (--dnf)"
    echo "     - System integrated via package manager"
    echo "     - May lag behind latest releases"
    echo "     - Easiest updates via 'dnf update'"
    echo ""
    echo "  3. Flatpak (--flatpak)"
    echo "     - Sandboxed application"
    echo "     - Auto-updates from Flathub"
    echo "     - Limited access to host tools (requires overrides)"
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

# Check if build dependencies are installed
check_build_dependencies() {
    step "Checking build dependencies..."

    local missing_deps=()
    local all_deps=("cmake" "gcc" "g++" "make" "git" "ninja-build" "gettext" "libtool" "libtool-ltdl-devel" "autoconf" "automake" "pkg-config")

    for dep in "${all_deps[@]}"; do
        if ! command -v "$dep" &> /dev/null && ! rpm -q "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done

    if [ ${#missing_deps[@]} -eq 0 ]; then
        success "All build dependencies are installed"
        return 0
    fi

    warning "Missing build dependencies: ${missing_deps[*]}"
    echo ""
    info "Install with: sudo dnf install ${missing_deps[*]}"
    echo ""

    if [[ "$DRY_RUN" == true ]]; then
        return 0
    fi

    read -p "Install missing dependencies now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo dnf install -y "${missing_deps[@]}"
        success "Dependencies installed"
        return 0
    else
        error "Cannot build without dependencies"
        return 1
    fi
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

# Build Neovim from source
build_from_source() {
    step "Building Neovim from source..."

    # Check dependencies first
    if ! check_build_dependencies; then
        return 1
    fi

    local latest_version
    latest_version=$(get_latest_version)

    if [[ -z "$latest_version" ]]; then
        error "Failed to get latest version from GitHub"
        return 1
    fi

    info "Building Neovim v$latest_version from source"

    if [[ "$DRY_RUN" == true ]]; then
        info "Would clone repository and build Neovim v$latest_version"
        info "Build type: Release (optimized)"
        info "Install prefix: $HOME/.local"
        return 0
    fi

    local build_dir="$HOME/.cache/nvim-build"
    local repo_dir="$build_dir/neovim"

    # Create build directory
    mkdir -p "$build_dir"

    # Clone or update repository
    if [[ -d "$repo_dir" ]]; then
        info "Updating existing Neovim repository..."
        cd "$repo_dir"
        git fetch --all
    else
        info "Cloning Neovim repository..."
        git clone https://github.com/neovim/neovim.git "$repo_dir"
        cd "$repo_dir"
    fi

    # Checkout latest stable tag
    info "Checking out v$latest_version..."
    git checkout "v$latest_version"

    # Clean previous builds
    info "Cleaning previous builds..."
    make distclean 2>/dev/null || true
    rm -rf build/ .deps/ 2>/dev/null || true

    # Configure build
    info "Configuring build (CMAKE_BUILD_TYPE=Release)..."
    cmake -B build -G Ninja \
        -D CMAKE_BUILD_TYPE=Release \
        -D CMAKE_INSTALL_PREFIX="$HOME/.local"

    # Build with parallel jobs
    local nproc_count=$(nproc)
    info "Building with $nproc_count parallel jobs..."
    cmake --build build --parallel "$nproc_count"

    # Install
    info "Installing to ~/.local/..."
    cmake --install build

    # Create desktop entry
    create_desktop_entry "$HOME/.local"

    success "Neovim v$latest_version built and installed from source"
    info "Build directory preserved at: $build_dir"
    return 0
}

# Install Neovim via DNF package manager
install_via_dnf() {
    step "Installing Neovim via DNF package manager..."

    if [[ "$DRY_RUN" == true ]]; then
        info "Would run: sudo dnf install neovim"
        return 0
    fi

    warning "DNF package may not be the latest version"
    info "Checking available version..."

    local dnf_version=$(dnf info neovim 2>/dev/null | grep "^Version" | awk '{print $3}')
    if [[ -n "$dnf_version" ]]; then
        info "DNF repository has: v$dnf_version"
    fi

    sudo dnf install -y neovim

    success "Neovim installed via DNF"
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

    # Install using selected method
    local install_success=false

    case "$INSTALL_METHOD" in
        source)
            info "Building from source..."
            if build_from_source; then
                install_success=true
            else
                error "Build from source failed"
                exit 1
            fi
            ;;
        dnf)
            info "Installing via DNF..."
            if install_via_dnf; then
                install_success=true
            else
                error "DNF installation failed"
                exit 1
            fi
            ;;
        flatpak)
            warning "Using Flatpak with known sandboxing limitations"
            info "You may need to install tools (rg, fd, etc.) separately"
            if install_via_flatpak; then
                install_success=true
            else
                error "Flatpak installation failed"
                exit 1
            fi
            ;;
        *)
            error "Unknown installation method: $INSTALL_METHOD"
            exit 1
            ;;
    esac
    
    # Test installation
    if ! test_installation; then
        error "Installation test failed"
        warning "You may need to run the rollback script"
        exit 1
    fi
    
    echo ""
    success "Neovim successfully updated to v$latest_version using $INSTALL_METHOD method!"
    echo ""
    info "Backup created in: $BACKUP_DIR/nvim-$CURRENT_DATE"
    info "To rollback if needed: $SCRIPT_DIR/rollback-nvim.sh"
    echo ""

    case "$INSTALL_METHOD" in
        source)
            info "Build artifacts preserved at: ~/.cache/nvim-build/"
            ;;
        flatpak)
            warning "Remember: Flatpak has sandbox limitations"
            info "Consider using: flatpak override --user io.neovim.nvim --filesystem=/usr/bin:ro"
            ;;
    esac

    echo ""
    info "Please restart your terminal and test your Neovim configuration"
}

# Run main function with all arguments
main "$@"
