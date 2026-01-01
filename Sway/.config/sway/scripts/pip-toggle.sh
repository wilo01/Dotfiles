#!/bin/bash
# Toggle PiP mode for focused window
# Cycles through corners: BR → TR → TL → BL → exit

PIPMARK="pip_window"
STATE_FILE="/tmp/sway-pip-state"
NOTIFY_ID=9999
MARGIN=10
WAYBAR_HEIGHT=30  # Adjust if your waybar is different height

notify() {
    notify-send -r $NOTIFY_ID -t 1500 "PiP" "$1"
}

get_window_info() {
    swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | "\(.rect.width) \(.rect.height) \(.rect.x) \(.rect.y)"'
}

get_output_for_window() {
    # Get the output name where the focused window is located
    local win_x win_y
    read _ _ win_x win_y <<< "$(get_window_info)"

    # Find output that contains this position
    swaymsg -t get_outputs | jq -r --argjson wx "${win_x:-0}" --argjson wy "${win_y:-0}" '
        .[] | select(.active) |
        select(.rect.x <= $wx and $wx < (.rect.x + .rect.width)) |
        select(.rect.y <= $wy and $wy < (.rect.y + .rect.height)) |
        "\(.rect.width) \(.rect.height)"
    ' | head -1
}

# Calculate corner position (keeps window fully on-screen)
calc_position() {
    local corner=$1
    read win_w win_h _ _ <<< "$(get_window_info)"
    read out_w out_h <<< "$(get_output_for_window)"

    # Fallback if output detection fails
    out_w=${out_w:-1920}
    out_h=${out_h:-1080}

    # Account for waybar at bottom
    local usable_h=$((out_h - WAYBAR_HEIGHT))

    case "$corner" in
        br) echo "$((out_w - win_w - MARGIN)) $((usable_h - win_h - MARGIN))";;
        tr) echo "$((out_w - win_w - MARGIN)) $MARGIN";;
        tl) echo "$MARGIN $MARGIN";;
        bl) echo "$MARGIN $((usable_h - win_h - MARGIN))";;
    esac
}

is_pip=$(swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | .marks // [] | contains(["'$PIPMARK'"])')

if [ "$is_pip" = "true" ]; then
    if [ -f "$STATE_FILE" ]; then
        source "$STATE_FILE"
    else
        corner="br"
    fi

    case "$corner" in
        br) next="tr";;
        tr) next="tl";;
        tl) next="bl";;
        bl) next="exit";;
    esac

    if [ "$next" = "exit" ]; then
        read width height _ _ <<< "$(get_window_info)"
        echo "pip_width=$width" > "$STATE_FILE"
        echo "pip_height=$height" >> "$STATE_FILE"
        swaymsg "unmark $PIPMARK; floating disable; sticky disable"
        notify "Disabled"
    else
        pos=$(calc_position "$next")
        echo "corner=$next" > "$STATE_FILE"
        [ -n "$pip_width" ] && echo "pip_width=$pip_width" >> "$STATE_FILE"
        [ -n "$pip_height" ] && echo "pip_height=$pip_height" >> "$STATE_FILE"
        swaymsg "move position $pos"
        notify "→ ${next^^}"
    fi
else
    if [ -f "$STATE_FILE" ]; then
        source "$STATE_FILE"
    fi

    width=${pip_width:-400}
    height=${pip_height:-300}

    echo "corner=br" > "$STATE_FILE"
    echo "pip_width=$width" >> "$STATE_FILE"
    echo "pip_height=$height" >> "$STATE_FILE"

    swaymsg "mark --add $PIPMARK; floating enable; sticky enable; resize set $width $height"
    # Position after resize so we have correct dimensions
    sleep 0.1
    pos=$(calc_position "br")
    swaymsg "move position $pos"
    notify "Enabled (↘)"
fi
