package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	internalJira "github.com/dariuszw/hlp/internal/jira"
)

// expandPath expands a leading ~ to the user's home directory
func expandPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}

// SourceRepo is a main checkout that task worktrees are created from.
type SourceRepo struct {
	Name string
	Path string
}

// reposRootSkip lists directory names under the repos root that are never
// source repos, whatever they contain.
var reposRootSkip = map[string]bool{"tasks": true}

// discoverSourceRepos lists the repos directly under root. "Is a repo" is a
// stat of .git, deliberately not `git rev-parse`, which walks up and would
// classify any directory inside an enclosing repo as a repo.
func discoverSourceRepos(root string) ([]SourceRepo, error) {
	root = expandPath(root)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos root %s: %w", root, err)
	}

	var repos []SourceRepo
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || strings.HasPrefix(name, ".") || reposRootSkip[name] {
			continue
		}
		path := filepath.Join(root, name)
		if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
			continue
		}
		repos = append(repos, SourceRepo{Name: name, Path: path})
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	return repos, nil
}

// resolveDefaultBase returns the repo's default branch. Not every clone has
// origin/HEAD set (a plain `git clone` of some repos leaves it unset), so this
// falls back through progressively more expensive lookups before guessing.
func resolveDefaultBase(source string) (string, error) {
	if base := symbolicOriginHead(source); base != "" {
		return base, nil
	}

	exec.Command("git", "-C", source, "remote", "set-head", "origin", "-a").Run()
	if base := symbolicOriginHead(source); base != "" {
		return base, nil
	}

	if out, err := exec.Command("git", "-C", source, "ls-remote", "--symref", "origin", "HEAD").Output(); err == nil {
		for line := range strings.SplitSeq(string(out), "\n") {
			if rest, found := strings.CutPrefix(line, "ref: refs/heads/"); found {
				if name, _, ok := strings.Cut(rest, "\t"); ok {
					return strings.TrimSpace(name), nil
				}
			}
		}
	}

	for _, candidate := range []string{"main", "master"} {
		if remoteRefExists(source, candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("cannot determine the default branch for %s: origin/HEAD is unset and neither origin/main nor origin/master exists", source)
}

func symbolicOriginHead(source string) string {
	out, err := exec.Command("git", "-C", source, "symbolic-ref", "--short", "refs/remotes/origin/HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "origin/")
}

func remoteRefExists(source, branch string) bool {
	return exec.Command("git", "-C", source, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+branch).Run() == nil
}

// refreshSource brings the main checkout up to date before a worktree is cut
// from it. The fetch is required — a stale origin/<base> would silently give a
// worktree from yesterday. Updating the checkout itself is best-effort: it
// fails whenever the clone is dirty or parked on a feature branch, and that
// does not matter, because addDetachedWorktree branches from the remote ref.
// It returns its warnings rather than printing them: the caller runs this under
// a spinner that owns the current line, and printing into it corrupts both.
func refreshSource(source, base string) ([]string, error) {
	if out, err := exec.Command("git", "-C", source, "fetch", "origin", "--prune").CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git fetch origin failed in %s: %s", source, strings.TrimSpace(string(out)))
	}

	name := filepath.Base(source)
	if out, err := exec.Command("git", "-C", source, "checkout", base).CombinedOutput(); err != nil {
		return []string{fmt.Sprintf("%s: could not check out %s, leaving the clone as-is (the worktree still comes from the freshly fetched origin/%s): %s",
			name, base, base, staleLockHint(firstLine(string(out)), source))}, nil
	}
	if out, err := exec.Command("git", "-C", source, "pull", "--ff-only").CombinedOutput(); err != nil {
		return []string{fmt.Sprintf("%s: could not fast-forward %s: %s",
			name, base, staleLockHint(firstLine(string(out)), source))}, nil
	}
	return nil, nil
}

// staleLockHint appends the fix for the most common cause of a refusal: an
// index.lock left behind by a git process that died, which no longer has an
// owner and will otherwise block every future run.
func staleLockHint(message, source string) string {
	lock := filepath.Join(source, ".git", "index.lock")
	if !strings.Contains(message, "index.lock") {
		return message
	}
	if info, err := os.Stat(lock); err == nil {
		return fmt.Sprintf("%s (lock is from %s — if no git process is running, remove it: rm %s)",
			message, info.ModTime().Format("2006-01-02 15:04"), lock)
	}
	return message
}

// addDetachedWorktree checks out origin/<base> at path with no branch. Detached
// is required rather than cosmetic: `git worktree add <path> master` is refused
// outright while master is checked out in the source clone. Claude creates the
// real branch when it starts work.
func addDetachedWorktree(source, path, base string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create worktree parent dir: %w", err)
	}
	args := []string{"-C", source, "worktree", "add", "--detach", path, "origin/" + base}
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add failed for %s: %s", filepath.Base(source), strings.TrimSpace(string(out)))
	}
	return nil
}

