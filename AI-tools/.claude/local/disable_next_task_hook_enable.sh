#!/bin/bash
# Enable the next-task hook by removing the disable flag

DISABLE_FILE="$HOME/.claude/local/disable_next_task_hook"

if [ -f "$DISABLE_FILE" ]; then
    rm "$DISABLE_FILE"
    echo "✅ Next-task hook ENABLED - will activate /tm:next-task on stop"
else
    echo "✅ Next-task hook already ENABLED"
fi