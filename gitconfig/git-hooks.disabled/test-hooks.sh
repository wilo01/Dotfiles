#!/bin/bash
# Exercises the protected-branch guard end to end in a throwaway repo with a local
# bare remote. Touches nothing outside its own temp directory.
#
#   ./test-hooks.sh

set -u

HOOKS_DIR=$(dirname "$(readlink -f "$0")")
WORK=$(mktemp -d)
PASS=0
FAIL=0

cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

ok()   { PASS=$((PASS + 1)); echo "  ✅ $*"; }
bad()  { FAIL=$((FAIL + 1)); echo "  ❌ $*"; }

expect_exit() {
   local want="$1" got="$2" what="$3"
   if [[ "$got" == "$want" ]]; then ok "$what (exit $got)"; else bad "$what — expected exit $want, got $got"; fi
}

echo "Hooks under test: $HOOKS_DIR"
echo "Configured core.hooksPath: $(git config --get core.hooksPath || echo '(unset)')"
echo "Scratch repo: $WORK"
echo

cd "$WORK" || exit 1
git init -q -b master repo
git init -q --bare remote.git
cd repo || exit 1
git config commit.gpgSign false
git config user.name "hook test"
git config user.email "hook@test.local"
git remote add origin "$WORK/remote.git"

echo "1. Commit on master"
echo hello > file.txt
git add file.txt
git commit -m "should be blocked" >/dev/null 2>"$WORK/out1"
expect_exit 1 $? "commit on master is refused"
grep -q 'BLOCKED' "$WORK/out1" && ok "refusal message shown" || bad "no BLOCKED message"
if grep -q 'no-verify' "$WORK/out1"; then
   bad "message leaks the bypass flag"
else
   ok "message does not mention any bypass flag"
fi

echo
echo "2. Commit on a working branch"
git switch -q -c feature/TEST-1-demo
git commit -q -m "should succeed" 2>/dev/null
expect_exit 0 $? "commit on feature/TEST-1-demo succeeds"
git log --oneline -1 >/dev/null 2>&1 && ok "commit landed, staged file survived the switch"

echo
echo "3. Back on master, the existing branch is offered"
# Step 1 was correctly blocked, so master has no commit yet and cannot be checked
# out. Anchor it to the work done on the feature branch.
git branch -f master HEAD
git switch -q master
echo more >> file.txt
git add file.txt
git commit -m "blocked again" >/dev/null 2>"$WORK/out3"
expect_exit 1 $? "second commit on master is refused"
if grep -q 'feature/TEST-1-demo' "$WORK/out3"; then
   ok "existing branch listed, so no second branch gets created"
else
   bad "existing branch not listed in the message"
fi

echo
echo "4. Push"
git switch -q feature/TEST-1-demo
git push -q origin master 2>"$WORK/out4"
expect_exit 1 $? "push to master is refused"
grep -q 'BLOCKED' "$WORK/out4" && ok "push refusal message shown" || bad "no BLOCKED message on push"

git push -q -u origin feature/TEST-1-demo 2>/dev/null
expect_exit 0 $? "push of the working branch succeeds"

echo
echo "5. Branches that must NOT be blocked"
for branch in Linux MacOS release/1.2; do
   git switch -q -c "$branch" 2>/dev/null
   "$HOOKS_DIR/pre-commit" >/dev/null 2>&1
   expect_exit 0 $? "'$branch' is not protected"
done

git checkout -q --detach 2>/dev/null
"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 0 $? "detached HEAD (rebase, bisect) is not blocked"

echo
echo "6. Repo-local hooks are still honoured"
git switch -q feature/TEST-1-demo
mkdir -p .git/hooks
printf '#!/bin/sh\necho LOCAL_RAN\nexit 0\n' > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
if "$HOOKS_DIR/pre-commit" 2>&1 | grep -q LOCAL_RAN; then
   ok "repo-local pre-commit still runs"
else
   bad "repo-local pre-commit was shadowed"
fi

printf '#!/bin/sh\nexit 7\n' > .git/hooks/pre-commit
"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 7 $? "repo-local hook failure propagates"

echo
echo "7. Per-repo opt-out"
rm -f .git/hooks/pre-commit
git switch -q master
ZERO=$(printf '0%.0s' {1..40})
push_master() { printf 'refs/heads/x %s refs/heads/master %s\n' "$(git rev-parse HEAD)" "$ZERO" \
                  | "$HOOKS_DIR/pre-push" origin "$WORK/remote.git" >/dev/null 2>&1; }

"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 1 $? "guard on by default"

git config hooks.branchGuardCommit false
"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 0 $? "branchGuardCommit=false disables the commit guard"
push_master
expect_exit 1 $? "  ...and leaves the push guard active"
git config --unset hooks.branchGuardCommit

git config hooks.branchGuardPush false
push_master
expect_exit 0 $? "branchGuardPush=false disables the push guard"
"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 1 $? "  ...and leaves the commit guard active"
git config --unset hooks.branchGuardPush

for value in false no off 0; do
   git config hooks.branchGuard "$value"
   "$HOOKS_DIR/pre-commit" >/dev/null 2>&1
   expect_exit 0 $? "branchGuard=$value disables both guards"
done
git config --unset hooks.branchGuard

"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 1 $? "guard re-enabled once the setting is removed"

git config hooks.protectedBranches "develop"
"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 0 $? "protectedBranches override narrows the list (master no longer protected)"
git config --unset hooks.protectedBranches

echo
echo "8. Husky co-existence"
git switch -q feature/TEST-1-demo
rm -f .git/hooks/pre-commit
mkdir -p .husky/_
printf '#!/bin/sh\n. "$(dirname "$0")/h"\n' > .husky/_/pre-commit
printf '#!/bin/sh\nsh "$(dirname "$0")/../pre-commit"\n' > .husky/_/h
printf '#!/bin/sh\necho HUSKY_LINT_STAGED_RAN\n' > .husky/pre-commit
chmod +x .husky/_/pre-commit .husky/_/h .husky/pre-commit

if "$HOOKS_DIR/pre-commit" 2>&1 | grep -q HUSKY_LINT_STAGED_RAN; then
   ok "husky hook still runs when the guard owns core.hooksPath"
else
   bad "husky hook was shadowed — lint-staged would silently stop running"
fi

git switch -q master
"$HOOKS_DIR/pre-commit" >/dev/null 2>&1
expect_exit 1 $? "guard still blocks master in a husky repo"

echo
echo "───────────────────────────────"
echo "  passed: $PASS   failed: $FAIL"
[[ "$FAIL" -eq 0 ]]
