package agent

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// tmuxSocket lets tests (and any risky experiment) run against an isolated tmux
// server. Fedora's tmux 3.7 can take the whole server down, so never test
// against the default socket.
func tmuxSocket() []string {
	if socket := os.Getenv("HLP_TMUX_SOCKET"); socket != "" {
		return []string{"-L", socket}
	}
	return nil
}

func tmuxCmd(args ...string) *exec.Cmd {
	return exec.Command("tmux", append(tmuxSocket(), args...)...)
}

// newSessionWithWindow creates a detached session whose first window is named
// after a repo, so every window in the session follows one naming rule.
func newSessionWithWindow(session, window, dir string) error {
	out, err := tmuxCmd("new-session", "-d", "-s", session, "-n", window, "-c", dir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux session %s: %s", session, strings.TrimSpace(string(out)))
	}
	return nil
}

// listWindowNames returns the window names in the session, in tmux order.
func listWindowNames(session string) []string {
	out, err := tmuxCmd("list-windows", "-t="+session, "-F", "#{window_name}").Output()
	if err != nil {
		return nil
	}
	var names []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			names = append(names, line)
		}
	}
	return names
}

func windowExists(session, window string) bool {
	return slices.Contains(listWindowNames(session), window)
}

// newWindow appends a window to an existing session.
func newWindow(session, window, dir string) error {
	out, err := tmuxCmd("new-window", "-d", "-t="+session, "-n", window, "-c", dir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create window %s in %s: %s", window, session, strings.TrimSpace(string(out)))
	}
	return nil
}

// killWindow removes a window by name. The target is session:window, which tmux
// resolves exactly thanks to the = prefix on the session part.
func killWindow(session, window string) error {
	out, err := tmuxCmd("kill-window", "-t="+session+":"+window).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill window %s in %s: %s", window, session, strings.TrimSpace(string(out)))
	}
	return nil
}

// sendKeysToWindow types a command into a specific window's active pane.
func sendKeysToWindow(session, window, command string) error {
	out, err := tmuxCmd("send-keys", "-t="+session+":"+window, command, "Enter").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send keys to %s:%s: %s", session, window, strings.TrimSpace(string(out)))
	}
	return nil
}

// windowPath reports a window's active pane working directory.
func windowPath(session, window string) string {
	out, err := tmuxCmd("display-message", "-p", "-t="+session+":"+window, "#{pane_current_path}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// windowCommands returns the foreground command of every pane in one window.
func windowCommands(session, window string) []string {
	out, err := tmuxCmd("list-panes", "-t="+session+":"+window, "-F", "#{pane_current_command}").Output()
	if err != nil {
		return nil
	}
	var cmds []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			cmds = append(cmds, line)
		}
	}
	return cmds
}

// claudeRunningInWindow reports whether a Claude process holds a pane in the
// window, used to avoid relaunching over a live session.
func claudeRunningInWindow(session, window string) bool {
	for _, cmd := range windowCommands(session, window) {
		if cmd == "claude" || cmd == "node" {
			return true
		}
	}
	return false
}

// syncWindows makes the session's windows match the task's repos: it appends a
// window for each new repo and removes windows for repos that are gone. Windows
// the user created by hand are left alone — only names in known are candidates
// for removal.
func syncWindows(session string, want []windowSpec, known map[string]bool) error {
	existing := map[string]bool{}
	for _, name := range listWindowNames(session) {
		existing[name] = true
	}

	wanted := map[string]bool{}
	for _, w := range want {
		wanted[w.Name] = true
		if existing[w.Name] {
			continue
		}
		if err := newWindow(session, w.Name, w.Dir); err != nil {
			return err
		}
	}

	for name := range existing {
		if !wanted[name] && known[name] {
			if err := killWindow(session, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// windowSpec is one repo's window: its name and working directory.
type windowSpec struct {
	Name string
	Dir  string
}

// selectWindow makes a window the session's active one, so attaching lands on
// the primary repo rather than wherever the last new-window left the cursor.
func selectWindow(session, window string) error {
	return tmuxCmd("select-window", "-t="+session+":"+window).Run()
}

// attachSession puts the user in the session, replacing the current process.
//
// Inside tmux, attaching is impossible ("sessions should be nested with care"),
// so the client is switched instead. Outside it, exec hands the terminal over
// directly rather than running tmux as a child of a command that is about to
// exit.
//
// It only ever runs on a real terminal: from a script, a pipe or a test there is
// no terminal to hand over, and exec would replace the caller with a tmux that
// has nothing to attach to.
func attachSession(session string) error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return errNotATerminal
	}

	if os.Getenv("TMUX") != "" {
		out, err := tmuxCmd("switch-client", "-t="+session).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to switch to session %s: %s", session, strings.TrimSpace(string(out)))
		}
		return nil
	}

	bin, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}
	args := append([]string{"tmux"}, tmuxSocket()...)
	args = append(args, "attach-session", "-t="+session)
	return syscall.Exec(bin, args, os.Environ())
}

// errNotATerminal marks the case where attaching was never possible, so the
// caller can fall back to printing the command instead of warning about it.
var errNotATerminal = errors.New("not a terminal")
