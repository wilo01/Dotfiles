#!/bin/bash

# Task Master Workflow Hook
# Automatically manages Task Master workflow integration

LOG_FILE="${HOME}/.Dotfiles/.claude/hooks/taskmaster.log"
PWD_DIR=$(pwd)

# Check if this is a Task Master project
if [[ ! -d ".taskmaster" ]]; then
    exit 0
fi

PROMPT_TEXT="$1"

# Log hook execution
echo "$(date): TaskMaster hook in $PWD_DIR" >> "$LOG_FILE"

# Task Master command detection and automation
detect_tm_intent() {
    local prompt="$1"
    
    if echo "$prompt" | grep -qi "next task\|what.*next\|ready.*work"; then
        echo "next-task"
    elif echo "$prompt" | grep -qi "complete.*task\|finish.*task\|done.*task"; then
        echo "complete-task"
    elif echo "$prompt" | grep -qi "show.*task\|task.*detail\|task.*info"; then
        echo "show-task"
    elif echo "$prompt" | grep -qi "add.*task\|new.*task\|create.*task"; then
        echo "add-task"
    elif echo "$prompt" | grep -qi "list.*task\|all.*task\|task.*status"; then
        echo "list-tasks"
    elif echo "$prompt" | grep -qi "expand.*task\|break.*down\|subtask"; then
        echo "expand-task"
    else
        echo "general"
    fi
}

# Provide Task Master suggestions
provide_tm_suggestions() {
    local intent="$1"
    
    case "$intent" in
        "next-task")
            echo "🎯 Suggested workflow:"
            echo "1. /tm/next - Get next available task"
            echo "2. Review task details and requirements"
            echo "3. Start implementation"
            ;;
        "complete-task")
            echo "✅ Task completion workflow:"
            echo "1. Verify implementation is complete"
            echo "2. Run tests if applicable"
            echo "3. /tm/done <task-id> - Mark as done"
            ;;
        "show-task")
            echo "📋 Task information:"
            echo "Use: /tm/show <task-id> for detailed view"
            echo "Or: task-master show <id> for full details"
            ;;
        "add-task")
            echo "➕ Adding new tasks:"
            echo "Use: task-master add-task --prompt='description'"
            echo "Add --research flag for AI-enhanced task creation"
            ;;
        "list-tasks")
            echo "📝 Task overview:"
            echo "Use: /tm/list for quick overview"
            echo "Or: task-master list --with-subtasks for detailed view"
            ;;
        "expand-task")
            echo "🔄 Task expansion:"
            echo "Use: task-master expand --id=<task-id>"
            echo "Add --research flag for better subtask generation"
            ;;
        "general")
            # Check current project status
            if command -v task-master >/dev/null 2>&1; then
                NEXT_TASK=$(task-master next 2>/dev/null | grep -o "Task [0-9.]*" | head -1)
                if [[ -n "$NEXT_TASK" ]]; then
                    echo "🎯 $NEXT_TASK is ready to work on"
                    echo "Use: /tm/next to get started"
                fi
            fi
            ;;
    esac
}

# Main execution
INTENT=$(detect_tm_intent "$PROMPT_TEXT")
provide_tm_suggestions "$INTENT"

# Log the intent detection
echo "$(date): Detected intent: $INTENT" >> "$LOG_FILE"