#!/bin/bash

# Claude Code Status Line Script - ZSH Prompt Style
# Format: OS filepath on git_icon branch | version | model | commit

# Read JSON input from stdin
input=$(cat)

# Extract information from JSON input
version=$(echo "$input" | jq -r '.version // ""')
model_name=$(echo "$input" | jq -r '.model.display_name // ""')
current_dir=$(echo "$input" | jq -r '.workspace.current_dir // ""')
project_dir=$(echo "$input" | jq -r '.workspace.project_dir // ""')

# Fallback for current directory
if [ -z "$current_dir" ] || [ "$current_dir" = "null" ]; then
    current_dir=$(pwd)
fi

# ANSI color codes to exactly match Powerlevel10k
# Using exact color codes from your p10k config
DIR_COLOR=$'\033[38;5;39m'       # Color 39 - bright cyan (POWERLEVEL9K_DIR_ANCHOR_FOREGROUND=39)
DIR_ANCHOR_COLOR=$'\033[38;5;39m' # Color 39 - bright cyan (POWERLEVEL9K_DIR_ANCHOR_FOREGROUND=39)
GIT_CLEAN=$'\033[38;5;76m'       # Color 76 - bright green (VCS_CLEAN_FOREGROUND=76)
GIT_MODIFIED=$'\033[38;5;178m'   # Color 178 - yellow/gold (VCS_MODIFIED_FOREGROUND=178)
GIT_UNTRACKED=$'\033[38;5;39m' # Color 39 - bright cyan (POWERLEVEL9K_DIR_ANCHOR_FOREGROUND=39)
BRIGHT_CYAN=$'\033[1;36m'        # Bright cyan for version
BRIGHT_BLUE=$'\033[1;34m'        # Bright blue for model
BRIGHT_WHITE=$'\033[1;37m'       # Bright white for commit and separators
GRAY=$'\033[0;90m'               # Gray for separators
RESET=$'\033[0m'

# OS Icon detection - using actual Nerd Font icons with Linux distribution support
get_os_icon() {
    case "$(uname -s)" in
        Linux*)
            # Check for specific Linux distributions
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                case "$ID" in
                    ubuntu)     echo $'\uf31b ';;      # Ubuntu logo (U+F31B)
                    fedora)     echo $'\uf30a ';;      # Fedora logo (U+F30A)
                    arch)       echo $'\uf303 ';;      # Arch logo (U+F303)
                    debian)     echo $'\uf306 ';;      # Debian logo (U+F306)
                    opensuse*)  echo $'\uf314 ';;      # openSUSE logo (U+F314)
                    centos)     echo $'\uf304 ';;      # CentOS logo (U+F304)
                    rhel)       echo $'\uf316 ';;      # Red Hat logo (U+F316)
                    *)          echo $'\uf17c ';;      # Generic Tux penguin (U+F17C)
                esac
            else
                echo $'\uf17c '                        # Generic Tux penguin (U+F17C)
            fi
            ;;
        Darwin*)    echo $'\uf179 ';;                  # Apple logo (U+F179)
        CYGWIN*|MINGW*) echo $'\uf17a ';;             # Windows logo (U+F17A)
        *)          echo $'\uf108 ';;                  # Generic computer (U+F108)
    esac
}

OS_ICON=$(get_os_icon)

# Git branch text - using "git" text instead of icon
GIT_BRANCH_TEXT="git "

# Folder icon - Nerd Font folder icon
FOLDER_ICON=$'\uf07b '

# Build status line components in order: [OS icon] [folder icon] filepath on [git icon] branch ⇡2 !4 | version | model | commit
status_components=()

