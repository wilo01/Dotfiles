#!/bin/bash
# Resize PIP window or adjust gaps if not in PIP mode
# Usage: pip-resize.sh [grow|shrink]

PIPMARK="pip_window"
STATE_FILE="/tmp/sway-pip-state"
NOTIFY_ID=9999
MARGIN=10
WAYBAR_HEIGHT=30
RESIZE_STEP=50
MIN_WIDTH=200
MIN_HEIGHT=113  # 16:9 ratio: 200 * 9/16 = 112.5
MAX_WIDTH=1280
MAX_HEIGHT=720  # 16:9 ratio (720p)

notify() {
    notify-send -r $NOTIFY_ID -t 1000 "PiP" "$1"
}

get_window_info() {
    swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | "\(.rect.width) \(.rect.height) \(.rect.x) \(.rect.y)"'
}

get_output_for_window() {
    local win_x win_y
    read _ _ win_x win_y <<< "$(get_window_info)"
    swaymsg -t get_outputs | jq -r --argjson wx "${win_x:-0}" --argjson wy "${win_y:-0}" '
        .[] | select(.active) |
        select(.rect.x <= $wx and $wx < (.rect.x + .rect.width)) |
        select(.rect.y <= $wy and $wy < (.rect.y + .rect.height)) |
        "\(.rect.width) \(.rect.height)"
    ' | head -1
}

calc_position() {
    local corner=$1 win_w=$2 win_h=$3
    read out_w out_h <<< "$(get_output_for_window)"
    out_w=${out_w:-1920}
    out_h=${out_h:-1080}
    local usable_h=$((out_h - WAYBAR_HEIGHT))

    case "$corner" in
        br) echo "$((out_w - win_w - MARGIN)) $((usable_h - win_h - MARGIN))";;
        tr) echo "$((out_w - win_w - MARGIN)) $MARGIN";;
        tl) echo "$MARGIN $MARGIN";;
        bl) echo "$MARGIN $((usable_h - win_h - MARGIN))";;
    esac
}

# Check if focused window is PIP
is_pip=$(swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | .marks // [] | contains(["'$PIPMARK'"])')

if [[ "$is_pip" != "true" ]]; then
    # Not PIP - fallback to gap adjustment
    case "$1" in
        grow)   swaymsg "gaps inner current plus 5" ;;
        shrink) swaymsg "gaps inner current minus 5" ;;
    esac
    exit 0
fi

# Load current state (|| true handles malformed files)
[ -f "$STATE_FILE" ] && source "$STATE_FILE" || true
corner=${corner:-br}

# Get current dimensions
read cur_w cur_h _ _ <<< "$(get_window_info)"

# Calculate new dimensions (16:9 aspect ratio)
case "$1" in
    grow)
        new_w=$((cur_w + RESIZE_STEP))
        new_h=$((new_w * 9 / 16))
        # Enforce maximum size
        if [ $new_w -gt $MAX_WIDTH ]; then
            new_w=$MAX_WIDTH
            new_h=$MAX_HEIGHT
        fi
        ;;
    shrink)
        new_w=$((cur_w - RESIZE_STEP))
        new_h=$((new_w * 9 / 16))
        # Enforce minimum size
        if [ $new_w -lt $MIN_WIDTH ]; then
            new_w=$MIN_WIDTH
            new_h=$MIN_HEIGHT
        fi
        ;;
    *)
        echo "Usage: $0 [grow|shrink]" >&2
        exit 1
        ;;
esac

# Apply resize
swaymsg "resize set $new_w $new_h"

# Reposition to maintain corner placement
sleep 0.1
pos=$(calc_position "$corner" "$new_w" "$new_h")
swaymsg "move position $pos"

# Update state file
echo "corner=$corner" > "$STATE_FILE"
echo "pip_width=$new_w" >> "$STATE_FILE"
echo "pip_height=$new_h" >> "$STATE_FILE"

notify "${new_w}x${new_h}"
