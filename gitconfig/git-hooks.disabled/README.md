# Protected-branch guard (disabled)

Blocks commits and pushes to `master` / `main` / `develop`. Currently **inactive** —
this directory is not referenced by `core.hooksPath`, so git never runs it.

| File | |
|---|---|
| `branch-guard.sh` | `is_protected_branch`, `guard_enabled`, `candidate_branches`, refusal messages, `delegate_to_local_hook` |
| `pre-commit` / `pre-push` | The hooks themselves |
| `install-guard.sh` | Repoints repos that set `core.hooksPath` locally (husky) at the guard |
| `test-hooks.sh` | 29-check regression suite, runs in a throwaway repo |

## Re-enable

```bash
cd ~/.Dotfiles
mv gitconfig/git-hooks.disabled gitconfig/git-hooks
git config --global core.hooksPath /home/dariuszw/gitconfig/git-hooks
~/gitconfig/git-hooks/install-guard.sh --apply
~/gitconfig/git-hooks/test-hooks.sh
```

The `install-guard.sh` step matters: ~10 repos (all the `tds-hexer` checkouts, plus
`devspace` and one `.githooks` repo) set `core.hooksPath` locally, and local config beats
global — without it the guard is silently inactive in exactly those repos.

## Disable again

```bash
~/gitconfig/git-hooks/install-guard.sh --revert   # hand husky repos their hooksPath back
git config --global --unset core.hooksPath
cd ~/.Dotfiles && mv gitconfig/git-hooks gitconfig/git-hooks.disabled
```

`--revert` only restores husky repos. `Dev/Other/branch-opener2-bakcup/branches/safe` uses
`.githooks` and must be reset by hand.

## Notes

- Per-repo opt-out while active: `git config hooks.branchGuard false`, or
  `hooks.branchGuardCommit` / `hooks.branchGuardPush` for one hook, or
  `hooks.protectedBranches "master maintenance/13.2AV"` to change the list.
- `Dev/Private` has `hooks.branchguard false` set and will stay exempt on re-enable.
- Husky's `prepare` script resets `core.hooksPath` on every `npm`/`pnpm install`, which
  silently deactivates the guard in that repo. Re-run `install-guard.sh --apply` after
  installing dependencies.
- The refusal text deliberately never names `--no-verify`, so an agent reading it does not
  learn the bypass. Real enforcement needs a Claude Code `PreToolUse` hook rejecting that
  flag; that was never installed.
