# Helper CLI - Execution Methods Guide

## Overview
The helper CLI now supports multiple ways to output and execute commands without polluting your clipboard.

## Available Options

### 1. `--exec` / `-e` Flag
Outputs the raw command for direct shell execution.

**Usage:**
```bash
# Execute directly using command substitution
$(helper jira "VIS-5280 Add feature" -e)

# Or use eval
eval $(helper stash "TAC-511 Fix issue" -e)
```

### 2. `--type` / `-t` Flag (Experimental)
Attempts to type the command directly to your terminal.

**Supported methods:**
- **Linux (X11)**: Uses `xdotool` if installed
- **macOS**: Uses AppleScript
- **Fallback**: Displays command for manual copy

**Usage:**
```bash
helper jira "VIS-5280 Add feature" --type
```

### 3. `--no-copy` Flag
Displays the command without copying to clipboard.

**Usage:**
```bash
helper jira "VIS-5280 Add feature" --no-copy
```

## Shell Integration

### Bash/Zsh Functions
Add these to your `.bashrc` or `.zshrc`:

```bash
# Execute JIRA branch command
hjira() {
    eval $(helper jira "$1" -e)
}

# Execute stash command
hstash() {
    eval $(helper stash "$1" -e)
}
```

Then use them like:
```bash
hjira "VIS-5280 Add new feature"
hstash "Work in progress"
```

### Aliases
Alternative approach using aliases:

```bash
alias hj='f() { eval $(helper jira "$1" -e); }; f'
alias hs='f() { eval $(helper stash "$1" -e); }; f'
```

## Examples

### Direct Execution
```bash
# Create and checkout branch
$(helper jira "VIS-5280 Add user auth" -e)

# Stash changes
$(helper stash "WIP: debugging issue" -e)
```

### With Shell Functions
```bash
# If you've added the functions to your shell config
hjira "TAC-511 Fix security issue"
hstash "Temporary changes"
```

### Type to Terminal (Linux with xdotool)
```bash
# Install xdotool first
sudo dnf install xdotool  # Fedora
sudo apt install xdotool  # Ubuntu/Debian

# Then use --type flag
helper jira "VIS-5280 Feature" --type
# Command appears in terminal, just press Enter
```

## Installation for Auto-Type

### Fedora/RHEL
```bash
sudo dnf install xdotool
```

### Ubuntu/Debian
```bash
sudo apt-get install xdotool
```

### macOS
No additional installation needed (uses built-in AppleScript).

## Benefits

1. **No Clipboard Pollution**: Your clipboard remains untouched
2. **Direct Execution**: Commands run immediately without copy-paste
3. **Shell Integration**: Seamless workflow with shell functions
4. **Cross-Platform**: Works on Linux, macOS, and in SSH sessions

## Command Reference

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--exec` | `-e` | Output for shell execution | `$(helper jira "text" -e)` |
| `--type` | `-t` | Type to terminal | `helper jira "text" -t` |
| `--no-copy` | | Display without copying | `helper jira "text" --no-copy` |
| `--copy` | | Copy to clipboard (default) | `helper jira "text"` |

## Tips

- Use `--exec` with `$()` for immediate execution
- Add shell functions for frequently used commands
- Install `xdotool` on Linux for auto-typing support
- Combine with git aliases for even faster workflow