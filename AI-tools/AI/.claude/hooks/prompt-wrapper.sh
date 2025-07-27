#!/bin/bash

# Alternative prompt capture wrapper
# This tries to capture prompts through environment variables or other methods

LOG_FILE="${HOME}/.Dotfiles/.claude/prompts/capture.log"

# Log all environment variables that might contain the prompt
{
    echo "$(date): === Hook Triggered ==="
    echo "Arguments: $*"
    echo "STDIN available: $([[ -t 0 ]] && echo "no" || echo "yes")"
    echo "Environment variables:"
    env | grep -i claude || echo "No claude env vars"
    env | grep -i prompt || echo "No prompt env vars"
    echo "=========================="
} >> "$LOG_FILE"

# Try the original capture method
exec /home/dariuszw/.Dotfiles/.claude/hooks/prompt-capture.sh "$@"