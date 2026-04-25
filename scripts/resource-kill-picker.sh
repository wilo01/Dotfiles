#!/usr/bin/env bash
# resource-kill-picker.sh — show top-15 processes by RSS or %CPU in a rofi
# menu and kill the chosen one. Bound to Sway hotkeys ($mod+Shift+m for
# memory, $mod+Shift+x for cpu); paired with pressure-watch.service which
# fires the notification that reminds you these hotkeys exist.
#
# Usage:
#   resource-kill-picker.sh [mem|cpu]
#   resource-kill-picker.sh --sort=mem
#   resource-kill-picker.sh --sort=cpu
#
# rofi keys:
#   Enter        → SIGTERM (graceful)
#   Alt+Enter    → SIGKILL (immediate)
#   Esc          → cancel

set -euo pipefail

SORT="${1:-mem}"
case "$SORT" in --sort=*) SORT="${SORT#--sort=}";; esac

case "$SORT" in
    mem) sort_key=-rss  ; label="memory" ;;
    cpu) sort_key=-pcpu ; label="cpu"    ;;
    *)   echo "usage: $0 [mem|cpu]" >&2 ; exit 2 ;;
esac

# Build rofi rows. Format:
#   <comm padded>  <rss>G  <cpu>%   pid=<pid>   -- <args>
# RSS is reported in KiB by ps; convert to GiB. Args trimmed to ARGS_MAX
# characters (default 200) — large enough to fit most realistic command
# lines without horizontal scroll, since rofi does not scroll.
ARGS_MAX=${PICKER_ARGS_MAX:-200}
rows=$(ps -eo pid=,rss=,pcpu=,comm=,args= --sort="$sort_key" --no-headers \
    | awk -v amax="$ARGS_MAX" 'NR<=15 {
        pid=$1; rss=$2/1048576; cpu=$3; comm=$4;
        args="";
        for (i=5; i<=NF; i++) args = args " " $i;
        if (length(args) > amax) args = substr(args,1,amax-3) "...";
        printf "%-22s %6.2fG %5s%%   pid=%d  --%s\n", comm, rss, cpu, pid, args;
    }')

# rofi window width: percent of screen. Override with PICKER_WIDTH=70 etc.
PICKER_WIDTH=${PICKER_WIDTH:-90}

choice=$(printf '%s\n' "$rows" \
    | rofi -dmenu -i \
           -p "kill ($label)" \
           -theme-str "window { width: ${PICKER_WIDTH}%; } listview { columns: 1; } element-text { font: \"monospace 11\"; }" \
           -kb-custom-1 "Alt+Return" \
           -mesg "Enter = SIGTERM | Alt+Enter = SIGKILL") || rc=$?
rc=${rc:-0}

# rofi exit codes: 0 = Enter, 1 = Esc/dismiss, 10 = -kb-custom-1 (Alt+Return)
[ "$rc" = 1 ] && exit 0
[ -z "$choice" ] && exit 0

pid=$(awk '{ for (i=1; i<=NF; i++) if ($i ~ /^pid=/) { sub("pid=","",$i); print $i; exit } }' <<<"$choice")
name=$(awk '{print $1}' <<<"$choice")

# ─────────────────────────────────────────────────────────────────────────────
# YOUR CODE — decide signal + escalation + safety policy.
#
# Inputs:
#   rc    — rofi exit code: 0 = Enter pressed, 10 = Alt+Enter (custom-1)
#   pid   — selected pid (integer)
#   name  — process comm string (for the confirmation toast)
#
# Trade-offs to weigh:
#   • Always TERM, never KILL: safest (lets process clean up), but a hung
#     process will ignore TERM and you'll have to escalate manually.
#   • Honour rc strictly: Enter → SIGTERM, Alt+Enter → SIGKILL. Matches
#     what the rofi prompt promised the user.
#   • Escalation: send TERM, sleep N, if still alive send KILL. Most "correct"
#     behaviour but slower and adds a sleep+recheck cycle.
#   • Critical-process guard: refuse to KILL `sway`, `pipewire`, `Xwayland`,
#     `swaync`, etc. — a misclick on those nukes your session.
#
# After you kill, fire a confirmation notification so you know it landed:
#   notify-send -a resource-kill-picker -t 3000 \
#     "killed $name" "pid=$pid signal=TERM/KILL"
#
# 5–10 lines total.
# ─────────────────────────────────────────────────────────────────────────────
do_kill() {
    local rc="$1" pid="$2" name="$3"
    local sig="TERM"
    [ "$rc" = 10 ] && sig="KILL"

    # Refuse SIGKILL on session-critical processes — a misclick on these
    # nukes your Sway session. SIGTERM is still allowed (most of these
    # ignore TERM anyway, so it's harmless).
    case "$sig:$name" in
        KILL:sway|KILL:swaync|KILL:pipewire|KILL:wireplumber|KILL:Xwayland|KILL:dbus-daemon)
            notify-send -a resource-kill-picker -u critical -t 5000 \
                "refused" "Will not SIGKILL session-critical process: $name (pid=$pid)"
            return 0 ;;
    esac

    if kill -"$sig" "$pid" 2>/dev/null; then
        notify-send -a resource-kill-picker -t 3000 \
            "killed $name" "pid=$pid signal=SIG$sig"
    else
        notify-send -a resource-kill-picker -u critical -t 4000 \
            "kill failed" "$name (pid=$pid) — already gone or no permission"
    fi
}

do_kill "$rc" "$pid" "$name"
