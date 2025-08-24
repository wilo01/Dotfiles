# Helper CLI - Terminal Paste Setup

## The Problem
You want commands to appear in your terminal as if you typed them, so you can just press Enter to execute, without using the clipboard.

## The Solution

### For Zsh Users (Recommended)

Add these functions to your `~/.zshrc`:

```bash
# Helper functions that put commands in the input buffer
helper-jira() {
    local cmd=$(helper jira "$@" --exec)
    print -z "$cmd"
}

helper-stash() {
    local cmd=$(helper stash "$@" --exec)
    print -z "$cmd"
}

# Shorter aliases
alias hj='helper-jira'
alias hs='helper-stash'
```

Then reload your shell:
```bash
source ~/.zshrc
```

#### Usage:
```bash
# Type this:
hj "VIS-5280 Add new feature"

# The command appears in your terminal:
git checkout -b VIS-5280-add-new-feature

# Just press Enter to execute!
```

### For Bash Users

Add these functions to your `~/.bashrc`:

```bash
# Helper functions with readline editing
helper-jira() {
    local cmd=$(helper jira "$@" --exec)
    read -e -i "$cmd" -p "" executed_cmd
    eval "$executed_cmd"
}

helper-stash() {
    local cmd=$(helper stash "$@" --exec)
    read -e -i "$cmd" -p "" executed_cmd
    eval "$executed_cmd"
}

# Shorter aliases
alias hj='helper-jira'
alias hs='helper-stash'
```

Then reload your shell:
```bash
source ~/.bashrc
```

#### Usage:
```bash
# Type this:
hj "VIS-5280 Add new feature"

# The command appears (you can edit it):
git checkout -b VIS-5280-add-new-feature

# Press Enter to execute or edit first then Enter
```

### For Fish Users

Add to your `~/.config/fish/config.fish`:

```fish
function helper-jira
    set cmd (helper jira $argv --exec)
    commandline -r $cmd
    commandline -f repaint
end

function helper-stash
    set cmd (helper stash $argv --exec)
    commandline -r $cmd
    commandline -f repaint
end

# Shorter aliases
alias hj='helper-jira'
alias hs='helper-stash'
```

## Advanced: Zsh Key Binding

For Zsh users who want to convert text directly with a hotkey:

```bash
# Add to ~/.zshrc
helper-jira-widget() {
    local text="${BUFFER}"
    BUFFER=$(helper jira "$text" --exec)
    zle end-of-line
}

zle -N helper-jira-widget
bindkey '^G' helper-jira-widget  # Ctrl+G to convert current line
```

Usage:
1. Type: `VIS-5280 Add feature`
2. Press `Ctrl+G`
3. Line becomes: `git checkout -b VIS-5280-add-feature`
4. Press Enter to execute

## Alternative Methods

### Method 1: Direct Execution
```bash
$(helper jira "VIS-5280" --exec)
```

### Method 2: Copy to Clipboard (Original)
```bash
helper jira "VIS-5280"  # Copies to clipboard
# Then paste with Ctrl+V or Cmd+V
```

### Method 3: Display Only
```bash
helper jira "VIS-5280" --no-copy  # Just displays the command
```

## Installation Script

Run this to automatically add functions to your shell config:

```bash
# Detect shell and add appropriate functions
if [ -n "$ZSH_VERSION" ]; then
    echo '
# Helper CLI functions
helper-jira() { print -z "$(helper jira "$@" --exec)"; }
helper-stash() { print -z "$(helper stash "$@" --exec)"; }
alias hj="helper-jira"
alias hs="helper-stash"
' >> ~/.zshrc
    echo "✅ Added to ~/.zshrc - Run: source ~/.zshrc"
    
elif [ -n "$BASH_VERSION" ]; then
    echo '
# Helper CLI functions  
helper-jira() {
    local cmd=$(helper jira "$@" --exec)
    read -e -i "$cmd" -p "" executed_cmd
    eval "$executed_cmd"
}
helper-stash() {
    local cmd=$(helper stash "$@" --exec)
    read -e -i "$cmd" -p "" executed_cmd
    eval "$executed_cmd"
}
alias hj="helper-jira"
alias hs="helper-stash"
' >> ~/.bashrc
    echo "✅ Added to ~/.bashrc - Run: source ~/.bashrc"
fi
```

## Benefits

✅ **No clipboard pollution** - Your clipboard stays untouched
✅ **Direct in terminal** - Command appears as if you typed it
✅ **Editable** - You can modify before executing
✅ **Fast** - Just type alias and press Enter
✅ **Shell-native** - Uses built-in shell features

## Troubleshooting

- **Command doesn't appear**: Make sure you sourced your shell config
- **Command executes immediately**: You're using the Bash version, which shows the command then waits for Enter
- **Zsh print -z not working**: Ensure you're in an interactive Zsh session
- **Fish commandline not working**: Update Fish to latest version