package dev

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	commitCopy   bool
	commitNoCopy bool
	commitExec   bool
)

var commitCmd = &cobra.Command{
	Use:   "commit [text]",
	Short: "Generate git branch checkout from JIRA ticket description",
	Long: `Generate a git checkout command for a feature branch from JIRA ticket.

Semantically indicates "committing" to work on a JIRA ticket by creating
a feature branch. Functionally identical to 'branch' command.

Examples:
  hlp commit "VIS-6968 Kiosk > TinyMC > NDA does not display HTML"
  hlp commit "TDT-123 Fix login bug" --no-copy
  hlp commit "VIS-1234 Update config" --exec`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		result := format.JiraBranch(text)

		if commitExec {
			fmt.Println(result)
			return
		}

		if !commitNoCopy && commitCopy {
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
	commitCmd.Flags().BoolVarP(&commitCopy, "copy", "c", true, "copy to clipboard")
	commitCmd.Flags().BoolVar(&commitNoCopy, "no-copy", false, "do not copy to clipboard")
	commitCmd.Flags().BoolVarP(&commitExec, "exec", "e", false, "output for shell execution")
}
