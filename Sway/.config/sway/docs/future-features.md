# Sway Enhancement Plans - Saved for Later Implementation

These are 4 features planned but not yet implemented. Each section is self-contained and can be implemented independently.

---

## Feature 1: Quick Output Swap (`Super+o`)

**Purpose**: Move current workspace to the other monitor instantly.

### Implementation

**File to modify**: `~/.config/sway/config`

**Script to create**: `~/.config/sway/scripts/swap-output.sh`

```bash
#!/bin/bash
# Move current workspace to the other output

LAPTOP="eDP-1"
CURRENT_OUTPUT=$(swaymsg -t get_workspaces | jq -r '.[] | select(.focused) | .output')

# Find the other output
if [ "$CURRENT_OUTPUT" = "$LAPTOP" ]; then
    # Find external monitor
    TARGET=$(swaymsg -t get_outputs | jq -r '.[] | select(.name != "eDP-1" and .active == true) | .name' | head -1)
else
    TARGET="$LAPTOP"
fi

if [ -n "$TARGET" ]; then
    swaymsg move workspace to output "$TARGET"
    notify-send "Workspace" "Moved to $TARGET"
else
    notify-send "Workspace" "No other output available"
fi
```

**Config addition**:
```
bindsym $mod+o exec ~/.config/sway/scripts/swap-output.sh
```

### Keybinding
| Shortcut | Action |
|----------|--------|
| `Super+o` | Move current workspace to other monitor |

---

## Feature 2: Quick Launcher with Leader Key (`Super+`` then 1-6)

**Purpose**: Open apps on specific workspaces using a two-key sequence.

### Implementation

**File to modify**: `~/.config/sway/config`

**Approach**: Use Sway's mode system for leader-key behavior.

**Config addition**:
```
# Quick launcher mode (leader key: Super+`)
mode "launcher" {
    # 1: Terminal on workspace 1
    bindsym 1 exec $term; workspace 1; mode "default"

    # 2: Code on workspace 1 (your existing behavior)
    bindsym 2 exec code; mode "default"

    # 3: Thunar on floating
    bindsym 3 exec thunar; mode "default"

    # 4: Spotify on workspace 20
    bindsym 4 exec flatpak run com.spotify.Client; workspace 20; mode "default"

    # 5: Chrome on workspace 5
    bindsym 5 exec google-chrome; workspace 5; mode "default"

    # 6: KeePass on workspace 19
    bindsym 6 exec keepassxc; workspace 19; mode "default"

    # Cancel with Escape or another Super+`
    bindsym Escape mode "default"
    bindsym grave mode "default"
    bindsym Return mode "default"
}

# Enter launcher mode with Super+`
bindsym $mod+grave mode "launcher"
```

**Note**: This replaces your current `cycle-focus.sh` binding on `$mod+grave`. You may want to move cycle-focus to a different key like `$mod+Tab+grave` or similar.

### Keybindings
| Sequence | Action |
|----------|--------|
| `Super+`` then `1` | Terminal → workspace 1 |
| `Super+`` then `2` | VS Code |
| `Super+`` then `3` | Thunar |
| `Super+`` then `4` | Spotify → workspace 20 |
| `Super+`` then `5` | Chrome → workspace 5 |
| `Super+`` then `6` | KeePass → workspace 19 |
| `Super+`` then `Esc` | Cancel |

### Customization
Edit the mode block to add/change apps. Format:
```
bindsym <key> exec <app>; [workspace N;] mode "default"
```

---

## Feature 3: Auto-lock on Lid Close

**Purpose**: Lock screen when laptop lid closes, even with external monitor connected.

### Implementation

**Approach**: Use systemd-logind + sway event handling.

**Option A: Systemd logind.conf (system-wide)**

Edit `/etc/systemd/logind.conf`:
```ini
HandleLidSwitch=lock
HandleLidSwitchExternalPower=lock
HandleLidSwitchDocked=lock
```

Then create a lock handler script.

**Option B: Sway bindswitch (recommended)**

**Config addition**:
```
# Lock on lid close (works even with external monitor)
bindswitch --reload --locked lid:on exec swaylock -i $default_bg -f
```

This uses Sway's native switch binding which:
- Works regardless of external monitor
- Respects your existing swaylock config
- `--reload` makes it work after config reload
- `--locked` allows it to trigger even when already locked

