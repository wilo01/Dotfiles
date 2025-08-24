#!/bin/bash
# Launch script for nvim-ai with proper environment

# Disable askpass dialogs that are causing GTK errors
export GIT_ASKPASS=""
export SSH_ASKPASS=""

# Launch Neovim with nvim-ai config
NVIM_APPNAME=nvim-ai nvim "$@"