package dev

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	branchCopy   bool
	branchNoCopy bool
	branchExec   bool
)

var branchCmd = &cobra.Command{
	Use:   "branch [text]",
	Short: "Generate JIRA branch name from text",
	Long: `Generate a git checkout command for a JIRA-style branch name.

Examples:
  hlp branch "VIS-1234 Add new feature"
  hlp branch "fix login bug" --no-copy
  hlp branch "TDT-99 Update config" --exec`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		result := format.JiraBranch(text)

		if branchExec {
			// Output for shell execution: $(hlp branch "text" --exec)
			fmt.Println(result)
			return
		}

		if !branchNoCopy && branchCopy {
			if err := clipboard.WriteAll(result); err != nil {
				fmt.Println(ui.Warning("Failed to copy to clipboard: " + err.Error()))
			} else {
				fmt.Println(ui.SuccessMsg("Copied to clipboard"))
			}
		}

		fmt.Println(ui.Code.Render(result))
	},
}

func init() {
	branchCmd.Flags().BoolVarP(&branchCopy, "copy", "c", true, "copy to clipboard")
	branchCmd.Flags().BoolVar(&branchNoCopy, "no-copy", false, "do not copy to clipboard")
	branchCmd.Flags().BoolVarP(&branchExec, "exec", "e", false, "output for shell execution")
}