### Additional Enhancement

To also turn off the laptop screen on lid close:
```
bindswitch --reload lid:on output eDP-1 disable
bindswitch --reload lid:off output eDP-1 enable
```

### Complete Config Block
```
# Lid close handling
bindswitch --reload --locked lid:on exec swaylock -i $default_bg -f
bindswitch --reload lid:on output eDP-1 disable
bindswitch --reload lid:off output eDP-1 enable
```

---

## Feature 4: PiP Window (Picture-in-Picture)

**Purpose**: Toggle any window into a small floating sticky window that cycles through corners.

### Implementation

**Script to create**: `~/.config/sway/scripts/pip-toggle.sh`

```bash
#!/bin/bash
# Toggle PiP mode for focused window
# Cycles through corners: BR → TR → TL → BL → BR...

PIPMARK="pip_window"
STATE_FILE="/tmp/sway-pip-corner"

# Screen dimensions (adjust for your setup or detect dynamically)
get_output_geometry() {
    swaymsg -t get_outputs | jq -r '.[] | select(.focused) | "\(.rect.width) \(.rect.height)"'
}

# Check if current window has PiP mark
is_pip=$(swaymsg -t get_tree | jq -r '.. | select(.focused? == true) | .marks // [] | contains(["'$PIPMARK'"])')

if [ "$is_pip" = "true" ]; then
    # Already PiP - cycle corner or exit
    current_corner=$(cat "$STATE_FILE" 2>/dev/null || echo "br")

    case "$current_corner" in
        br) next="tr"; pos="80ppt 0";;
        tr) next="tl"; pos="0 0";;
        tl) next="bl"; pos="0 80ppt";;
        bl) next="exit";;
    esac

    if [ "$next" = "exit" ]; then
        # Exit PiP mode
        swaymsg "unmark $PIPMARK; floating disable; sticky disable"
        rm -f "$STATE_FILE"
        notify-send "PiP" "Disabled"
    else
        # Move to next corner
        echo "$next" > "$STATE_FILE"
        swaymsg "move position $pos"
        notify-send "PiP" "Corner: $next"
    fi
else
    # Enter PiP mode
    echo "br" > "$STATE_FILE"
    swaymsg "mark --add $PIPMARK; floating enable; sticky enable; resize set 400 300; move position 80ppt 80ppt"
    notify-send "PiP" "Enabled (bottom-right)"
fi
```

**Config addition**:
```
# PiP toggle (cycles corners, then exits)
bindsym $mod+p exec ~/.config/sway/scripts/pip-toggle.sh
```

### Keybinding
| Shortcut | Action |
|----------|--------|
| `Super+p` (1st) | Enter PiP mode (bottom-right) |
| `Super+p` (2nd) | Move to top-right |
| `Super+p` (3rd) | Move to top-left |
| `Super+p` (4th) | Move to bottom-left |
| `Super+p` (5th) | Exit PiP mode |

### PiP Window Properties
- **Size**: 400x300 pixels (adjustable in script)
- **Sticky**: Follows you across workspaces
- **Floating**: Doesn't affect tiling layout
- **Marked**: Uses sway marks for state tracking

---

## Quick Reference - All New Keybindings

| Shortcut | Feature | Action |
|----------|---------|--------|
| `Super+o` | Output Swap | Move workspace to other monitor |
| `Super+`` | Quick Launch | Enter launcher mode |
| `Super+p` | PiP | Toggle/cycle PiP window |
| (lid close) | Auto-lock | Lock screen automatically |

---

## Implementation Order (Recommended)

1. **Auto-lock** - Single config line, no script needed
2. **Output Swap** - Simple script, high utility
3. **PiP Window** - Medium complexity script
4. **Quick Launcher** - Requires binding conflict resolution

---

## Files Summary

| File | Action | Feature |
|------|--------|---------|
| `~/.config/sway/config` | Modify | All features |
| `~/.config/sway/scripts/swap-output.sh` | Create | Output Swap |
| `~/.config/sway/scripts/pip-toggle.sh` | Create | PiP Window |

---

## Status

- [ ] Feature 1: Quick Output Swap
- [ ] Feature 2: Quick Launcher
- [ ] Feature 3: Auto-lock on Lid Close
- [x] Feature 4: PiP Window

*Last updated: 2024-12-31*
