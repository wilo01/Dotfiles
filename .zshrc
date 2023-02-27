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
plugins=(git zsh-autosuggestions zsh-syntax-highlighting evalcache git-extras debian tmux screen history extract colorize web-search docker)

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
# if [[ -n $SSH_CONNECTION ]]; then
#   export EDITOR='vim'
# else
#   export EDITOR='mvim'
# fi

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

alias LS="echo la -lha -F --show-control-chars --time-style=locale --color=auto ; la -lha -F --show-control-chars --time-style=locale --color=auto"
alias ls="echo la -lha -F --show-control-chars --time-style=locale --color=auto ; la -lha -F --show-control-chars --time-style=locale --color=auto"
alias CD="cd"
alias sshkey="echo cat ~/.ssh/id_ed25519.pub ; cat ~/.ssh/id_ed25519.pub"
alias cl="clear"
alias cds="echo cd ~/Dev/branch-opener/branches/safe ; cd ~/Dev/branch-opener/branches/safe"
alias cdb="echo cd ~/Dev/branch-opener/app/ ; cd ~/Dev/branch-opener/app/"
alias cdy="echo cd ~/Dev/branch-opener/branches/safe/test/Cypress ; cd ~/Dev/branch-opener/branches/safe/test/Cypress"
alias sqldev="echo ~/SQLDeveloper/opt/sqldeveloper/sqldeveloper.sh ; ~/SQLDeveloper/opt/sqldeveloper/sqldeveloper.sh" 
alias gnome-terminal='gnome-terminal --full-screen'
alias br="echo npm start at: ; cdb ; sleep 3 ; killall node ; xdg-open http://localhost:3333/static/ ; npm start"
alias bo="echo npm start at: ; cdb ; sleep 3 ; killall node ; xdg-open http://localhost:3333/static/ ; npm start"
alias BR="echo npm start at: ; cdb ; sleep 3 ; killall node ; xdg-open http://localhost:3333/static/ ; npm start"
alias BO="echo npm start at: ; cdb ; sleep 3 ; killall node ; xdg-open http://localhost:3333/static/ ; npm start"
alias cy="echo npx cypress open at: ; cdy ; sleep 3 ; npx cypress open"
alias git_lens="git log --graph --oneline --decorate ; echo git log --graph --oneline --decorate"
alias git_graph="git log --graph --oneline --decorate ; echo git log --graph --oneline --decorate"
alias git_last="git log -1 --stat ; echo git log -1 --stat"
alias git_stash="echo git stash ; echo Git stash changes; git stash"
alias git_apply="echo git stash apply ; echo Git stash apply last changes; git stash apply"
alias git_drop="echo git stash clear ; echo Clear all stash container ; git stash clear"
alias git_clear="echo git stash clear ; echo Clear all stash container ; git stash clear"
alias git_pop="echo git stash pop ; echo Apply last stash ; git stash pop"
alias open="echo xdg-open; xdg-open"
alias liqui_valid="echo cd ~/Dev/branch-opener/branches/safe/source/server/database/ ; echo ./liquibase --defaultsFile=validate.liquibase.properties validate ; cd ~/branch-opener/branches/safe/source/server/database/ ; ./liquibase --defaultsFile=validate.liquibase.properties validate"
alias zshrc="echo sudo nvim ~/.zshrc ; sudo nvim ~/.zshrc "

export PATH="/usr/lib/jvm/java-8-openjdk-amd64/bin:$PATH"
export PATH="/home/dariusz/bin/Sencha/Cmd:$PATH"