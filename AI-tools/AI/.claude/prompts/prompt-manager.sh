#!/bin/bash

# Claude Code Prompt Manager
# Automatically saves and manages prompts used in Claude Code sessions

PROMPTS_DIR="${HOME}/.Dotfiles/.claude/prompts"
SESSIONS_DIR="${PROMPTS_DIR}/sessions"
INDEX_FILE="${PROMPTS_DIR}/prompt-index.json"

# Create directory structure
mkdir -p "${SESSIONS_DIR}"
mkdir -p "${PROMPTS_DIR}/saved"
mkdir -p "${PROMPTS_DIR}/templates"

# Initialize index file if it doesn't exist
if [[ ! -f "${INDEX_FILE}" ]]; then
    echo '{"sessions": [], "saved_prompts": [], "tags": [], "last_updated": ""}' > "${INDEX_FILE}"
fi

# Function to save a prompt
save_prompt() {
    local prompt_text="$1"
    local title="$2"
    local tags="$3"
    local timestamp=$(date '+%Y-%m-%d_%H-%M-%S')
    local session_id="${timestamp}_$(echo "$title" | tr ' ' '_' | tr -cd '[:alnum:]_-')"
    
    # Create session file
    cat > "${SESSIONS_DIR}/${session_id}.md" << EOF
# ${title}

**Date:** $(date '+%Y-%m-%d %H:%M:%S')
**Tags:** ${tags}
**Session ID:** ${session_id}

## Prompt

\`\`\`
${prompt_text}
\`\`\`

## Context

- Working Directory: $(pwd)
- Git Branch: $(git branch --show-current 2>/dev/null || echo "N/A")
- Git Status: $(git status --porcelain 2>/dev/null | wc -l) modified files

## Notes

<!-- Add implementation notes, results, or follow-up actions here -->

EOF

    # Update index
    update_index "${session_id}" "${title}" "${tags}"
    
    echo "Prompt saved: ${SESSIONS_DIR}/${session_id}.md"
}

# Function to update the index
update_index() {
    local session_id="$1"
    local title="$2"
    local tags="$3"
    local timestamp=$(date -Iseconds)
    
    # Use jq to update the index (install if needed)
    if command -v jq &> /dev/null; then
        local temp_file=$(mktemp)
        jq --arg id "$session_id" \
           --arg title "$title" \
           --arg tags "$tags" \
           --arg timestamp "$timestamp" \
           '.sessions += [{"id": $id, "title": $title, "tags": $tags, "timestamp": $timestamp}] | .last_updated = $timestamp' \
           "$INDEX_FILE" > "$temp_file" && mv "$temp_file" "$INDEX_FILE"
    fi
}

# Function to search prompts
search_prompts() {
    local query="$1"
    echo "Searching for: $query"
    echo "===================="
    
    if [[ -z "$query" ]]; then
        # List all prompts
        find "${SESSIONS_DIR}" -name "*.md" -exec basename {} .md \; | sort -r | head -20
    else
        # Search in content
        grep -l -i "$query" "${SESSIONS_DIR}"/*.md 2>/dev/null | while read -r file; do
            basename=$(basename "$file" .md)
            title=$(grep "^# " "$file" | head -1 | sed 's/^# //')
            echo "$basename: $title"
        done
    fi
}

# Function to show a specific prompt
show_prompt() {
    local prompt_id="$1"
    local file="${SESSIONS_DIR}/${prompt_id}.md"
    
    if [[ -f "$file" ]]; then
        cat "$file"
    else
        echo "Prompt not found: $prompt_id"
        echo "Available prompts:"
        search_prompts ""
    fi
}

# Function to tag a prompt
tag_prompt() {
    local prompt_id="$1"
    local new_tags="$2"
    local file="${SESSIONS_DIR}/${prompt_id}.md"
    
    if [[ -f "$file" ]]; then
        sed -i "s/\*\*Tags:\*\* .*/\*\*Tags:\*\* ${new_tags}/" "$file"
        echo "Tags updated for $prompt_id"
    else
        echo "Prompt not found: $prompt_id"
    fi
}

# Function to create a template
create_template() {
    local template_name="$1"
    local template_content="$2"
    
    cat > "${PROMPTS_DIR}/templates/${template_name}.md" << EOF
# ${template_name} Template

## Description
<!-- Describe when to use this template -->

## Template
\`\`\`
${template_content}
\`\`\`

## Variables
<!-- Document any variables that should be replaced -->
- \${PROJECT_NAME} - Name of the current project
- \${CONTEXT} - Additional context or requirements

## Usage
\`\`\`bash
prompt-manager use-template ${template_name}
\`\`\`
EOF

    echo "Template created: ${PROMPTS_DIR}/templates/${template_name}.md"
}

# Main command dispatcher
case "$1" in
    save)
        if [[ -z "$2" ]]; then
            echo "Usage: prompt-manager save 'prompt text' [title] [tags]"
            exit 1
        fi
        save_prompt "$2" "${3:-Untitled}" "${4:-general}"
        ;;
    search)
        search_prompts "$2"
        ;;
    show)
        show_prompt "$2"
        ;;
    tag)
        tag_prompt "$2" "$3"
        ;;
    template)
        create_template "$2" "$3"
        ;;
    list)
        search_prompts ""
        ;;
    recent)
        find "${SESSIONS_DIR}" -name "*.md" -printf '%T@ %p\n' | sort -nr | head -10 | while read -r timestamp file; do
            basename=$(basename "$file" .md)
            title=$(grep "^# " "$file" | head -1 | sed 's/^# //')
            date=$(date -d "@${timestamp%.*}" '+%Y-%m-%d %H:%M')
            echo "$date - $basename: $title"
        done
        ;;
    stats)
        echo "Prompt Statistics"
        echo "================="
        echo "Total sessions: $(find "${SESSIONS_DIR}" -name "*.md" | wc -l)"
        echo "Templates: $(find "${PROMPTS_DIR}/templates" -name "*.md" 2>/dev/null | wc -l)"
        echo "Disk usage: $(du -sh "${PROMPTS_DIR}" | cut -f1)"
        echo ""
        echo "Recent activity:"
        find "${SESSIONS_DIR}" -name "*.md" -printf '%T@ %p\n' | sort -nr | head -5 | while read -r timestamp file; do
            date=$(date -d "@${timestamp%.*}" '+%Y-%m-%d')
            echo "  $date - $(basename "$file" .md)"
        done
        ;;
    help|*)
        cat << EOF
Claude Code Prompt Manager

USAGE:
    prompt-manager <command> [options]

COMMANDS:
    save 'text' [title] [tags]  Save a prompt with optional title and tags
    search [query]              Search prompts by content (empty = list all)
    show <id>                   Display a specific prompt by ID
    tag <id> <tags>            Update tags for a prompt
    template <name> <content>   Create a reusable template
    list                        List all saved prompts
    recent                      Show 10 most recent prompts
    stats                       Show usage statistics
    help                        Show this help message

EXAMPLES:
    prompt-manager save 'Fix the login bug in auth.js' 'Auth Fix' 'bug,auth,javascript'
    prompt-manager search 'database'
    prompt-manager show 2025-01-26_12-30-45_Auth_Fix
    prompt-manager recent

FILES:
    Sessions: ~/.Dotfiles/.claude/prompts/sessions/
    Templates: ~/.Dotfiles/.claude/prompts/templates/
    Index: ~/.Dotfiles/.claude/prompts/prompt-index.json
EOF
        ;;
esac