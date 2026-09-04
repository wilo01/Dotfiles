#!/usr/bin/env bash
# Keybinding cheatsheet, bound to prefix ?.
#
# The key list is read live from tmux rather than from a hand-kept file, so it
# cannot drift from .tmux.conf. Descriptions come from the -N notes on each
# bind-key; a note that starts with one of the KEY_GROUPS below marks the binding as
# ours and files it under that heading. Anything without such a note is a stock
# tmux default and stays hidden until 'a' is pressed.
#
# tmux caveat this works around: `list-keys -N` covers only the prefix and root
# tables, and prints every noted key as "C-a <key>" even when it is a root-table
# binding that takes no prefix. So the authoritative key list comes from the raw
# `list-keys -T <table>` output, and -N is used purely as a key -> note lookup.

set -u

tmux_bin="${TMUX_BIN:-/usr/bin/tmux}"
[ -x "$tmux_bin" ] || exit 0

KEY_GROUPS=(session window pane copy popup project claude harpoon config)

C_HEAD=$'\e[38;2;122;162;247m\e[1m'
C_KEY=$'\e[38;2;255;158;100m'
C_TEXT=$'\e[38;2;169;177;214m'
C_DIM=$'\e[38;2;86;95;137m'
C_HIT=$'\e[38;2;158;206;106m'
C_OFF=$'\e[0m'

SHOW_ALL=0
FILTER=""
TOP=0
KEY=""
declare -a ROWS=()
declare -a VIEW=()
declare -A NOTE_PREFIX=()
declare -A NOTE_ROOT=()
declare -A HAS_PREFIX=()
declare -A HAS_ROOT=()

is_group() {
    local g
    for g in "${KEY_GROUPS[@]}"; do
        [[ "$1" == "$g" ]] && return 0
    done
    return 1
}

trim_left() {
    local s=$1
    printf '%s' "${s#"${s%%[![:space:]]*}"}"
}

table_keys() {
    "$tmux_bin" list-keys -T "$1" 2>/dev/null |
        sed -E "s/^bind-key +(-r )?-T $1 +([^ ]+) .*/\2/"
}

