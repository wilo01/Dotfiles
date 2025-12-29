#!/usr/bin/env bash
#
# Dotfiles Bootstrap Script
# Full system setup for a fresh machine
#
# Usage:
#   ./scripts/bootstrap.sh           # Interactive mode
#   ./scripts/bootstrap.sh --yes     # Unattended mode
#   ./scripts/bootstrap.sh --skip-packages
#   ./scripts/bootstrap.sh --only=zsh,nvim,tmux
#

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Source libraries
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/detect-os.sh"
source "$SCRIPT_DIR/lib/packages.sh"
source "$SCRIPT_DIR/installers/zsh.sh"
source "$SCRIPT_DIR/installers/fonts.sh"

# Default options
AUTO_YES="${AUTO_YES:-false}"
DRY_RUN=false
SKIP_PACKAGES=false
SKIP_STOW=false
SKIP_ZSH=false
SKIP_FONTS=false
ONLY_COMPONENTS=()

# Dry-run wrapper - prints command instead of executing
dry_run() {
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "  ${CYAN}[DRY-RUN]${NC} $*"
        return 0
    else
        "$@"
    fi
}

# All stow packages
STOW_PACKAGES=(
    # Shell essentials
    zsh bash profile bin
    # Git
    gitconfig
    # Terminal
    tmux alacritty kitty wezterm ghostty
    # Editor
    nvim vscode
    # Tools
    lazygit htop taskwarrior yazi-conf rofi
    # Desktop
    Sway regolith gtk
    # Misc
    yarn spotify sqldev Browser recatest IMU
    # Meta (claude config)
    .Dotfiles
)

# Packages that need special handling
SPECIAL_PACKAGES=(zsh nvim scripts .Dotfiles Browser)

# Usage
usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
    --yes, -y           Run in unattended mode (no prompts)
    --dry-run, -n       Show what would be done without executing
    --skip-packages     Skip system package installation
    --skip-stow         Skip stow operations
    --skip-zsh          Skip zsh/oh-my-zsh setup
    --skip-fonts        Skip Nerd Fonts installation
    --only=COMPONENTS   Only install specific components (comma-separated)
                        Valid: packages,zsh,stow,fonts,helper-go,all
    -h, --help          Show this help message

Examples:
    $(basename "$0")                    # Interactive full install
    $(basename "$0") --dry-run          # Show what would happen
    $(basename "$0") --yes              # Unattended full install
    $(basename "$0") --only=zsh,fonts   # Only zsh and fonts
    $(basename "$0") --skip-packages    # Skip apt/dnf installs
EOF
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --yes|-y)
                AUTO_YES=true
                export AUTO_YES
                shift
                ;;
            --dry-run|-n)
                DRY_RUN=true
                AUTO_YES=true  # Don't prompt in dry-run
                export AUTO_YES DRY_RUN
                shift
                ;;
            --skip-packages)
                SKIP_PACKAGES=true
                shift
                ;;
            --skip-stow)
                SKIP_STOW=true
                shift
                ;;
            --skip-zsh)
                SKIP_ZSH=true
                shift
                ;;
            --skip-fonts)
                SKIP_FONTS=true
                shift
                ;;
            --only=*)
                IFS=',' read -ra ONLY_COMPONENTS <<< "${1#*=}"
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Check if component should run
should_run() {
    local component="$1"

    # If --only specified, check if this component is in the list
    if [[ ${#ONLY_COMPONENTS[@]} -gt 0 ]]; then
        for c in "${ONLY_COMPONENTS[@]}"; do
            [[ "$c" == "$component" || "$c" == "all" ]] && return 0
        done
        return 1
    fi

    # Otherwise, check skip flags
    case "$component" in
        packages) [[ "$SKIP_PACKAGES" != "true" ]] ;;
        zsh)      [[ "$SKIP_ZSH" != "true" ]] ;;
        stow)     [[ "$SKIP_STOW" != "true" ]] ;;
        fonts)    [[ "$SKIP_FONTS" != "true" ]] ;;
        *)        return 0 ;;
    esac
}

# Stow a single package with conflict handling
stow_package() {
    local pkg="$1"
    local target="${2:-$HOME}"

    # Skip if package doesn't exist
    if [[ ! -d "$DOTFILES_DIR/$pkg" ]]; then
        warn "Package $pkg not found, skipping"
        return 0
    fi

    info "Stowing $pkg..."

    # Try stow, handle conflicts
    if ! stow -d "$DOTFILES_DIR" -t "$target" --no-folding "$pkg" 2>/dev/null; then
        warn "Conflicts detected for $pkg, attempting --adopt"
        if stow -d "$DOTFILES_DIR" -t "$target" --adopt --no-folding "$pkg"; then
            success "Stowed $pkg (with adopt)"
        else
            error "Failed to stow $pkg"
            return 1
        fi
    else
        success "Stowed $pkg"
    fi
}

