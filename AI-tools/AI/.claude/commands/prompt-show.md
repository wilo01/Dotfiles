# Show Saved Prompt

Display the full content of a saved Claude Code prompt by its ID.

## Usage
`/prompt-show <prompt-id>`

## Examples
- `/prompt-show 2025-01-26_12-30-45_Auth_Fix`

## Steps

1. Use the prompt manager to show the specific prompt:
   ```bash
   /home/dariuszw/.Dotfiles/.claude/prompts/prompt-manager.sh show "$ARGUMENTS"
   ```

2. If the prompt is not found, list available prompts and suggest using `/prompt-search` to find the right ID.

3. Display the full prompt content including metadata, context, and any notes.