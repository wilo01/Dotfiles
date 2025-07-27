# Claude Code Prompt Manager

An automated system for saving, organizing, and managing prompts used in Claude Code sessions.

## Features

- **Automatic Capture**: Prompts are automatically saved when submitted to Claude Code
- **Smart Tagging**: Auto-tags prompts based on content (bug-fix, feature, refactor, etc.)
- **Search & Retrieval**: Easily search through saved prompts by content or tags
- **Templates**: Predefined templates for common prompt patterns
- **Slash Commands**: Integrated slash commands for quick access within Claude Code

## Quick Start

The system is automatically activated through Claude Code hooks. Your prompts will be saved automatically to:
```
~/.Dotfiles/.claude/prompts/sessions/
```

## Commands

### System-wide Command
```bash
prompt-manager help                    # Show all available commands
prompt-manager search [query]          # Search saved prompts
prompt-manager show <id>              # Display specific prompt
prompt-manager recent                 # Show 10 most recent prompts
prompt-manager stats                  # Show usage statistics
```

### Claude Code Slash Commands
- `/prompt-search [query]` - Search through saved prompts
- `/prompt-show <id>` - Display a specific prompt
- `/prompt-save <title> [tags]` - Manually save a prompt
- `/prompt-stats` - Show prompt statistics

## File Structure

```
~/.Dotfiles/.claude/prompts/
├── sessions/           # Automatically saved prompts
├── templates/          # Reusable prompt templates  
├── saved/             # Manually saved prompts
├── prompt-manager.sh  # Main script
└── prompt-index.json  # Search index
```

## Automatic Features

### Auto-Capture Hook
Prompts are automatically captured when submitted to Claude Code through the `user_prompt_submit` hook configured in `settings.local.json`.

### Smart Tagging
Prompts are automatically tagged based on content:
- `bug-fix` - Contains words like "bug", "fix", "error", "issue"
- `feature` - Contains "feature", "implement", "add", "create"
- `refactor` - Contains "refactor", "improve", "optimize"
- `testing` - Contains "test", "testing", "spec"
- `documentation` - Contains "document", "readme", "doc"

### Context Capture
Each saved prompt includes:
- Timestamp and session ID
- Current working directory
- Git branch and status
- Space for implementation notes

## Templates

Pre-built templates for common scenarios:

### Debug Session Template
Use when troubleshooting issues:
```bash
prompt-manager template debug-session
```

### Code Review Template  
Use when requesting code reviews:
```bash
prompt-manager template code-review
```

### Feature Implementation Template
Use when implementing new features:
```bash
prompt-manager template feature-implementation
```

## Usage Examples

### Searching Prompts
```bash
# Search for database-related prompts
prompt-manager search database

# Find all bug fixes
prompt-manager search bug

# List recent prompts
prompt-manager recent
```

### Viewing Prompts
```bash
# Show a specific prompt
prompt-manager show 2025-01-26_14-30-45_Database_Fix

# See usage statistics
prompt-manager stats
```

### Within Claude Code
```
/prompt-search authentication
/prompt-show 2025-01-26_14-30-45_Auth_Implementation
/prompt-stats
```

## Manual Saving

While prompts are auto-captured, you can manually save important prompts:

```bash
prompt-manager save "Implement user authentication with JWT tokens" "JWT Auth Implementation" "auth,security,jwt"
```

## Configuration

### Hook Configuration
The auto-capture is configured via `~/.Dotfiles/.claude/settings.local.json`:

```json
{
  "hooks": {
    "user_prompt_submit": {
      "command": "/home/dariuszw/.Dotfiles/.claude/hooks/prompt-capture.sh",
      "enabled": true
    }
  }
}
```

### Disable Auto-Capture
To temporarily disable automatic prompt capture:
```bash
# Edit settings.local.json and set enabled: false
# Or rename the hook script to disable it
```

## Integration with Dotfiles

The prompt manager integrates with your dotfiles setup:
- Main script: `~/.Dotfiles/.claude/prompts/prompt-manager.sh`
- System wrapper: `~/Dotfiles/scripts/prompt-manager` 
- Hook script: `~/.Dotfiles/.claude/hooks/prompt-capture.sh`

Make sure your `$PATH` includes the scripts directory to use `prompt-manager` globally.

## Troubleshooting

### Check Hook Status
```bash
# View recent captures
tail ~/.Dotfiles/.claude/prompts/capture.log

# Test hook manually
~/.Dotfiles/.claude/hooks/prompt-capture.sh "test prompt"
```

### Verify Installation
```bash
# Check if prompt manager is executable
ls -la ~/.Dotfiles/.claude/prompts/prompt-manager.sh

# Test basic functionality
prompt-manager stats
```

### Rebuild Index
If search isn't working properly:
```bash
# The index rebuilds automatically, but you can check its status
cat ~/.Dotfiles/.claude/prompts/prompt-index.json
```

## Tips

1. **Regular Review**: Use `prompt-manager recent` to review your recent prompting patterns
2. **Template Creation**: Create custom templates for your common use cases
3. **Tagging Strategy**: Use consistent tags to make searching more effective
4. **Archive Old Prompts**: Periodically clean up old prompts to keep the system fast

## Dependencies

- `bash` - Main scripting environment
- `jq` - JSON processing (optional, for index management)
- `grep` - Text searching
- `find` - File operations
- `date` - Timestamp generation

The system gracefully degrades if optional dependencies are missing.