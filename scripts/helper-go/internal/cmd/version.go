package cmd

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersionInfo sets the version information from build ldflags
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)

		fmt.Printf("%s %s\n", labelStyle.Render("Version:"), valueStyle.Render(version))
		fmt.Printf("%s %s\n", labelStyle.Render("Commit:"), valueStyle.Render(commit))
		fmt.Printf("%s %s\n", labelStyle.Render("Built:"), valueStyle.Render(date))
		fmt.Printf("%s %s\n", labelStyle.Render("Go:"), valueStyle.Render(runtime.Version()))
		fmt.Printf("%s %s/%s\n", labelStyle.Render("Platform:"), valueStyle.Render(runtime.GOOS), valueStyle.Render(runtime.GOARCH))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
