package qr

import (
	"fmt"
	"net/url"

	"github.com/dariuszw/hlp/internal/ui"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var (
	size    int
	urlOnly bool
)

var QrCmd = &cobra.Command{
	Use:   "qrcode <data>",
	Short: "Generate QR code for sharing data between devices",
	Long: `Generate and display a QR code in the terminal.
Useful for sharing passwords, URLs, or text between devices.

Examples:
  hlp qrcode "mypassword123"           # Display QR in terminal
  hlp qrcode --url "mypassword123"     # Show URL only (no ASCII QR)
  hlp qrcode -s 300 "secret"           # Custom size for URL`,
	Args: cobra.ExactArgs(1),
	Run:  runQr,
}

func init() {
	QrCmd.Flags().IntVarP(&size, "size", "s", 150, "QR code size in pixels (for URL)")
	QrCmd.Flags().BoolVarP(&urlOnly, "url", "u", false, "Show URL only (no terminal QR)")
}

func runQr(cmd *cobra.Command, args []string) {
	data := args[0]

	if !urlOnly {
		qr, err := qrcode.New(data, qrcode.Medium)
		if err != nil {
			fmt.Println(ui.Error("Failed to generate QR: " + err.Error()))
			return
		}
		fmt.Println(qr.ToSmallString(false))
	}

	qrURL := generateQRURL(data, size)
	fmt.Println(ui.Muted.Render("URL: " + qrURL))
}

func generateQRURL(data string, size int) string {
	return fmt.Sprintf("https://qrgen.tdscloud.ie/?data=%s&size=%d",
		url.QueryEscape(data), size)
}
