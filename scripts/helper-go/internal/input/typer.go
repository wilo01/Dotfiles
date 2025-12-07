package input

import (
	"os"
	"os/exec"
	"time"

	"github.com/atotto/clipboard"
)

// Typer handles keyboard input automation
type Typer interface {
	Type(text string) error
	TypeWithDelay(text string, delay time.Duration) error
	IsSupported() bool
	Backend() string
}

// typer implements Typer
type typer struct {
	backend string
}

// NewTyper creates a new Typer based on the current environment
func NewTyper() Typer {
	t := &typer{}

	// Detect available backend
	if isWayland() {
		if hasCommand("wtype") {
			t.backend = "wtype"
		} else if hasCommand("ydotool") {
			t.backend = "ydotool"
		} else {
			t.backend = "clipboard"
		}
	} else {
		// X11 or other
		if hasCommand("xdotool") {
			t.backend = "xdotool"
		} else {
			t.backend = "clipboard"
		}
	}

	return t
}

// Type types the text using the detected backend
func (t *typer) Type(text string) error {
	return t.TypeWithDelay(text, 0)
}

// TypeWithDelay types the text after a delay
func (t *typer) TypeWithDelay(text string, delay time.Duration) error {
	if delay > 0 {
		time.Sleep(delay)
	}

	switch t.backend {
	case "wtype":
		return exec.Command("wtype", text).Run()
	case "ydotool":
		return exec.Command("ydotool", "type", text).Run()
	case "xdotool":
		return exec.Command("xdotool", "type", "--", text).Run()
	default:
		// Fallback to clipboard
		return clipboard.WriteAll(text)
	}
}

// IsSupported returns true if keyboard automation is available
func (t *typer) IsSupported() bool {
	return t.backend != "clipboard"
}

// Backend returns the name of the backend being used
func (t *typer) Backend() string {
	return t.backend
}

// isWayland checks if running under Wayland
func isWayland() bool {
	return os.Getenv("XDG_SESSION_TYPE") == "wayland" ||
		os.Getenv("WAYLAND_DISPLAY") != ""
}

// hasCommand checks if a command is available in PATH
func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
