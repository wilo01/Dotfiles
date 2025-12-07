package dev

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	stashCopy   bool
	stashNoCopy bool
	stashExec   bool
)

var stashCmd = &cobra.Command{
	Use:   "stash [message]",
	Short: "Generate git stash command with formatted message",
	Long: `Generate a git stash push command with a formatted message.

Examples:
  hlp stash "work in progress"
  hlp stash "VIS-1234 saving changes" --exec`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		result := format.Stash(text)

		if stashExec {
			fmt.Println(result)
			return
		}

		if !stashNoCopy && stashCopy {
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
	stashCmd.Flags().BoolVarP(&stashCopy, "copy", "c", true, "copy to clipboard")
	stashCmd.Flags().BoolVar(&stashNoCopy, "no-copy", false, "do not copy to clipboard")
	stashCmd.Flags().BoolVarP(&stashExec, "exec", "e", false, "output for shell execution")
}
