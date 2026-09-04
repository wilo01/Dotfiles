# Enable Powerlevel10k instant prompt. Should stay close to the top of ~/.zshrc.
# Initialization code that may require console input (password prompts, [y/n]
# confirmations, etc.) must go above this block; everything else may go below.
if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

# If you come from bash you might have to change your $PATH.
# export PATH=$HOME/bin:/usr/local/bin:$PATH

# Path to your oh-my-zsh installation.
export ZSH="$HOME/.oh-my-zsh"

# Set name of the theme to load --- if set to "random", it will
# load a random theme each time oh-my-zsh is loaded, in which case,
# to know which specific one was loaded, run: echo $RANDOM_THEME
# See https://github.com/ohmyzsh/ohmyzsh/wiki/Themes
ZSH_THEME="powerlevel10k/powerlevel10k"
POWERLEVEL9K_MODE="nerdfont-complete"
plugins=(git zsh-autosuggestions zsh-syntax-highlighting evalcache git-extras debian tmux screen history extract colorize web-search docker direnv)

# Set list of themes to pick from when loading at random
# Setting this variable when ZSH_THEME=random will cause zsh to load
# a theme from this variable instead of looking in $ZSH/themes/
# If set to an empty array, this variable will have no effect.
# ZSH_THEME_RANDOM_CANDIDATES=( "robbyrussell" "agnoster" )

# Uncomment the following line to use case-sensitive completion.
# CASE_SENSITIVE="true"

# Uncomment the following line to use hyphen-insensitive completion.
# Case-sensitive completion must be off. _ and - will be interchangeable.
# HYPHEN_INSENSITIVE="true"

# Uncomment one of the following lines to change the auto-update behavior
# zstyle ':omz:update' mode disabled  # disable automatic updates
# zstyle ':omz:update' mode auto      # update automatically without asking
# zstyle ':omz:update' mode reminder  # just remind me to update when it's time

# Uncomment the following line to change how often to auto-update (in days).
# zstyle ':omz:update' frequency 13

# Uncomment the following line if pasting URLs and other text is messed up.
# DISABLE_MAGIC_FUNCTIONS="true"

# Uncomment the following line to disable colors in ls.
# DISABLE_LS_COLORS="true"

# Uncomment the following line to disable auto-setting terminal title.
# DISABLE_AUTO_TITLE="true"

# Uncomment the following line to enable command auto-correction.
# ENABLE_CORRECTION="true"

# Uncomment the following line to display red dots whilst waiting for completion.
# You can also set it to another string to have that shown instead of the default red dots.
# e.g. COMPLETION_WAITING_DOTS="%F{yellow}waiting...%f"
# Caution: this setting can cause issues with multiline prompts in zsh < 5.7.1 (see #5765)
# COMPLETION_WAITING_DOTS="true"

# Uncomment the following line if you want to disable marking untracked files
# under VCS as dirty. This makes repository status check for large repositories
# much, much faster.
# DISABLE_UNTRACKED_FILES_DIRTY="true"

# Uncomment the following line if you want to change the command execution time
# stamp shown in the history command output.
# You can set one of the optional three formats:
# "mm/dd/yyyy"|"dd.mm.yyyy"|"yyyy-mm-dd"
# or set a custom format using the strftime function format specifications,
# see 'man strftime' for details.
# HIST_STAMPS="mm/dd/yyyy"

# Would you like to use another custom folder than $ZSH/custom?
# ZSH_CUSTOM=/path/to/new-custom-folder

# Which plugins would you like to load?
# Standard plugins can be found in $ZSH/plugins/
# Custom plugins may be added to $ZSH_CUSTOM/plugins/
# Example format: plugins=(rails git textmate ruby lighthouse)
# Add wisely, as too many plugins slow down shell startup.

source $ZSH/oh-my-zsh.sh

# User configuration

# export MANPATH="/usr/local/man:$MANPATH"

# You may need to manually set your language environment
# export LANG=en_US.UTF-8

# Preferred editor for local and remote sessions
if [[ -n $SSH_CONNECTION ]]; then
  if command -v nvim >/dev/null 2>&1; then
    export EDITOR='nvim'
  else
    export EDITOR='vim'
  fi
else
  export EDITOR='nvim'
fi

# Compilation flags
# export ARCHFLAGS="-arch x86_64"

