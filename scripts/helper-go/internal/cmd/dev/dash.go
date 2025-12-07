package dev

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	dashCopy   bool
	dashNoCopy bool
)

var dashCmd = &cobra.Command{
	Use:   "dash [text]",
	Short: "Convert text to dash-separated format",
	Long: `Convert text to lowercase dash-separated format, suitable for filenames and URLs.

Examples:
  hlp dash "Hello World"
  Output: hello-world

  hlp dash "My Feature Name"
  Output: my-feature-name`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		result := format.Dash(text)

		if !dashNoCopy && dashCopy {
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
	dashCmd.Flags().BoolVarP(&dashCopy, "copy", "c", true, "copy to clipboard")
	dashCmd.Flags().BoolVar(&dashNoCopy, "no-copy", false, "do not copy to clipboard")
}
