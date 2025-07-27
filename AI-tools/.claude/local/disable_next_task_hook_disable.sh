#!/bin/bash
# Disable the next-task hook by creating the disable flag

DISABLE_FILE="$HOME/.claude/local/disable_next_task_hook"

touch "$DISABLE_FILE"
echo "🚫 Next-task hook DISABLED - will NOT activate /tm:next-task on stop"