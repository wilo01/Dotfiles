#!/bin/bash

# Claude Code Prompt Capture Hook
# Automatically captures prompts when submitted to Claude Code

PROMPT_MANAGER="${HOME}/.Dotfiles/.claude/prompts/prompt-manager.sh"
LOG_FILE="${HOME}/.Dotfiles/.claude/prompts/capture.log"

# Ensure prompt manager exists
if [[ ! -x "$PROMPT_MANAGER" ]]; then
    echo "$(date): Prompt manager not found at $PROMPT_MANAGER" >> "$LOG_FILE"
    exit 0
fi

# Extract prompt from the input (try multiple methods)
PROMPT_TEXT="$1"

# If no argument, try reading from stdin
if [[ -z "$PROMPT_TEXT" ]] && [[ ! -t 0 ]]; then
    PROMPT_TEXT=$(cat)
fi

# Debug logging
echo "$(date): Hook called with args: '$*'" >> "$LOG_FILE"
echo "$(date): Hook received prompt: '${PROMPT_TEXT:0:100}...'" >> "$LOG_FILE"

# Skip if prompt is empty or too short
if [[ -z "$PROMPT_TEXT" ]] || [[ ${#PROMPT_TEXT} -lt 10 ]]; then
    exit 0
fi

# Skip common commands that shouldn't be saved
case "$PROMPT_TEXT" in
    "/help"|"/clear"|"/exit"|"help"|"exit"|"quit")
        exit 0
        ;;
    *"/agents"*|*"/models"*|*"/settings"*)
        exit 0
        ;;
esac

# Generate a title from the prompt (first line, truncated)
TITLE=$(echo "$PROMPT_TEXT" | head -1 | cut -c1-50 | sed 's/[^a-zA-Z0-9 ]//g' | xargs)
if [[ -z "$TITLE" ]]; then
    TITLE="Auto-captured prompt"
fi

# Determine tags based on content
TAGS="auto-captured"
if echo "$PROMPT_TEXT" | grep -qi "bug\|fix\|error\|issue"; then
    TAGS="$TAGS,bug-fix"
fi
if echo "$PROMPT_TEXT" | grep -qi "feature\|implement\|add\|create"; then
    TAGS="$TAGS,feature"
fi
if echo "$PROMPT_TEXT" | grep -qi "refactor\|improve\|optimize"; then
    TAGS="$TAGS,refactor"
fi
if echo "$PROMPT_TEXT" | grep -qi "test\|testing\|spec"; then
    TAGS="$TAGS,testing"
fi
if echo "$PROMPT_TEXT" | grep -qi "document\|readme\|doc"; then
    TAGS="$TAGS,documentation"
fi

# Save the prompt
"$PROMPT_MANAGER" save "$PROMPT_TEXT" "$TITLE" "$TAGS" >> "$LOG_FILE" 2>&1

# Log the capture
echo "$(date): Captured prompt - $TITLE" >> "$LOG_FILE"