#!/usr/bin/env bash
# prefix + W: jump to an agent — Claude Code or Copilot CLI — that wants attention.
#
# Exactly one blocked pane is unambiguous, so it switches straight there.
# Anything else — several blocked, or none blocked but others running — opens a
# menu ordered most-urgent first, rather than guessing which one was meant.

set -u

tmux_bin="${TMUX_BIN:-/usr/bin/tmux}"
[ -x "$tmux_bin" ] || exit 0

# shellcheck source=claude-icons.sh
. "$(dirname "$0")/claude-icons.sh"

here_pane=$("$tmux_bin" display-message -p '#{pane_id}' 2>/dev/null)

# Excluding the pane rather than the window: a Claude and a Copilot often share one
# window, and sitting in one of them is no reason to hide the other.
mapfile -t rows < <(
   "$(dirname "$0")/claude-waiting.sh" |
      while IFS=$'\t' read -r state session index name title age agent pane; do
         [ "$pane" = "$here_pane" ] && continue
         printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
            "$state" "$session" "$index" "$name" "$title" "$age" "$agent" "$pane"
      done
)

if [ "${#rows[@]}" -eq 0 ]; then
   "$tmux_bin" display-message "no other agent panes"
   exit 0
fi

# A pane id is a valid switch-client target and moves session, window and pane in one
# step, which is what makes a per-agent row jumpable when two agents share a window.
jump() {
   "$tmux_bin" switch-client -t "$1"
}

mapfile -t blocked < <(printf '%s\n' "${rows[@]}" | grep '^waiting	')
if [ "${#blocked[@]}" -eq 1 ]; then
   IFS=$'\t' read -r _ _ _ _ _ _ _ pane <<<"${blocked[0]}"
   jump "$pane"
   exit 0
fi

# A bell or a moon is an invitation, not a block, so only a genuinely blocked
# session ever gets jumped to without asking. The header names whichever signal
# the rows are actually carrying.
fresh_count() {
   local count=0 row state age
   for row in "${rows[@]}"; do
      IFS=$'\t' read -r state _ _ _ _ age _ _ <<<"$row"
      case "$state" in waiting | running) continue ;; esac
      [ "$(claude_age_or_oldest "$age")" -lt "$CLAUDE_FRESH_SECS" ] && count=$((count + 1))
   done
   printf '%s' "$count"
}

title=' agents '
if [ "${#blocked[@]}" -gt 1 ]; then
   title=" ${#blocked[@]} waiting "
else
   fresh=$(fresh_count)
   [ "$fresh" -gt 0 ] && title=" $fresh done "
fi

menu=(display-menu -T "$title" -x C -y C)
position=0
for row in "${rows[@]}"; do
   IFS=$'\t' read -r state session index name pane_title age agent pane <<<"$row"
   position=$((position + 1))
   key=$([ "$position" -le 9 ] && printf '%s' "$position" || printf '')
   short=${pane_title:0:80}

   # The moon buckets carry no style, so the trailing reset is only worth emitting
   # when something was actually set.
   style=$(claude_style_for "$state")
   reset=$([ -n "$style" ] && printf '#[default]')

   # Menu labels are expanded as formats, so a literal # in a title ("PR #127") has
   # to be doubled or tmux reads it as the start of a sequence. Copilot summaries
   # carry issue and PR numbers routinely, so this is the common case, not the edge.
   label=$(printf '%s%s %s  %s:%s %s: "%s"%s' \
      "$style" "$(claude_icon_for "$state" "$age")" "$(claude_agent_badge "$agent")" \
      "$session" "$index" "$name" "${short//\#/##}" "$reset")
   menu+=("$label" "$key" "switch-client -t '${pane}'")
done

"$tmux_bin" "${menu[@]}"
