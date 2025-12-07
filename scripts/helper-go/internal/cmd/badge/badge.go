package badge

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/dariuszw/hlp/internal/input"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
)

const (
	defaultRFIDBadge = "@4324325553#"
	defaultQRBadge   = "$4324325553#"
)

var (
	badgeCopy  bool
	badgeType  bool
	badgeDelay int
)

// BadgeCmd is the badge command
var BadgeCmd = &cobra.Command{
	Use:   "badge [id]",
	Short: "Type or copy badge string",
	Long: `Type or copy a badge string with a prefix.

The badge format is: @<id>#

Examples:
  hlp badge 123456       # Uses default ID
  hlp badge 999999       # Uses custom ID
  hlp badge --type       # Type using keyboard automation
  hlp badge --copy       # Copy to clipboard only`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := "4324325553" // default
		if len(args) > 0 {
			id = args[0]
		}

		badge := fmt.Sprintf("@%s#", id)
		handleOutput(badge, badgeType, badgeCopy, badgeDelay)
	},
}

// RFIDCmd is the rfid command
var RFIDCmd = &cobra.Command{
	Use:   "rfid [id]",
	Short: "Type or copy RFID badge string",
	Long: `Type or copy an RFID badge string.

The badge format is: @<id>#

Examples:
  hlp rfid               # Uses default RFID
  hlp rfid 999999        # Uses custom ID`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var badge string
		if len(args) > 0 {
			badge = fmt.Sprintf("@%s#", args[0])
		} else {
			badge = defaultRFIDBadge
		}

		handleOutput(badge, badgeType, badgeCopy, badgeDelay)
	},
}

// QRCmd is the qr command
var QRCmd = &cobra.Command{
	Use:   "qr [id]",
	Short: "Type or copy QR badge string",
	Long: `Type or copy a QR badge string.

The badge format is: $<id>#

Examples:
  hlp qr                 # Uses default QR
  hlp qr 999999          # Uses custom ID`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var badge string
		if len(args) > 0 {
			badge = fmt.Sprintf("$%s#", args[0])
		} else {
			badge = defaultQRBadge
		}

		handleOutput(badge, badgeType, badgeCopy, badgeDelay)
	},
}

func init() {
	// Badge command flags
	BadgeCmd.Flags().BoolVarP(&badgeCopy, "copy", "c", false, "copy to clipboard only")
	BadgeCmd.Flags().BoolVarP(&badgeType, "type", "t", true, "type using keyboard automation")
	BadgeCmd.Flags().IntVarP(&badgeDelay, "delay", "d", 2, "delay in seconds before typing")

	// RFID command flags
	RFIDCmd.Flags().BoolVarP(&badgeCopy, "copy", "c", false, "copy to clipboard only")
	RFIDCmd.Flags().BoolVarP(&badgeType, "type", "t", true, "type using keyboard automation")
	RFIDCmd.Flags().IntVarP(&badgeDelay, "delay", "d", 2, "delay in seconds before typing")

	// QR command flags
	QRCmd.Flags().BoolVarP(&badgeCopy, "copy", "c", false, "copy to clipboard only")
	QRCmd.Flags().BoolVarP(&badgeType, "type", "t", true, "type using keyboard automation")
	QRCmd.Flags().IntVarP(&badgeDelay, "delay", "d", 2, "delay in seconds before typing")
}

func handleOutput(text string, doType bool, doCopy bool, delay int) {
	if doCopy && !doType {
		// Copy only
		if err := clipboard.WriteAll(text); err != nil {
			fmt.Println(ui.Error("Failed to copy: " + err.Error()))
			return
		}
		fmt.Println(ui.SuccessMsg("Copied to clipboard: " + text))
		return
	}

	if doType {
		typer := input.NewTyper()
		if typer.IsSupported() {
			fmt.Printf("Typing in %d seconds using %s...\n", delay, typer.Backend())
			if err := typer.TypeWithDelay(text, time.Duration(delay)*time.Second); err != nil {
				fmt.Println(ui.Warning("Typing failed, copied to clipboard instead"))
				clipboard.WriteAll(text)
			} else {
				fmt.Println(ui.SuccessMsg("Typed: " + text))
			}
		} else {
			// Fallback to clipboard
			clipboard.WriteAll(text)
			fmt.Println(ui.Warning("Keyboard automation not available, copied to clipboard"))
		}
		return
	}

	// Default: just print
	fmt.Println(text)
}

// RegisterTopLevel registers badge commands at the top level
func RegisterTopLevel(rootCmd *cobra.Command) {
	rootCmd.AddCommand(BadgeCmd)
	rootCmd.AddCommand(RFIDCmd)
	rootCmd.AddCommand(QRCmd)
}
