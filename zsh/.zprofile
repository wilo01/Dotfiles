if [[ "$OSTYPE" == darwin* ]]; then
  eval "$(/opt/homebrew/bin/brew shellenv)"
else
  # prepend ~/bin and ~/.local/bin to $PATH unless already there
  [[ "$PATH" =~ "$HOME/bin" ]]         || PATH="$HOME/bin:$PATH"
  [[ "$PATH" =~ "$HOME/.local/bin:" ]] || PATH="$HOME/.local/bin:$PATH"
  export PATH

  [[ -f "$HOME/.go/env" ]] && source "$HOME/.go/env"
fi
