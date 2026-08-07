#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HARPOON="$CURRENT_DIR/scripts/harpoon.sh"

for i in 1 2 3 4 5 6 7 8 9; do
    tmux bind-key "$i" run-shell "$HARPOON go $i"
done

tmux bind-key E run-shell "$HARPOON add '#{session_name}' '#{session_path}'"

tmux bind-key -r J run-shell "$HARPOON next '#{session_name}'"
tmux bind-key -r K run-shell "$HARPOON prev '#{session_name}'"

tmux bind-key -r t run-shell "$HARPOON breakout '#{pane_current_path}' '#{pane_id}'"

tmux bind-key e display-popup -E -w 70% -h 60% -T ' harpoon ' "$HARPOON menu"