// worktreeSeedFiles are per-repo local files git does not track but a working
// checkout needs.
var worktreeSeedFiles = []string{".env", ".env.local", filepath.Join(".claude", "settings.local.json")}

// seedWorktree copies untracked local config from the source checkout into a
// fresh worktree. Best-effort by design: a repo without a .env is normal.
func seedWorktree(source, worktree string) {
	for _, rel := range worktreeSeedFiles {
		copyFileIfExists(filepath.Join(source, rel), filepath.Join(worktree, rel))
	}
}

// seedTaskRoot copies the shared repos-root config into the task umbrella dir so
// Claude picks up the same instructions it would in the repos root.
func seedTaskRoot(reposRoot, taskRoot string) {
	reposRoot = expandPath(reposRoot)
	for _, rel := range []string{"CLAUDE.md", ".env", ".env.local"} {
		copyFileIfExists(filepath.Join(reposRoot, rel), filepath.Join(taskRoot, rel))
	}
	copyDirIfExists(filepath.Join(reposRoot, ".claude"), filepath.Join(taskRoot, ".claude"))
}

func copyFileIfExists(src, dst string) {
	info, err := os.Stat(src)
	if err != nil || info.IsDir() {
		return
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return
	}
	os.WriteFile(dst, data, info.Mode().Perm())
}

func copyDirIfExists(src, dst string) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyDirIfExists(s, d)
			continue
		}
		copyFileIfExists(s, d)
	}
}

