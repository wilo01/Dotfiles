# Claude Code Hooks System

Intelligent context-aware hooks that provide project-specific guidance and workflow automation.

## Hook Architecture

### Master Hook (`master-hook.sh`)
- **Entry point** that coordinates all other hooks
- Analyzes project context and prompt content
- Routes to appropriate specialized hooks
- Provides consolidated output

### Specialized Hooks

#### Project Context Hook (`project-context-hook.sh`)
- Detects project type (TDS Suite, dotfiles, Node.js, Python, etc.)
- Provides relevant tech stack information
- Suggests appropriate commands and workflows

#### Task Master Hook (`taskmaster-workflow-hook.sh`)
- **Activates in**: Projects with `.taskmaster/` directory
- Detects Task Master intents in prompts
- Suggests appropriate workflow commands
- Shows current task status when relevant

#### Git Workflow Hook (`git-workflow-hook.sh`)
- **Activates in**: Git repositories (`.git/` directory)
- Analyzes current repository state
- Provides contextual git guidance
- Suggests next actions based on git status

#### Legacy Hooks
- `prompt-capture.sh` - Captures prompts for analysis
- `prompt-wrapper.sh` - Alternative capture method

## Hook Activation Logic

```bash
# Always active
project-context-hook.sh

# Conditional activation
if [[ -d ".taskmaster" ]] || prompt contains "task|todo|tm/"; then
    taskmaster-workflow-hook.sh
fi

if [[ -d ".git" ]] || prompt contains "git|commit|branch"; then
    git-workflow-hook.sh
fi
```

## Configuration

Hooks are automatically configured through Claude Code settings. To enable the master hook system:

1. **Set in `.claude/settings.json`:**
```json
{
  "hooks": {
    "user-prompt-submit": "~/.Dotfiles/.claude/hooks/master-hook.sh"
  }
}
```

2. **Or use environment variable:**
```bash
export CLAUDE_USER_PROMPT_SUBMIT_HOOK="$HOME/.Dotfiles/.claude/hooks/master-hook.sh"
```

## Example Outputs

### TDS Suite Project
```
🤖 Claude Code Context Assistant
================================
🔗 Running project-context-hook.sh...
🏢 TDS Suite Project
Key paths: source/ui/, source/server/, test/Cypress/
Tech: ExtJS, Oracle APEX, Java, Liquibase
Commands: npm run devDocker, sencha app build

🔗 Running git-workflow-hook.sh...
📍 Branch: feature/auth-system
📝 3 modified files
💡 Stage changes: git add . or git add <file>
```

### Task Master Project with Task Intent
```
🤖 Claude Code Context Assistant
================================
🔗 Running taskmaster-workflow-hook.sh...
🎯 Suggested workflow:
1. /tm/next - Get next available task
2. Review task details and requirements
3. Start implementation

🎯 Task 1.2 is ready to work on
Use: /tm/next to get started
```

### Dotfiles Project
```
🤖 Claude Code Context Assistant
================================
🔗 Running project-context-hook.sh...
⚙️  Dotfiles Management
Key commands: stow, stow -D, stow -R
Focus: Configuration, symlinks, environment setup
💡 Dotfiles Tips: Use 'stow -D' before 'stow -R', check conflicts with 'stow -n'
```

## Benefits

1. **Context Awareness**: Automatically provides relevant information based on current project
2. **Workflow Guidance**: Suggests appropriate next steps and commands
3. **Intelligent Routing**: Only activates relevant hooks to minimize noise
4. **Extensible**: Easy to add new specialized hooks for different project types
5. **Non-Intrusive**: Provides helpful context without overwhelming output

## Adding New Hooks

1. Create hook script in `/home/dariuszw/.Dotfiles/.claude/hooks/`
2. Make executable: `chmod +x hook-name.sh`
3. Add detection logic to `master-hook.sh`
4. Follow existing patterns for consistent output format

## Logging

All hooks log to individual files in `/home/dariuszw/.Dotfiles/.claude/hooks/`:
- `master.log` - Master hook coordination
- `context.log` - Project context detection  
- `taskmaster.log` - Task Master workflow events
- `git-workflow.log` - Git workflow guidance