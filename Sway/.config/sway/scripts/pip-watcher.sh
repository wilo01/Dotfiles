#!/bin/bash
# Watch for PiP window workspace changes and reset PiP mode

PIPMARK="pip_window"
STATE_FILE="/tmp/sway-pip-state"
LOCK_FILE="/tmp/sway-pip-watcher.lock"

# Ensure only one instance runs
exec 200>"$LOCK_FILE"
flock -n 200 || exit 0

# Track PiP window's workspace
pip_workspace=""

# Find workspace containing PiP-marked window
get_pip_workspace() {
    swaymsg -t get_tree | jq -r '
        .nodes[].nodes[] |
        select(.. | .marks? // [] | contains(["'"$PIPMARK"'"])) |
        .name
    ' 2>/dev/null | head -1
}

# Find workspace for a specific container ID
get_workspace_for_container() {
    local cid=$1
    swaymsg -t get_tree | jq -r --argjson cid "$cid" '
        .nodes[].nodes[] |
        select(.. | .id? == $cid) |
        .name
    ' 2>/dev/null | head -1
}

# Subscribe to window events
swaymsg -t subscribe '["window"]' | while read -r event; do
    change=$(echo "$event" | jq -r '.change // empty')
    has_mark=$(echo "$event" | jq -r '.container.marks // [] | contains(["'"$PIPMARK"'"])')

    if [ "$has_mark" = "true" ] && [ "$change" = "move" ]; then
        container_id=$(echo "$event" | jq -r '.container.id')
        new_workspace=$(get_workspace_for_container "$container_id")

        # Only reset if workspace actually changed
        if [ -n "$pip_workspace" ] && [ -n "$new_workspace" ] && [ "$pip_workspace" != "$new_workspace" ]; then
            swaymsg "[con_id=$container_id] unmark $PIPMARK; floating disable; sticky disable"
            rm -f "$STATE_FILE"
            notify-send -r 9999 -t 1500 "PiP" "Reset (moved to WS)"
            pip_workspace=""
        fi
    elif [ "$has_mark" = "true" ] && [ "$change" = "mark" ]; then
        # Window just got PiP mark - track its workspace
        pip_workspace=$(get_pip_workspace)
    fi
done
