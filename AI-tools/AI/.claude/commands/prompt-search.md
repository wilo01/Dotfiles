# Search Saved Prompts

Search through your saved Claude Code prompts and display relevant results.

## Usage
`/prompt-search [query]`

## Examples
- `/prompt-search database` - Find prompts related to database work
- `/prompt-search` - List recent prompts

## Steps

1. Use the prompt manager to search for prompts:
   ```bash
   /home/dariuszw/.Dotfiles/.claude/prompts/prompt-manager.sh search "$ARGUMENTS"
   ```

2. If no query provided, show recent prompts instead:
   ```bash
   /home/dariuszw/.Dotfiles/.claude/prompts/prompt-manager.sh recent
   ```

3. Display results in a readable format with instructions on how to view full prompts.

4. Suggest using `/prompt-show <id>` to view a specific prompt in detail.