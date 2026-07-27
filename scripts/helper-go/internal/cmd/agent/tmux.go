package agent

import (
	"fmt"
	"os/exec"
	"strings"
)

// agentStartedEnvVar marks a session where an agent launch was already sent
const agentStartedEnvVar = "HLP_AGENT_STARTED"

func tmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// sessionExists reports whether a tmux session with the exact name exists
// (the = prefix disables tmux prefix-matching).
func sessionExists(name string) bool {
	return exec.Command("tmux", "has-session", "-t="+name).Run() == nil
}

// sessionPath returns the session's working directory.
// display-message takes a PANE target: "=name:" = exact session, current window.
func sessionPath(name string) (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "-t="+name+":", "#{session_path}").Output()
	if err != nil {
		return "", fmt.Errorf("failed to read session path: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// newSession creates a detached session with the given name and working dir
func newSession(name, dir string) error {
	out, err := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", dir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux session %s: %s", name, strings.TrimSpace(string(out)))
	}
	return nil
}

// sendKeys types a command into the session's active pane and presses Enter.
// send-keys takes a PANE target: "=name" alone is not resolved as a session,
// so the trailing ":" (exact session, current window) is required.
func sendKeys(name, command string) error {
	out, err := exec.Command("tmux", "send-keys", "-t="+name+":", command, "Enter").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send keys to %s: %s", name, strings.TrimSpace(string(out)))
	}
	return nil
}

// paneCommands returns the current foreground command of every pane in the session
func paneCommands(name string) []string {
	out, err := exec.Command("tmux", "list-panes", "-s", "-t="+name, "-F", "#{pane_current_command}").Output()
	if err != nil {
		return nil
	}
	var cmds []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			cmds = append(cmds, line)
		}
	}
	return cmds
}

// agentRunning reports whether a claude agent appears active in the session
func agentRunning(name string) bool {
	for _, cmd := range paneCommands(name) {
		if cmd == "claude" || cmd == "node" {
			return true
		}
	}
	return false
}

// setSessionEnv sets a session-scoped tmux environment variable
func setSessionEnv(name, key, value string) error {
	return exec.Command("tmux", "set-environment", "-t="+name, key, value).Run()
}

// getSessionEnv reads a session-scoped tmux environment variable ("" if unset)
func getSessionEnv(name, key string) string {
	out, err := exec.Command("tmux", "show-environment", "-t="+name, key).Output()
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(out))
	if value, found := strings.CutPrefix(line, key+"="); found {
		return value
	}
	return ""
}

// killSession terminates the tmux session
func killSession(name string) error {
	out, err := exec.Command("tmux", "kill-session", "-t="+name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill session %s: %s", name, strings.TrimSpace(string(out)))
	}
	return nil
}

// interruptSession sends Escape + Ctrl-C to the active pane (graceful stop)
func interruptSession(name string) error {
	if err := exec.Command("tmux", "send-keys", "-t="+name+":", "Escape").Run(); err != nil {
		return err
	}
	return exec.Command("tmux", "send-keys", "-t="+name+":", "C-c").Run()
}
