#!/bin/bash
set -euo pipefail

# Watch for output changes and trigger workspace swap
# Runs in background, started by sway config

SCRIPT_DIR="$(dirname "$(realpath "$0")")"
LOCKFILE="/tmp/sway-monitor-watcher.lock"

# Single instance via flock (kernel-managed, no stale PID issues)
exec 200>"$LOCKFILE"
flock -n 200 || { echo "Already running" >&2; exit 0; }

# Clean exit on signals
cleanup() { exit 0; }
trap cleanup SIGTERM SIGINT SIGPIPE

# Event loop with improved debounce (process substitution = single bash process)
while read -r event; do  # [ ] TODO: event appears unused. Verify use (or export if used externally).; event appears unused. Verify use (or export if used externally).
    sleep 1
    # Drain events accumulated during debounce
    while read -r -t 0.1 _; do :; done
    "$SCRIPT_DIR/swap-workspaces.sh" || echo "Swap failed" >&2
done < <(swaymsg -t subscribe '["output"]')
