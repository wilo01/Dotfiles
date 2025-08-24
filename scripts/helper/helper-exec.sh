#!/bin/bash
# Helper script to execute helper commands directly
# Add this to your shell config: source /path/to/helper-exec.sh

# Function to execute jira branch command
hjira() {
    if [ -z "$1" ]; then
        echo "Usage: hjira 'ticket description'"
        return 1
    fi
    
    # Get the command and execute it
    eval $(helper jira "$1" -e)
}

# Function to execute stash command
hstash() {
    if [ -z "$1" ]; then
        echo "Usage: hstash 'stash message'"
        return 1
    fi
    
    # Get the command and execute it
    eval $(helper stash "$1" -e)
}

# Alias versions (alternative approach)
alias helper-jira-exec='f() { eval $(helper jira "$1" -e); }; f'
alias helper-stash-exec='f() { eval $(helper stash "$1" -e); }; f'

echo "Helper exec functions loaded. Use:"
echo "  hjira 'ticket text'    - Execute JIRA branch command"
echo "  hstash 'stash text'    - Execute stash command"
echo ""
echo "Or use --exec flag:"
echo "  \$(helper jira 'text' -e)"
echo "  eval \$(helper stash 'text' -e)"