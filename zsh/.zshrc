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
alias cdy="~/bin/.local/scripts/tmux-sessionizer ~/Dev/branch-opener/branches/tds-suite/test/Cypress"
# CODE actions
alias code_ks="echo code ~/Dev/branch-opener/branches/tds-suite/source/ui-kiosk/app/global/Settings.js ; code ~/Dev/branch-opener/branches/tds-suite/source/ui-kiosk/app/global/Settings.js"
alias code_ka="echo code ~/Dev/branch-opener/branches/tds-suite/source/ui-kiosk/app/Application.js ; code ~/Dev/branch-opener/branches/tds-suite/source/ui-kiosk/app/Application.js"
alias code_ksc="echo code ~/Dev/branch-opener/branches/tds-suite/source/ui-kiosk/app/view/settings/SettingsController.js ; code ~/Dev/branch-opener/branches/tds-suite/source/ui-kiosk/app/view/settings/SettingsController.js"
alias kiosk_settings="echo open kiosk settings at: ; code_ks ; sleep 1 ; code_ka ; sleep 1 ; code_ksc ;"
alias liqui_valid="echo cd ~/Dev/branch-opener/branches/tds-suite/source/server/database/ ; echo ./liquibase --defaultsFile=validate.liquibase.properties validate ; cd ~/Dev/branch-opener/branches/tds-suite/source/server/database/ ; ./liquibase --defaultsFile=validate.liquibase.properties validate"
alias sqldev="echo ~/SQLDeveloper/opt/sqldeveloper/sqldeveloper.sh ; ~/SQLDeveloper/opt/sqldeveloper/sqldeveloper.sh"
alias br='echo npm start at: ; echo ~/Dev/branch-opener/app/ ; if [[ -n "$(find ~/Dev/branch-opener/app/apex/kiosk/bdb/ -maxdepth 0 -type f -o -type d -printf '%s')" ]]; then echo "Removing content from ~/Dev/branch-opener/app/apex/kiosk/bdb/" ; rm -rf ~/Dev/branch-opener/app/apex/kiosk/bdb/* ; else echo "No content found in ~/Dev/branch-opener/app/apex/kiosk/bdb/, skipping removal." ; fi ; ls ~/Dev/branch-opener/app/apex/kiosk/bdb/ ; cd ~/Dev/branch-opener/app/ ; sleep 1 ; xdg-open http://localhost:3333/static/ ; npm start'
function hx() {
    local variant="${1:-dev}"
    pkill -f "pnpm run dev"
    echo "pnpm start at: ~/tds-hexer/"
    echo "Running: pnpm run $variant"
    cd ~/tds-hexer/
    pnpm install
    pnpm run "$variant"
    xdg-open http://localhost:3005/safe/
}
alias ksw='echo kiosk start at: ; echo ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/ ; cd ~/tds-branch-opener/branches/tds-suite/source/ui-kiosk/ ; sencha app watch ; xdg-open http://localhost:3005/kiosk/'
alias vsw='echo visitor-web-app start at: ; echo ~/tds-branch-opener/branches/tds-visitor-web-app/ui ; cd ~/tds-branch-opener/branches/tds-visitor-web-app/ui ; sencha app watch ; xdg-open http://localhost:3005/kiosk/'
alias cy="echo Cypress open at: ; cdy ; sleep 1 ; echo ./node_modules/cypress/bin/cypress open ; ./node_modules/cypress/bin/cypress open"
alias cy_all="echo Cypress run all tests at: ; cdy ; sleep 1 ; echo npx cypress run --headless --spec cypress/integration/tdsvisitor/rt/*.js ; npx cypress run --headless --spec cypress/integration/tdsvisitor/rt/*.js"
alias docker_start_trunk="echo cd ~/Dev/branch-opener/branches/safe ; echo sudo docker start -ai trunk ; cd ~/Dev/branch-opener/branches/safe && sudo docker start -ai trunk"
alias liquibaseLocalDockerUpdate="echo cd ~/Dev/branch-opener/branches/safe ; echo npm run liquibaseLocalDockerUpdate ; cd ~/Dev/branch-opener/branches/safe && npm run liquibaseLocalDockerUpdate"
alias npmliquibaseLocalDockerUpdate="echo cd ~/Dev/branch-opener/branches/safe ; echo npm run liquibaseLocalDockerUpdate ; cd ~/Dev/branch-opener/branches/safe && npm run liquibaseLocalDockerUpdate"
alias apex_remove="echo rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/* ; rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/*"
alias remove_apex="echo rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/* ; rm -rf ~/Dev/branch-opener/app/apex/backoffice/bdb/*"
alias apex_zip="echo zip -r rt.zip ~/Dev/branch-opener/branches/tds-suite/source/server/rt/* ; zip -r rt.zip ~/Dev/branch-opener/branches/tds-suite/source/server/rt/* && "
alias zip_apex="echo zip -r rt.zip ~/Dev/branch-opener/branches/tds-suite/source/server/rt/* ; zip -r rt.zip ~/Dev/branch-opener/branches/tds-suite/source/server/rt/* && "
# alias csp_hash="echo sha256-$(echo -n "$(xclip -o)" | openssl sha256 -binary | openssl base64)"
# Git
alias git_lens="git log --graph --oneline --decorate ; echo git log --graph --oneline --decorate"
alias git_graph="git log --graph --oneline --decorate ; echo git log --graph --oneline --decorate"
alias git_last="git log -1 --stat ; echo git log -1 --stat"
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
alias filepath='realpath'
alias rm="sudo rm"
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
export PATH="$JAVA_HOME/bin:$PATH"

export PATH="$HOME/bin/Sencha/Cmd:$PATH"
export PATH="$HOME/.cargo/bin:$PATH"
export PATH=$PATH:/usr/local/go/bin
export PATH="$PATH:$HOME/bin/.local/scripts"
export PATH="$HOME/.local/bin:$PATH"

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
[[ -f "$HOME/.linuxbrew/bin/brew" ]] && eval "$("$HOME/.linuxbrew/bin/brew" shellenv)"
export LD_LIBRARY_PATH=/opt/oracle/instantclient_21_14:$LD_LIBRARY_PATH
export PATH=$LD_LIBRARY_PATH:$PATH

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
