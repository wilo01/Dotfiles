package cmd

import (
	"fmt"
	"os"

	"github.com/dariuszw/hlp/internal/cmd/badge"
	"github.com/dariuszw/hlp/internal/cmd/dev"
	"github.com/dariuszw/hlp/internal/cmd/git"
	"github.com/dariuszw/hlp/internal/cmd/jira"
	"github.com/dariuszw/hlp/internal/cmd/namefinder"
	"github.com/dariuszw/hlp/internal/cmd/standup"
	"github.com/dariuszw/hlp/internal/cmd/timesheet"
	"github.com/dariuszw/hlp/internal/config"
	"github.com/dariuszw/hlp/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "hlp",
	Short: "Developer utility CLI",
	Long: ui.Logo() + `A comprehensive developer utility CLI for JIRA workflow automation,
							time tracking, badge generation, and name availability checking.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/hlp/config.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")

	// Bind flags to viper
	viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Register commands at top level
	dev.RegisterTopLevel(rootCmd)
	badge.RegisterTopLevel(rootCmd)
	rootCmd.AddCommand(git.GitCmd)
	rootCmd.AddCommand(jira.JiraCmd)
	rootCmd.AddCommand(standup.StandupCmd)
	rootCmd.AddCommand(timesheet.TimesheetCmd)
	rootCmd.AddCommand(namefinder.NameFinderCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configDir := config.GetConfigDir()
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintln(os.Stderr, ui.Error("Failed to create config directory: "+err.Error()))
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Environment variable support
	viper.SetEnvPrefix("HLP")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, ui.Warning("Config file error: "+err.Error()))
		}
	}

	// Note: NO_COLOR handling is done via lipgloss auto-detection

	// Load config struct
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Warning("Failed to load config: "+err.Error()))
		cfg = config.Default()
	}
}

// GetConfig returns the current configuration
func GetConfig() *config.Config {
	if cfg == nil {
		cfg = config.Default()
	}
	return cfg
}