# The C-arrows are bound in both prefix and root, so they appear twice in the
# -N stream — prefix first, then root — with different notes. Splitting the
# stream by which table still needs a note for that key keeps them apart.
load_notes() {
    NOTE_PREFIX=(); NOTE_ROOT=(); HAS_PREFIX=(); HAS_ROOT=()
    local line key note
    while IFS= read -r key; do HAS_PREFIX["$key"]=1; done < <(table_keys prefix)
    while IFS= read -r key; do HAS_ROOT["$key"]=1; done < <(table_keys root)
    while IFS= read -r line; do
        line=${line#C-a }
        key=${line%% *}
        note=$(trim_left "${line#"$key"}")
        [[ -n "$key" && -n "$note" ]] || continue
        if [[ -n "${HAS_PREFIX[$key]:-}" && -z "${NOTE_PREFIX[$key]:-}" ]]; then
            NOTE_PREFIX["$key"]=$note
        elif [[ -n "${HAS_ROOT[$key]:-}" ]]; then
            NOTE_ROOT["$key"]=$note
        fi
    done < <("$tmux_bin" list-keys -N 2>/dev/null)
}

# Raw command text is only a fallback label, so squeeze it onto one line.
abbrev() {
    local cmd=${1//$'\t'/ }
    printf '%s' "${cmd:0:70}"
}

load_rows() {
    ROWS=()
    local table label line rest key cmd note group
    for table in prefix root; do
        if [[ $table == prefix ]]; then label="prefix + "; else label=""; fi
        while IFS= read -r line; do
            rest=${line#*-T $table }
            [[ "$rest" == "$line" ]] && continue
            key=${rest%% *}
            cmd=$(trim_left "${rest#"$key"}")
            if [[ $table == prefix ]]; then
                note=${NOTE_PREFIX[$key]:-}
            else
                note=${NOTE_ROOT[$key]:-}
            fi
            group=${note%%:*}
            if [[ -n "$note" ]] && is_group "$group"; then
                ROWS+=("$group"$'\t'"$label$key"$'\t'"${note#*: }")
            elif ((SHOW_ALL)); then
                ROWS+=("tmux defaults"$'\t'"$label$key"$'\t'"${note:-$(abbrev "$cmd")}")
            fi
        done < <("$tmux_bin" list-keys -T "$table" 2>/dev/null)
    done
}

# Flatten the grouped rows into the exact lines to paint, headers included, so
# scrolling and filtering only ever deal with one flat array.
build_view() {
    VIEW=()
    local wanted row rowgroup key desc shown
    for wanted in "${KEY_GROUPS[@]}" "tmux defaults"; do
        shown=0
        for row in "${ROWS[@]}"; do
            rowgroup=${row%%$'\t'*}
            [[ "$rowgroup" == "$wanted" ]] || continue
            key=${row#*$'\t'}
            key=${key%%$'\t'*}
            desc=${row##*$'\t'}
            if [[ -n "$FILTER" ]]; then
                shopt -s nocasematch
                if [[ "$key $desc" != *"$FILTER"* ]]; then
                    shopt -u nocasematch
                    continue
                fi
                shopt -u nocasematch
            fi
            if ((shown == 0)); then
                ((${#VIEW[@]})) && VIEW+=("S")
                VIEW+=("H"$'\t'"$wanted")
                shown=1
            fi
            VIEW+=("R"$'\t'"$key"$'\t'"$desc")
        done
    done
}

term_lines() { tput lines 2>/dev/null || echo 24; }
term_cols() { tput cols 2>/dev/null || echo 80; }

body_height() {
    local h=$(( $(term_lines) - 2 ))
    ((h < 3)) && h=3
    printf '%s' "$h"
}

draw() {
    local cols body i kind a b pad width
    cols=$(term_cols)
    body=$(body_height)
    width=$((cols - 26))
    ((width < 20)) && width=20

    printf '\e[2J\e[H'
    if ((${#VIEW[@]} == 0)); then
        printf '\n  %sno bindings match "%s"%s\n' "$C_DIM" "$FILTER" "$C_OFF"
    fi
    for ((i = TOP; i < TOP + body && i < ${#VIEW[@]}; i++)); do
        kind=${VIEW[i]%%$'\t'*}
        case "$kind" in
            H)
                a=${VIEW[i]#*$'\t'}
                printf '  %s%s%s\n' "$C_HEAD" "${a^^}" "$C_OFF"
                ;;
            R)
                a=${VIEW[i]#*$'\t'}
                b=${a#*$'\t'}
                a=${a%%$'\t'*}
                pad=$((20 - ${#a}))
                ((pad < 1)) && pad=1
                printf '    %s%s%s%*s%s%s%s\n' \
                    "$C_KEY" "$a" "$C_OFF" "$pad" "" "$C_TEXT" "${b:0:width}" "$C_OFF"
                ;;
            *) printf '\n' ;;
        esac
    done

    printf '\e[%d;1H\e[2K' "$(term_lines)"
    [[ -n "$FILTER" ]] && printf '  %s/%s%s ' "$C_HIT" "$FILTER" "$C_OFF"
    printf '  %sj/k scroll   / filter   a %s defaults   q quit%s' \
        "$C_DIM" "$(((SHOW_ALL)) && printf hide || printf show)" "$C_OFF"
}

read_key() {
    local rest
    KEY=""
    IFS= read -rsn1 KEY || return 1
    if [[ "$KEY" == $'\x1b' ]]; then
        rest=""
        read -rsn2 -t 0.05 rest || true
        case "$rest" in
            '[A') KEY=k ;;
            '[B') KEY=j ;;
            '[5') read -rsn1 -t 0.05 rest || true; KEY=K ;;
            '[6') read -rsn1 -t 0.05 rest || true; KEY=J ;;
            *) KEY=q ;;
        esac
    fi
}

prompt_filter() {
    printf '\e[%d;1H\e[2K  %s/%s' "$(term_lines)" "$C_HIT" "$C_OFF"
    printf '\e[?25h'
    IFS= read -r FILTER
    printf '\e[?25l'
    TOP=0
}

main() {
    local body max
    load_notes
    load_rows
    build_view
    printf '\e[?25l'
    trap 'printf "\e[?25h\e[0m"' EXIT

    while :; do
        body=$(body_height)
        max=$((${#VIEW[@]} - body))
        ((max < 0)) && max=0
        ((TOP > max)) && TOP=$max
        ((TOP < 0)) && TOP=0

        draw
        read_key || break
        case "$KEY" in
            j) ((TOP < max)) && ((TOP++)) ;;
            k) ((TOP > 0)) && ((TOP--)) ;;
            J) TOP=$((TOP + body)) ;;
            K) TOP=$((TOP - body)) ;;
            g) TOP=0 ;;
            G) TOP=$max ;;
            /) prompt_filter; build_view ;;
            a) SHOW_ALL=$((1 - SHOW_ALL)); TOP=0; load_rows; build_view ;;
            q | $'\r' | "") break ;;
        esac
    done
}

case "${1:-}" in
    "" | menu) main ;;
    list) load_notes; load_rows; build_view; printf '%s\n' "${VIEW[@]}" ;;
    *) echo "usage: keys-help.sh [menu|list]" >&2; exit 1 ;;
esac
