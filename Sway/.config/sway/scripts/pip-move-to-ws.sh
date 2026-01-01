#!/bin/bash
# PiP-aware workspace move - resets PiP before moving container
# Usage: pip-move-to-ws.sh <workspace> [output]

WORKSPACE=$1
OUTPUT=$2
PIPMARK="pip_window"
STATE_FILE="/tmp/sway-pip-state"

# Check if focused window has PiP mark
is_pip=$(swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | .marks // [] | contains(["'$PIPMARK'"])')

if [ "$is_pip" = "true" ]; then
    # Get current window size before resetting
    read cur_w cur_h <<< "$(swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | "\(.rect.width) \(.rect.height)"')"

    # Reset PiP
    swaymsg "unmark $PIPMARK; floating disable; sticky disable"

    # Preserve size in state file (only clear corner)
    echo "pip_width=${cur_w:-400}" > "$STATE_FILE"
    echo "pip_height=${cur_h:-300}" >> "$STATE_FILE"

    notify-send -r 9999 -t 1500 "PiP" "Reset (moved to WS)"
fi

# Perform the workspace move
swaymsg "move container to workspace $WORKSPACE; workspace $WORKSPACE"

# Optionally move workspace to specific output (for WS 9-20)
if [ -n "$OUTPUT" ]; then
    swaymsg "move workspace to output $OUTPUT"
fi
