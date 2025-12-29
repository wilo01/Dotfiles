#!/usr/bin/env bash
# Common utilities for bootstrap scripts

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Logging functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

header() {
    echo ""
    echo -e "${BOLD}${CYAN}=== $1 ===${NC}"
    echo ""
}

# Interactive confirmation
# Returns 0 (true) if user confirms or AUTO_YES is set
confirm() {
    local message="${1:-Continue?}"

    if [[ "${AUTO_YES:-false}" == "true" ]]; then
        info "$message [auto-yes]"
        return 0
    fi

    printf "%b%s%b [Y/n] " "$YELLOW" "$message" "$NC"
    read -r response
    [[ -z "$response" || "$response" =~ ^[Yy] ]]
}

# Check if command exists
command_exists() {
    command -v "$1" &>/dev/null
}

# Run command with sudo if needed
run_sudo() {
    if [[ $EUID -ne 0 ]]; then
        sudo "$@"
    else
        "$@"
    fi
}

# Check if running as root
check_not_root() {
    if [[ $EUID -eq 0 ]]; then
        error "Do not run this script as root. It will use sudo when needed."
        exit 1
    fi
}
