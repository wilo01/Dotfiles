#!/bin/bash
# Check the status of the next-task hook

DISABLE_FILE="$HOME/.claude/local/disable_next_task_hook"

if [ -f "$DISABLE_FILE" ]; then
    echo "🚫 Next-task hook is DISABLED"
else
    echo "✅ Next-task hook is ENABLED"
fi