# Set personal aliases, overriding those provided by oh-my-zsh libs,
# plugins, and themes. Aliases can be placed here, though oh-my-zsh
# users are encouraged to define aliases within the ZSH_CUSTOM folder.
# For a full list of active aliases, run `alias`.
#
# Example aliases
# alias zshconfig="mate ~/.zshrc"
# alias ohmyzsh="mate ~/.oh-my-zsh"

# To customize prompt, run `p10k configure` or edit ~/.p10k.zsh.
[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh

# nvim switcher
#
# To add new nvim config (for Flatpak Neovim):
# 1. Clone/create config: ~/.Dotfiles/nvim/.config/nvim-<name>/
# 2. Run: nvims-update (or nvu) - automatically syncs configs
#
# Manual sync (if needed):
# - Stow: cd ~/.Dotfiles && stow nvim
# - Flatpak: ln -s ~/.config/nvim-<name> ~/.var/app/io.neovim.nvim/config/nvim-<name>
#
alias nvim-vimscript="NVIM_APPNAME=nvim-vimscript nvim"
alias nvim-reddit="NVIM_APPNAME=nvim-reddit nvim"

function nvims() {
    if ! command -v fzf &>/dev/null; then
        echo "Error: fzf is not installed"
        return 1
    fi
    local config_path="$HOME/.config"
    local items=()
    local numbered_items=()
    local configs=()

    for config_dir in "$config_path"/nv*; do
        if [[ -d $config_dir ]]; then
            configs+=("${config_dir##*/}")
        fi
    done

    IFS=$'\n' configs=($(sort <<<"${configs[*]}"))
    unset IFS

    local i=1
    local bind_cmds=""
    for name in "${configs[@]}"; do
        items+=("$name")
        numbered_items+=("$i) $name")

        if [[ $i -le 9 ]]; then
            if [[ -n "$bind_cmds" ]]; then
                bind_cmds+=","
            fi

            local pos=$((i - 1))
            bind_cmds+="$i:pos($i)+accept"
        fi
        ((i++))
    done

    local selected
    selected=$(printf "%s\n" "${numbered_items[@]}" | \
        fzf --prompt=" Neovim Config  " \
            --height=50% \
            --layout=reverse \
            --border \
            --exit-0 \
            --bind="$bind_cmds")

    if [[ -z "$selected" ]]; then
        echo "Nothing selected"
        return 0
    fi

    local config=$(echo "$selected" | sed 's/^[0-9]*) //')

    if [[ "$config" == "nvim" ]]; then
        NVIM_APPNAME="" nvim $@
    else
        NVIM_APPNAME="$config" nvim $@
    fi
}

bindkey -s ^a "nvims\n"

function nvims-update() {
    local dotfiles_path="$HOME/.Dotfiles"
    local config_path="$HOME/.config"
    local flatpak_config_path="$HOME/.var/app/io.neovim.nvim/config"
    local original_dir="$(pwd)"
    local synced=0
    local skipped=0
    local errors=0

    echo "🔄 Syncing Neovim configurations..."
    echo ""

    if [[ ! -d "$dotfiles_path" ]]; then
        echo "❌ Error: $dotfiles_path directory not found"
        return 1
    fi

    if ! command -v stow >/dev/null 2>&1; then
        echo "❌ Error: stow command not found. Please install stow."
        return 1
    fi

    echo "📦 Running stow to sync configs to ~/.config/..."
    cd "$dotfiles_path" || { echo "❌ Failed to cd to $dotfiles_path"; return 1; }

    if stow nvim 2>/dev/null; then
        echo "✅ Stow completed successfully"
    else
        echo "⚠️  Stow completed with warnings (configs may already be linked)"
    fi
    echo ""

    mkdir -p "$flatpak_config_path" 2>/dev/null
    echo "🔗 Creating symlinks for Flatpak Neovim..."
    echo ""

    for config_dir in "$config_path"/nvim*; do
        if [[ -d "$config_dir" ]]; then
            local config_name="${config_dir##*/}"
            local flatpak_link="$flatpak_config_path/$config_name"

            if [[ -L "$flatpak_link" ]]; then
                local current_target="$(readlink "$flatpak_link")"
                if [[ "$current_target" == "$config_dir" ]]; then
                    echo "⏭️  $config_name (already synced)"
                    ((skipped++))
                else
                    echo "⚠️  $config_name (updating symlink)"
                    rm "$flatpak_link"
                    if ln -s "$config_dir" "$flatpak_link" 2>/dev/null; then
                        echo "✅ $config_name (updated)"
                        ((synced++))
                    else
                        echo "❌ $config_name (failed to update)"
                        ((errors++))
                    fi
                fi
            elif [[ -e "$flatpak_link" ]]; then
                echo "❌ $config_name (path exists but is not a symlink)"
                ((errors++))
            else
                if ln -s "$config_dir" "$flatpak_link" 2>/dev/null; then
                    echo "✅ $config_name (synced)"
                    ((synced++))
                else
                    echo "❌ $config_name (failed to create symlink)"
                    ((errors++))
                fi
            fi
        fi
    done

    cd "$original_dir" || true

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📊 Summary: $synced synced, $skipped skipped, $errors errors"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $errors -gt 0 ]]; then
        return 1
    fi
}

bindkey "^[[1;2C" forward-word
bindkey "^[[1;2D" backward-word

# cd & ls movements
alias ls="ls --group-directories-first --color=auto"
alias ll="ls -lha -F --show-control-chars --time-style=locale --color=auto"
# alias cd="~/bin/.local/scripts/tmux-sessionizer"
# alias CD="~/bin/.local/scripts/tmux-sessionizer"
alias CD="cd"
alias cl="clear"
alias cds="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/safe"
alias cdkio="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/kiosk-chrome-app"
alias cdkiosk="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/kiosk-chrome-app"
alias cdweb="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/visitor-web-app"
alias cdwapp="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/visitor-web-app"
alias cdwebapp="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/visitor-web-app"
alias cdb="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/app/"
alias cdy="~/bin/.local/scripts/tmux-sessionizer ~/tds-branch-opener/branches/tds-suite/test/Cypress"
# CODE actions
alias code_ks="echo code ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/app/global/Settings.js ; code ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/app/global/Settings.js"
alias code_ka="echo code ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/app/Application.js ; code ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/app/Application.js"
alias code_ksc="echo code ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/app/view/settings/SettingsController.js ; code ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/app/view/settings/SettingsController.js"
alias kiosk_settings="echo open kiosk settings at: ; code_ks ; sleep 1 ; code_ka ; sleep 1 ; code_ksc ;"
alias liqui_valid="echo cd ~/tds-branch-opener/branches/tds-suite/source/server/database/ ; echo ./liquibase --defaultsFile=validate.liquibase.properties validate ; cd ~/tds-branch-opener/branches/tds-suite/source/server/database/ ; ./liquibase --defaultsFile=validate.liquibase.properties validate"
alias sqldev="echo ~/SQLDeveloper/opt/sqldeveloper/sqldeveloper.sh ; ~/SQLDeveloper/opt/sqldeveloper/sqldeveloper.sh"
alias br='echo npm start at: ; echo ~/Dev/branch-opener/app/ ; if [[ -n "$(find ~/Dev/branch-opener/app/apex/kiosk/bdb/ -maxdepth 0 -type f -o -type d -printf '%s')" ]]; then echo "Removing content from ~/Dev/branch-opener/app/apex/kiosk/bdb/" ; rm -rf ~/Dev/branch-opener/app/apex/kiosk/bdb/* ; else echo "No content found in ~/Dev/branch-opener/app/apex/kiosk/bdb/, skipping removal." ; fi ; ls ~/Dev/branch-opener/app/apex/kiosk/bdb/ ; cd ~/Dev/branch-opener/app/ ; sleep 1 ; xdg-open http://localhost:3333/static/ ; npm start'
function hx() {
    local variant
    case "${1:-dev}" in
        dev|local|"") variant="dev" ;;
        cloud|c|dev:cloud) variant="dev:cloud" ;;
        *) echo "[hx] Unknown variant '$1' (use: hx | hx cloud)" >&2; return 1 ;;
    esac
    local env="${HX_ENV:-dev}"

    if [[ $variant == dev ]] && ! curl -sf http://localhost:80/api/status >/dev/null 2>&1; then
        echo "[hx] Starting Infisical..."
        docker compose -f ~/.infisical/docker-compose.yml up -d
        echo -n "[hx] Waiting for Infisical..."
        local i=0
        until curl -sf http://localhost:80/api/status >/dev/null 2>&1; do
            (( i++ )); [[ $i -ge 30 ]] && { echo " timed out." >&2; return 1; }
            echo -n "."; sleep 1
        done
        echo " ready"
    fi

    pkill -f "pnpm run dev"
    echo "pnpm start at: ~/tds-branch-opener/tds-hexer/"
    echo "Pulling latest changes..."
    echo "Running: pnpm run $variant (env: $env)"
    cd ~/tds-branch-opener/tds-hexer/
    if command git rev-parse --abbrev-ref '@{upstream}' >/dev/null 2>&1; then
        command git pull || echo "[hx] git pull failed — continuing anyway" >&2
    else
        echo "[hx] No upstream for branch $(command git branch --show-current) — skipping pull"
    fi
    pnpm install
    pnpm --dir src/web install
    if [[ $variant == dev ]]; then
        infisical run --env="$env" --domain=http://localhost:80 -- pnpm run "$variant"
    else
        if [[ -f .env ]]; then
            echo "[hx] .env found in $PWD — dev:cloud must not use it (dotenv in server.js loads it for any unset vars)." >&2
            echo "[hx] Move it aside first: mv .env .env.local-only" >&2
            return 1
        fi
        pnpm run "$variant"
    fi
    xdg-open https://trunk.acrid.dev:3443/safe
}
alias ksw='echo kiosk start at: ; echo ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/ ; cd ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/ ; sencha app watch'
alias vsw='echo visitor-web-app start at: ; echo ~/tds-branch-opener/branches/tds-visitor-web-app/ui ; cd ~/tds-branch-opener/branches/tds-visitor-web-app/ui ; sencha app watch'
alias cy="echo Cypress open at: ; cdy ; sleep 1 ; echo ./node_modules/cypress/bin/cypress open ; ./node_modules/cypress/bin/cypress open"
alias cy_all="echo Cypress run all tests at: ; cdy ; sleep 1 ; echo npx cypress run --headless --spec cypress/integration/tdsvisitor/rt/*.js ; npx cypress run --headless --spec cypress/integration/tdsvisitor/rt/*.js"
alias docker_start_trunk="echo cd ~/Dev/branch-opener/branches/safe ; echo sudo docker start -ai trunk ; cd ~/Dev/branch-opener/branches/safe && sudo docker start -ai trunk"
alias liquibaseLocalDockerUpdate="echo cd ~/Dev/branch-opener/branches/safe ; echo npm run liquibaseLocalDockerUpdate ; cd ~/Dev/branch-opener/branches/safe && npm run liquibaseLocalDockerUpdate"
alias npmliquibaseLocalDockerUpdate="echo cd ~/Dev/branch-opener/branches/safe ; echo npm run liquibaseLocalDockerUpdate ; cd ~/Dev/branch-opener/branches/safe && npm run liquibaseLocalDockerUpdate"
alias apex_remove="echo rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/* ; rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/*"
alias remove_apex="echo rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/* ; rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/*"
alias apex_zip="echo zip -r rt.zip ~/tds-branch-opener/branches/tds-suite/source/server/rt/* ; zip -r rt.zip ~/tds-branch-opener/branches/tds-suite/source/server/rt/* && "
alias zip_apex="echo zip -r rt.zip ~/tds-branch-opener/branches/tds-suite/source/server/rt/* ; zip -r rt.zip ~/tds-branch-opener/branches/tds-suite/source/server/rt/* && "
# alias csp_hash="echo sha256-$(echo -n "$(xclip -o)" | openssl sha256 -binary | openssl base64)"
# Git
alias git_lens="git log --graph --oneline --decorate ; echo git log --graph --oneline --decorate"
alias git_graph="git log --graph --oneline --decorate ; echo git log --graph --oneline --decorate"
alias git_last="git log -1 --stat ; echo git log -1 --stat"
GIT_DIFF_STATE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/git_diff"
function git_default_remote_branch() {
   local head candidate
   head=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null) && { echo "$head"; return }
   for candidate in origin/master origin/main; do
      git rev-parse --verify --quiet "$candidate" >/dev/null && { echo "$candidate"; return }
   done
   echo origin/master
}
function git_diff_state_file() {
   local top
   top=$(git rev-parse --show-toplevel 2>/dev/null) || return 1
   mkdir -p "$GIT_DIFF_STATE_DIR"
   echo "$GIT_DIFF_STATE_DIR/${top//\//_}.json"
}
function git_diff_server_alive() {
   local file="$1" pid port
   [[ -f "$file" ]] || return 1
   pid=$(jq -r '.pid' "$file" 2>/dev/null)
   port=$(jq -r '.port' "$file" 2>/dev/null)
   kill -0 "$pid" 2>/dev/null || return 1
   curl -sf -o /dev/null --max-time 2 "http://localhost:${port}/api/diff"
}
function git_diff() {
   git rev-parse --is-inside-work-tree >/dev/null 2>&1 || {
      echo "git_diff: not inside a git repository"
      return 1
   }

   local base
   if [[ -n "$1" && "$1" != -* ]]; then
      base="$1"
      shift
   else
      base="$(git_default_remote_branch)"
   fi

   local state url output handshake
   state="$(git_diff_state_file)"

   if git_diff_server_alive "$state"; then
      url=$(jq -r '.url' "$state")
      echo "git_diff: already watching this repo -> ${url}"
      xdg-open "$url" >/dev/null 2>&1 &!
      return 0
   fi

   git fetch --quiet origin "${base#origin/}" 2>/dev/null

   local -a runner
   if command -v difit >/dev/null 2>&1; then runner=(difit); else runner=(npx -y difit@5); fi
   output=$("${runner[@]}" . "$base" --merge-base --include-untracked --keep-alive --no-open --background "$@" 2>&1)
   handshake=$(echo "$output" | grep -o '{"port".*}' | tail -1)

   if [[ -z "$handshake" ]]; then
      echo "git_diff: difit failed to start"
      echo "$output"
      return 1
   fi

   url=$(echo "$handshake" | jq -r '.url')
   echo "$handshake" | jq \
      --arg repo "$(git rev-parse --show-toplevel)" \
      --arg branch "$(git rev-parse --abbrev-ref HEAD)" \
      --arg base "$base" \
      '. + {repo: $repo, branch: $branch, base: $base}' > "$state"

   echo "git_diff: $(git rev-parse --abbrev-ref HEAD) vs ${base} -> ${url}"
   xdg-open "$url" >/dev/null 2>&1 &!
}
function git_diff_list() {
   local file rows=""
   local -a files
   files=("$GIT_DIFF_STATE_DIR"/*.json(N))

   for file in $files; do
      if git_diff_server_alive "$file"; then
         rows+="$(jq -r '[.url, .branch, .repo, .pid] | @tsv' "$file")"$'\n'
      else
         rm -f "$file"
      fi
   done

   [[ -n "$rows" ]] || { echo "git_diff: nothing running"; return 0 }

   printf 'URL\tBRANCH\tREPO\tPID\n%s' "$rows" | sed "s|${HOME}|~|g" | column -t -s $'\t'
}
function git_diff_stop() {
   local -a files
   if [[ "$1" == "--all" || "$1" == "-a" ]]; then
      files=("$GIT_DIFF_STATE_DIR"/*.json(N))
   else
      files=("${(@f)$(git_diff_state_file)}") || {
         echo "git_diff_stop: not inside a git repository (use --all)"
         return 1
      }
   fi

   local file pid
   for file in $files; do
      [[ -f "$file" ]] || continue
      pid=$(jq -r '.pid' "$file")
      kill "$pid" 2>/dev/null && echo "git_diff: stopped $(jq -r '.branch' "$file") (pid ${pid})"
      rm -f "$file"
   done
}
alias git_hash="echo Get current branch hash; echo git rev-parse HEAD ; echo ; git rev-parse HEAD"
alias git_hash_10-2av="echo Get 10.2AV branch hash; echo git rev-parse maintenance/10.2AV ; echo ; git rev-parse maintenance/10.2AV"
alias git_hash_11av="echo Get 11AV branch hash; echo git rev-parse maintenance/11AV ; echo ; git rev-parse maintenance/11AV"
alias git_hash_11-1av="echo Get 11.1AV branch hash; echo git rev-parse maintenance/11.1AV ; echo ; git rev-parse maintenance/11.1AV"
alias git_hash_12av="echo Get 12AV branch hash; echo git rev-parse maintenance/12AV ; echo ; git rev-parse maintenance/12AV"
alias git_stash="echo git stash ; echo Git stash changes; git stash"
alias git_apply="echo git stash apply ; echo Git stash apply last changes; git stash apply"
alias git_drop="echo git stash clear ; echo Clear all stash container ; git stash clear"
alias git_stash_clear="echo git stash clear ; echo Clear all stash container ; git stash clear"
alias git_pop="echo git stash pop ; echo Apply last stash ; git stash pop"
alias git_clear="echo git restore . ; echo Git clear changes ; git restore . "
alias git_clean="echo git restore . ; echo Git clear changes ; git restore . "
alias git_branch="echo git branch --show-current ; echo Git show current branch ; echo ; git branch --show-current ; echo ;"
alias git_amend="echo git commit --amend ; echo Git undo commit ; git commit --amend"
alias git_undo="echo git reset --soft HEAD~1 ; echo Git undo commit ; git reset --soft HEAD~1"
alias git_reset="echo git reset --soft HEAD~1 ; echo Git undo commit ; git reset --soft HEAD~1"
alias git_merge_abort="echo git merge --abort ; echo Git abort merge ; git merge --abort"
alias git_abort_merge="echo git merge --abort ; echo Git abort merge ; git merge --abort"
alias git_undo_merge="echo git merge --abort ; echo Git abort merge ; git merge --abort"
alias git_merge_undo="echo git merge --abort ; echo Git abort merge ; git merge --abort"

# Git wrapper - bisect helper + status after add
function git() {
    # Handle git bisect stop/exit
    if [[ $1 == "bisect" && ($2 == "stop" || $2 == "exit") ]]; then
        echo "❗ 'git bisect reset' is the proper way to exit bisect mode. Executing it for you now..."
        command git bisect reset
        return $?
    fi

    # Execute git command
    command git "$@"
    local ret=$?

    # Show status after add
    if [[ "$1" == "add" && $ret -eq 0 ]]; then
        command git status
    fi

    return $ret
}

# gh wrapper - pull latest master before pr checkout
function gh() {
    command git status
    if [[ "$1" == "pr" && "$2" == "checkout" ]]; then
        echo "Syncing master before PR checkout..."
        command git checkout master && command git pull || {
            echo "Failed to sync master. Aborting PR checkout."
            return 1
        }
    fi
    command gh "$@"
}

# Linux Setup
alias sshkey="echo cat ~/.ssh/id_ed25519.pub ; cat ~/.ssh/id_ed25519.pub"
alias ssh_key="echo cat ~/.ssh/id_ed25519.pub ; cat ~/.ssh/id_ed25519.pub"
alias open="echo xdg-open; xdg-open"
alias gnome-terminal='gnome-terminal --full-screen'
alias zshrc="echo nvim ~/.zshrc ; nvim ~/.zshrc "
alias recat="echo ~/recatest/recatest_run.sh ; ~/recatest/recatest_run.sh"
alias clear_cache="echo free -h ; echo ; echo Before clean:; free -h ; echo ; echo After clean: ; echo sync \&\& echo 3 \| sudo tee /proc/sys/vm/drop_caches \&\& free -h ; sync && echo 3 | sudo tee /proc/sys/vm/drop_caches && free -h"
alias tm='task-master'
alias taskmaster='task-master'
alias gemini-fast='gemini -m gemini-2.0-flash'
alias clauded='claude --dangerously-skip-permissions'
alias claude-sonnet='claude --model claude-sonnet-4-20250514'
alias claude-opus='claude --model claude-opus-4-1-20250805'
alias claude-fast='claude-haiku'
alias claude-haiku='claude --model claude-3-5-haiku-20241022'
# Personal Claude Code account (isolated config dir; first run will prompt /login)
alias claude-daro='CLAUDE_CONFIG_DIR="$HOME/.claude-personal" claude'
# Minimal Claude Code profile: no skills/rules/commands
alias claude-mini='CLAUDE_CONFIG_DIR="$HOME/.claude-mini" claude'
# alias fcc-claude='fcc-claude'

# Auto-start fcc-server if not already running
fcc-claude() {
    local original_dir="$PWD"
    local log_dir="$HOME/tmp/fcc-logs"

    # Check if fcc-server is already healthy
    if ! curl -s -o /dev/null --max-time 2 http://127.0.0.1:8082/health 2>/dev/null; then
        echo "Starting fcc-server in background..."
        mkdir -p "$log_dir"
        cd "$log_dir" || return 1
        nohup fcc-server > "$log_dir/server-stdout.log" 2>&1 &
        cd "$original_dir" || return 1
        # Wait up to 10s for it to become healthy
        for i in $(seq 1 10); do
            if curl -s -o /dev/null --max-time 1 http://127.0.0.1:8082/health 2>/dev/null; then
                echo "fcc-server is ready"
                break
            fi
            sleep 1
        done
    fi
    command fcc-claude "$@"
}
filepath() { realpath "${1:-.}"; }
alias rm="sudo rm"
function /bin/rm { sudo /usr/bin/rm "$@" }
function /usr/bin/rm { sudo /usr/bin/rm "$@" }
alias rm_nvim="echo 'Removing Neovim data, cache, state, and lazy-lock.json...' ; command rm -rf ~/.local/share/nvim ~/.local/state/nvim ~/.cache/nvim ~/.config/nvim/lazy-lock.json ~/.var/app/io.neovim.nvim/cache/nvim ~/.var/app/io.neovim.nvim/data/nvim && echo 'Neovim reset complete! Restart nvim to reinstall plugins.'"
alias nvim_rm="echo 'Removing Neovim data, cache, state, and lazy-lock.json...' ; command rm -rf ~/.local/share/nvim ~/.local/state/nvim ~/.cache/nvim ~/.config/nvim/lazy-lock.json ~/.var/app/io.neovim.nvim/cache/nvim ~/.var/app/io.neovim.nvim/data/nvim && echo 'Neovim reset complete! Restart nvim to reinstall plugins.'"
# alias xsave="echo '$(xclip -selection clipboard -o)' >> ~/.clipboard_history ; cat ~/.clipboard_history"

# lazygit and lazydocker aliases
alias lg="lazygit"
alias ld="lazydocker"
alias dnf="sudo dnf"

# hlp (helper) build alias
alias hlp-build="(cd ~/.Dotfiles/scripts/helper-go && go build -o ~/go/bin/hlp ./cmd/hlp) && echo 'Built: ~/go/bin/hlp'"
alias hlp-install="cd ~/.Dotfiles/scripts/helper-go && go install ./cmd/hlp && echo 'Installed to: $(go env GOPATH)/bin/hlp'"

# Other exports
export MANPAGER='nvim +Man!'
export USE_BUILTIN_RIPGREP=1
# export MANWIDTH=999
# export JAVA_HOME="/usr/lib/jvm/java-11-openjdk/"
export JAVA_HOME="/usr/lib/jvm/java-21-openjdk"
# export JAVA_HOME="/usr/lib/jvm/jdk1.8.0_471"  # for branch-opener (needs Java 8)
export PATH="$JAVA_HOME/bin:$PATH"

export PATH="$HOME/bin/Sencha/Cmd:$PATH"
export PATH="$HOME/.cargo/bin:$PATH"
export PATH=$PATH:/usr/local/go/bin
export PATH="$PATH:$HOME/bin/.local/scripts"
export PATH="$HOME/.local/bin:$PATH"
export PATH="$HOME/.npm-global/bin:$PATH"

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
[[ -f "$HOME/.linuxbrew/bin/brew" ]] && eval "$("$HOME/.linuxbrew/bin/brew" shellenv)"
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_14:$LD_LIBRARY_PATH
export PATH=$LD_LIBRARY_PATH:$PATH
export PATH=/home/dariuszw/.opencode/bin:$PATH
alias opencode='infisical run --domain=http://localhost --projectId=fe560626-91f2-427b-a402-8473e61943ad --env=dev --path=/opencode --path=/opencode/mcp -- /home/dariuszw/.opencode/bin/opencode'
function hlp-token {
   local json token
   json=$(infisical run --domain=http://localhost --projectId=fe560626-91f2-427b-a402-8473e61943ad --env=dev --path=/hlp -- hlp token) || return 1
   token=$(printf '%s' "$json" | jq -r '.accessToken // empty')
   if [[ -z "$token" ]]; then
      echo "hlp-token: no accessToken in response: $json" >&2
      return 1
   fi
   export SUITE_API_TOKEN="$token"
   if [[ -t 1 ]]; then
      printf '%s\n' "$json" | jq .
      echo "✓ exported: SUITE_API_TOKEN=$SUITE_API_TOKEN" >&2
   else
      printf '%s\n' "$json"
   fi
}

# pnpm
export PNPM_HOME="$HOME/.local/share/pnpm"
case ":$PATH:" in
  *":$PNPM_HOME:"*) ;;
  *) export PATH="$PNPM_HOME:$PATH" ;;
esac
# pnpm end
export GOTOOLCHAIN=auto
export PATH=$PATH:$(go env GOPATH)/bin
command -v direnv >/dev/null 2>&1 && eval "$(direnv hook zsh)"

export PATH=$PATH:$HOME/.spicetify
setopt ignore_eof

export PYENV_ROOT="$HOME/.pyenv"
[[ -d $PYENV_ROOT/bin ]] && export PATH="$PYENV_ROOT/bin:$PATH"
command -v pyenv >/dev/null 2>&1 && eval "$(pyenv init --path)"
command -v pyenv >/dev/null 2>&1 && eval "$(pyenv init -)"
command -v pyenv >/dev/null 2>&1 && eval "$(pyenv virtualenv-init -)"

export BROWSER="google-chrome --profile-directory=Default"

# -----------------------------------------------------------------------------
# Git stale lock cleanup - runs before any git command
# -----------------------------------------------------------------------------
_git_cleanup_stale_lock() {
   local cmd="$1"
   # Only run for git commands
   [[ "$cmd" != git\ * ]] && return

   # Get git dir (works even in subdirectories)
   local git_dir
   git_dir=$(git rev-parse --git-dir 2>/dev/null) || return

   local lock_file="$git_dir/index.lock"
   [[ ! -f "$lock_file" ]] && return

   # Check if any process holds the lock
   if lsof "$lock_file" &>/dev/null; then
      echo "⚠️  Lock file in use by another process: $lock_file" >&2
      return
   fi

   # Show lock file info
   local lock_age
   lock_age=$(stat -c %Y "$lock_file" 2>/dev/null)
   local now=$(date +%s)
   local age_mins=$(( (now - lock_age) / 60 ))

   echo "" >&2
   echo "🔒 Stale git lock detected: $lock_file" >&2
   echo "   Age: ${age_mins} minutes (created: $(date -d @$lock_age '+%H:%M:%S'))" >&2
   echo "" >&2

   # Ask for confirmation
   echo -n "Remove stale lock and continue? [Y/n] " >&2
   read -r response
   if [[ "$response" =~ ^[Nn] ]]; then
      echo "Aborted. Lock file kept." >&2
      return 1
   fi

   rm -f "$lock_file"
   echo "🔓 Removed stale lock. Continuing..." >&2
}
autoload -Uz add-zsh-hook
add-zsh-hook preexec _git_cleanup_stale_lock

[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh
alias ga="git add \"\$@\" && git status"
export INFISICAL_API_URL="http://localhost"

infisical-run() {
   local compose_dir="$HOME/.infisical"
   local api_url="http://localhost/api/status"
   local max_wait=30

   echo "▶ Starting Infisical backend..."
   (cd "$compose_dir" && docker compose up -d backend 2>&1)

   echo "⏳ Waiting for Infisical API to be ready..."
   local i=0
   until curl -sf "$api_url" > /dev/null 2>&1; do
      if (( i >= max_wait )); then
         echo "✗ Infisical did not start within ${max_wait}s — check: docker logs infisical-backend"
         return 1
      fi
      sleep 1
      (( i++ ))
   done

   echo "✓ Infisical is ready (${i}s)"
   echo "→ Opening http://localhost in browser..."
   xdg-open "http://localhost" 2>/dev/null || echo "  Open manually: http://localhost"
}
alias infisical-stop='(cd ~/.infisical && docker compose stop backend) && echo "Infisical stopped"'
alias isync='~/.Dotfiles/scripts/infisical-sync'

# Oracle SQLcl for the Identity Builder oracle-sqlcl MCP server
export SQLCL_PATH=/home/dariuszw/.local/sqlcl/bin/sql

if command -v wt >/dev/null 2>&1; then eval "$(command wt config shell init zsh)"; fi

# >>> grok installer >>>
export PATH="$HOME/.grok/bin:$PATH"
fpath=(~/.grok/completions/zsh $fpath)
autoload -Uz compinit && compinit -C
# <<< grok installer <<<
export PATH="/opt/sqlcl/bin:$PATH"
