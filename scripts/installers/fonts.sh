#!/usr/bin/env bash
# Nerd Fonts installation

# Source common utilities only if not already loaded
if ! declare -f info &>/dev/null; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    source "$SCRIPT_DIR/../lib/common.sh"
fi

# Default fonts to install
FONTS=(
    "FiraCode"
    "FiraMono"
    "JetBrainsMono"
)

# Font installation directory
get_font_dir() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "$HOME/Library/Fonts"
    else
        echo "$HOME/.local/share/fonts"
    fi
}

# Install a single Nerd Font
install_nerd_font() {
    local font_name="$1"
    local font_dir
    font_dir=$(get_font_dir)
    local tmp_dir
    tmp_dir=$(mktemp -d -t nerd-fonts-XXXXXX)

    info "Installing $font_name Nerd Font..."

    # Create font directory
    mkdir -p "$font_dir"

    # Download font with timeout
    local download_url="https://github.com/ryanoasis/nerd-fonts/releases/latest/download/${font_name}.zip"

    if ! curl --connect-timeout 30 --max-time 300 -fsSL "$download_url" -o "$tmp_dir/${font_name}.zip"; then
        error "Failed to download $font_name"
        rm -rf "$tmp_dir"
        return 1
    fi

    # Extract font
    if ! unzip -q "$tmp_dir/${font_name}.zip" -d "$tmp_dir/${font_name}"; then
        error "Failed to extract $font_name"
        rm -rf "$tmp_dir"
        return 1
    fi

    # Copy font files (ttf and otf)
    if ! find "$tmp_dir/${font_name}" -type f \( -name "*.ttf" -o -name "*.otf" \) \
        ! -name "*Windows*" \
        -exec cp {} "$font_dir/" \; ; then
        error "Failed to copy font files for $font_name"
        rm -rf "$tmp_dir"
        return 1
    fi

    # Cleanup
    rm -rf "$tmp_dir"

    success "$font_name installed"
}

# Refresh font cache
refresh_font_cache() {
    if command_exists fc-cache; then
        info "Refreshing font cache..."
        fc-cache -fv >/dev/null 2>&1
        success "Font cache refreshed"
    fi
}

# Install all configured fonts
install_all_fonts() {
    header "Installing Nerd Fonts"

    for font in "${FONTS[@]}"; do
        install_nerd_font "$font"
    done

    refresh_font_cache
}

# Install specific fonts
install_fonts() {
    local fonts=("$@")

    if [[ ${#fonts[@]} -eq 0 ]]; then
        fonts=("${FONTS[@]}")
    fi

    header "Installing Nerd Fonts"

    for font in "${fonts[@]}"; do
        install_nerd_font "$font"
    done

    refresh_font_cache
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    install_all_fonts
fi
