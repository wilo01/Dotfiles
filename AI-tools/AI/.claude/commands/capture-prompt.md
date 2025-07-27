# Manually Capture Current Prompt

Manually save the current conversation context as a prompt for future reference.

## Usage
`/capture-prompt [title] [tags]`

## Examples
- `/capture-prompt "Tmux Configuration" tmux,config`
- `/capture-prompt "Debug Session" bug,debugging`

## Steps

1. Prompt the user to provide the prompt text they want to save if not clear from context.

2. Extract title and tags from arguments, using defaults if not provided.

3. Use the prompt manager to save:
   ```bash
   /home/dariuszw/.Dotfiles/.claude/prompts/prompt-manager.sh save "$PROMPT_TEXT" "$TITLE" "$TAGS"
   ```

4. Confirm the prompt was saved and provide the session ID for future reference.