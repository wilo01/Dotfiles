#!/bin/bash
# Git Hooks Common Library
# Shared functions and utilities for git hooks

# -----------------------------------------------------------------------------
# Configuration Management
# -----------------------------------------------------------------------------

# Load hook configuration from git config with defaults
load_hook_config() {
    # AI Configuration
    AI_MAX_TIMEOUT=$(git config --local hooks.aiMaxTimeout || echo "${AI_MAX_TIMEOUT:-60}")
    AI_INACTIVITY_TIMEOUT=$(git config --local hooks.aiInactivityTimeout || echo "${AI_INACTIVITY_TIMEOUT:-30}")
    AI_SHOW_PROGRESS=$(git config --local hooks.aiShowProgress || echo "${AI_SHOW_PROGRESS:-true}")
    AI_PARALLEL_MODE=$(git config --local hooks.aiParallelMode || echo "${AI_PARALLEL_MODE:-false}")
    AI_DEBUG=$(git config --local hooks.aiDebug || echo "${AI_DEBUG:-false}")
    
    # Hook Settings
    ENABLE_GLOBAL_HOOKS=$(git config --local hooks.enableGlobalHooks || echo "true")
    ENABLE_LOCAL_HOOKS=$(git config --local hooks.enableLocalHooks || echo "false")
    ENABLE_AI_COMMIT=$(git config --local hooks.enableAiCommit || echo "false")
    
    # File Paths
    HOOKS_LOCAL_PATH=$(git config --local hooks.hooksLocalPath | sed "s|^~|$HOME|")
    HOOKS_LOCAL_FILENAME=$(git config --local hooks.hooksLocalFilename)
    
    export AI_MAX_TIMEOUT AI_INACTIVITY_TIMEOUT AI_SHOW_PROGRESS AI_PARALLEL_MODE AI_DEBUG
    export ENABLE_GLOBAL_HOOKS ENABLE_LOCAL_HOOKS ENABLE_AI_COMMIT
    export HOOKS_LOCAL_PATH HOOKS_LOCAL_FILENAME
}

# Debug configuration if enabled
debug_config() {
    if [[ "$AI_DEBUG" == "true" ]]; then
        log_info "🔧 AI Hook Configuration:"
        log_info "  Max Timeout: ${AI_MAX_TIMEOUT}s"
        log_info "  Inactivity Timeout: ${AI_INACTIVITY_TIMEOUT}s"
        log_info "  Show Progress: $AI_SHOW_PROGRESS"
        log_info "  Parallel Mode: $AI_PARALLEL_MODE"
    fi
}

# -----------------------------------------------------------------------------
# Logging Utilities
# -----------------------------------------------------------------------------

# Color codes for terminal output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_RESET='\033[0m'

# Log functions with consistent formatting
log_info() {
    echo "$@" >&2
}

log_success() {
    echo "✅ $*" >&2
}

log_warning() {
    echo "⚠️  $*" >&2
}

log_error() {
    echo "❌ $*" >&2
}

log_debug() {
    if [[ "$AI_DEBUG" == "true" ]]; then
        echo "🔍 $*" >&2
    fi
}

# -----------------------------------------------------------------------------
# Progress Indicators
# -----------------------------------------------------------------------------

# Spinner characters for progress indication
readonly SPINNER_CHARS='⣾⣽⣻⢿⡿⣟⣯⣷'

# Show progress with spinner
# Usage: show_progress "message" elapsed_seconds
show_progress() {
    local message="$1"
    local elapsed="${2:-0}"
    local spinner_index=$(( elapsed % 8 ))
    
    if [[ "$AI_SHOW_PROGRESS" == "true" ]]; then
        printf "\r${SPINNER_CHARS:$spinner_index:1} %s... (%ds)" "$message" "$elapsed" >&2
    fi
}

# Clear progress line
clear_progress() {
    if [[ "$AI_SHOW_PROGRESS" == "true" ]]; then
        printf "\r%-60s\r" " " >&2
    fi
}

# -----------------------------------------------------------------------------
# Process Management
# -----------------------------------------------------------------------------

