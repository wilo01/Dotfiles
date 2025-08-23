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
# Outputs: Process output to stdout, timing info to stderr via PROCESS_ELAPSED_TIME variable
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
            PROCESS_ELAPSED_TIME=$elapsed
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
                PROCESS_ELAPSED_TIME=$elapsed
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
    
    # Calculate final elapsed time
    local end_time=$(date +%s)
    PROCESS_ELAPSED_TIME=$((end_time - start_time))
    
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
# Analytics & Performance Tracking
# -----------------------------------------------------------------------------

# Analytics file location
readonly ANALYTICS_FILE="${HOOKS_DIR:-$(dirname "${BASH_SOURCE[0]}")}/.git-hooks-analytics.json"
readonly ANALYTICS_MAX_ENTRIES=50

# Initialize analytics file if it doesn't exist
init_analytics() {
    if [[ ! -f "$ANALYTICS_FILE" ]]; then
        cat > "$ANALYTICS_FILE" <<EOF
{
    "claude": {
        "total_calls": 0,
        "successful_calls": 0,
        "failed_calls": 0,
        "total_time": 0,
        "wins": 0,
        "recent_failures": [],
        "last_success": null,
        "disabled_until": null
    },
    "gemini": {
        "total_calls": 0,
        "successful_calls": 0,
        "failed_calls": 0,
        "total_time": 0,
        "wins": 0,
        "recent_failures": [],
        "last_success": null,
        "disabled_until": null
    },
    "history": []
}
EOF
    fi
}

