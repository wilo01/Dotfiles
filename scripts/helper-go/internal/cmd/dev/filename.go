package dev

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/format"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

var (
	filenameCopy   bool
	filenameNoCopy bool
	filenameNoExt  bool
)

var filenameCmd = &cobra.Command{
	Use:   "filename [text]",
	Short: "Generate clean markdown filename",
	Long: `Generate a clean filename from text, suitable for markdown files.

Examples:
  hlp filename "VIS-1234 My Report"
  Output: vis-1234-my-report.md

  hlp filename "Meeting Notes" --no-ext
  Output: meeting-notes`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]

		var result string
		if filenameNoExt {
			result = format.Dash(text)
		} else {
			result = format.Filename(text)
		}

		if !filenameNoCopy && filenameCopy {
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
	filenameCmd.Flags().BoolVarP(&filenameCopy, "copy", "c", true, "copy to clipboard")
	filenameCmd.Flags().BoolVar(&filenameNoCopy, "no-copy", false, "do not copy to clipboard")
	filenameCmd.Flags().BoolVar(&filenameNoExt, "no-ext", false, "omit .md extension")
}