# Stow all packages
stow_all_packages() {
    header "Stowing Configuration Packages"

    cd "$DOTFILES_DIR"

    local failed=()

    for pkg in "${STOW_PACKAGES[@]}"; do
        # Skip special packages that need pre-handling
        if [[ " ${SPECIAL_PACKAGES[*]} " =~ " $pkg " ]]; then
            continue
        fi

        if ! stow_package "$pkg"; then
            failed+=("$pkg")
        fi
    done

    # Handle special packages
    # zsh: already handled by install_zsh_full
    if [[ -d "$DOTFILES_DIR/zsh" ]]; then
        stow_package "zsh"
    fi

    # nvim
    if [[ -d "$DOTFILES_DIR/nvim" ]]; then
        stow_package "nvim"
    fi

    # .Dotfiles (claude config)
    if [[ -d "$DOTFILES_DIR/.Dotfiles" ]]; then
        stow_package ".Dotfiles"
    fi

    if [[ ${#failed[@]} -gt 0 ]]; then
        warn "Failed to stow: ${failed[*]}"
    fi
}

# Build helper-go
build_helper_go() {
    header "Building helper-go CLI"

    local helper_dir="$DOTFILES_DIR/scripts/helper-go"

    if [[ ! -d "$helper_dir" ]]; then
        warn "helper-go directory not found, skipping"
        return 0
    fi

    if ! command_exists go; then
        warn "Go not installed, skipping helper-go build"
        return 0
    fi

    # Use subshell to avoid polluting PWD
    (
        cd "$helper_dir" || exit 1

        if [[ -f "Makefile" ]]; then
            info "Running make build..."
            make build
            success "helper-go built successfully"

            if confirm "Install helper-go to GOPATH/bin?"; then
                make install
                success "helper-go installed"
            fi
        else
            warn "No Makefile found in helper-go"
        fi
    )
}

# Install Python helper
install_python_helper() {
    header "Installing Python Helper CLI"

    local helper_dir="$DOTFILES_DIR/scripts/helper"

    if [[ ! -d "$helper_dir" ]]; then
        warn "helper directory not found, skipping"
        return 0
    fi

    if ! command_exists pip3; then
        warn "pip3 not installed, skipping Python helper"
        return 0
    fi

    # Use subshell to avoid polluting PWD
    (
        cd "$helper_dir" || exit 1

        if [[ -f "pyproject.toml" ]] || [[ -f "setup.py" ]]; then
            info "Installing Python helper in development mode..."
            pip3 install -e . --user
            success "Python helper installed"
        else
            warn "No pyproject.toml or setup.py found"
        fi
    )
}

# Print summary
print_summary() {
    header "Bootstrap Complete!"

    echo -e "${GREEN}What was done:${NC}"
    should_run packages && echo "  - System packages installed"
    should_run zsh && echo "  - oh-my-zsh + plugins + powerlevel10k installed"
    should_run stow && echo "  - Configuration packages stowed"
    should_run fonts && echo "  - Nerd Fonts installed"
    should_run helper-go && echo "  - helper-go CLI built"

    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "  1. Restart your terminal (or run: exec zsh)"
    echo "  2. Run 'p10k configure' if you want to reconfigure the prompt"
    echo "  3. Open Neovim and let plugins install (:Lazy)"
    echo ""
}

# Dry-run: show packages info
dry_run_packages() {
    local os="$1"
    header "System Packages (dry-run)"

    info "Would install core packages:"
    echo "  ${PACKAGES_CORE[*]}"

    info "Would install shell packages:"
    echo "  ${PACKAGES_SHELL[*]}"

    info "Would install dev packages:"
    echo "  ${PACKAGES_DEV[*]} go python-pip"

    info "Would install optional packages:"
    echo "  ${PACKAGES_OPTIONAL[*]}"

    local pm
    pm=$(get_package_manager "$os")
    info "Package manager: $pm"
}

# Plugin repository URLs for dry-run display
declare -A PLUGIN_REPOS=(
    [zsh-autosuggestions]="https://github.com/zsh-users/zsh-autosuggestions"
    [zsh-syntax-highlighting]="https://github.com/zsh-users/zsh-syntax-highlighting"
    [evalcache]="https://github.com/mroth/evalcache"
)

# Dry-run: show zsh info
dry_run_zsh() {
    header "Zsh Setup (dry-run)"

    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        info "oh-my-zsh: Already installed at ~/.oh-my-zsh"
    else
        dry_run "sh -c \"\$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)\""
    fi

    local plugins_dir="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins"
    local themes_dir="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes"

    for plugin in zsh-autosuggestions zsh-syntax-highlighting evalcache; do
        if [[ -d "$plugins_dir/$plugin" ]]; then
            info "$plugin: Already installed"
        else
            dry_run "git clone ${PLUGIN_REPOS[$plugin]} $plugins_dir/$plugin"
        fi
    done

    if [[ -d "$themes_dir/powerlevel10k" ]]; then
        info "powerlevel10k: Already installed"
    else
        dry_run "git clone --depth=1 https://github.com/romkatv/powerlevel10k $themes_dir/powerlevel10k"
    fi
}

# Dry-run: show stow info
dry_run_stow() {
    header "Stow Packages (dry-run)"

    info "Would stow ${#STOW_PACKAGES[@]} packages:"

    local existing=()
    local missing=()

    for pkg in "${STOW_PACKAGES[@]}"; do
        if [[ -d "$DOTFILES_DIR/$pkg" ]]; then
            existing+=("$pkg")
        else
            missing+=("$pkg")
        fi
    done

    echo -e "  ${GREEN}Found:${NC} ${existing[*]}"
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo -e "  ${YELLOW}Missing:${NC} ${missing[*]}"
    fi

    # Check for potential conflicts
    info "Checking for potential conflicts..."
    local conflicts=()
    for pkg in "${existing[@]}"; do
        local conflict_output
        conflict_output=$(stow -d "$DOTFILES_DIR" -t "$HOME" --no-folding -n "$pkg" 2>&1 || true)
        if echo "$conflict_output" | grep -qiE "conflict|existing target"; then
            conflicts+=("$pkg")
        fi
    done

    if [[ ${#conflicts[@]} -gt 0 ]]; then
        warn "Potential conflicts: ${conflicts[*]}"
        info "These would use --adopt to resolve"
    else
        success "No conflicts detected"
    fi
}

# Dry-run: show fonts info
dry_run_fonts() {
    header "Nerd Fonts (dry-run)"

    info "Would install fonts:"
    for font in "${FONTS[@]}"; do
        dry_run "curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/latest/download/${font}.zip"
    done

    local font_dir
    font_dir=$(get_font_dir)
    info "Font directory: $font_dir"
}

# Dry-run: show helper-go info
dry_run_helper_go() {
    header "Helper Tools (dry-run)"

    local helper_go_dir="$DOTFILES_DIR/scripts/helper-go"
    local helper_py_dir="$DOTFILES_DIR/scripts/helper"

    if [[ -d "$helper_go_dir" ]]; then
        if command_exists go; then
            dry_run "cd $helper_go_dir && make build"
            dry_run "cd $helper_go_dir && make install"
        else
            warn "Go not installed - would skip helper-go build"
        fi
    fi

    if [[ -d "$helper_py_dir" ]]; then
        if command_exists pip3; then
            dry_run "pip3 install -e $helper_py_dir --user"
        else
            warn "pip3 not installed - would skip Python helper"
        fi
    fi
}

# Main function
main() {
    parse_args "$@"

    check_not_root

    header "Dotfiles Bootstrap"
    info "Dotfiles directory: $DOTFILES_DIR"

    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${BOLD}${YELLOW}>>> DRY-RUN MODE - No changes will be made <<<${NC}"
        echo ""
    fi

    # Detect OS
    local os
    os=$(detect_os)
    info "Detected OS: $(get_os_name) ($os)"

    if [[ "$os" == "unknown" ]]; then
        error "Unsupported operating system"
        exit 1
    fi

    # In dry-run mode, show what would happen
    if [[ "$DRY_RUN" == "true" ]]; then
        should_run packages && dry_run_packages "$os"
        should_run zsh && dry_run_zsh
        should_run stow && dry_run_stow
        should_run fonts && dry_run_fonts
        should_run helper-go && dry_run_helper_go

        header "Dry-Run Complete"
        info "Run without --dry-run to execute these operations"
        exit 0
    fi

    # Confirm before proceeding
    if ! confirm "Proceed with installation?"; then
        info "Aborted"
        exit 0
    fi

    # Install system packages
    if should_run packages; then
        install_all_packages "$os"
    fi

    # Install zsh + oh-my-zsh + plugins
    if should_run zsh; then
        install_zsh_full
    fi

    # Stow packages
    if should_run stow; then
        stow_all_packages
    fi

    # Install fonts
    if should_run fonts; then
        install_all_fonts
    fi

    # Build helper tools
    if should_run helper-go; then
        build_helper_go
        install_python_helper
    fi

    # Set default shell
    if should_run zsh; then
        set_default_shell
    fi

    print_summary
}

# Run main
main "$@"
