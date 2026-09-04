#!/usr/bin/env bash
# Prints one "state<TAB>session<TAB>window_index<TAB>window_name<TAB>pane_title<TAB>age<TAB>agent<TAB>pane_id"
# row per agent running in tmux — Claude Code and GitHub Copilot CLI both — ordered
# waiting > running > freshest-first. Age is seconds since the session last finished a
# turn. Consumers do their own filtering: claude-status.sh drops the attached session,
# claude-jump.sh drops the current pane.
#
# Rows are keyed per agent rather than per window, so a window running a Claude beside
# a Copilot reports both and each can be jumped to individually.
#
# Panes with no state of their own fall back to what the process is doing, so
# sessions report before any hook has ever fired for them.

set -u

tmux_bin="${TMUX_BIN:-/usr/bin/tmux}"
[ -x "$tmux_bin" ] || exit 0

now=$(date +%s)

# Claude parks its pane title on this glyph when resting and cycles spinner
# characters through it while working, which is the only signal available before
# the hooks are installed.
idle_glyph=$'✳'

# A Claude blocked on a question or a permission prompt parks its title on the
# same glyph as an idle one, so the only way to tell them apart without the hooks
# is the prompt footer it draws. "Esc to cancel" means it is waiting on an answer;
# note that a working Claude shows "esc to interrupt", which must not match.
claude_pane_is_blocked() {
   "$tmux_bin" capture-pane -p -t "$1" -S -15 2>/dev/null |
      grep -qE 'Esc to cancel|Do you want to'
}

# Copilot leaves its pane title on the session summary in every state, so unlike
# Claude there is no title glyph to read and the footer is the whole signal. Its
# strings are its own: a blocked Copilot draws "esc to cancel" in lower case where
# Claude uses "Esc to cancel", and a working one draws "esc interrupt" where Claude
# draws "esc to interrupt" — near-misses that would cross-match if the two agents
# shared one matcher, hence the separate function.
copilot_pane_state() {
   local tail
   tail=$("$tmux_bin" capture-pane -p -t "$1" -S -15 2>/dev/null)
   if printf '%s' "$tail" | grep -qE 'Do you want to allow|\(Esc to stop\)|esc to cancel'; then
      printf 'waiting'
   elif printf '%s' "$tail" | grep -qF 'esc interrupt'; then
      printf 'running'
   else
      printf 'idle'
   fi
}

"$tmux_bin" list-panes -a -F \
   '#{pane_id}|#{session_name}|#{window_index}|#{window_name}|#{@claude_state}|#{@claude_done_at}|#{pane_current_command}|#{window_activity}|#{pane_title}' 2>/dev/null |
   while IFS='|' read -r pane_id session index name inherited inherited_at command activity title; do
      case "$command" in
         claude | copilot) ;;
         *) continue ;;
      esac

      # Both options are window-inherited in this format, so an empty pair proves
      # the pane holds neither and the lookup can be skipped entirely. When either
      # is set they are read back in one call rather than forking per option.
      own="" own_at=""
      if [ -n "$inherited" ] || [ -n "$inherited_at" ]; then
         while read -r option value; do
            case "$option" in
               @claude_state) own="$value" ;;
               @claude_done_at) own_at="$value" ;;
            esac
         done < <("$tmux_bin" show-options -qp -t "$pane_id" 2>/dev/null)
      fi

      label="$title"

      if [ -n "$own" ]; then
         state="$own"
      elif [ "$command" = "claude" ]; then
         case "$title" in
            "$idle_glyph"*)
               if claude_pane_is_blocked "$pane_id"; then state="waiting"; else state="idle"; fi
               ;;
            *) state="running" ;;
         esac
      else
         state=$(copilot_pane_state "$pane_id")
         # Copilot suffixes every title with the product name, which is the one thing
         # a list of Copilots does not need repeated on every row.
         label="${title% - GitHub Copilot}"
      fi

      # Sessions predating the hooks — and any profile not running them — have no
      # stamp of their own, so how long the window has been quiet stands in for
      # when its agent last finished.
      finished_at="${own_at:-$activity}"
      age=$((now - finished_at))
      [ "$age" -lt 0 ] && age=0

      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
         "$state" "$session" "$index" "$name" "$label" "$age" "$command" "$pane_id"
   done |
   awk -F'\t' '
      function rank(s) {
         return s == "waiting" ? 3 : (s == "running" ? 2 : 1)
      }
      {
         # Keyed on the agent within the window, not the pane, so two Claudes side by
         # side are one row while a Claude beside a Copilot is two. The row follows
         # whichever of its panes is most urgent, and the age of the one that
         # finished most recently.
         key = $2 SUBSEP $3 SUBSEP $4 SUBSEP $7
         if (!(key in state)) {
            order[++n] = key
            state[key] = $1; sess[key] = $2; idx[key] = $3; nm[key] = $4
            label[key] = $5; age[key] = $6; agent[key] = $7; pane[key] = $8
         }
         if (rank($1) > rank(state[key])) {
            state[key] = $1; label[key] = $5; pane[key] = $8
         }
         if ($6 + 0 < age[key] + 0) age[key] = $6
      }
      END {
         for (r = 3; r >= 2; r--)
            for (i = 1; i <= n; i++)
               if (rank(state[order[i]]) == r) print row(order[i])

         # Neither blocked nor working, so there is no urgency to rank the tail by
         # and it goes newest-first instead — the session that just finished sits
         # directly under the running ones, the long-abandoned sink to the bottom.
         for (i = 1; i <= n; i++)
            if (rank(state[order[i]]) == 1) rest[++rest_n] = order[i]

         for (a = 1; a <= rest_n; a++) {
            best = a
            for (b = a + 1; b <= rest_n; b++)
               if (age[rest[b]] + 0 < age[rest[best]] + 0) best = b
            swap = rest[a]; rest[a] = rest[best]; rest[best] = swap
            print row(rest[a])
         }
      }
      function row(key) {
         return state[key] "\t" sess[key] "\t" idx[key] "\t" nm[key] "\t" \
                label[key] "\t" age[key] "\t" agent[key] "\t" pane[key]
      }'
