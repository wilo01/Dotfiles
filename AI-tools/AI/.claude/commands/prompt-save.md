# Save Prompt Manually

Manually save a prompt with custom title and tags.

## Usage
`/prompt-save <title> [tags]`

## Examples
- `/prompt-save "Fix authentication bug" bug,auth,security`
- `/prompt-save "Database migration script" database,migration`

## Steps

1. Prompt the user for the prompt text if not provided in context:
   "Please provide the prompt text you want to save."

2. Extract title from the first argument and tags from the second argument (if provided).

3. Use the prompt manager to save:
   ```bash
   /home/dariuszw/.Dotfiles/.claude/prompts/prompt-manager.sh save "$PROMPT_TEXT" "$TITLE" "$TAGS"
   ```

4. Confirm the prompt was saved and provide the generated ID for future reference.