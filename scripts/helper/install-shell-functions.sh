#!/bin/bash
# Shell functions for helper CLI that enable "paste to terminal" functionality
# Add this to your .bashrc or .zshrc:
# source /path/to/install-shell-functions.sh

# For Bash
if [ -n "$BASH_VERSION" ]; then
    # Function that prints command and waits for Enter
    helper-jira() {
        local cmd=$(helper jira "$@" --exec)
        # Use read with -e to enable readline (allows editing)
        read -e -i "$cmd" -p "" executed_cmd
        # Execute what the user confirmed/edited
        eval "$executed_cmd"
    }
    
    helper-stash() {
        local cmd=$(helper stash "$@" --exec)
        # Use read with -e to enable readline (allows editing)
        read -e -i "$cmd" -p "" executed_cmd
        # Execute what the user confirmed/edited
        eval "$executed_cmd"
    }
fi

# For Zsh
if [ -n "$ZSH_VERSION" ]; then
    # Zsh version - uses vared for line editing
    helper-jira() {
        local cmd=$(helper jira "$@" --exec)
        # Put command in buffer and let user edit/execute
        print -z "$cmd"
    }
    
    helper-stash() {
        local cmd=$(helper stash "$@" --exec)
        # Put command in buffer and let user edit/execute
        print -z "$cmd"
    }
fi

# Alternative: Widget-based approach for Zsh (more advanced)
if [ -n "$ZSH_VERSION" ]; then
    # Define a widget that gets JIRA branch command
    helper-jira-widget() {
        local text="${BUFFER}"
        if [ -z "$text" ]; then
            echo "Usage: Type ticket text, then press the bound key"
            return
        fi
        BUFFER=$(helper jira "$text" --exec)
        zle end-of-line
    }
    
    # Register the widget (uncomment to use)
    # zle -N helper-jira-widget
    # bindkey '^J' helper-jira-widget  # Ctrl+J to convert current line to JIRA command
fi

echo "Helper shell functions installed!"
echo ""
echo "For Bash: Commands will appear in readline for editing"
echo "For Zsh: Commands will appear in the input buffer"
echo ""
echo "Usage:"
echo "  helper-jira 'VIS-5280 Add feature'  # Command appears, press Enter to run"
echo "  helper-stash 'Work in progress'     # Command appears, press Enter to run"