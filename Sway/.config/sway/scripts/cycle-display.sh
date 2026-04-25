#!/bin/bash
set -euo pipefail

# Cycle display profiles via kanshi: Both -> External only -> Internal only -> Both
# Triggered by F7 (XF86Display).
#
# Adding a new external monitor: add three profiles in ~/.config/kanshi/config
# (docked_X / ext_only_X / int_only_X) AND extend make_to_family() below.

LAPTOP="eDP-1"

# Adwaita has no laptop glyph; fall through to breeze for the "internal only" icon.
LAPTOP_ICON="/usr/share/icons/breeze-dark/devices/24/computer-laptop-symbolic.svg"

command -v jq        >/dev/null || { echo "jq required" >&2; exit 1; }
command -v swaymsg   >/dev/null || { echo "swaymsg required" >&2; exit 1; }
command -v kanshictl >/dev/null || { echo "kanshictl required" >&2; exit 1; }

notify() {
    local icon="$1"; shift
    command -v notify-send >/dev/null && notify-send --icon="$icon" "$@" || echo "$*" >&2
}

# Map a monitor manufacturer to its kanshi profile family suffix.
make_to_family() {
    case "$1" in
        Samsung*) echo "samsung" ;;
        Iiyama*)  echo "iiyama"  ;;
        *)        echo ""        ;;
    esac
}

OUTPUTS=$(swaymsg -t get_outputs)

# Prefer the currently-active external; fall back to any non-eDP-1 entry.
EXT_JSON=$(echo "$OUTPUTS" | jq -c \
    "[.[] | select(.name != \"$LAPTOP\")] | (map(select(.active == true)) + .)[0] // empty")

if [[ -z "$EXT_JSON" ]]; then
    notify "dialog-warning-symbolic" "Display" "No external monitor connected"
    exit 0
fi

EXT_MAKE=$(echo "$EXT_JSON" | jq -r '.make')
EXT_ACTIVE=$(echo "$EXT_JSON" | jq -r '.active')
EDP_ACTIVE=$(echo "$OUTPUTS" | jq -r ".[] | select(.name == \"$LAPTOP\") | .active")

FAMILY=$(make_to_family "$EXT_MAKE")
if [[ -z "$FAMILY" ]]; then
    notify "dialog-error-symbolic" "Display" "Unknown external: $EXT_MAKE — add a profile and update cycle-display.sh"
    exit 1
fi

case "$EDP_ACTIVE,$EXT_ACTIVE" in
    true,true)
        kanshictl switch "ext_only_$FAMILY"
        notify "video-display-symbolic" "Display" "External only"
        ;;
    false,true)
        kanshictl switch "int_only_$FAMILY"
        notify "$LAPTOP_ICON" "Display" "Internal only"
        ;;
    true,false)
        kanshictl switch "docked_$FAMILY"
        notify "video-joined-displays-symbolic" "Display" "Both"
        ;;
    *)
        kanshictl switch "docked_$FAMILY"
        notify "video-joined-displays-symbolic" "Display" "Both (recovery)"
        ;;
esac
