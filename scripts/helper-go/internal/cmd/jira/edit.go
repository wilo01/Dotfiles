package jira

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open worklogs CSV in neovim",
	Long: `Open the worklogs CSV file in neovim for manual editing.

The file path is configurable via jira.worklogs_file in config.yaml.
Default: ~/.config/hlp/worklogs.csv`,
	Run: runEdit,
}

func runEdit(_ *cobra.Command, _ []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(ui.Error("Failed to load config: " + err.Error()))
		return
	}

	path := cfg.Jira.WorklogsFile
	if path == "" {
		path = "~/.config/hlp/worklogs.csv"
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(ui.Error("Failed to resolve home directory: " + err.Error()))
			return
		}
		path = home + path[1:]
	}

	editor := exec.Command("nvim", path)
	editor.Stdin = os.Stdin
	editor.Stdout = os.Stdout
	editor.Stderr = os.Stderr
	if err := editor.Run(); err != nil {
		fmt.Println(ui.Error("Failed to open editor: " + err.Error()))
	}
}
