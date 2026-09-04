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
	return tmuxCmd("has-session", "-t="+name).Run() == nil
}

// setSessionEnv sets a session-scoped tmux environment variable
func setSessionEnv(name, key, value string) error {
	return tmuxCmd("set-environment", "-t="+name, key, value).Run()
}

// killSession terminates the tmux session
func killSession(name string) error {
	out, err := tmuxCmd("kill-session", "-t="+name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill session %s: %s", name, strings.TrimSpace(string(out)))
	}
	return nil
}

// interruptSession sends Escape + Ctrl-C to the active pane (graceful stop)
func interruptSession(name string) error {
	if err := tmuxCmd("send-keys", "-t="+name+":", "Escape").Run(); err != nil {
		return err
	}
	return tmuxCmd("send-keys", "-t="+name+":", "C-c").Run()
}