// worktreeDirty reports whether a worktree has uncommitted changes, so removal
// can refuse rather than discard work.
func worktreeDirty(path string) bool {
	out, err := exec.Command("git", "-C", path, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// worktreeUnpushed reports whether the worktree is on a branch with commits
// that exist nowhere on origin. A detached worktree that was never branched has
// nothing to lose and reports false.
func worktreeUnpushed(path string) bool {
	out, err := exec.Command("git", "-C", path, "rev-list", "--count", "HEAD", "--not", "--remotes").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != "0"
}

// removeWorktree detaches a worktree from its source repo. git refuses dirty
// worktrees unless forced, and that refusal is kept deliberately.
func removeWorktree(source, path string, force bool) error {
	args := []string{"-C", source, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove failed for %s: %s", path, strings.TrimSpace(string(out)))
	}
	return nil
}

// worktreeHead describes what a worktree currently has checked out, for status
// output: a branch name, or "detached@<short-sha>".
func worktreeHead(path string) string {
	out, err := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "?"
	}
	head := strings.TrimSpace(string(out))
	if head != "HEAD" {
		return head
	}
	short, err := exec.Command("git", "-C", path, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "detached"
	}
	return "detached@" + strings.TrimSpace(string(short))
}

// cloneURL turns a bare repo name into a clone URL via the configured template,
// and passes anything that already looks like a URL through untouched.
func cloneURL(nameOrURL, template string) string {
	if strings.Contains(nameOrURL, "://") || strings.Contains(nameOrURL, "@") {
		return nameOrURL
	}
	return strings.ReplaceAll(template, "{{REPO}}", nameOrURL)
}

// repoNameFromURL derives the directory name a clone URL would produce.
func repoNameFromURL(url string) string {
	name := strings.TrimSuffix(url, ".git")
	if idx := strings.LastIndexAny(name, "/:"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

// cloneRepo clones into the repos root so the new repo shows up in the picker.
func cloneRepo(reposRoot, url string) (SourceRepo, error) {
	reposRoot = expandPath(reposRoot)
	name := repoNameFromURL(url)
	dest := filepath.Join(reposRoot, name)

	if _, err := os.Stat(dest); err == nil {
		return SourceRepo{}, fmt.Errorf("%s already exists", dest)
	}
	if err := os.MkdirAll(reposRoot, 0o755); err != nil {
		return SourceRepo{}, fmt.Errorf("failed to create repos root: %w", err)
	}

	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return SourceRepo{}, fmt.Errorf("git clone %s failed: %w", url, err)
	}
	return SourceRepo{Name: name, Path: dest}, nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

// listBaseBranches returns the branches worth cutting a worktree from: the
// repo's default, then every remote branch matching one of the configured
// long-lived patterns, most recently committed first.
//
// It filters rather than listing everything on purpose — tds-suite alone has
// over 4600 remote branches, and a base is a release line, never someone's
// feature branch.
func listBaseBranches(source, defaultBranch string, patterns []string) []string {
	branches := []string{defaultBranch}
	if len(patterns) == 0 {
		return branches
	}

	args := []string{"-C", source, "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)"}
	for _, p := range patterns {
		args = append(args, "refs/remotes/origin/"+strings.TrimPrefix(p, "origin/"))
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return branches
	}

	seen := map[string]bool{defaultBranch: true}
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		name := strings.TrimPrefix(strings.TrimSpace(line), "origin/")
		if name == "" || name == "HEAD" || seen[name] {
			continue
		}
		seen[name] = true
		branches = append(branches, name)
	}
	return branches
}

// fetchDescription pulls a ticket's body as plain text. It is advisory context
// for the picker, so a failure degrades to no description rather than blocking
// the run.
func fetchDescription(client *internalJira.Client, key string) string {
	description, err := client.GetTicketDescription(key)
	if err != nil {
		return ""
	}
	return description
}

// repoMentionStopWords are alias fragments too common in TDS ticket text to
// carry any signal — "suite" appears in the product name on nearly every ticket.
var repoMentionStopWords = map[string]bool{"suite": true, "internal": true, "cpp": true}

// mentionedRepos flags the repos a ticket's text names, as a hint for which
// worktrees the work needs. It matches the full repo name and, when distinctive
// enough, the part after the "tds-" prefix.
//
// The ticket key is stripped first: without that, "SUITE-9250" would mark
// tds-suite on every ticket in the project.
func mentionedRepos(ticket *internalJira.Ticket, description string, sources []SourceRepo) map[string]bool {
	haystack := strings.ToLower(ticket.Summary + "\n" + description)
	haystack = strings.ReplaceAll(haystack, strings.ToLower(ticket.Key), " ")

	mentioned := map[string]bool{}
	for _, source := range sources {
		for _, alias := range repoAliases(source.Name) {
			if containsWord(haystack, alias) {
				mentioned[source.Name] = true
				break
			}
		}
	}
	return mentioned
}

// repoAliases are the names a ticket might call a repo by: the directory name
// itself, the name without the "tds-" prefix, and the distinguishing first
// segment of that — prose says "the kiosk app", never "tds-kiosk-chrome-app".
//
// Short or generic fragments are dropped, since a hint that fires on every
// ticket is worse than no hint.
func repoAliases(repoName string) []string {
	name := strings.ToLower(repoName)
	aliases := []string{name}

	suffix := strings.TrimPrefix(name, "tds-")
	for _, candidate := range []string{suffix, firstSegment(suffix)} {
		if candidate == name || len(candidate) < 4 || repoMentionStopWords[candidate] {
			continue
		}
		if !slices.Contains(aliases, candidate) {
			aliases = append(aliases, candidate)
		}
	}
	return aliases
}

func firstSegment(name string) string {
	if idx := strings.IndexByte(name, '-'); idx > 0 {
		return name[:idx]
	}
	return name
}

// containsWord reports whether needle appears in haystack bounded by something
// other than a letter or digit, so "api" does not match "rapid".
func containsWord(haystack, needle string) bool {
	for offset := 0; ; {
		idx := strings.Index(haystack[offset:], needle)
		if idx < 0 {
			return false
		}
		start := offset + idx
		end := start + len(needle)
		if !isWordByte(haystack, start-1) && !isWordByte(haystack, end) {
			return true
		}
		offset = start + 1
	}
}

func isWordByte(s string, i int) bool {
	if i < 0 || i >= len(s) {
		return false
	}
	c := s[i]
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}
