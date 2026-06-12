package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

type ignoreResult struct {
	pattern string
	added   bool
}

var (
	ignoreList  bool
	ignoreCheck bool
	ignoreEdit  bool
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore <pattern> [pattern...]",
	Short: "Add patterns to global gitignore",
	Long: `Add one or more patterns to the global gitignore file.

Examples:
  hlp git ignore .taskmaster/ todo.md .env
  hlp git ignore --list
  hlp git ignore --edit
  hlp git ignore --check .env todo.md`,
	Run: runIgnore,
}

func init() {
	ignoreCmd.Flags().BoolVarP(&ignoreList, "list", "l", false, "show current gitignore contents")
	ignoreCmd.Flags().BoolVarP(&ignoreCheck, "check", "c", false, "check if patterns exist without adding")
	ignoreCmd.Flags().BoolVarP(&ignoreEdit, "edit", "e", false, "open gitignore in nvim")
}

func runIgnore(cmd *cobra.Command, args []string) {
	path, err := gitignorePath()
	if err != nil {
		fmt.Println(ui.Error("Failed to resolve gitignore path: " + err.Error()))
		return
	}

	if ignoreEdit {
		editor := exec.Command("nvim", path)
		editor.Stdin = os.Stdin
		editor.Stdout = os.Stdout
		editor.Stderr = os.Stderr
		if err := editor.Run(); err != nil {
			fmt.Println(ui.Error("Failed to open editor: " + err.Error()))
		}
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(ui.Error("Failed to read " + path + ": " + err.Error()))
		return
	}

	existing := parsePatterns(string(content))

	if ignoreList {
		fmt.Println(ui.Header("Global gitignore"))
		fmt.Println()
		fmt.Println(string(content))
		return
	}

	if len(args) == 0 {
		fmt.Println(ui.Error("Provide at least one pattern, or use --list"))
		return
	}

	if ignoreCheck {
		for _, p := range args {
			if existing[strings.TrimSpace(p)] {
				fmt.Println(ui.SuccessMsg(p + " — present"))
			} else {
				fmt.Println(ui.Warning(p + " — not found"))
			}
		}
		return
	}

	var results []ignoreResult
	var toAdd []string
	for _, p := range args {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if existing[p] {
			results = append(results, ignoreResult{p, false})
		} else {
			results = append(results, ignoreResult{p, true})
			toAdd = append(toAdd, p)
		}
	}

	if len(toAdd) > 0 {
		text := string(content)
		if len(text) > 0 && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		if len(text) > 0 && !strings.HasSuffix(text, "\n\n") {
			text += "\n"
		}
		text += strings.Join(toAdd, "\n") + "\n"

		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			fmt.Println(ui.Error("Failed to write " + path + ": " + err.Error()))
			return
		}
	}

	newText := string(content)
	if len(newText) > 0 && !strings.HasSuffix(newText, "\n") {
		newText += "\n"
	}
	if len(toAdd) > 0 {
		newText += "\n" + strings.Join(toAdd, "\n") + "\n"
	}

	showDiff(string(content), newText, results)
}

func showDiff(old, new string, results []ignoreResult) {
	diff := buildUnifiedDiff(old, new, results)
	if diff == "" {
		return
	}

	delta, err := exec.LookPath("delta")
	if err != nil {
		fmt.Print(diff)
		return
	}

	cmd := exec.Command(delta, "--file-style", "omit", "--hunk-header-style", "line-number")
	cmd.Stdin = strings.NewReader(diff)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Print(diff)
	}
}

func buildUnifiedDiff(old, new string, results []ignoreResult) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var sb strings.Builder
	sb.WriteString("--- a/.gitignore\n")
	sb.WriteString("+++ b/.gitignore\n")

	existingSet := make(map[string]bool)
	for _, r := range results {
		if !r.added {
			existingSet[r.pattern] = true
		}
	}

	for _, r := range results {
		if r.added {
			continue
		}
		for i, line := range oldLines {
			if strings.TrimSpace(line) == r.pattern {
				writeContextHunk(&sb, oldLines, i)
				break
			}
		}
	}

	addedCount := 0
	for _, r := range results {
		if r.added {
			addedCount++
		}
	}
	if addedCount > 0 {
		oldEnd := len(oldLines)
		newEnd := len(newLines)
		ctx := 7
		oldStart := oldEnd - ctx
		if oldStart < 0 {
			oldStart = 0
		}
		newStart := newEnd - addedCount - ctx
		if newStart < 0 {
			newStart = 0
		}

		sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			oldStart+1, oldEnd-oldStart,
			newStart+1, newEnd-newStart))

		for j := oldStart; j < oldEnd; j++ {
			sb.WriteString(" " + oldLines[j] + "\n")
		}
		for _, r := range results {
			if r.added {
				sb.WriteString("+" + r.pattern + "\n")
			}
		}
	}

	return sb.String()
}

func writeContextHunk(sb *strings.Builder, lines []string, matchIdx int) {
	start := matchIdx - 7
	if start < 0 {
		start = 0
	}
	end := matchIdx + 8
	if end > len(lines) {
		end = len(lines)
	}

	sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@ already exists\n",
		start+1, end-start, start+1, end-start))

	for j := start; j < end; j++ {
		sb.WriteString(" " + lines[j] + "\n")
	}
}

func gitignorePath() (string, error) {
	out, err := exec.Command("git", "config", "--global", "core.excludesfile").Output()
	if err != nil {
		return "", fmt.Errorf("core.excludesfile not set: %w", err)
	}
	path := strings.TrimSpace(string(out))
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = home + path[1:]
	}
	return path, nil
}

func parsePatterns(content string) map[string]bool {
	existing := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		existing[line] = true
	}
	return existing
}
