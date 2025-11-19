# Neovim Manager - Future Improvements Plan

> **Status**: Planned for future implementation
> **Date**: 2025-11-19
> **Current Version**: nvim-update.sh with source/dnf/flatpak support

## Table of Contents
- [Overview](#overview)
- [Current State Analysis](#current-state-analysis)
- [Improvement 1: GitHub API Robustness](#improvement-1-github-api-robustness)
- [Improvement 2: Comprehensive Dependency Management](#improvement-2-comprehensive-dependency-management)
- [Improvement 3: Unified Script Architecture](#improvement-3-unified-script-architecture)
- [Improvement 4: Healthcheck Integration](#improvement-4-healthcheck-integration)
- [Implementation Phases](#implementation-phases)
- [Code Examples](#code-examples)

---

## Overview

This document outlines planned improvements to the Neovim installation and update scripts. The goal is to create a unified management system that handles:

- ✅ Fresh system installations (all dependencies)
- ✅ Updates on existing systems
- ✅ Dependency checking and installation
- ✅ Healthcheck-driven troubleshooting
- ✅ Cross-platform support (Fedora/Ubuntu)

**Key Philosophy**: One script to rule them all, but maintain simplicity and reliability.

---

## Current State Analysis

### Existing Scripts

1. **`nvim-update.sh`** (Main script - working well)
   - Supports: source build, DNF, Flatpak
   - Has backup/restore functionality
   - Default: build from source
   - Status: ✅ Production ready

2. **`install-neovim.sh`** (Basic installer)
   - Simple source build script
   - Minimal dependency checking
   - Status: Functional but basic

3. **`rollback-nvim.sh`** (Backup restoration)
   - Status: ✅ Keep as-is

4. **`test-nvim.sh`** (Installation testing)
   - Status: ✅ Keep as-is

### Current Healthcheck Status (v0.11.5)

**Working Perfectly:**
- ✅ Telescope (rg, fd both found)
- ✅ LSP servers (all configured)
- ✅ Treesitter (all parsers installed)
- ✅ Git integration (lazygit, delta)
- ✅ Mason (all language toolchains)

**Minor Warnings (non-critical):**
- ⚠️ luarocks lua version mismatch (ignorable)
- ⚠️ Missing optional: Composer, PHP, Julia
- ⚠️ LaTeX rendering tools (tectonic/pdflatex)
- ⚠️ Ruby neovim gem

**Conclusion**: Current setup works excellently. Improvements are for fresh installs and edge cases.

---

## Improvement 1: GitHub API Robustness

### Problem
Current implementation is fragile:
```bash
get_latest_version() {
    curl -s https://api.github.com/repos/neovim/neovim/releases/latest | \
        grep '"tag_name"' | \
        cut -d'"' -f4 | \
        sed 's/v//'
}
```

**Issues:**
- No error handling
- No validation of response
- No rate limit detection
- Silent failures return empty string

### Solution: Robust API Handling

```bash
get_latest_version() {
    local max_retries=3
    local retry_delay=2
    local latest_version=""

    step "Fetching latest Neovim version from GitHub API..."

    for ((attempt=1; attempt<=max_retries; attempt++)); do
        if [[ "$VERBOSE" == "true" ]]; then
            info "Attempt $attempt of $max_retries..."
        fi

        # Use GitHub API with proper error handling
        local api_response
        api_response=$(curl -s -f \
            -H "Accept: application/vnd.github.v3+json" \
            --connect-timeout 10 \
            --max-time 30 \
            "https://api.github.com/repos/neovim/neovim/releases/latest" 2>&1)

        local curl_exit=$?

        # Check curl exit status
        if [[ $curl_exit -ne 0 ]]; then
            warning "Network request failed (exit code: $curl_exit)"
            if [[ $attempt -lt $max_retries ]]; then
                info "Retrying in ${retry_delay}s..."
                sleep $retry_delay
                continue
            else
                error "Failed to fetch version after $max_retries attempts"
                return 1
            fi
        fi

        # Validate JSON response (requires jq)
        if ! echo "$api_response" | jq empty 2>/dev/null; then
            warning "Invalid JSON response from GitHub API"
            if [[ $attempt -lt $max_retries ]]; then
                info "Retrying in ${retry_delay}s..."
                sleep $retry_delay
                continue
            else
                error "GitHub API returned invalid JSON"
                return 1
            fi
        fi

        # Check for API rate limiting
        if echo "$api_response" | jq -e '.message' 2>/dev/null | grep -qi "rate limit"; then
            error "GitHub API rate limit exceeded"
            local reset_time=$(echo "$api_response" | jq -r '.documentation_url // "unknown"')
            warning "See: $reset_time"
            return 1
        fi

        # Extract version
        latest_version=$(echo "$api_response" | jq -r '.tag_name // empty' | sed 's/^v//')

        # Validate version format (semantic versioning: X.Y.Z)
        if [[ ! "$latest_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            warning "Invalid version format received: '$latest_version'"
            if [[ $attempt -lt $max_retries ]]; then
                info "Retrying in ${retry_delay}s..."
                sleep $retry_delay
                continue
            else
                error "Could not extract valid version number"
                return 1
            fi
        fi

        # Success!
        success "Latest Neovim version: v$latest_version"
        echo "$latest_version"
        return 0
    done

    error "Failed to get latest version after all retries"
    return 1
}
```

**Benefits:**
- ✅ Retry logic (3 attempts with backoff)
- ✅ JSON validation using `jq`
- ✅ Rate limit detection
- ✅ Semantic version validation
- ✅ Connection timeouts (10s connect, 30s total)
- ✅ Detailed error messages
- ✅ Graceful degradation

**New Dependency:** Requires `jq` package

---

## Improvement 2: Comprehensive Dependency Management

### Complete Dependency Matrix

#### 1. Build Dependencies (Critical)

**Purpose**: Required to compile Neovim from source

**Fedora/RHEL:**
```bash
BUILD_DEPS_FEDORA=(
    "git" "cmake" "ninja-build" "gcc" "gcc-c++" "make"
    "libtool" "autoconf" "automake" "pkgconfig" "unzip"
    "gettext-devel" "libtool-ltdl-devel"
)
```

**Ubuntu/Debian:**
```bash
BUILD_DEPS_UBUNTU=(
    "git" "cmake" "ninja-build" "gcc" "g++" "make"
    "libtool" "libtool-bin" "autoconf" "automake"
    "pkg-config" "unzip" "gettext"
)
```

#### 2. Runtime Dependencies (Required for plugins)

**Core Tools:**
```bash
RUNTIME_TOOLS=(
    "rg:ripgrep"           # Telescope live grep
    "fd:fd-find"           # Telescope file finding
    "git:git"              # Version control
    "curl:curl"            # API calls, downloads
    "jq:jq"                # JSON parsing
)
```

**Git Enhancement:**
```bash
GIT_TOOLS=(
    "delta:git-delta"      # Better git diffs
    "lazygit:lazygit"      # Git TUI
    "bat:bat"              # File preview
)
```

#### 3. Language Toolchains (For LSP servers)

**Node.js Ecosystem:**
```bash
NODEJS_TOOLS=(
    "node:nodejs"          # Runtime for LSP servers
    "npm:npm"              # Package manager
)
```

**LSP Servers requiring Node.js:**
- `bash-language-server`
- `typescript-language-server`
- `vscode-langservers-extracted` (eslint)
- `docker-langserver-nodejs`
- `yaml-language-server`

**Python Ecosystem:**
```bash
PYTHON_TOOLS=(
    "python3:python3"      # Python runtime
    "pip3:python3-pip"     # Package manager
)
```

**LSP Servers requiring Python:**
- `pyright`
- `ruff`

**Go Ecosystem:**
```bash
GO_TOOLS=(
    "go:golang"            # Go runtime
)
```

**LSP Servers requiring Go:**
- `gopls`
- `golangci-lint`

**Rust Ecosystem:**
```bash
RUST_TOOLS=(
    "cargo:cargo"          # Rust package manager
    "rustc:rust"           # Rust compiler
)
```

#### 4. Optional Dependencies (Enhanced features)

**Markdown & Documentation:**
```bash
MARKDOWN_TOOLS=(
    "chrome:google-chrome-stable"  # Markdown preview
    "firefox:firefox"              # Alternative browser
)
```

**Image Rendering (Snacks.nvim):**
```bash
IMAGE_TOOLS=(
    "magick:ImageMagick"   # Image conversion
    "gs:ghostscript"       # PDF rendering
    "mmdc:mermaid-cli"     # Mermaid diagrams
)
```

**LaTeX Support:**
```bash
LATEX_TOOLS=(
    "tectonic:tectonic"    # LaTeX compiler
    "pdflatex:texlive"     # Alternative
)
```

**Treesitter:**
```bash
TREESITTER_TOOLS=(
    "tree-sitter:tree-sitter-cli"  # Parser generator
    "cc:gcc"                        # C compiler for parsers
)
```

### Dependency Checking Function

```bash
check_system_dependencies() {
    local dep_category="$1"    # build|runtime|optional|language|all
    local install_missing="$2" # true|false
    local os_type="$3"         # fedora|ubuntu|auto

    step "Checking ${dep_category} dependencies..."

    # Auto-detect OS if not specified
    if [[ "$os_type" == "auto" || -z "$os_type" ]]; then
        os_type=$(detect_os)
    fi

    local missing_build=()
    local missing_runtime=()
    local missing_optional=()
    local missing_language=()

    # Build dependencies check
    if [[ "$dep_category" == "build" || "$dep_category" == "all" ]]; then
        if [[ "$os_type" == "fedora" ]]; then
            check_packages_fedora "${BUILD_DEPS_FEDORA[@]}" missing_build
        elif [[ "$os_type" == "ubuntu" ]]; then
            check_packages_ubuntu "${BUILD_DEPS_UBUNTU[@]}" missing_build
        fi
    fi

    # Runtime tools check (OS-agnostic)
    if [[ "$dep_category" == "runtime" || "$dep_category" == "all" ]]; then
        for tool_spec in "${RUNTIME_TOOLS[@]}"; do
            local cmd="${tool_spec%%:*}"
            local pkg="${tool_spec##*:}"

            if ! command -v "$cmd" &> /dev/null; then
                missing_runtime+=("$pkg")
            fi
        done
    fi

    # Optional tools check
    if [[ "$dep_category" == "optional" || "$dep_category" == "all" ]]; then
        for tool_spec in "${GIT_TOOLS[@]}" "${IMAGE_TOOLS[@]}" "${LATEX_TOOLS[@]}"; do
            local cmd="${tool_spec%%:*}"
            local pkg="${tool_spec##*:}"

            if ! command -v "$cmd" &> /dev/null; then
                missing_optional+=("$pkg")
            fi
        done
    fi

    # Language toolchains check
    if [[ "$dep_category" == "language" || "$dep_category" == "all" ]]; then
        for tool_spec in "${NODEJS_TOOLS[@]}" "${PYTHON_TOOLS[@]}" "${GO_TOOLS[@]}" "${RUST_TOOLS[@]}"; do
            local cmd="${tool_spec%%:*}"
            local pkg="${tool_spec##*:}"

            if ! command -v "$cmd" &> /dev/null; then
                missing_language+=("$pkg")
            fi
        done
    fi

    # Report findings
    local total_missing=$((${#missing_build[@]} + ${#missing_runtime[@]} + ${#missing_optional[@]} + ${#missing_language[@]}))

    if [[ $total_missing -eq 0 ]]; then
        success "All ${dep_category} dependencies are satisfied!"
        return 0
    fi

    # Display missing dependencies
    if [[ ${#missing_build[@]} -gt 0 ]]; then
        warning "Missing build dependencies:"
        printf '  - %s\n' "${missing_build[@]}"
    fi

    if [[ ${#missing_runtime[@]} -gt 0 ]]; then
        warning "Missing runtime dependencies:"
        printf '  - %s\n' "${missing_runtime[@]}"
    fi

    if [[ ${#missing_optional[@]} -gt 0 ]]; then
        info "Missing optional dependencies (enhance features):"
        printf '  - %s\n' "${missing_optional[@]}"
    fi

    if [[ ${#missing_language[@]} -gt 0 ]]; then
        info "Missing language toolchains (for LSP servers):"
        printf '  - %s\n' "${missing_language[@]}"
    fi

    # Offer to install if requested
    if [[ "$install_missing" == "true" ]]; then
        echo ""
        read -p "Install missing dependencies? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            install_dependencies "$os_type" missing_build missing_runtime missing_optional missing_language
        fi
    fi

    return 1
}

# OS Detection
detect_os() {
    if [[ -f /etc/fedora-release ]]; then
        echo "fedora"
    elif [[ -f /etc/lsb-release ]] && grep -qi ubuntu /etc/lsb-release; then
        echo "ubuntu"
    elif [[ -f /etc/debian_version ]]; then
        echo "ubuntu"  # Treat Debian as Ubuntu for package names
    else
        echo "unknown"
    fi
}

# Install dependencies based on OS
install_dependencies() {
    local os_type="$1"
    shift
    local -n build_ref=$1
    local -n runtime_ref=$2
    local -n optional_ref=$3
    local -n language_ref=$4

    if [[ "$os_type" == "fedora" ]]; then
        if [[ ${#build_ref[@]} -gt 0 || ${#runtime_ref[@]} -gt 0 ]]; then
            info "Installing critical dependencies..."
            sudo dnf install -y "${build_ref[@]}" "${runtime_ref[@]}"
        fi

        if [[ ${#optional_ref[@]} -gt 0 ]]; then
            read -p "Install optional dependencies? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                sudo dnf install -y "${optional_ref[@]}"
            fi
        fi

        if [[ ${#language_ref[@]} -gt 0 ]]; then
            read -p "Install language toolchains for LSP servers? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                sudo dnf install -y "${language_ref[@]}"
            fi
        fi
    elif [[ "$os_type" == "ubuntu" ]]; then
        sudo apt update
        if [[ ${#build_ref[@]} -gt 0 || ${#runtime_ref[@]} -gt 0 ]]; then
            info "Installing critical dependencies..."
            sudo apt install -y "${build_ref[@]}" "${runtime_ref[@]}"
        fi
        # Similar optional/language prompts as Fedora
    fi
}
```

---

## Improvement 3: Unified Script Architecture

### Command-Based Interface Design

```bash
#!/bin/bash
# nvim-manager.sh - Unified Neovim Installation and Management Script
# Usage: ./nvim-manager.sh <command> [options]

# Commands:
#   install       Fresh installation with dependency management
#   update        Update existing Neovim installation
#   deps          Check and manage dependencies
#   health        Parse healthcheck and suggest fixes
#   rollback      Restore from backup
#   test          Run test suite
#
# Options:
#   --method=<source|dnf|flatpak>  Installation method (default: source)
#   --deps-only                     Install dependencies without Neovim
#   --skip-deps                     Skip dependency installation
#   --skip-optional                 Skip optional dependencies
#   --check-health                  Run healthcheck after operation
#   --dry-run                       Show what would be done
#   --verbose                       Detailed output

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMAND="${1:-help}"
shift || true

# Parse command
case "$COMMAND" in
    install)
        cmd_install "$@"
        ;;
    update)
        cmd_update "$@"
        ;;
    deps|dependencies)
        cmd_dependencies "$@"
        ;;
    health|check)
        cmd_health "$@"
        ;;
    rollback)
        cmd_rollback "$@"
        ;;
    test)
        cmd_test "$@"
        ;;
    help|-h|--help)
        show_help
        ;;
    *)
        error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac
```

### Command Implementations

```bash
# Fresh installation
cmd_install() {
    step "Starting fresh Neovim installation..."

    # Detect OS
    local os_type=$(detect_os)
    info "Detected OS: $os_type"

    # Check and install dependencies
    if [[ "$SKIP_DEPS" != "true" ]]; then
        check_system_dependencies "all" "true" "$os_type"
    fi

    # Proceed with Neovim installation
    if [[ "$DEPS_ONLY" != "true" ]]; then
        case "$INSTALL_METHOD" in
            source)
                build_from_source
                ;;
            dnf)
                install_via_dnf
                ;;
            flatpak)
                install_via_flatpak
                ;;
        esac
    fi

    # Run healthcheck if requested
    if [[ "$CHECK_HEALTH" == "true" ]]; then
        cmd_health
    fi
}

# Update existing installation
cmd_update() {
    step "Updating Neovim installation..."

    # Check current version
    local current_version=$(get_current_version)
    local latest_version=$(get_latest_version)

    if [[ "$current_version" == "$latest_version" ]]; then
        success "Neovim is already up to date (v$current_version)"
        return 0
    fi

    info "Current: v$current_version"
    info "Latest: v$latest_version"

    # Create backup
    create_backup

    # Perform update based on method
    case "$INSTALL_METHOD" in
        source)
            build_from_source
            ;;
        dnf)
            install_via_dnf
            ;;
        flatpak)
            install_via_flatpak
            ;;
    esac

    # Test installation
    test_installation
}

# Dependency management
cmd_dependencies() {
    local action="${1:-check}"  # check|install|list

    case "$action" in
        check)
            check_system_dependencies "all" "false" "auto"
            ;;
        install)
            check_system_dependencies "all" "true" "auto"
            ;;
        list)
            list_all_dependencies
            ;;
        *)
            error "Unknown deps action: $action"
            exit 1
            ;;
    esac
}

# Healthcheck analysis
cmd_health() {
    step "Running Neovim healthcheck..."

    if ! command -v nvim &> /dev/null; then
        error "Neovim is not installed"
        return 1
    fi

    # Run healthcheck and capture output
    local health_output
    health_output=$(timeout 30s nvim --headless -c "checkhealth" -c "quit" 2>&1 || true)

    # Parse for issues
    local has_errors=false
    local missing_tools=()

    while IFS= read -r line; do
        if echo "$line" | grep -qi "ERROR\|not found\|missing"; then
            has_errors=true
            # Extract tool name if possible
            local tool=$(echo "$line" | grep -oP '`\K[^`]+' || echo "$line" | grep -oP '\[\K[^\]]+' || true)
            if [[ -n "$tool" ]]; then
                missing_tools+=("$tool")
            fi
        fi
    done <<< "$health_output"

    if [[ "$has_errors" == "true" ]]; then
        warning "Healthcheck found issues"

        if [[ ${#missing_tools[@]} -gt 0 ]]; then
            echo ""
            info "Missing tools detected:"
            printf '  - %s\n' "${missing_tools[@]}"
            echo ""

            read -p "Install missing tools? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                install_missing_tools "${missing_tools[@]}"
            fi
        fi
    else
        success "Healthcheck passed - all systems operational!"
    fi

    # Optionally display full output
    if [[ "$VERBOSE" == "true" ]]; then
        echo ""
        info "Full healthcheck output:"
        echo "$health_output"
    fi
}
```

---

## Improvement 4: Healthcheck Integration

### Parse Healthcheck Output

```bash
parse_nvim_healthcheck() {
    step "Analyzing Neovim healthcheck output..."

    local health_output
    health_output=$(timeout 30s nvim --headless -c "checkhealth" -c "quit" 2>&1 || true)

    # Create structured analysis
    local -A issues
    issues[missing_tools]=""
    issues[warnings]=""
    issues[errors]=""

    # Parse line by line
    local current_plugin=""
    while IFS= read -r line; do
        # Detect plugin section
        if echo "$line" | grep -q "^=="; then
            current_plugin=$(echo "$line" | grep -oP '==\s*\K[^:]+')
        fi

        # Detect errors
        if echo "$line" | grep -qi "ERROR"; then
            issues[errors]+="[$current_plugin] $line\n"

            # Extract tool name
            if echo "$line" | grep -qi "not found"; then
                local tool=$(echo "$line" | grep -oP '`\K[^`]+' || echo "$line" | awk '{print $NF}')
                issues[missing_tools]+="$tool "
            fi
        fi

        # Detect warnings
        if echo "$line" | grep -qi "WARNING"; then
            issues[warnings]+="[$current_plugin] $line\n"
        fi
    done <<< "$health_output"

    # Report findings
    if [[ -n "${issues[errors]}" ]]; then
        warning "Errors found:"
        echo -e "${issues[errors]}"
    fi

    if [[ -n "${issues[warnings]}" ]]; then
        info "Warnings found:"
        echo -e "${issues[warnings]}"
    fi

    if [[ -n "${issues[missing_tools]}" ]]; then
        echo ""
        warning "Missing tools: ${issues[missing_tools]}"

        # Map tools to packages
        suggest_packages "${issues[missing_tools]}"
    fi
}

# Suggest package installation based on missing tools
suggest_packages() {
    local missing_tools="$1"
    local os_type=$(detect_os)

    # Tool to package mapping
    declare -A tool_to_package_fedora
    tool_to_package_fedora[rg]="ripgrep"
    tool_to_package_fedora[fd]="fd-find"
    tool_to_package_fedora[lazygit]="lazygit"
    tool_to_package_fedora[delta]="git-delta"
    tool_to_package_fedora[bat]="bat"
    tool_to_package_fedora[node]="nodejs"
    tool_to_package_fedora[npm]="npm"
    tool_to_package_fedora[python3]="python3"
    tool_to_package_fedora[pip3]="python3-pip"

    local packages_to_install=()

    for tool in $missing_tools; do
        if [[ -n "${tool_to_package_fedora[$tool]}" ]]; then
            packages_to_install+=("${tool_to_package_fedora[$tool]}")
        fi
    done

    if [[ ${#packages_to_install[@]} -gt 0 ]]; then
        echo ""
        if [[ "$os_type" == "fedora" ]]; then
            info "Install with: sudo dnf install ${packages_to_install[*]}"
        elif [[ "$os_type" == "ubuntu" ]]; then
            info "Install with: sudo apt install ${packages_to_install[*]}"
        fi
    fi
}
```

---

## Implementation Phases

### Phase 1: Core Improvements (Low Risk)
**Focus**: Enhance existing `nvim-update.sh` without breaking changes

**Tasks:**
1. ✅ Add improved `get_latest_version()` function
2. ✅ Add `check_system_dependencies()` function
3. ✅ Add `--fresh-install` flag
4. ✅ Add OS detection
5. ✅ Test on current system

**Time Estimate**: 2-3 hours
**Risk**: Low - adds functionality without removing existing code

### Phase 2: Dependency Management (Medium Risk)
**Focus**: Comprehensive dependency checking and installation

**Tasks:**
1. ✅ Create complete dependency matrices (Fedora/Ubuntu)
2. ✅ Implement package installation functions
3. ✅ Add interactive prompts for optional tools
4. ✅ Test on fresh VM or container

**Time Estimate**: 3-4 hours
**Risk**: Medium - requires testing on fresh systems

### Phase 3: Command Interface (Optional)
**Focus**: Unified command-based interface

**Tasks:**
1. ✅ Create `nvim-manager.sh` wrapper script
2. ✅ Implement command routing
3. ✅ Add `install`, `update`, `deps`, `health` commands
4. ✅ Create backward compatibility symlinks
5. ✅ Update documentation

**Time Estimate**: 4-5 hours
**Risk**: Low - wrapper around existing functionality

### Phase 4: Advanced Features (Optional)
**Focus**: Healthcheck integration and automation

**Tasks:**
1. ✅ Implement healthcheck parsing
2. ✅ Add tool-to-package mapping
3. ✅ Create suggestion system
4. ✅ Add automation flags

**Time Estimate**: 3-4 hours
**Risk**: Low - optional feature additions

---

## Code Examples

### Example 1: Fresh Install on New Fedora System

```bash
# Run with all features
./nvim-update.sh --fresh-install --check-health

# Output:
# === Neovim Installation Script ===
#
# ℹ Detected OS: Fedora 41
# ▶ Checking all dependencies...
# ✓ All build dependencies are satisfied
# ⚠ Missing runtime dependencies:
#   - ripgrep
#   - fd-find
# ℹ Missing optional dependencies:
#   - git-delta
#   - lazygit
#
# Install missing dependencies? (y/n) y
# ▶ Installing dependencies...
# ✓ Dependencies installed
#
# ▶ Building Neovim from source...
# ✓ Neovim v0.11.5 installed successfully
#
# ▶ Running healthcheck...
# ✓ All checks passed!
```

### Example 2: Update Existing Installation

```bash
# Standard update
./nvim-update.sh

# Output:
# === Neovim Update Script ===
#
# ℹ Current version: v0.11.4
# ℹ Latest version: v0.11.5
# ▶ Creating backup...
# ✓ Backup created
# ▶ Building from source...
# ✓ Neovim successfully updated to v0.11.5
```

### Example 3: Dependency Check Only

```bash
# Check what's missing
./nvim-update.sh --deps-only --dry-run

# Output:
# === Dependency Check ===
#
# ✓ Build dependencies: satisfied
# ✓ Runtime tools: satisfied
# ℹ Optional tools missing:
#   - lazydocker (Docker TUI)
#   - tectonic (LaTeX compiler)
# ℹ Language toolchains missing:
#   - julia (for Julia LSP)
```

---

## Migration Path

### For Current Users
No changes required - existing script continues to work:
```bash
./nvim-update.sh              # Still works as before
./nvim-update.sh --build-source  # Still works
```

### For Fresh Installs
Use new features:
```bash
./nvim-update.sh --fresh-install  # Handles all deps
```

### For Future Unified Manager
Optional migration:
```bash
./nvim-manager.sh install     # New interface
./nvim-manager.sh update      # New interface
```

---

## Testing Strategy

### Test Environments
1. **Existing System** (Current Fedora 41)
   - Test update functionality
   - Verify no regressions
   - Confirm improved error handling

2. **Fresh Fedora VM**
   - Test fresh install
   - Verify all dependencies installed
   - Confirm Neovim builds and works

3. **Fresh Ubuntu VM** (Optional)
   - Test cross-platform support
   - Verify package name mappings
   - Confirm apt commands work

### Test Cases
- ✅ Update with no version change
- ✅ Update with new version available
- ✅ Fresh install with missing dependencies
- ✅ GitHub API failure scenarios
- ✅ Invalid version responses
- ✅ Network timeout handling
- ✅ Dependency installation prompts
- ✅ Healthcheck parsing accuracy

---

## Benefits Summary

### For Fresh Systems
- ✅ One-command installation of everything
- ✅ Automatic dependency detection
- ✅ Guided installation of optional tools
- ✅ Post-install health verification

### For Existing Systems
- ✅ Improved reliability (API validation)
- ✅ Better error messages
- ✅ Dependency gap detection
- ✅ No breaking changes

### For Maintenance
- ✅ Easier troubleshooting
- ✅ Healthcheck-driven fixes
- ✅ Cross-platform support
- ✅ Comprehensive documentation

---

## References

### Key Dependencies

**Critical (must have):**
- git, cmake, ninja-build, gcc, make
- ripgrep, fd-find, curl, jq

**Important (for LSP):**
- nodejs, npm, python3, pip3, go, cargo

**Optional (enhanced features):**
- git-delta, lazygit, bat, lazydocker
- tree-sitter-cli, ImageMagick, ghostscript

### External Resources
- [Neovim Build Prerequisites](https://github.com/neovim/neovim/wiki/Building-Neovim#build-prerequisites)
- [Mason.nvim Requirements](https://github.com/williamboman/mason.nvim#requirements)
- [Telescope.nvim Dependencies](https://github.com/nvim-telescope/telescope.nvim#suggested-dependencies)

---

## Conclusion

This plan provides a comprehensive roadmap for enhancing the Neovim installation experience. The phased approach allows for incremental implementation while maintaining stability of the current working system.

**Current Status**: nvim-update.sh works excellently for updates and source builds.

**Future Goal**: Universal installation manager handling fresh installs, updates, and dependency management across platforms.

**Next Steps**: Implement Phase 1 when ready - low risk, high value improvements to API handling and dependency checking.

---

**Last Updated**: 2025-11-19
**Version**: 1.0
**Status**: Ready for implementation
