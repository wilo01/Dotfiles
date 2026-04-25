#!/usr/bin/env bash
# pressure-watch.sh — long-running PSI watcher with custom notification body.
# Replaces psi-notify so the notification can include the hotkey hint.
#
# Reads /proc/pressure/{memory,cpu}; on threshold cross fires a swaync
# notification that says which resource and reminds the user of the rofi
# hotkeys ($mod+Shift+m for memory, $mod+Shift+x for cpu).
#
# Anti-spam: rising-edge + 10-min cooldown, tracked independently per resource.

set -euo pipefail

# ── Tunables (override via systemd unit Environment= or env) ────────────────
INTERVAL_SEC=${INTERVAL_SEC:-5}
MEM_SOME_AVG10_PCT=${MEM_SOME_AVG10_PCT:-10.00}
CPU_SOME_AVG60_PCT=${CPU_SOME_AVG60_PCT:-50.00}
COOLDOWN_SEC=${COOLDOWN_SEC:-600}

# ── State ───────────────────────────────────────────────────────────────────
declare -A state          # state[mem]=high|normal, state[cpu]=high|normal
declare -A last_alert_ts  # last_alert_ts[mem|cpu]=epoch
state[mem]=normal; state[cpu]=normal
last_alert_ts[mem]=0; last_alert_ts[cpu]=0

# ── Helpers ─────────────────────────────────────────────────────────────────
psi_value() {
    # $1=resource (memory|cpu), $2=line (some|full), $3=window (avg10|avg60|avg300)
    # /proc/pressure/<resource> lines look like:
    #   some avg10=0.51 avg60=0.65 avg300=1.20 total=12345
    #   full avg10=0.00 ...
    # so we match the LINE column ($1 in awk) against $2 here (some|full).
    awk -v which="$2" -v win="$3" '
        $1 == which { for (i=1;i<=NF;i++) if (index($i,win"=")==1) {
            split($i,a,"="); print a[2]; exit
        } }' "/proc/pressure/$1"
}

cmp_ge() {
    # bash can't do float comparison; use awk. $1>=$2 ?
    awk -v a="$1" -v b="$2" 'BEGIN{exit (a+0 >= b+0 ? 0 : 1)}'
}

notify_alert() {
    # $1=kind (mem|cpu), $2=value, $3=threshold
    local kind="$1" value="$2" thresh="$3"
    local title body hotkey icon
    case "$kind" in
        mem) title="Memory pressure"
             hotkey='$mod+Shift+m'
             # freedesktop-named icon (theme-resolved). Falls back to absolute
             # Breeze SVG if the active icon theme doesn't ship a "memory" name
             # (Adwaita doesn't; Breeze is installed system-wide on this box).
             if [ -r /usr/share/icons/breeze/devices/64/memory.svg ]; then
                 icon=/usr/share/icons/breeze/devices/64/memory.svg
             else
                 icon=memory
             fi
             body=$(printf "PSI memory.some.avg10 = %s%% (threshold %s%%)\nPress %s to open the kill picker." \
                          "$value" "$thresh" "$hotkey") ;;
        cpu) title="CPU pressure"
             hotkey='$mod+Shift+x'
             if [ -r /usr/share/icons/breeze/devices/64/cpu.svg ]; then
                 icon=/usr/share/icons/breeze/devices/64/cpu.svg
             else
                 icon=cpu
             fi
             body=$(printf "PSI cpu.some.avg60 = %s%% (threshold %s%%)\nPress %s to open the kill picker." \
                          "$value" "$thresh" "$hotkey") ;;
    esac
    notify-send -a pressure-watch -u critical -t 0 -i "$icon" \
        --hint=string:x-canonical-private-synchronous:"pressure-$kind" \
        "$title" "$body"
}

evaluate() {
    # $1=kind (mem|cpu), $2=current value, $3=threshold
    local kind="$1" value="$2" thresh="$3"
    local now; now=$(date +%s)
    if cmp_ge "$value" "$thresh"; then
        # high
        if [ "${state[$kind]}" = normal ]; then
            notify_alert "$kind" "$value" "$thresh"
            state[$kind]=high
            last_alert_ts[$kind]=$now
        elif [ $((now - ${last_alert_ts[$kind]})) -ge "$COOLDOWN_SEC" ]; then
            notify_alert "$kind" "$value" "$thresh"
            last_alert_ts[$kind]=$now
        fi
    else
        state[$kind]=normal
    fi
}

# ── Main loop ───────────────────────────────────────────────────────────────
trap 'echo "pressure-watch exiting"; exit 0' TERM INT
echo "pressure-watch started (interval=${INTERVAL_SEC}s mem=${MEM_SOME_AVG10_PCT}% cpu=${CPU_SOME_AVG60_PCT}% cooldown=${COOLDOWN_SEC}s)"

while true; do
    mem_val=$(psi_value memory some avg10)
    cpu_val=$(psi_value cpu     some avg60)
    evaluate mem "$mem_val" "$MEM_SOME_AVG10_PCT"
    evaluate cpu "$cpu_val" "$CPU_SOME_AVG60_PCT"
    sleep "$INTERVAL_SEC"
done
