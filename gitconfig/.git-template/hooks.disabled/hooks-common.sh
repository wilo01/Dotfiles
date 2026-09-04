#!/bin/bash
# Git Hooks Common Library
# Shared functions and utilities for git hooks

# -----------------------------------------------------------------------------
# Configuration Management
# -----------------------------------------------------------------------------

# Load hook configuration from git config with defaults
load_hook_config() {
   # AI Configuration
   AI_MAX_TIMEOUT=$(git config --local hooks.aiMaxTimeout || echo "${AI_MAX_TIMEOUT:-120}")
   AI_INACTIVITY_TIMEOUT=$(git config --local hooks.aiInactivityTimeout || echo "${AI_INACTIVITY_TIMEOUT:-60}")
   AI_SHOW_PROGRESS=$(git config --local hooks.aiShowProgress || echo "${AI_SHOW_PROGRESS:-true}")
   AI_DEBUG=$(git config --local hooks.aiDebug || echo "${AI_DEBUG:-false}")

   # Hook Settings
   ENABLE_GLOBAL_HOOKS=$(git config --local hooks.enableGlobalHooks || echo "true")
   ENABLE_LOCAL_HOOKS=$(git config --local hooks.enableLocalHooks || echo "false")
   ENABLE_AI_COMMIT=$(git config --local hooks.enableAiCommit || echo "true")

   # Maintenance branch prompt timeout (seconds)
   MAINTENANCE_TIMEOUT=$(git config --local hooks.maintenanceTimeout || echo "${MAINTENANCE_TIMEOUT:-10}")

   # AI Command Path - auto-detect with fallback chain
   # Prefer opencode-commit (free cloud model) over ollama-commit (local)
   AI_LLM_CMD=$(git config --local hooks.aiLlmCmd 2>/dev/null)
   if [[ -z "$AI_LLM_CMD" ]]; then
      if [[ -x "$HOME/.Dotfiles/bin/opencode-commit" ]]; then
         AI_LLM_CMD="$HOME/.Dotfiles/bin/opencode-commit"
      elif command -v opencode-commit >/dev/null 2>&1; then
         AI_LLM_CMD=$(command -v opencode-commit)
      elif [[ -x "$HOME/.Dotfiles/bin/ollama-commit" ]]; then
         AI_LLM_CMD="$HOME/.Dotfiles/bin/ollama-commit"
      elif command -v ollama-commit >/dev/null 2>&1; then
         AI_LLM_CMD=$(command -v ollama-commit)
      else
         AI_LLM_CMD="$HOME/.Dotfiles/bin/ollama-commit"
      fi
   fi

   # AI Model selection - default depends on which backend is active
   AI_LLM_MODEL=$(git config --local hooks.aiModel 2>/dev/null)
   if [[ -z "$AI_LLM_MODEL" ]]; then
      case "$(basename "$AI_LLM_CMD")" in
         opencode-commit)
            AI_LLM_MODEL="${OPENCODE_MODEL:-opencode/deepseek-v4-flash-free}"
            ;;
         *)
            AI_LLM_MODEL="${OLLAMA_MODEL:-qwen3:1.7b}"
            ;;
      esac
   fi

   # File Paths
   HOOKS_LOCAL_PATH=$(git config --local hooks.hooksLocalPath | sed "s|^~|$HOME|")
   HOOKS_LOCAL_FILENAME=$(git config --local hooks.hooksLocalFilename)

   export AI_MAX_TIMEOUT AI_INACTIVITY_TIMEOUT AI_SHOW_PROGRESS AI_DEBUG
   export ENABLE_GLOBAL_HOOKS ENABLE_LOCAL_HOOKS ENABLE_AI_COMMIT MAINTENANCE_TIMEOUT
   export AI_LLM_CMD AI_LLM_MODEL
   export HOOKS_LOCAL_PATH HOOKS_LOCAL_FILENAME
}

# Debug configuration if enabled
debug_config() {
   if [[ "$AI_DEBUG" == "true" ]]; then
      log_info "🔧 AI Hook Configuration:"
      log_info "  Backend: $(basename "$AI_LLM_CMD")"
      log_info "  Model: $AI_LLM_MODEL"
      log_info "  Max Timeout: ${AI_MAX_TIMEOUT}s"
      log_info "  Inactivity Timeout: ${AI_INACTIVITY_TIMEOUT}s"
      log_info "  Show Progress: $AI_SHOW_PROGRESS"
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
export -f get_current_branch get_jira_tag get_project_root has_staged_changes get_diff_stats
export -f safe_write_file safe_read_file
export -f command_exists validate_commands
export -f shell_escape trim
export -f is_maintenance_branch is_jira_branch lookup_commit_from_history
export -f scan_staged_for_credentials run_gitleaks_scan format_credential_warning run_credential_scan
