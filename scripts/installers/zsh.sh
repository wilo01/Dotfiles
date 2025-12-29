#!/usr/bin/env bash
# Zsh installation: oh-my-zsh + plugins + powerlevel10k

# Source common utilities only if not already loaded
if ! declare -f info &>/dev/null; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    source "$SCRIPT_DIR/../lib/common.sh"
fi

# oh-my-zsh installation
install_oh_my_zsh() {
    header "Installing Oh-My-Zsh"

    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        warn "oh-my-zsh already installed at ~/.oh-my-zsh"
        return 0
    fi

    info "Downloading and installing oh-my-zsh..."
    RUNZSH=no CHSH=no KEEP_ZSHRC=yes \
        sh -c "$(curl --connect-timeout 30 --max-time 120 -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        success "oh-my-zsh installed successfully"
    else
        error "Failed to install oh-my-zsh"
        return 1
    fi
}

# Install zsh plugins
install_zsh_plugins() {
    header "Installing Zsh Plugins"

    local plugins_dir="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins"

    # zsh-autosuggestions
    if [[ ! -d "$plugins_dir/zsh-autosuggestions" ]]; then
        info "Installing zsh-autosuggestions..."
        git clone --depth 1 https://github.com/zsh-users/zsh-autosuggestions "$plugins_dir/zsh-autosuggestions"
        success "zsh-autosuggestions installed"
    else
        warn "zsh-autosuggestions already installed"
    fi

    # zsh-syntax-highlighting
    if [[ ! -d "$plugins_dir/zsh-syntax-highlighting" ]]; then
        info "Installing zsh-syntax-highlighting..."
        git clone --depth 1 https://github.com/zsh-users/zsh-syntax-highlighting "$plugins_dir/zsh-syntax-highlighting"
        success "zsh-syntax-highlighting installed"
    else
        warn "zsh-syntax-highlighting already installed"
    fi

    # evalcache
    if [[ ! -d "$plugins_dir/evalcache" ]]; then
        info "Installing evalcache..."
        git clone --depth 1 https://github.com/mroth/evalcache "$plugins_dir/evalcache"
        success "evalcache installed"
    else
        warn "evalcache already installed"
    fi
}

# Install powerlevel10k theme
install_powerlevel10k() {
    header "Installing Powerlevel10k Theme"

    local themes_dir="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/themes"

    if [[ ! -d "$themes_dir/powerlevel10k" ]]; then
        info "Installing powerlevel10k..."
        git clone --depth=1 https://github.com/romkatv/powerlevel10k "$themes_dir/powerlevel10k"
        success "powerlevel10k installed"
    else
        warn "powerlevel10k already installed"
    fi
}

# Set zsh as default shell
set_default_shell() {
    header "Setting Zsh as Default Shell"

    local current_shell
    current_shell=$(getent passwd "$USER" | cut -d: -f7)
    local zsh_path
    zsh_path=$(command -v zsh)

    if [[ -z "$zsh_path" ]]; then
        error "zsh not found in PATH"
        return 1
    fi

    if [[ "$current_shell" == "$zsh_path" ]]; then
        success "Zsh is already the default shell"
        return 0
    fi

    if confirm "Change default shell to zsh?"; then
        if ! chsh -s "$zsh_path"; then
            error "Failed to change shell. Ensure $zsh_path is in /etc/shells"
            return 1
        fi
        success "Default shell changed to zsh"
        info "You'll need to log out and back in for this to take effect"
    else
        warn "Skipping shell change"
    fi
}

# Main installation function
install_zsh_full() {
    install_oh_my_zsh
    install_zsh_plugins
    install_powerlevel10k
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    install_zsh_full
    set_default_shell
fi
