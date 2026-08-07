#!/usr/bin/env bash
set -u

DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/tmux-harpoon"
LIST="$DATA_DIR/list"

mkdir -p "$DATA_DIR"
touch "$LIST"

display() {
    tmux display-message "harpoon: $1"
}

session_format() {
    local format=$1
    if [[ -n "${HARPOON_CLIENT:-}" ]]; then
        tmux display-message -c "$HARPOON_CLIENT" -p "$format"
    else
        tmux display-message -p "$format"
    fi
}

switch_to() {
    local name=$1
    if [[ -n "${HARPOON_CLIENT:-}" ]]; then
        tmux switch-client -c "$HARPOON_CLIENT" -t "$name"
    else
        tmux switch-client -t "$name"
    fi
}

cmd_add() {
    local name path
    if [[ $# -ge 2 ]]; then
        name=$1
        path=$2
    else
        name=$(session_format '#{session_name}')
        path=$(session_format '#{session_path}')
    fi
    if cut -f1 "$LIST" | grep -qxF "$name"; then
        display "$name already pinned"
        return 0
    fi
    printf '%s\t%s\n' "$name" "$path" >>"$LIST"
    display "pinned $name (slot $(wc -l <"$LIST"))"
}

resolve_line() {
    local line=$1 slot=$2
    local name=${line%%$'\t'*}
    local path=${line#*$'\t'}
    if tmux has-session -t="$name" 2>/dev/null; then
        switch_to "$name"
    elif [[ "$path" != "$name" && -d "$path" ]]; then
        tmux new-session -ds "$name" -c "$path"
        switch_to "$name"
    else
        display "slot $slot gone ($name)"
        return 1
    fi
}

cmd_go() {
    local n=$1
    local line
    line=$(sed -n "${n}p" "$LIST")
    if [[ -z "$line" ]]; then
        display "slot $n empty"
        return 0
    fi
    resolve_line "$line" "$n"
}

cmd_breakout() {
    local path=$1 pane=$2
    local name placeholder=""
    name=$(printf %s "$path" | grep -oE '[a-zA-Z]+-[0-9]+' | head -1 | tr '[:lower:]' '[:upper:]')
    [[ -n "$name" ]] || name=$(basename "$path" | tr . _)
    if ! tmux has-session -t="$name" 2>/dev/null; then
        placeholder=$(tmux new-session -dP -F '#{window_id}' -s "$name" -c "$path")
    fi
    switch_to "$name"
    tmux break-pane -s "$pane" -t "$name:"
    [[ -n "$placeholder" ]] && tmux kill-window -t "$placeholder"
    cmd_add "$name" "$path"
}

cmd_cycle() {
    local dir=$1 current=${2:-}
    local -a lines=()
    local line
    while IFS= read -r line; do
        [[ -n "$line" ]] && lines+=("$line")
    done <"$LIST"
    local count=${#lines[@]}
    if ((count == 0)); then
        display "no pins"
        return 0
    fi
    [[ -n "$current" ]] || current=$(session_format '#{session_name}')
    local idx=0 i name
    for ((i = 0; i < count; i++)); do
        name=${lines[i]%%$'\t'*}
        [[ "$name" == "$current" ]] && { idx=$((i + 1)); break; }
    done
    local target
    if [[ "$dir" == next ]]; then
        target=$((idx % count + 1))
    elif ((idx <= 1)); then
        target=$count
    else
        target=$((idx - 1))
    fi
    resolve_line "${lines[target - 1]}" "$target"
}

cmd_normalize() {
    local tmp
    tmp=$(mktemp "$DATA_DIR/.list.XXXXXX")
    awk -F'\t' 'NF && $1 != "" && !seen[$1]++' "$LIST" >"$tmp"
    mv "$tmp" "$LIST"
}

menu_load() {
    MENU_ENTRIES=()
    local line
    while IFS= read -r line; do
        [[ -n "$line" ]] && MENU_ENTRIES+=("$line")
    done <"$LIST"
}

menu_save() {
    if ((${#MENU_ENTRIES[@]})); then
        printf '%s\n' "${MENU_ENTRIES[@]}" >"$LIST"
    else
        : >"$LIST"
    fi
}

menu_draw() {
    printf '\e[2J\e[H'
    printf '\n'
    if ((${#MENU_ENTRIES[@]} == 0)); then
        printf '   \e[2mempty — press a to pin the current session\e[0m\n'
    else
        local i name marker
        for ((i = 0; i < ${#MENU_ENTRIES[@]}; i++)); do
            name=${MENU_ENTRIES[i]%%$'\t'*}
            marker=""
            tmux has-session -t="$name" 2>/dev/null || marker=$' \e[2m(dead)\e[0m'
            if ((i + 1 == MENU_SEL)); then
                printf '  \e[7m %d  %-30s \e[0m%b\n' "$((i + 1))" "$name" "$marker"
            else
                printf '   %d  %-30s %b\n' "$((i + 1))" "$name" "$marker"
            fi
        done
    fi
    printf '\n  \e[2mj/k move   J/K reorder   d unpin   a pin   enter jump   q quit\e[0m\n'
    if [[ -n "$MENU_MSG" ]]; then
        printf '  \e[33m%s\e[0m\n' "$MENU_MSG"
        MENU_MSG=""
    fi
}

menu_jump() {
    local slot=$1
    ((slot >= 1 && slot <= ${#MENU_ENTRIES[@]})) || return 0
    menu_save
    if resolve_line "${MENU_ENTRIES[slot - 1]}" "$slot" 2>/dev/null; then
        exit 0
    fi
    MENU_MSG="slot $slot gone (${MENU_ENTRIES[slot - 1]%%$'\t'*})"
}

menu_swap() {
    local a=$1 b=$2
    local tmp=${MENU_ENTRIES[a - 1]}
    MENU_ENTRIES[a - 1]=${MENU_ENTRIES[b - 1]}
    MENU_ENTRIES[b - 1]=$tmp
    MENU_SEL=$b
    menu_save
}

menu_read_key() {
    local rest
    KEY=""
    IFS= read -rsn1 KEY || return 1
    if [[ "$KEY" == $'\x1b' ]]; then
        rest=""
        read -rsn2 -t 0.05 rest || true
        case "$rest" in
            '[A') KEY=k ;;
            '[B') KEY=j ;;
            *)    KEY=q ;;
        esac
    fi
}

cmd_menu() {
    if [[ $# -ge 1 && -n "$1" && "${1:0:1}" != '#' ]]; then
        HARPOON_CLIENT=$1
    else
        HARPOON_CLIENT=$(tmux display-message -p '#{client_name}' 2>/dev/null || true)
    fi
    export HARPOON_CLIENT
    MENU_SEL=1
    MENU_MSG=""
    menu_load
    printf '\e[?25l'
    trap 'printf "\e[?25h"' EXIT

    local n
    while :; do
        ((MENU_SEL > ${#MENU_ENTRIES[@]})) && MENU_SEL=${#MENU_ENTRIES[@]}
        ((MENU_SEL < 1)) && MENU_SEL=1
        menu_draw
        menu_read_key || break
        case "$KEY" in
            j) ((MENU_SEL < ${#MENU_ENTRIES[@]})) && ((MENU_SEL++)) ;;
            k) ((MENU_SEL > 1)) && ((MENU_SEL--)) ;;
            J) ((MENU_SEL < ${#MENU_ENTRIES[@]})) && menu_swap "$MENU_SEL" "$((MENU_SEL + 1))" ;;
            K) ((MENU_SEL > 1)) && menu_swap "$MENU_SEL" "$((MENU_SEL - 1))" ;;
            d)
                if ((${#MENU_ENTRIES[@]})); then
                    MENU_ENTRIES=("${MENU_ENTRIES[@]:0:MENU_SEL-1}" "${MENU_ENTRIES[@]:MENU_SEL}")
                    menu_save
                fi
                ;;
            a)
                cmd_add >/dev/null 2>&1
                menu_load
                ;;
            "" | $'\r') menu_jump "$MENU_SEL" ;;
            [1-9])
                n=$KEY
                ((n <= ${#MENU_ENTRIES[@]})) && menu_jump "$n"
                ;;
            q) break ;;
        esac
    done
}

case "${1:-}" in
    add)       shift; cmd_add "$@" ;;
    go)        cmd_go "$2" ;;
    next)      cmd_cycle next "${2:-}" ;;
    prev)      cmd_cycle prev "${2:-}" ;;
    breakout)  cmd_breakout "$2" "$3" ;;
    menu)      shift; cmd_menu "$@" ;;
    normalize) cmd_normalize ;;
    *)         echo "usage: harpoon.sh add|go <n>|next|prev|breakout <path> <pane>|menu [client]|normalize" >&2; exit 1 ;;
esac