# Record AI performance
record_ai_performance() {
    local ai_name="$1"
    local success="$2"  # true/false
    local response_time="$3"
    local is_winner="${4:-false}"  # true/false for race mode
    
    init_analytics
    
    # Read current analytics
    local analytics=$(cat "$ANALYTICS_FILE")
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Update using jq if available, otherwise use python
    if command_exists jq; then
        analytics=$(echo "$analytics" | jq \
            --arg ai "$ai_name" \
            --arg success "$success" \
            --arg time "$response_time" \
            --arg winner "$is_winner" \
            --arg ts "$timestamp" \
            '
            .[$ai].total_calls += 1 |
            if $success == "true" then
                .[$ai].successful_calls += 1 |
                .[$ai].total_time += ($time | tonumber) |
                .[$ai].last_success = $ts |
                .[$ai].recent_failures = []
            else
                .[$ai].failed_calls += 1 |
                .[$ai].recent_failures += [$ts] |
                .[$ai].recent_failures = .[$ai].recent_failures[-3:]
            end |
            if $winner == "true" then
                .[$ai].wins += 1
            end |
            .history += [{
                "ai": $ai,
                "success": ($success == "true"),
                "time": ($time | tonumber),
                "winner": ($winner == "true"),
                "timestamp": $ts
            }] |
            .history = .history[-50:]
            ')
    elif command_exists python3; then
        analytics=$(python3 -c "
import json
import sys

data = json.loads('''$analytics''')
ai = '$ai_name'
success = '$success' == 'true'
time = float('$response_time') if '$response_time' != 'failed' else 0
winner = '$is_winner' == 'true'
ts = '$timestamp'

data[ai]['total_calls'] += 1
if success:
    data[ai]['successful_calls'] += 1
    data[ai]['total_time'] += time
    data[ai]['last_success'] = ts
    data[ai]['recent_failures'] = []
else:
    data[ai]['failed_calls'] += 1
    data[ai]['recent_failures'].append(ts)
    data[ai]['recent_failures'] = data[ai]['recent_failures'][-3:]

if winner:
    data[ai]['wins'] += 1

data['history'].append({
    'ai': ai,
    'success': success,
    'time': time,
    'winner': winner,
    'timestamp': ts
})
data['history'] = data['history'][-50:]

print(json.dumps(data, indent=2))
")
    fi
    
    # Check if AI should be disabled (3 consecutive failures)
    local recent_failures_count=$(echo "$analytics" | grep -o "\"$ai_name\".*recent_failures.*\[.*\]" | grep -o "\"20" | wc -l)
    if [[ "$recent_failures_count" -ge 3 ]]; then
        local disable_until=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ")
        if command_exists jq; then
            analytics=$(echo "$analytics" | jq --arg ai "$ai_name" --arg until "$disable_until" '.[$ai].disabled_until = $until')
        fi
    fi
    
    # Save updated analytics
    echo "$analytics" > "$ANALYTICS_FILE"
}

# Get AI performance stats
get_ai_stats() {
    local ai_name="$1"
    
    init_analytics
    
    if command_exists jq; then
        jq -r --arg ai "$ai_name" '.[$ai]' "$ANALYTICS_FILE"
    elif command_exists python3; then
        python3 -c "
import json
data = json.load(open('$ANALYTICS_FILE'))
print(json.dumps(data['$ai_name'], indent=2))
"
    else
        cat "$ANALYTICS_FILE"
    fi
}

# Check if AI is disabled
is_ai_disabled() {
    local ai_name="$1"
    
    init_analytics
    
    local disabled_until=""
    if command_exists jq; then
        disabled_until=$(jq -r --arg ai "$ai_name" '.[$ai].disabled_until // ""' "$ANALYTICS_FILE")
    elif command_exists python3; then
        disabled_until=$(python3 -c "
import json
data = json.load(open('$ANALYTICS_FILE'))
print(data['$ai_name'].get('disabled_until', ''))
")
    fi
    
    if [[ -n "$disabled_until" ]] && [[ "$disabled_until" != "null" ]]; then
        local current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        if [[ "$current_time" < "$disabled_until" ]]; then
            return 0  # AI is disabled
        else
            # Re-enable AI
            if command_exists jq; then
                local analytics=$(jq --arg ai "$ai_name" '.[$ai].disabled_until = null | .[$ai].recent_failures = []' "$ANALYTICS_FILE")
                echo "$analytics" > "$ANALYTICS_FILE"
            fi
        fi
    fi
    
    return 1  # AI is not disabled
}

# Get best performing AI
get_best_ai() {
    init_analytics
    
    local claude_disabled=$(is_ai_disabled "claude" && echo "true" || echo "false")
    local gemini_disabled=$(is_ai_disabled "gemini" && echo "true" || echo "false")
    
    # If both are disabled, return empty
    if [[ "$claude_disabled" == "true" ]] && [[ "$gemini_disabled" == "true" ]]; then
        echo ""
        return
    fi
    
    # If one is disabled, return the other
    if [[ "$claude_disabled" == "true" ]]; then
        echo "gemini"
        return
    elif [[ "$gemini_disabled" == "true" ]]; then
        echo "claude"
        return
    fi
    
    # Calculate performance scores
    if command_exists jq; then
        jq -r '
            if .claude.total_calls == 0 and .gemini.total_calls == 0 then
                "claude"
            else
                [.claude, .gemini] |
                map({
                    name: (if . == .claude then "claude" else "gemini" end),
                    score: (
                        if .total_calls == 0 then
                            0
                        else
                            (.successful_calls / .total_calls * 100) +
                            (if .successful_calls > 0 then (50 - (.total_time / .successful_calls)) else 0 end)
                        end
                    )
                }) |
                max_by(.score) |
                .name
            end
        ' "$ANALYTICS_FILE"
    else
        # Default to claude if no json processor available
        echo "claude"
    fi
}

# Calculate adaptive timeout based on historical data
get_adaptive_timeout() {
    local ai_name="$1"
    local default_timeout="${2:-30}"
    
    init_analytics
    
    if command_exists jq; then
        local avg_time=$(jq -r --arg ai "$ai_name" '
            if .[$ai].successful_calls > 0 then
                (.[$ai].total_time / .[$ai].successful_calls)
            else
                0
            end
        ' "$ANALYTICS_FILE")
        
        if [[ "$avg_time" != "0" ]]; then
            # Set timeout to 1.5x average + 5s buffer
            local timeout=$(echo "$avg_time * 1.5 + 5" | bc 2>/dev/null || python3 -c "print(int($avg_time * 1.5 + 5))")
            # Ensure within bounds (10-60 seconds)
            if [[ "$timeout" -lt 10 ]]; then
                echo "10"
            elif [[ "$timeout" -gt 60 ]]; then
                echo "60"
            else
                echo "$timeout"
            fi
        else
            echo "$default_timeout"
        fi
    else
        echo "$default_timeout"
    fi
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
export -f init_analytics record_ai_performance get_ai_stats is_ai_disabled get_best_ai get_adaptive_timeout