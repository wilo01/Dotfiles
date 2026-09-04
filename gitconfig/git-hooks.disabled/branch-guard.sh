#!/bin/bash
# Shared helpers for the protected-branch guard hooks.
#
# The refusal messages are read by autonomous agents as much as by humans, so they
# are written as instructions ("switch to this branch") rather than as errors, and
# they deliberately never name a bypass flag.

PROTECTED_BRANCHES_DEFAULT="master main develop"

# Per-repo opt-out. $1 is the hook suffix: Commit or Push.
#
#   git config hooks.branchGuard false        # disable both hooks here
#   git config hooks.branchGuardCommit false  # disable the commit guard only
#   git config hooks.branchGuardPush false    # disable the push guard only
#
# --type=bool normalises false/no/off/0, so any of those spellings work.
guard_enabled() {
   local setting
   for setting in hooks.branchGuard "hooks.branchGuard${1}"; do
      [[ "$(git config --type=bool --get "$setting" 2>/dev/null || echo true)" == "false" ]] && return 1
   done
   return 0
}

# Returns 0 if $1 names a branch that must not be committed or pushed to directly.
is_protected_branch() {
   local branch="${1,,}"
   local protected
   protected=$(git config --get hooks.protectedBranches || echo "$PROTECTED_BRANCHES_DEFAULT")

   local candidate
   for candidate in $protected; do
      [[ "$branch" == "${candidate,,}" ]] && return 0
   done
   return 1
}

current_branch() {
   git symbolic-ref --short HEAD 2>/dev/null || echo "HEAD"
}

# Recent local branches that are legitimate commit targets, most recent first.
candidate_branches() {
   local name age
   while IFS='|' read -r name age; do
      is_protected_branch "$name" && continue
      printf '%-38s (%s)\n' "$name" "$age"
   done < <(git for-each-ref --sort=-committerdate --count=8 \
               --format='%(refname:short)|%(committerdate:relative)' refs/heads/) \
      | head -n 5
}

refuse_commit() {
   local branch="$1"
   local candidates
   candidates=$(candidate_branches)

   {
      echo ""
      echo "✋ BLOCKED — direct commit to protected branch '${branch}' is not allowed."
      echo ""
      echo "   HOW TO PROCEED (in this order):"
      echo ""
      if [[ -n "$candidates" ]]; then
         echo "   1. A working branch for this task may already exist. If one of these is"
         echo "      yours, switch to it and commit there — do NOT make another branch:"
         echo ""
         sed 's/^/          /' <<<"$candidates"
         echo ""
         echo "          git switch <branch-from-the-list-above>"
         echo "          git commit ..."
         echo ""
         echo "   2. Only if NO branch for this task exists yet, create exactly ONE:"
      else
         echo "   1. Create exactly ONE working branch for this task:"
      fi
      echo ""
      echo "          git switch -c feature/TICKET-123-short-description"
      echo "          git commit ..."
      echo ""
      echo "   RULES:"
      echo "   - One branch per task. Once you have branched out, every further commit"
      echo "     goes onto that same branch. Do not branch again."
      echo "   - Do not attempt to bypass this check. Switch branches instead."
      echo ""
      echo "   Your staged changes are preserved across 'git switch' — nothing is lost."
      echo ""
   } >&2
   return 1
}

refuse_push() {
   local branch="$1" remote="$2"
   local working
   working=$(current_branch)
   [[ "$working" == "HEAD" ]] && working="feature/TICKET-123-short-description"

   {
      echo ""
      echo "✋ BLOCKED — pushing to protected branch '${branch}' on '${remote}' is not allowed."
      echo ""
      echo "   Push your working branch instead and open a pull request:"
      echo ""
      echo "       git push -u ${remote} ${working}"
      echo "       gh pr create --draft"
      echo ""
      echo "   RULES:"
      echo "   - Never push directly to master / main / develop."
      echo "   - Do not attempt to bypass this check."
      echo ""
   } >&2
   return 1
}

# core.hooksPath shadows every repo's own .git/hooks, so re-run the repo-local hook
# by hand to avoid silently disabling it.
delegate_to_local_hook() {
   local hook_name="$1"
   shift

   [[ "${GIT_BRANCH_GUARD_ACTIVE:-}" == "1" ]] && return 0

   local top
   top=$(git rev-parse --show-toplevel 2>/dev/null)

   # .husky/_/<hook> is husky's own runner, which in turn runs .husky/<hook>. Taking
   # over core.hooksPath from husky would otherwise silently kill lint-staged.
   local candidate
   for candidate in \
      "$(git rev-parse --git-dir)/hooks/${hook_name}" \
      "${top:+$top/.husky/_/${hook_name}}" \
      "${top:+$top/.husky/${hook_name}}"
   do
      [[ -n "$candidate" && -x "$candidate" ]] || continue
      GIT_BRANCH_GUARD_ACTIVE=1 "$candidate" "$@"
      return $?
   done
   return 0
}