# 1. Add OS icon and filepath (using tilde notation) - cyan color like P10k
if [ -n "$current_dir" ]; then
    # Convert to tilde notation if in home directory
    if [[ "$current_dir" == "$HOME"* ]]; then
        display_path="~${current_dir#$HOME}"
    else
        display_path="$current_dir"
    fi
    
    # Check if we're in a git repo to combine filepath and git info
    if [ -d "$current_dir/.git" ] || git -C "$current_dir" rev-parse --git-dir > /dev/null 2>&1; then
        # Get current branch name
        branch=$(git -C "$current_dir" branch --show-current 2>/dev/null)
        
        if [ -z "$branch" ]; then
            # Handle detached HEAD or other cases
            branch="HEAD"
        fi
        
        # Combine filepath and git section without separator: OS_ICON + DIR_COLOR + FOLDER_ICON + filepath + " on" (default) + git_text (green) + branch (green)
        combined_section="${OS_ICON}${DIR_COLOR}${FOLDER_ICON}${display_path}${RESET} on ${GIT_CLEAN}${GIT_BRANCH_TEXT}${branch}${RESET}"
        
        # Get commits ahead/behind (using P10k logic)
        upstream=$(git -C "$current_dir" rev-parse --abbrev-ref @{upstream} 2>/dev/null)
        if [ -n "$upstream" ]; then
            # P10k uses different logic: behind = commits to pull, ahead = commits to push
            behind=$(git -C "$current_dir" rev-list --count HEAD..@{upstream} 2>/dev/null)
            ahead=$(git -C "$current_dir" rev-list --count @{upstream}..HEAD 2>/dev/null)
            
            # ⇣ for commits behind (to pull) - green like P10k
            if [ "$behind" -gt 0 ]; then
                combined_section+=" ${GIT_CLEAN}⇣${behind}${RESET}"
            fi
            
            # ⇡ for commits ahead (to push) - green like P10k  
            if [ "$ahead" -gt 0 ]; then
                combined_section+=" ${GIT_CLEAN}⇡${ahead}${RESET}"
            fi
        fi
        
        # Get git stash count (*) - green like P10k (color 76)
        stash_count=$(git -C "$current_dir" stash list 2>/dev/null | wc -l)
        if [ "$stash_count" -gt 0 ]; then
            combined_section+=" ${GIT_CLEAN}*${stash_count}${RESET}"
        fi
        
        # Get unstaged files count (!) - yellow like P10k
        unstaged=$(git -C "$current_dir" diff --name-only 2>/dev/null | wc -l)
        if [ "$unstaged" -gt 0 ]; then
            combined_section+=" ${GIT_MODIFIED}!${unstaged}${RESET}"
        fi
        
        # Get untracked files count (?) - green like P10k (color 76)
        untracked=$(git -C "$current_dir" ls-files --others --exclude-standard 2>/dev/null | wc -l)
        if [ "$untracked" -gt 0 ]; then
            combined_section+=" ${GIT_UNTRACKED}?${untracked}${RESET}"
        fi
        
        status_components+=("$combined_section")
    else
        # No git repo - just add filepath
        status_components+=("${OS_ICON}${DIR_COLOR}${FOLDER_ICON}${display_path}${RESET}")
    fi
fi

# 2. Add version if available (bright cyan)
if [ -n "$version" ] && [ "$version" != "null" ]; then
    status_components+=("${BRIGHT_CYAN}󰓹 v${version}${RESET}")
fi

# 3. Add model if available (bright blue)
if [ -n "$model_name" ] && [ "$model_name" != "null" ]; then
    status_components+=("${BRIGHT_BLUE}󰧠 ${model_name}${RESET}")
fi

# 4. Add project name if available (using basename of project_dir)
if [ -n "$project_dir" ] && [ "$project_dir" != "null" ] && [ "$project_dir" != "$current_dir" ]; then
    project_name=$(basename "$project_dir")
    status_components+=("${BRIGHT_WHITE}󰉋 ${project_name}${RESET}")
fi

# 5. Add commit hash identifier if in git repo (bright white)
if [ -d "$current_dir/.git" ] || git -C "$current_dir" rev-parse --git-dir > /dev/null 2>&1; then
    commit_hash=$(git -C "$current_dir" rev-parse --short HEAD 2>/dev/null)
    if [ -n "$commit_hash" ]; then
        status_components+=("${BRIGHT_WHITE}󰿘 ${commit_hash}${RESET}")
    fi
fi

# Assembly logic: join all components with | separator
final_status=""
for i in "${!status_components[@]}"; do
    if [ $i -gt 0 ]; then
        final_status+=" ${GRAY}│${RESET} "
    fi
    final_status+="${status_components[$i]}"
done

echo "$final_status"
