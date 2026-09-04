#!/usr/bin/env bash
# Sourced by claude-jump.sh and claude-status.sh so the picker and the status bar
# render the same session identically. Both take (state, age_in_seconds).
#
# Blocked and working sessions are urgency signals and get coloured Nerd Font
# glyphs (MesloLGS NF), which are monochrome and so take the fg colour. Everything
# else is an age signal and gets an emoji moon instead: terminals draw emoji from
# their own palette and ignore the fg, so those rows are deliberately emitted with
# no style at all rather than a colour that would silently do nothing.
#
# Bucket edges are overridable so retuning the ladder needs no edits here.

: "${CLAUDE_FRESH_SECS:=300}"
: "${CLAUDE_FULL_SECS:=900}"
: "${CLAUDE_GIBBOUS_SECS:=3600}"
: "${CLAUDE_HALF_SECS:=14400}"
: "${CLAUDE_CRESCENT_SECS:=43200}"

# An unknown or non-numeric age reads as oldest rather than newest, so a session
# whose age could not be resolved sinks to the bottom instead of masquerading as
# freshly finished.
claude_age_or_oldest() {
   case "${1:-}" in
      '' | *[!0-9]*) printf '%s' "$((CLAUDE_CRESCENT_SECS + 1))" ;;
      *) printf '%s' "$1" ;;
   esac
}

# Bell U+F0F3, cog U+F013, written as escapes so the icons survive editing tools
# that mangle Private Use Area characters.
claude_icon_for() {
   local state="$1" age
   case "$state" in
      waiting) printf '\uf0f3' ; return ;;
      running) printf '\uf013' ; return ;;
   esac

   age=$(claude_age_or_oldest "${2:-}")
   if [ "$age" -lt "$CLAUDE_FRESH_SECS" ]; then printf '🔔'
   elif [ "$age" -lt "$CLAUDE_FULL_SECS" ]; then printf '🌕'
   elif [ "$age" -lt "$CLAUDE_GIBBOUS_SECS" ]; then printf '🌔'
   elif [ "$age" -lt "$CLAUDE_HALF_SECS" ]; then printf '🌓'
   elif [ "$age" -lt "$CLAUDE_CRESCENT_SECS" ]; then printf '🌒'
   else printf '🌑'
   fi
}

# Which agent a row belongs to, so a window running a Claude beside a Copilot reads
# as two distinguishable rows rather than two identical ones. The state icon already
# carries the urgency; this only carries the identity.
#
# Claude's is its own resting asterisk U+2733; Copilot's is the Nerd Font GitHub mark
# U+F09B, written as escapes so both survive editing tools that mangle Private Use
# Area characters.
claude_agent_badge() {
   case "${1:-}" in
      copilot) printf '\uf09b' ;;
      *) printf '\u2733' ;;
   esac
}

# Prints nothing for the moon buckets. Callers must tolerate an empty style —
# appending a #[default] reset after one is dead weight.
claude_style_for() {
   case "$1" in
      waiting) printf '#[fg=#f7768e,bold]' ;;
      running) printf '#[fg=#e0af68,bold]' ;;
   esac
}
