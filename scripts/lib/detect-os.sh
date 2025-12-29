#!/usr/bin/env bash
# OS detection for bootstrap scripts

# Detect the operating system
# Returns: fedora, debian, ubuntu, or unknown
detect_os() {
    if [[ -f /etc/fedora-release ]]; then
        echo "fedora"
    elif [[ -f /etc/debian_version ]]; then
        # Check if it's Ubuntu specifically
        if grep -qi ubuntu /etc/os-release 2>/dev/null; then
            echo "ubuntu"
        else
            echo "debian"
        fi
    elif [[ -f /etc/arch-release ]]; then
        echo "arch"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    else
        echo "unknown"
    fi
}

# Get the package manager command
get_package_manager() {
    local os="$1"
    case "$os" in
        fedora)
            echo "dnf"
            ;;
        debian|ubuntu)
            echo "apt"
            ;;
        arch)
            echo "pacman"
            ;;
        macos)
            echo "brew"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Get OS pretty name
get_os_name() {
    if [[ -f /etc/os-release ]]; then
        source /etc/os-release
        echo "$PRETTY_NAME"
    else
        echo "Unknown OS"
    fi
}
