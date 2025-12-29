#!/usr/bin/env bash
# Package installation for bootstrap scripts

# Source common utilities only if not already loaded
if ! declare -f info &>/dev/null; then
    _PACKAGES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    source "$_PACKAGES_DIR/common.sh"
    source "$_PACKAGES_DIR/detect-os.sh"
fi

# Core packages needed for dotfiles
PACKAGES_CORE=(
    git
    stow
    curl
    wget
    unzip
)

# Shell and terminal packages
PACKAGES_SHELL=(
    zsh
    tmux
)

# Development tools
PACKAGES_DEV=(
    neovim
    ripgrep
    fzf
    fd
)

# Optional but recommended
PACKAGES_OPTIONAL=(
    htop
    lazygit
    yazi
)

# Fedora-specific package names
declare -A FEDORA_PACKAGES=(
    [fd]="fd-find"
    [go]="golang"
    [python-pip]="python3-pip"
)

# Debian/Ubuntu-specific package names
declare -A DEBIAN_PACKAGES=(
    [fd]="fd-find"
    [go]="golang-go"
    [python-pip]="python3-pip"
    [ripgrep]="ripgrep"
)

# Translate package name for specific OS
translate_package() {
    local pkg="$1"
    local os="$2"

    case "$os" in
        fedora)
            echo "${FEDORA_PACKAGES[$pkg]:-$pkg}"
            ;;
        debian|ubuntu)
            echo "${DEBIAN_PACKAGES[$pkg]:-$pkg}"
            ;;
        *)
            echo "$pkg"
            ;;
    esac
}

# Install packages using the appropriate package manager
install_packages() {
    local os="$1"
    shift
    local packages=("$@")

    local pm
    pm=$(get_package_manager "$os")

    # Translate package names for this OS
    local translated=()
    for pkg in "${packages[@]}"; do
        translated+=("$(translate_package "$pkg" "$os")")
    done

    info "Installing: ${translated[*]}"

    case "$pm" in
        dnf)
            run_sudo dnf install -y "${translated[@]}"
            ;;
        apt)
            run_sudo apt update
            run_sudo apt install -y "${translated[@]}"
            ;;
        pacman)
            run_sudo pacman -S --noconfirm "${translated[@]}"
            ;;
        brew)
            brew install "${translated[@]}"
            ;;
        *)
            error "Unknown package manager: $pm"
            return 1
            ;;
    esac
}

# Install all core packages
install_core_packages() {
    local os="$1"
    header "Installing core packages"
    install_packages "$os" "${PACKAGES_CORE[@]}"
}

# Install shell packages
install_shell_packages() {
    local os="$1"
    header "Installing shell packages"
    install_packages "$os" "${PACKAGES_SHELL[@]}"
}

# Install development packages
install_dev_packages() {
    local os="$1"
    header "Installing development packages"
    install_packages "$os" "${PACKAGES_DEV[@]}" go python-pip
}

# Install optional packages
install_optional_packages() {
    local os="$1"
    header "Installing optional packages"

    # Some packages may not be in default repos
    for pkg in "${PACKAGES_OPTIONAL[@]}"; do
        if ! install_packages "$os" "$pkg"; then
            warn "Could not install $pkg - may need manual installation"
        fi
    done
}

# Install all packages
install_all_packages() {
    local os="$1"
    install_core_packages "$os"
    install_shell_packages "$os"
    install_dev_packages "$os"
    install_optional_packages "$os"
}