# Monitor a process with activity timeout
# Usage: monitor_process_with_timeout cmd description [max_timeout] [inactivity_timeout]
monitor_process_with_timeout() {
    local cmd="$1"
    local description="${2:-Process}"
    local max_timeout="${3:-$AI_MAX_TIMEOUT}"
    local inactivity_timeout="${4:-$AI_INACTIVITY_TIMEOUT}"
    
    local temp_file=$(mktemp)
    local error_file=$(mktemp)
    local last_size=0
    local last_activity=$(date +%s)
    local start_time=$(date +%s)
    
    # Clean up temp files on exit
    trap "rm -f '$temp_file' '$error_file'" RETURN
    
    # Start command in background
    eval "$cmd" > "$temp_file" 2>"$error_file" &
    local pid=$!
    
    # Monitor loop
    while kill -0 "$pid" 2>/dev/null; do
        local current_size=$(stat -f%z "$temp_file" 2>/dev/null || stat -c%s "$temp_file" 2>/dev/null || echo 0)
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        
        # Check max timeout
        if (( elapsed > max_timeout )); then
            kill_process_tree "$pid"
            log_warning "$description exceeded maximum timeout (${max_timeout}s)"
            return 124  # timeout exit code
        fi
        
        # Check activity
        if [[ "$current_size" != "$last_size" ]]; then
            last_activity=$current_time
            last_size=$current_size
        else
            local inactive_time=$((current_time - last_activity))
            if (( inactive_time > inactivity_timeout )); then
                kill_process_tree "$pid"
                log_warning "$description timed out after ${inactivity_timeout}s of inactivity"
                return 124  # timeout exit code
            fi
        fi
        
        show_progress "Waiting for $description" "$elapsed"
        sleep 0.2
    done
    
    clear_progress
    
    # Wait for process and get exit code
    wait "$pid" 2>/dev/null
    local exit_code=$?
    
    # Show errors if any
    if [[ -s "$error_file" ]] && [[ "$AI_DEBUG" == "true" ]]; then
        log_warning "$description error output:"
        cat "$error_file" >&2
    fi
    
    # Output result
    cat "$temp_file"
    
    return $exit_code
}

# Kill a process and all its children
kill_process_tree() {
    local pid="$1"
    local signal="${2:-TERM}"
    
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        # Try graceful termination first
        kill -"$signal" "$pid" 2>/dev/null
        sleep 0.5
        
        # Force kill if still running
        if kill -0 "$pid" 2>/dev/null; then
            kill -KILL "$pid" 2>/dev/null
        fi
    fi
}

# -----------------------------------------------------------------------------
# Git Utilities
# -----------------------------------------------------------------------------

# Get current branch name
get_current_branch() {
    git symbolic-ref --short HEAD 2>/dev/null || echo "HEAD"
}

# Get JIRA tag from branch name (first two dash-separated parts)
get_jira_tag() {
    get_current_branch | cut -d'-' -f1,2
}

# Get project root directory
get_project_root() {
    git rev-parse --show-toplevel 2>/dev/null || pwd
}

# Check if there are staged changes
has_staged_changes() {
    git diff --cached --quiet 2>/dev/null
    [[ $? -ne 0 ]]
}

# Get diff statistics
get_diff_stats() {
    local files_changed=$(git diff --staged --name-status 2>/dev/null | wc -l)
    local lines_added=$(git diff --staged --numstat 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    local lines_deleted=$(git diff --staged --numstat 2>/dev/null | awk '{sum+=$2} END {print sum+0}')
    
    echo "$files_changed $lines_added $lines_deleted"
}

# -----------------------------------------------------------------------------
# File Operations
# -----------------------------------------------------------------------------

# Safely write to a file with backup
safe_write_file() {
    local file="$1"
    local content="$2"
    
    # Create directory if it doesn't exist
    local dir=$(dirname "$file")
    [[ -d "$dir" ]] || mkdir -p "$dir"
    
    # Backup existing file
    if [[ -f "$file" ]]; then
        cp "$file" "${file}.bak"
    fi
    
    # Write new content
    echo "$content" > "$file"
}

# Read file safely with fallback
safe_read_file() {
    local file="$1"
    local default="${2:-}"
    
    if [[ -f "$file" ]] && [[ -r "$file" ]]; then
        cat "$file"
    else
        echo "$default"
    fi
}

# -----------------------------------------------------------------------------
# Validation Utilities
# -----------------------------------------------------------------------------

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Validate required commands
validate_commands() {
    local commands=("$@")
    local missing=()
    
    for cmd in "${commands[@]}"; do
        if ! command_exists "$cmd"; then
            missing+=("$cmd")
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required commands: ${missing[*]}"
        return 1
    fi
    
    return 0
}

# -----------------------------------------------------------------------------
# String Utilities
# -----------------------------------------------------------------------------

# Escape string for use in shell commands
shell_escape() {
    printf '%q' "$1"
}

# Trim whitespace from string
trim() {
    local var="$*"
    var="${var#"${var%%[![:space:]]*}"}"   # remove leading whitespace
    var="${var%"${var##*[![:space:]]}"}"   # remove trailing whitespace
    echo -n "$var"
}

# -----------------------------------------------------------------------------
# Export Functions
# -----------------------------------------------------------------------------

# Export all functions for use in subshells
export -f load_hook_config debug_config
export -f log_info log_success log_warning log_error log_debug
export -f show_progress clear_progress
export -f monitor_process_with_timeout kill_process_tree
export -f get_current_branch get_jira_tag get_project_root has_staged_changes get_diff_stats
export -f safe_write_file safe_read_file
export -f command_exists validate_commands
export -f shell_escape trim