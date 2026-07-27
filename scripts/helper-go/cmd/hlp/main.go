package main

import (
	"fmt"
	"os"

	"github.com/dariuszw/hlp/internal/cmd"
	"github.com/dariuszw/hlp/internal/ui"
)

// Version information - set by ldflags during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error(err.Error()))
		os.Exit(1)
	}
}
