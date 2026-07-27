#!/usr/bin/env bash
#
# fix-home-paths.sh — rewrite a stale, hardcoded home path to the CURRENT user's
# home inside THIS dotfiles repo.
#
# The replacement value is derived at runtime from the CLI ($HOME / id -un), so
# the script is not tied to any one username or machine — run it on a fresh
# checkout after a rename or on a new host and it normalises the paths to
# whoever is running it.
#
# NOTE: the target files (SQLDeveloper XML/.properties, a Lua lockfile, etc.) do
# NOT expand shell variables. So we write the *expanded* path (e.g.
# /home/dariuszw), never the literal string "$HOME".
#
# Usage:
#   ./scripts/fix-home-paths.sh                    # dry-run, auto-detect old user
#   ./scripts/fix-home-paths.sh --apply            # write changes
#   ./scripts/fix-home-paths.sh --old-user=dariusz # dry-run for a specific old user
#   ./scripts/fix-home-paths.sh --old-user=dariusz --apply
#
set -euo pipefail

# --- resolve where we are (works regardless of cwd or repo name/location) ------
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CURRENT_USER="$(id -un)"
NEW_HOME="$HOME"                       # expanded, e.g. /home/dariuszw

# Names that look like home paths but must NEVER be rewritten:
#   - the current user (already correct)
#   - vendored / reference configs (e.g. tjdevries' nvim dotfiles)
#   - generic placeholders found in example/template files
IGNORE_USERS=("$CURRENT_USER" tjdevries user test root)

# Directories that hold logs, caches, vendored deps, or runtime/session state —
# these contain incidental "/home/<name>/" strings as *data*, not config, so we
# must never rewrite inside them. (The repo embeds a submodule that stows the
# whole ~/.claude tree, including Claude Code runtime state: backup transcripts,
# file-history snapshots, per-session task/plan/tool-result files.)
EXCLUDE_DIRS=(.git .oh-my-zsh claude-bkp node_modules .cache .npm .cargo
              .serena .playwright-mcp .mypy_cache .venv logs dist build
              file-history projects tasks plans oracle.javatools.cache)

# File patterns that are data/logs/binary caches, never config.
EXCLUDE_GLOBS=("*.jsonl" "*.log" "*.jdb")

# Build the grep --exclude-dir / --exclude argument arrays once.
GREP_EXCL=()
for d in "${EXCLUDE_DIRS[@]}"; do GREP_EXCL+=(--exclude-dir="$d"); done
for g in "${EXCLUDE_GLOBS[@]}"; do GREP_EXCL+=(--exclude="$g"); done

APPLY=0
OLD_USER=""

for arg in "$@"; do
   case "$arg" in
      --apply)        APPLY=1 ;;
      --old-user=*)   OLD_USER="${arg#*=}" ;;
      -h|--help)
         grep '^#' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
         exit 0 ;;
      *) echo "fix-home-paths: unknown argument: $arg" >&2; exit 2 ;;
   esac
done

is_ignored() {
   local name="$1"
   local ignore
   for ignore in "${IGNORE_USERS[@]}"; do
      [[ "$name" == "$ignore" ]] && return 0
   done
   return 1
}

# --- auto-detect the old user if not supplied ----------------------------------
if [[ -z "$OLD_USER" ]]; then
   mapfile -t candidates < <(
      grep -rohE '/home/[A-Za-z0-9._-]+/' "$REPO_ROOT" \
         --binary-files=text "${GREP_EXCL[@]}" \
      | sed -E 's#^/home/([^/]+)/$#\1#' \
      | sort -u
   )

   detected=()
   for name in "${candidates[@]:-}"; do
      [[ -z "$name" ]] && continue
      is_ignored "$name" && continue
      detected+=("$name")
   done

   if [[ ${#detected[@]} -eq 0 ]]; then
      echo "fix-home-paths: no stale /home/<user>/ paths found — nothing to do."
      exit 0
   elif [[ ${#detected[@]} -gt 1 ]]; then
      echo "fix-home-paths: multiple candidate old users found:" >&2
      printf '  - %s\n' "${detected[@]}" >&2
      echo "Re-run with --old-user=<name> to pick one." >&2
      exit 3
   fi
   OLD_USER="${detected[0]}"
fi

OLD_PREFIX="/home/${OLD_USER}/"
NEW_PREFIX="${NEW_HOME%/}/"

if [[ "$OLD_PREFIX" == "$NEW_PREFIX" ]]; then
   echo "fix-home-paths: old path ($OLD_PREFIX) already equals current home — nothing to do."
   exit 0
fi

# --- find affected files (force text so SQLDeveloper XMLs aren't skipped) -------
mapfile -d '' -t files < <(
   grep -rlZ --binary-files=text "${GREP_EXCL[@]}" \
      "$OLD_PREFIX" "$REPO_ROOT" 2>/dev/null || true
)

if [[ ${#files[@]} -eq 0 ]]; then
   echo "fix-home-paths: no occurrences of ${OLD_PREFIX} in ${REPO_ROOT} — nothing to do."
   exit 0
fi

echo "Repo:     $REPO_ROOT"
echo "Rewrite:  ${OLD_PREFIX}  ->  ${NEW_PREFIX}"
echo "Files:    ${#files[@]}"
echo

# --- preview (always) ----------------------------------------------------------
for f in "${files[@]}"; do
   grep -nH --binary-files=text "$OLD_PREFIX" "$f"
done

if [[ "$APPLY" -ne 1 ]]; then
   echo
   echo "Dry-run only. Re-run with --apply to write these changes."
   exit 0
fi

# --- apply (prefix has a trailing slash, so no collision with /home/<old>w/) ----
for f in "${files[@]}"; do
   sed -i "s#${OLD_PREFIX}#${NEW_PREFIX}#g" "$f"
done

echo
echo "Applied. Rewrote ${OLD_PREFIX} -> ${NEW_PREFIX} in ${#files[@]} file(s)."
