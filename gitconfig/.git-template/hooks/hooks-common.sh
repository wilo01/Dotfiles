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
   AI_INACTIVITY_TIMEOUT=$(git config --local hooks.aiInactivityTimeout || echo "${AI_INACTIVITY_TIMEOUT:-60}")
   AI_SHOW_PROGRESS=$(git config --local hooks.aiShowProgress || echo "${AI_SHOW_PROGRESS:-true}")
   AI_PARALLEL_MODE=$(git config --local hooks.aiParallelMode || echo "${AI_PARALLEL_MODE:-true}")
   AI_DEBUG=$(git config --local hooks.aiDebug || echo "${AI_DEBUG:-false}")

   # Hook Settings
   ENABLE_GLOBAL_HOOKS=$(git config --local hooks.enableGlobalHooks || echo "true")
   ENABLE_LOCAL_HOOKS=$(git config --local hooks.enableLocalHooks || echo "false")
   ENABLE_AI_COMMIT=$(git config --local hooks.enableAiCommit || echo "false")

   # Maintenance branch prompt timeout (seconds)
   MAINTENANCE_TIMEOUT=$(git config --local hooks.maintenanceTimeout || echo "${MAINTENANCE_TIMEOUT:-10}")

   # AI Command Paths - auto-detect if not configured
   AI_CLAUDE_CMD=$(git config --local hooks.aiClaudeCmd 2>/dev/null)
   if [[ -z "$AI_CLAUDE_CMD" ]]; then
      AI_CLAUDE_CMD=$(which claude 2>/dev/null || echo "/usr/local/bin/claude")
   fi

   AI_GEMINI_CMD=$(git config --local hooks.aiGeminiCmd 2>/dev/null)
   if [[ -z "$AI_GEMINI_CMD" ]]; then
      AI_GEMINI_CMD=$(which gemini 2>/dev/null || echo "/usr/local/bin/gemini")
   fi

   # File Paths (validation will be done later if validate_safe_path is available)
   HOOKS_LOCAL_PATH=$(git config --local hooks.hooksLocalPath | sed "s|^~|$HOME|")
   HOOKS_LOCAL_FILENAME=$(git config --local hooks.hooksLocalFilename)

   export AI_MAX_TIMEOUT AI_INACTIVITY_TIMEOUT AI_SHOW_PROGRESS AI_PARALLEL_MODE AI_DEBUG
   export ENABLE_GLOBAL_HOOKS ENABLE_LOCAL_HOOKS ENABLE_AI_COMMIT MAINTENANCE_TIMEOUT
   export AI_CLAUDE_CMD AI_GEMINI_CMD
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
   local spinner_index=$((elapsed % 8))

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
# Usage: monitor_process_with_timeout cmd description [max_timeout] [inactivity_timeout] [time_output_file]
# Outputs: Process output to stdout, timing info to stderr via PROCESS_ELAPSED_TIME variable
# Note: When called through a pipe, PROCESS_ELAPSED_TIME won't propagate. Use time_output_file instead.
monitor_process_with_timeout() {
   local cmd="$1"
   local description="${2:-Process}"
   local max_timeout="${3:-$AI_MAX_TIMEOUT}"
   local inactivity_timeout="${4:-$AI_INACTIVITY_TIMEOUT}"
   local time_output_file="${5:-}"

   # Create temp files securely with error handling
   local temp_file
   temp_file=$(mktemp) || {
      log_error "Failed to create temp file for $description"
      return 1
   }
   local error_file
   error_file=$(mktemp) || {
      rm -f "$temp_file"
      log_error "Failed to create error file for $description"
      return 1
   }

   # Set restrictive permissions on temp files
   chmod 600 "$temp_file" "$error_file"

   local last_size=0
   local last_activity start_time
   last_activity=$(date +%s)
   start_time=$(date +%s)

   # Clean up temp files on exit (variables captured at trap definition time)
   # shellcheck disable=SC2064
   trap "rm -f '$temp_file' '$error_file'" RETURN INT TERM

   # Start command in background - execute as array to prevent injection
   # Split command into array to safely execute without shell expansion
   local cmd_array
   read -ra cmd_array <<<"$cmd"
   "${cmd_array[@]}" >"$temp_file" 2>"$error_file" &
   local pid=$!

   # Monitor loop
   local current_size current_time elapsed
   while kill -0 "$pid" 2>/dev/null; do
      current_size=$(stat -f%z "$temp_file" 2>/dev/null || stat -c%s "$temp_file" 2>/dev/null || echo 0)
      current_time=$(date +%s)
      elapsed=$((current_time - start_time))

      # Check max timeout
      if ((elapsed > max_timeout)); then
         kill_process_tree "$pid"
         log_warning "$description exceeded maximum timeout (${max_timeout}s)"
         PROCESS_ELAPSED_TIME=$elapsed
         # Write time to file before returning (for pipe contexts)
         # Only write to files in /tmp (created by mktemp)
         if [[ -n "$time_output_file" ]] && [[ "$time_output_file" == /tmp/* ]]; then
            echo "$PROCESS_ELAPSED_TIME" > "$time_output_file"
         fi
         return 124 # timeout exit code
      fi

      # Check activity
      if [[ "$current_size" != "$last_size" ]]; then
         last_activity=$current_time
         last_size=$current_size
      else
         local inactive_time=$((current_time - last_activity))
         if ((inactive_time > inactivity_timeout)); then
            kill_process_tree "$pid"
            log_warning "$description timed out after ${inactivity_timeout}s of inactivity"
            PROCESS_ELAPSED_TIME=$elapsed
            # Write time to file before returning (for pipe contexts)
            # Only write to files in /tmp (created by mktemp)
            if [[ -n "$time_output_file" ]] && [[ "$time_output_file" == /tmp/* ]]; then
               echo "$PROCESS_ELAPSED_TIME" > "$time_output_file"
            fi
            return 124 # timeout exit code
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
   local end_time
   end_time=$(date +%s)
   PROCESS_ELAPSED_TIME=$((end_time - start_time))

   # Show errors if any
   if [[ -s "$error_file" ]] && [[ "$AI_DEBUG" == "true" ]]; then
      log_warning "$description error output:"
      cat "$error_file" >&2
   fi

   # Output result
   cat "$temp_file"

   # Write elapsed time to file if requested (for pipe contexts where variable won't propagate)
   # Only write to files in /tmp (created by mktemp)
   if [[ -n "$time_output_file" ]] && [[ "$time_output_file" == /tmp/* ]]; then
      echo "$PROCESS_ELAPSED_TIME" > "$time_output_file"
   fi

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
   ! git diff --cached --quiet 2>/dev/null
}

# Get diff statistics
get_diff_stats() {
   local files_changed lines_added lines_deleted
   files_changed=$(git diff --staged --name-status 2>/dev/null | wc -l)
   lines_added=$(git diff --staged --numstat 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
   lines_deleted=$(git diff --staged --numstat 2>/dev/null | awk '{sum+=$2} END {print sum+0}')

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
   local dir
   dir=$(dirname "$file")
   [[ -d "$dir" ]] || mkdir -p "$dir"

   # Backup existing file
   if [[ -f "$file" ]]; then
      cp "$file" "${file}.bak"
   fi

   # Write new content
   echo "$content" >"$file"
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

# Validate that a path is safe (no traversal attempts)
validate_safe_path() {
   local path="$1"

   # Reject empty paths
   if [[ -z "$path" ]]; then
      return 1
   fi

   # Reject paths with directory traversal
   if [[ "$path" =~ \.\. ]]; then
      log_debug "Path validation failed: contains .."
      return 1
   fi

   # Reject paths with null bytes
   if [[ "$path" =~ $'\0' ]]; then
      log_debug "Path validation failed: contains null byte"
      return 1
   fi

   # Make path absolute if relative
   if [[ "$path" != /* ]]; then
      path="$(pwd)/$path"
   fi

   # Verify path exists and is readable (optional check)
   # Uncomment if you want to enforce existence
   # if [[ ! -e "$path" ]]; then
   #     log_debug "Path validation failed: does not exist"
   #     return 1
   # fi

   echo "$path"
   return 0
}

# Validate AI command is safe to execute
validate_ai_command() {
   local cmd="$1"

   # Check if command exists
   if [[ ! -x "$cmd" ]]; then
      log_debug "AI command validation failed: not executable"
      return 1
   fi

   # Check if command is in expected location or configured AI commands
   if [[ "$cmd" != /usr/local/bin/* ]] && [[ "$cmd" != /usr/bin/* ]] && [[ "$cmd" != "$AI_CLAUDE_CMD"* ]] && [[ "$cmd" != "$AI_GEMINI_CMD"* ]]; then
      log_debug "AI command validation failed: unexpected location"
      return 1
   fi

   return 0
}

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
   var="${var#"${var%%[![:space:]]*}"}" # remove leading whitespace
   var="${var%"${var##*[![:space:]]}"}" # remove trailing whitespace
   echo -n "$var"
}

# -----------------------------------------------------------------------------
# Analytics & Performance Tracking
# -----------------------------------------------------------------------------

# Analytics file location
readonly ANALYTICS_FILE="${HOOKS_DIR:-$(dirname "${BASH_SOURCE[0]}")}/.git-hooks-analytics.json"
# shellcheck disable=SC2034  # Used in jq expressions for history trimming
readonly ANALYTICS_MAX_ENTRIES=50

# Initialize analytics file if it doesn't exist or is corrupted
init_analytics() {
    local needs_init=false

    # Check if file exists
    if [[ ! -f "$ANALYTICS_FILE" ]]; then
        needs_init=true
        log_debug "Analytics file does not exist, creating..."
    # Check if file is valid JSON (handles corrupted/empty files)
    elif ! jq -e . "$ANALYTICS_FILE" >/dev/null 2>&1; then
        needs_init=true
        log_warning "Analytics file corrupted, recreating..."
    fi

    if [[ "$needs_init" == "true" ]]; then
        # Use file locking to prevent race conditions during initialization
        local lock_file="${ANALYTICS_FILE}.init.lock"
        local lock_acquired=false

        # Try to acquire lock with short timeout (10 * 0.1s = 1 second)
        local max_wait=10
        local waited=0
        while [[ $waited -lt $max_wait ]]; do
            if mkdir "$lock_file" 2>/dev/null; then
                lock_acquired=true
                break
            fi
            sleep 0.1
            waited=$((waited + 1))
        done

        if [[ "$lock_acquired" != "true" ]]; then
            # Another process is initializing, wait and check again
            sleep 0.5
            if [[ -f "$ANALYTICS_FILE" ]] && jq -e . "$ANALYTICS_FILE" >/dev/null 2>&1; then
                return 0  # File was created by another process
            fi
            log_debug "Failed to acquire init lock, proceeding anyway"
        fi

        # Double-check file wasn't created while we waited for lock
        if [[ -f "$ANALYTICS_FILE" ]] && jq -e . "$ANALYTICS_FILE" >/dev/null 2>&1; then
            [[ "$lock_acquired" == "true" ]] && rmdir "$lock_file" 2>/dev/null
            return 0
        fi

        # Remove corrupted file if it exists
        [[ -f "$ANALYTICS_FILE" ]] && rm -f "$ANALYTICS_FILE"

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
        # Release lock
        [[ "$lock_acquired" == "true" ]] && rmdir "$lock_file" 2>/dev/null
    fi
}

# Record AI performance with file locking
record_ai_performance() {
    local ai_name="$1"
    local success="$2"  # true/false
    local response_time="$3"
    local is_winner="${4:-false}"  # true/false for race mode

    # Ensure response_time is numeric (default to 0 if not)
    if ! [[ "$response_time" =~ ^[0-9]+$ ]]; then
        response_time="0"
    fi

    init_analytics

    # Use file locking to prevent race conditions
    local lock_file="${ANALYTICS_FILE}.lock"
    local lock_acquired=false

    # Try to acquire lock with timeout (50 * 0.1s = 5 seconds)
    local max_wait=50
    local waited=0
    while [[ $waited -lt $max_wait ]]; do
        if mkdir "$lock_file" 2>/dev/null; then
            lock_acquired=true
            break
        fi
        sleep 0.1
        waited=$((waited + 1))
    done

    if [[ "$lock_acquired" != "true" ]]; then
        log_debug "Failed to acquire analytics lock after 5s"
        return 1
    fi

    # Ensure lock is released on exit (variable captured at trap definition time)
    # shellcheck disable=SC2064
    trap "rmdir '$lock_file' 2>/dev/null" RETURN

    # Read current analytics
    local analytics timestamp
    analytics=$(cat "$ANALYTICS_FILE")
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

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
        # Pipe analytics through stdin to avoid shell quoting issues
        analytics=$(echo "$analytics" | python3 -c "
import json
import sys

data = json.load(sys.stdin)
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
    local recent_failures_count=0
    if command_exists jq; then
        recent_failures_count=$(echo "$analytics" | jq -r --arg ai "$ai_name" '.[$ai].recent_failures | length' 2>/dev/null || echo "0")
    fi
    if [[ "$recent_failures_count" -ge 3 ]]; then
        local disable_until
        disable_until=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ")
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
        local current_time
        current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        if [[ "$current_time" < "$disabled_until" ]]; then
            return 0  # AI is disabled
        else
            # Re-enable AI
            if command_exists jq; then
                local analytics
                analytics=$(jq --arg ai "$ai_name" '.[$ai].disabled_until = null | .[$ai].recent_failures = []' "$ANALYTICS_FILE")
                echo "$analytics" > "$ANALYTICS_FILE"
            fi
        fi
    fi

    return 1  # AI is not disabled
}

# Get best performing AI
get_best_ai() {
    init_analytics

    local claude_disabled gemini_disabled
    claude_disabled=$(is_ai_disabled "claude" && echo "true" || echo "false")
    gemini_disabled=$(is_ai_disabled "gemini" && echo "true" || echo "false")

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

    # Calculate performance scores and check for draw
    if command_exists jq; then
        jq -r '
            if .claude.total_calls == 0 and .gemini.total_calls == 0 then
                "claude"
            else
                def calc_success_rate(ai):
                    if ai.total_calls > 0 then (ai.successful_calls / ai.total_calls * 100) | floor else 0 end;
                def calc_avg_time(ai):
                    if ai.successful_calls > 0 then (ai.total_time / ai.successful_calls) | floor else 0 end;

                # Calculate metrics for both
                (.claude | {
                    name: "claude",
                    success_rate: calc_success_rate(.),
                    avg_time: calc_avg_time(.),
                    wins: .wins
                }) as $claude_stats |
                (.gemini | {
                    name: "gemini",
                    success_rate: calc_success_rate(.),
                    avg_time: calc_avg_time(.),
                    wins: .wins
                }) as $gemini_stats |

                # Check for draw condition
                if $claude_stats.success_rate == $gemini_stats.success_rate and
                   $claude_stats.avg_time == $gemini_stats.avg_time then
                    "draw"
                # Otherwise compare scores
                else
                    [$claude_stats, $gemini_stats] |
                    map({
                        name: .name,
                        score: (.success_rate + (if .avg_time > 0 then (50 - .avg_time) else 0 end))
                    }) |
                    max_by(.score) |
                    .name
                end
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
        local avg_time
        avg_time=$(jq -r --arg ai "$ai_name" '
            if .[$ai].successful_calls > 0 then
                (.[$ai].total_time / .[$ai].successful_calls)
            else
                0
            end
        ' "$ANALYTICS_FILE")

        if [[ "$avg_time" != "0" ]]; then
            # Set timeout to 1.5x average + 5s buffer
            local timeout
            timeout=$(echo "$avg_time * 1.5 + 5" | bc 2>/dev/null || python3 -c "print(int($avg_time * 1.5 + 5))")
            # Convert to integer for comparison
            timeout=${timeout%.*}
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

# Display brief statistics summary
show_brief_stats() {
    init_analytics

    if command_exists jq; then
        echo ""
        echo "📊 Quick Stats:"

        # Claude stats
        local claude_stats
        claude_stats=$(jq -r '
            .claude |
            "  Claude: " +
            (if .total_calls > 0 then
                ((.successful_calls / .total_calls * 100) | floor | tostring) + "% success (" +
                (.successful_calls | tostring) + "/" + (.total_calls | tostring) + "), "
            else
                "No calls, "
            end) +
            "Avg: " +
            (if .successful_calls > 0 then
                ((.total_time / .successful_calls) | floor | tostring) + "s"
            else
                "N/A"
            end) +
            ", Wins: " + (.wins | tostring) +
            (if .disabled_until then " ⚠️ DISABLED" else "" end)
        ' "$ANALYTICS_FILE")

        # Gemini stats
        local gemini_stats
        gemini_stats=$(jq -r '
            .gemini |
            "  Gemini: " +
            (if .total_calls > 0 then
                ((.successful_calls / .total_calls * 100) | floor | tostring) + "% success (" +
                (.successful_calls | tostring) + "/" + (.total_calls | tostring) + "), "
            else
                "No calls, "
            end) +
            "Avg: " +
            (if .successful_calls > 0 then
                ((.total_time / .successful_calls) | floor | tostring) + "s"
            else
                "N/A"
            end) +
            ", Wins: " + (.wins | tostring) +
            (if .disabled_until then " ⚠️ DISABLED" else "" end)
        ' "$ANALYTICS_FILE")

        echo "$claude_stats"
        echo "$gemini_stats"

        # Best performer
        local best
        best=$(get_best_ai)
        if [[ -n "$best" ]]; then
            if [[ "$best" == "draw" ]]; then
                # Check who has more wins in draw scenario
                local claude_wins gemini_wins
                claude_wins=$(jq -r '.claude.wins' "$ANALYTICS_FILE")
                gemini_wins=$(jq -r '.gemini.wins' "$ANALYTICS_FILE")
                if [[ "$claude_wins" -gt "$gemini_wins" ]]; then
                    echo "  Best performer: Draw (Claude leads with $claude_wins wins)"
                elif [[ "$gemini_wins" -gt "$claude_wins" ]]; then
                    echo "  Best performer: Draw (Gemini leads with $gemini_wins wins)"
                else
                    echo "  Best performer: Draw"
                fi
            else
                echo "  Best performer: ${best^}"
            fi
        fi
        echo ""
    elif command_exists python3; then
        python3 -c "
import json
try:
    with open('$ANALYTICS_FILE') as f:
        data = json.load(f)

    print()
    print('📊 Quick Stats:')

    for ai in ['claude', 'gemini']:
        stats = data[ai]
        if stats['total_calls'] > 0:
            success_rate = int(stats['successful_calls'] / stats['total_calls'] * 100)
            success_str = f\"{success_rate}% success ({stats['successful_calls']}/{stats['total_calls']})\"
        else:
            success_str = 'No calls'

        if stats['successful_calls'] > 0:
            avg_time = int(stats['total_time'] / stats['successful_calls'])
            avg_str = f'{avg_time}s'
        else:
            avg_str = 'N/A'

        disabled = ' ⚠️ DISABLED' if stats.get('disabled_until') else ''
        print(f\"  {ai.capitalize()}: {success_str}, Avg: {avg_str}, Wins: {stats['wins']}{disabled}\")

    print()
except:
    pass
"
    fi
}

# -----------------------------------------------------------------------------
# Maintenance Branch Detection
# -----------------------------------------------------------------------------

# Check if current branch is a maintenance branch pattern
# Matches: *-13.1AV, *-13-1av, *-20AV, *-25.2av (case insensitive)
# Supports any version number with optional dot or dash separators
is_maintenance_branch() {
   local branch
   branch=$(get_current_branch)
   # Match patterns like *-13.1AV, *-20AV, *-25.2av (any version number)
   if [[ "${branch,,}" =~ -[0-9]+([.-][0-9]+)?av$ ]]; then
      return 0
   fi
   return 1
}

# Check if current branch follows JIRA pattern (XXX-1234)
# Matches: VIS-1234, TDT-5678, ABC-123, etc. (2+ uppercase letters + dash + numbers)
is_jira_branch() {
   local branch
   branch=$(get_current_branch)
   # Match pattern: 2+ uppercase letters, dash, 1+ digits (at branch start)
   if [[ "$branch" =~ ^[A-Z]{2,}-[0-9]+ ]]; then
      return 0
   fi
   return 1
}

# Lookup commit message from Commits.md by JIRA tag
# Supports formats: "JIRA: VIS-1234", "JIRA: #VIS-1234", or standalone "VIS-1234" at line start
# Configure path via COMMITS_FILE environment variable
lookup_commit_from_history() {
   local jira_tag="$1"
   local commits_file="${COMMITS_FILE:-$HOME/Dev/Private/Commits.md}"

   if [[ ! -f "$commits_file" ]] || [[ ! -r "$commits_file" ]]; then
      return 1
   fi

   # Extract base JIRA (without version suffix) e.g., VIS-1234 from VIS-1234-something-13.1AV
   local base_jira
   base_jira=$(echo "$jira_tag" | grep -oE '^[A-Z]+-[0-9]+')

   if [[ -z "$base_jira" ]]; then
      return 1
   fi

   # Search for JIRA tag at line start - use tac to find the LAST (most recent) match
   # Filter out metadata lines like "- Commit branch HASH" and "- Commit HASH"
   local result
   result=$(tac "$commits_file" | grep -m 1 -B 15 "^${base_jira}" | tac | grep -A 5 "^Logs:" | grep "^- " | grep -v "Commit.*HASH" | head -10)

   if [[ -n "$result" ]]; then
      echo "$result"
      return 0
   fi
   return 1
}

# -----------------------------------------------------------------------------
# Credential Detection
# -----------------------------------------------------------------------------

# Common credential patterns (regex for grep -E)
# Each pattern is designed to minimize false positives while catching real secrets
CREDENTIAL_PATTERNS=(
    # AWS
    'AKIA[0-9A-Z]{16}'                                    # AWS Access Key ID
    'aws_secret_access_key\s*=\s*[A-Za-z0-9/+=]{40}'      # AWS Secret Key

    # OpenAI
    'sk-[a-zA-Z0-9]{48}'                                  # OpenAI API Key
    'sk-proj-[a-zA-Z0-9_-]{80,}'                          # OpenAI Project Key

    # Anthropic
    'sk-ant-[a-zA-Z0-9_-]{90,}'                           # Anthropic API Key

    # Google
    'AIza[0-9A-Za-z_-]{35}'                               # Google API Key

    # GitHub
    'ghp_[a-zA-Z0-9]{36}'                                 # GitHub Personal Access Token
    'gho_[a-zA-Z0-9]{36}'                                 # GitHub OAuth Token
    'ghr_[a-zA-Z0-9]{36}'                                 # GitHub Refresh Token
    'ghs_[a-zA-Z0-9]{36}'                                 # GitHub Server Token
    'github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}'          # GitHub Fine-grained PAT

    # GitLab
    'glpat-[a-zA-Z0-9_-]{20,}'                            # GitLab Personal Token

    # Atlassian/JIRA
    'ATATT3x[a-zA-Z0-9_-]{100,}'                          # Atlassian API Token

    # Slack
    'xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*'   # Slack Token

    # Stripe
    'sk_live_[0-9a-zA-Z]{24}'                             # Stripe Live Secret Key
    'rk_live_[0-9a-zA-Z]{24}'                             # Stripe Live Restricted Key

    # Private Keys
    '-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----'
    '-----BEGIN PGP PRIVATE KEY BLOCK-----'

    # Database Connection Strings (with embedded passwords)
    'mongodb(\+srv)?://[^:]+:[^@]+@'                      # MongoDB
    'postgres(ql)?://[^:]+:[^@]+@'                        # PostgreSQL
    'mysql://[^:]+:[^@]+@'                                # MySQL

    # Generic patterns (more restrictive to reduce false positives)
    'password\s*[:=]\s*["\x27][^"\x27]{8,64}["\x27]'      # password = "..."
    'api[_-]?key\s*[:=]\s*["\x27][a-zA-Z0-9_-]{20,}["\x27]' # api_key = "..."
    'secret\s*[:=]\s*["\x27][^"\x27]{8,64}["\x27]'        # secret = "..."
    'Bearer\s+[a-zA-Z0-9_-]{20,}'                         # Bearer tokens
)

# File patterns to skip (allowlist)
CREDENTIAL_SKIP_PATTERNS=(
    '\.example$'
    '\.template$'
    '\.sample$'
    '\.md$'
    '\.lock$'
    '\.min\.js$'
    'package-lock\.json$'
    'yarn\.lock$'
    'go\.sum$'
    'Cargo\.lock$'
    # Skip git hooks themselves (contain pattern definitions)
    '\.git-template/hooks/'
    '\.githooks/'
    'hooks-common\.sh$'
    'pre-commit$'
    # Skip gitleaks config
    '\.gitleaks\.toml$'
)

# Scan staged files for credentials
# Returns: 0 if no findings, 1 if findings detected
# Output: Formatted warnings to stderr
scan_staged_for_credentials() {
    local findings=()
    local finding_count=0

    # Get list of staged files (only added/modified, not deleted)
    local staged_files
    staged_files=$(git diff --cached --name-only --diff-filter=AM 2>/dev/null)

    if [[ -z "$staged_files" ]]; then
        return 0
    fi

    # Build combined regex pattern for faster matching
    local combined_pattern=""
    for pattern in "${CREDENTIAL_PATTERNS[@]}"; do
        if [[ -n "$combined_pattern" ]]; then
            combined_pattern+="|"
        fi
        combined_pattern+="($pattern)"
    done

    # Process each file
    while IFS= read -r file; do
        # Skip if file matches allowlist patterns
        local skip=false
        for pattern in "${CREDENTIAL_SKIP_PATTERNS[@]}"; do
            if [[ "$file" =~ $pattern ]]; then
                skip=true
                break
            fi
        done
        [[ "$skip" == "true" ]] && continue

        # Get staged content and scan with grep -n for line numbers
        local matches
        matches=$(git show ":$file" 2>/dev/null | grep -nE "$combined_pattern" 2>/dev/null | head -20)

        if [[ -n "$matches" ]]; then
            while IFS= read -r match; do
                local line_num line_content masked_line
                line_num=$(echo "$match" | cut -d: -f1)
                line_content=$(echo "$match" | cut -d: -f2-)
                # Mask the sensitive part for display
                masked_line=$(echo "$line_content" | sed -E 's/([a-zA-Z0-9_-]{8})[a-zA-Z0-9_-]{10,}/\1***/g')
                findings+=("  File: $file:$line_num")
                findings+=("    > ${masked_line:0:100}")
                ((finding_count++))
            done <<< "$matches"
        fi
    done <<< "$staged_files"

    # Display findings if any
    if [[ $finding_count -gt 0 ]]; then
        format_credential_warning "${findings[@]}"
        return 1
    fi

    return 0
}

# Run gitleaks scan if available
# Returns: 0 if no findings or gitleaks not installed, 1 if findings
run_gitleaks_scan() {
    if ! command_exists gitleaks; then
        log_debug "gitleaks not installed, skipping"
        return 0
    fi

    local config_file="${HOOKS_DIR}/../.gitleaks.toml"
    local gitleaks_args=("git" "--staged" "--no-banner" "--exit-code" "1" "--verbose")

    if [[ -f "$config_file" ]]; then
        gitleaks_args+=("--config" "$config_file")
    fi

    # Run gitleaks and capture output
    local output
    output=$(gitleaks "${gitleaks_args[@]}" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 1 ]]; then
        echo "" >&2
        echo "============================================================" >&2
        echo "  GITLEAKS: Potential secrets detected" >&2
        echo "============================================================" >&2

        # Parse and display findings with masked secrets
        # Format: Finding: RuleID, Secret: xxx, File: path, Line: N
        echo "$output" | while IFS= read -r line; do
            # Skip info/warning lines, show finding details
            if [[ "$line" =~ ^Finding: ]] || [[ "$line" =~ ^Secret: ]] || \
               [[ "$line" =~ ^File: ]] || [[ "$line" =~ ^Line: ]] || \
               [[ "$line" =~ ^RuleID: ]] || [[ "$line" =~ ^Match: ]]; then
                # Mask secrets in output (show first 8 chars + ***)
                local masked_line
                masked_line=$(echo "$line" | sed -E 's/(Secret:\s*[a-zA-Z0-9_-]{8})[a-zA-Z0-9_-]{10,}/\1***/g')
                echo "  $masked_line" >&2
            elif [[ "$line" =~ leaks\ found ]]; then
                echo "" >&2
                echo "  $line" >&2
            fi
        done

        echo "------------------------------------------------------------" >&2
        return 1
    fi

    return 0
}

# Format credential warning output
format_credential_warning() {
    local findings=("$@")

    echo "" >&2
    echo "==============================================================" >&2
    echo "  CREDENTIAL WARNING - Potential secrets detected" >&2
    echo "==============================================================" >&2
    echo "" >&2

    for finding in "${findings[@]}"; do
        echo "$finding" >&2
    done

    echo "" >&2
    echo "--------------------------------------------------------------" >&2
    echo "  Tip: Use environment variables or a secrets manager" >&2
    echo "  To skip once: git commit --no-verify" >&2
    echo "==============================================================" >&2
}

# Main credential scan entry point
# Runs both gitleaks (if available) and custom patterns
run_credential_scan() {
    local enable_scan use_gitleaks
    enable_scan=$(git config --local hooks.enableCredentialScan 2>/dev/null || echo "true")
    use_gitleaks=$(git config --local hooks.useGitleaks 2>/dev/null || echo "true")

    if [[ "$enable_scan" != "true" ]]; then
        log_debug "Credential scanning disabled"
        return 0
    fi

    local has_findings=false

    # Run gitleaks first if enabled and available
    if [[ "$use_gitleaks" == "true" ]]; then
        if ! run_gitleaks_scan; then
            has_findings=true
        fi
    fi

    # Always run custom patterns (catches things gitleaks might miss)
    if ! scan_staged_for_credentials; then
        has_findings=true
    fi

    # Return status (but don't block - warning mode)
    if [[ "$has_findings" == "true" ]]; then
        return 1
    fi

    return 0
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
export -f command_exists validate_commands validate_safe_path validate_ai_command
export -f shell_escape trim
export -f init_analytics record_ai_performance get_ai_stats is_ai_disabled get_best_ai get_adaptive_timeout show_brief_stats
export -f is_maintenance_branch is_jira_branch lookup_commit_from_history
export -f scan_staged_for_credentials run_gitleaks_scan format_credential_warning run_credential_scan
