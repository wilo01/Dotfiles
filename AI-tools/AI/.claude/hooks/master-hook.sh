#!/bin/bash

# Master Hook Coordinator
# Intelligently routes to appropriate specialized hooks based on context

LOG_FILE="${HOME}/.Dotfiles/.claude/hooks/master.log"
HOOKS_DIR="${HOME}/.Dotfiles/.claude/hooks"
PWD_DIR=$(pwd)
PROMPT_TEXT="$1"

# Log master hook execution
echo "$(date): Master hook triggered in $PWD_DIR" >> "$LOG_FILE"

# Determine which hooks to run based on context
determine_active_hooks() {
    local hooks=()
    
    # Always run project context hook
    hooks+=("project-context-hook.sh")
    
    # Task Master projects
    if [[ -d ".taskmaster" ]]; then
        hooks+=("taskmaster-workflow-hook.sh")
    fi
    
    # Git repositories
    if [[ -d ".git" ]]; then
        hooks+=("git-workflow-hook.sh")
    fi
    
    # Prompt-based hook selection
    if echo "$PROMPT_TEXT" | grep -qi "git\|commit\|branch\|push\|pull"; then
        if [[ ! " ${hooks[@]} " =~ " git-workflow-hook.sh " ]]; then
            hooks+=("git-workflow-hook.sh")
        fi
    fi
    
    if echo "$PROMPT_TEXT" | grep -qi "task\|todo\|next\|tm/"; then
        if [[ ! " ${hooks[@]} " =~ " taskmaster-workflow-hook.sh " ]]; then
            hooks+=("taskmaster-workflow-hook.sh")
        fi
    fi
    
    # Agent completion detection
    if echo "$PROMPT_TEXT" | grep -qi -E "(analysis complete|review complete|task complete|scope complete|analysis summary|testing effort.*hours|estimated.*testing|success criteria|code quality|security.*issues|root cause|solution.*implemented|query.*results|data.*insights)"; then
        hooks+=("agent-completion-hook.sh")
    fi

    printf '%s\n' "${hooks[@]}"
}

# Execute hooks with error handling
execute_hook() {
    local hook_name="$1"
    local hook_path="$HOOKS_DIR/$hook_name"
    
    if [[ -x "$hook_path" ]]; then
        echo "🔗 Running $hook_name..."
        "$hook_path" "$PROMPT_TEXT" 2>/dev/null || {
            echo "$(date): Hook $hook_name failed" >> "$LOG_FILE"
        }
        echo
    else
        echo "$(date): Hook $hook_name not found or not executable" >> "$LOG_FILE"
    fi
}

# Main execution
echo "🤖 Claude Code Context Assistant"
echo "================================"

# Get and execute appropriate hooks
ACTIVE_HOOKS=($(determine_active_hooks))

for hook in "${ACTIVE_HOOKS[@]}"; do
    execute_hook "$hook"
done

# Log execution summary
echo "$(date): Executed hooks: ${ACTIVE_HOOKS[*]}" >> "$LOG_FILE"

# Provide general tips based on prompt analysis
if echo "$PROMPT_TEXT" | grep -qi "help\|guide\|how.*to"; then
    echo "💡 General Tips:"
    echo "• Use /tm/ slash commands for Task Master workflows"
    echo "• Check CLAUDE.md for project-specific context"
    echo "• Use /clear between different tasks for focus"
fi
