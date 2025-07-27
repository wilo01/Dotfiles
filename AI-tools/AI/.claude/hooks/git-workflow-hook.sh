#!/bin/bash

# Git Workflow Hook
# Provides contextual git guidance based on repository state

LOG_FILE="${HOME}/.Dotfiles/.claude/hooks/git-workflow.log"
PWD_DIR=$(pwd)

# Check if this is a git repository
if [[ ! -d ".git" ]]; then
    exit 0
fi

PROMPT_TEXT="$1"

# Log hook execution
echo "$(date): Git workflow hook in $PWD_DIR" >> "$LOG_FILE"

# Get current git status
get_git_status() {
    local status_output=""
    
    # Branch info
    local current_branch=$(git branch --show-current 2>/dev/null)
    local main_branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@')
    
    # Status checks
    local staged_files=$(git diff --cached --name-only 2>/dev/null | wc -l)
    local modified_files=$(git diff --name-only 2>/dev/null | wc -l)
    local untracked_files=$(git ls-files --others --exclude-standard 2>/dev/null | wc -l)
    
    echo "📍 Branch: $current_branch"
    
    if [[ $staged_files -gt 0 ]]; then
        echo "📦 $staged_files staged files ready to commit"
    fi
    
    if [[ $modified_files -gt 0 ]]; then
        echo "📝 $modified_files modified files"
    fi
    
    if [[ $untracked_files -gt 0 ]]; then
        echo "❓ $untracked_files untracked files"
    fi
    
    # Suggest actions based on state
    if [[ $staged_files -gt 0 ]]; then
        echo "💡 Ready to commit: git commit -m 'message'"
    elif [[ $modified_files -gt 0 ]] || [[ $untracked_files -gt 0 ]]; then
        echo "💡 Stage changes: git add . or git add <file>"
    fi
}

# Detect git-related intent
detect_git_intent() {
    local prompt="$1"
    
    if echo "$prompt" | grep -qi "commit\|check.*in"; then
        echo "commit"
    elif echo "$prompt" | grep -qi "push\|upload\|remote"; then
        echo "push"
    elif echo "$prompt" | grep -qi "pull\|fetch\|update"; then
        echo "pull"
    elif echo "$prompt" | grep -qi "branch\|checkout\|switch"; then
        echo "branch"
    elif echo "$prompt" | grep -qi "merge\|rebase"; then
        echo "merge"
    elif echo "$prompt" | grep -qi "status\|what.*changed"; then
        echo "status"
    elif echo "$prompt" | grep -qi "diff\|changes\|compare"; then
        echo "diff"
    elif echo "$prompt" | grep -qi "stash\|save.*work"; then
        echo "stash"
    else
        echo "general"
    fi
}

# Provide git workflow guidance
provide_git_guidance() {
    local intent="$1"
    
    case "$intent" in
        "commit")
            echo "📝 Commit workflow:"
            echo "1. Review changes: git diff"
            echo "2. Stage files: git add ."
            echo "3. Commit: git commit -m 'type: description'"
            echo "4. Conventional commits: feat:, fix:, docs:, style:, refactor:, test:, chore:"
            ;;
        "push")
            echo "🚀 Push workflow:"
            echo "1. Ensure commits are ready: git log --oneline -3"
            echo "2. Push: git push origin <branch>"
            echo "3. First push: git push -u origin <branch>"
            ;;
        "pull")
            echo "⬇️  Pull/Update workflow:"
            echo "1. Check status: git status"
            echo "2. Stash if needed: git stash"
            echo "3. Pull: git pull origin <branch>"
            echo "4. Apply stash: git stash pop"
            ;;
        "branch")
            echo "🌿 Branch workflow:"
            echo "1. Create: git checkout -b feature/name"
            echo "2. Switch: git checkout <branch>"
            echo "3. List: git branch -a"
            ;;
        "status"|"general")
            get_git_status
            ;;
    esac
}

# Main execution
INTENT=$(detect_git_intent "$PROMPT_TEXT")
provide_git_guidance "$INTENT"

# Log the intent detection
echo "$(date): Detected git intent: $INTENT" >> "$LOG_FILE"