package dev

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	prCopy   bool
	prNoCopy bool
)

var prCmd = &cobra.Command{
	Use:   "pr [title]",
	Short: "Generate PR title with uppercased ticket numbers",
	Long: `Generate a pull request title with JIRA ticket numbers uppercased.

Examples:
  hlp pr "vis-1234 add new feature"
  Output: VIS-1234 add new feature

  hlp pr "fix bug in tdt-99 module"
  Output: fix bug in TDT-99 module`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		result := format.PRTitle(text)

		if !prNoCopy && prCopy {
			if err := clipboard.WriteAll(result); err != nil {
				fmt.Println(ui.Warning("Failed to copy to clipboard: " + err.Error()))
			} else {
				fmt.Println(ui.SuccessMsg("Copied to clipboard"))
			}
		}

		fmt.Println(result)
	},
}

func init() {
	prCmd.Flags().BoolVarP(&prCopy, "copy", "c", true, "copy to clipboard")
	prCmd.Flags().BoolVar(&prNoCopy, "no-copy", false, "do not copy to clipboard")
}
