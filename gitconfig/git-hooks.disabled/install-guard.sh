#!/bin/bash
# Makes the protected-branch guard effective in every local repo.
#
# The guard is global via core.hooksPath, but a repo that sets core.hooksPath
# LOCALLY wins — husky does exactly that. This repoints those repos at the guard;
# the guard then delegates into husky's runner so lint-staged still runs.
#
# Husky's `prepare` script resets core.hooksPath on every npm/pnpm install, so
# re-run this after installing dependencies in a husky repo.
#
#   ./install-guard.sh            # report only
#   ./install-guard.sh --apply    # make the changes
#   ./install-guard.sh --revert   # hand husky its hooksPath back

set -u

GUARD_DIR=$(dirname "$(readlink -f "$0")")
MODE="report"
case "${1:-}" in
   --apply)  MODE="apply" ;;
   --revert) MODE="revert" ;;
   "")       ;;
   *)        echo "usage: $0 [--apply|--revert]" >&2; exit 2 ;;
esac

SEARCH_ROOTS=("$HOME/Dev" "$HOME/tds-branch-opener" "$HOME/.Dotfiles" "$HOME/homeassistant")

find_repos() {
   local root
   for root in "${SEARCH_ROOTS[@]}"; do
      [[ -d "$root" ]] || continue
      find "$root" -maxdepth 5 -name .git -not -path '*/node_modules/*' 2>/dev/null
   done | sed 's|/\.git$||' | sort -u
}

overridden=0
fixed=0
disabled=0

while read -r repo; do
   local_path=$(git -C "$repo" config --local --get core.hooksPath 2>/dev/null)

   if [[ "$(git -C "$repo" config --type=bool --get hooks.branchGuard 2>/dev/null)" == "false" ]]; then
      echo "  opted out : ${repo#"$HOME"/}"
      disabled=$((disabled + 1))
      continue
   fi

   if [[ "$MODE" == "revert" ]]; then
      if [[ "$local_path" == "$GUARD_DIR" ]] && [[ -d "$repo/.husky/_" ]]; then
         git -C "$repo" config core.hooksPath .husky/_
         echo "  reverted  : ${repo#"$HOME"/}"
         fixed=$((fixed + 1))
      fi
      continue
   fi

   [[ -z "$local_path" || "$local_path" == "$GUARD_DIR" ]] && continue

   overridden=$((overridden + 1))
   if [[ "$MODE" == "apply" ]]; then
      git -C "$repo" config core.hooksPath "$GUARD_DIR"
      echo "  repointed : ${repo#"$HOME"/}  (was $local_path)"
      fixed=$((fixed + 1))
   else
      echo "  UNGUARDED : ${repo#"$HOME"/}  ->  $local_path"
   fi
done < <(find_repos)

echo
case "$MODE" in
   report)
      if [[ "$overridden" -eq 0 ]]; then
         echo "All repos are covered by the guard. ($disabled explicitly opted out.)"
      else
         echo "$overridden repo(s) override core.hooksPath and are NOT guarded."
         echo "Run '$0 --apply' to fix them."
      fi
      ;;
   apply)  echo "Repointed $fixed repo(s) at the guard. ($disabled opted out.)" ;;
   revert) echo "Reverted $fixed repo(s) to husky." ;;
esac
