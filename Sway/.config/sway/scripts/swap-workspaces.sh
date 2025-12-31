#!/bin/bash
set -euo pipefail

# Swap workspaces between laptop (eDP-1) and external monitor
# WS 1-7 → External monitor, WS 8-20 → Laptop

LAPTOP="eDP-1"

# Dependency checks
command -v jq >/dev/null || { echo "jq required" >&2; exit 1; }
command -v swaymsg >/dev/null || { echo "swaymsg required" >&2; exit 1; }

# Notification helper (graceful fallback if notify-send missing)
notify() {
    command -v notify-send >/dev/null && notify-send "$@" || echo "$*" >&2
}

# Find external monitor (not eDP-1, must be active)
EXTERNAL=$(swaymsg -t get_outputs | jq -r '.[] | select(.name != "eDP-1" and .active == true) | .name' | head -1)

if [[ -z "$EXTERNAL" ]]; then
    notify "Swap Workspaces" "No external monitor connected"
    exit 0
fi

# Move workspaces 1-7 to external monitor (|| true for non-existent workspaces)
for ws in {1..7}; do
    swaymsg "[workspace=$ws]" move workspace to output "$EXTERNAL" 2>/dev/null || true
done

# Move workspaces 8-20 to laptop
for ws in {8..20}; do
    swaymsg "[workspace=$ws]" move workspace to output "$LAPTOP" 2>/dev/null || true
done

notify "Swap Workspaces" "WS 1-7 → $EXTERNAL, WS 8-20 → $LAPTOP"
