#!/usr/bin/env bash
# status-right renderer: one chip per other tmux session, showing its most urgent
# agent — Claude Code or Copilot CLI, whichever wants attention most — as a red bell
# blocked on input, a yellow cog working, and otherwise a moon ageing from a fresh
# bell through to a dead new moon. Prints nothing when there is nothing to report.
#
# A chip names the session, not the agent, so which of the two is blocked is a
# question for the picker; here they fold together deliberately.
#
# Takes the attached session name as $1 so a client never lists itself.

set -u

here="${1:-}"

# shellcheck source=claude-icons.sh
. "$(dirname "$0")/claude-icons.sh"

# status-interval is 1, so this is re-run on every redraw of every attached
# client. Ages are measured in minutes at the finest, so serving a few seconds
# stale costs nothing and saves the whole pipeline on most redraws.
: "${CLAUDE_STATUS_CACHE_SECS:=3}"

# Keyed on the asking session, because each client's chips exclude a different
# one — a shared file would have clients serving each other's output.
cache="${TMPDIR:-/tmp}/claude-status-${UID}-${here:-none}"

if [ -f "$cache" ]; then
   age=$(( $(date +%s) - $(stat -c %Y "$cache" 2>/dev/null || echo 0) ))
   if [ "$age" -ge 0 ] && [ "$age" -lt "$CLAUDE_STATUS_CACHE_SECS" ]; then
      cat "$cache"
      exit 0
   fi
fi

render() {
   "$(dirname "$0")/claude-waiting.sh" |
      awk -F'\t' -v here="$here" '
         function rank(s) {
            return s == "waiting" ? 3 : (s == "running" ? 2 : 1)
         }
         $2 == here { next }
         {
            # One chip per session, so its windows and both agents fold onto the
            # most urgent state and the age of whichever finished most recently.
            if (!($2 in state)) { order[++n] = $2; state[$2] = $1; age[$2] = $6 }
            if (rank($1) > rank(state[$2])) state[$2] = $1
            if ($6 + 0 < age[$2] + 0) age[$2] = $6
         }
         END {
            # Emitted in the same order the picker uses: blocked, then working,
            # then the rest newest-first.
            for (r = 3; r >= 2; r--)
               for (i = 1; i <= n; i++)
                  if (rank(state[order[i]]) == r) print state[order[i]] "\t" order[i] "\t" age[order[i]]

            for (i = 1; i <= n; i++)
               if (rank(state[order[i]]) == 1) rest[++rest_n] = order[i]

            for (a = 1; a <= rest_n; a++) {
               best = a
               for (b = a + 1; b <= rest_n; b++)
                  if (age[rest[b]] + 0 < age[rest[best]] + 0) best = b
               swap = rest[a]; rest[a] = rest[best]; rest[best] = swap
               print state[rest[a]] "\t" rest[a] "\t" age[rest[a]]
            }
         }' |
      while IFS=$'\t' read -r state session age; do
         # The moon buckets carry no style, so the trailing reset is only worth
         # emitting when something was actually set.
         style=$(claude_style_for "$state")
         reset=$([ -n "$style" ] && printf '#[default]')
         printf '%s%s %s%s ' "$style" "$(claude_icon_for "$state" "$age")" "$session" "$reset"
      done
}

# Written via a temp file so a concurrent redraw never reads a half-written chip
# line, and printed from the file so what is cached is exactly what was shown.
tmp="${cache}.$$"
render >"$tmp" 2>/dev/null
mv -f "$tmp" "$cache" 2>/dev/null
cat "$cache" 2>/dev/null
