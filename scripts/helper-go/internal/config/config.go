package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Jira        JiraConfig      `mapstructure:"jira" yaml:"jira"`
	Timesheet   TimesheetConfig `mapstructure:"timesheet" yaml:"timesheet"`
	Sheets      SheetsConfig    `mapstructure:"google_sheets" yaml:"google_sheets"`
	Preferences Preferences     `mapstructure:"preferences" yaml:"preferences"`
	Dev         DevConfig       `mapstructure:"dev" yaml:"dev"`
}

// DevConfig holds developer workflow settings
type DevConfig struct {
	CommitsFile string `mapstructure:"commits_file" yaml:"commits_file"`
}

// JiraConfig holds JIRA connection settings
type JiraConfig struct {
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Email   string `mapstructure:"email" yaml:"email"`
	// APIToken is stored in keyring, not config file
}

// TimesheetConfig holds Timesheet plugin settings
type TimesheetConfig struct {
	JiraURL string `mapstructure:"jira_url" yaml:"jira_url"`
	// Session tokens stored securely
}

// SheetsConfig holds Google Sheets settings
type SheetsConfig struct {
	SheetID     string `mapstructure:"sheet_id" yaml:"sheet_id"`
	DailyTabURL string `mapstructure:"daily_tab_url" yaml:"daily_tab_url"`
}

// ScheduleSlot defines a named time slot for scheduled batch runs
type ScheduleSlot struct {
	Name string `mapstructure:"name" yaml:"name"` // "morning", "afternoon"
	Time string `mapstructure:"time" yaml:"time"` // "08:45"
}

// ScheduleConfig holds scheduling configuration for batch operations
type ScheduleConfig struct {
	Slots          []ScheduleSlot `mapstructure:"slots" yaml:"slots"`                       // Named time slots
	Loop           bool           `mapstructure:"loop" yaml:"loop"`                         // Continue after each run
	Timezone       string         `mapstructure:"timezone" yaml:"timezone"`                 // "Europe/Dublin"
	WorkHoursStart string         `mapstructure:"work_hours_start" yaml:"work_hours_start"` // "08:00"
	WorkHoursEnd   string         `mapstructure:"work_hours_end" yaml:"work_hours_end"`     // "16:00"
}

// Preferences holds user preferences
type Preferences struct {
	AutoDetectContext     bool           `mapstructure:"auto_detect_context" yaml:"auto_detect_context"`
	DefaultDuration       string         `mapstructure:"default_duration" yaml:"default_duration"`
	WorkHoursStart        string         `mapstructure:"work_hours_start" yaml:"work_hours_start"`
	WorkHoursEnd          string         `mapstructure:"work_hours_end" yaml:"work_hours_end"`
	ExpectedHoursPerDay   string         `mapstructure:"expected_hours_per_day" yaml:"expected_hours_per_day"`
	RichOutput            bool           `mapstructure:"rich_output" yaml:"rich_output"`
	CopyToClipboard       bool           `mapstructure:"copy_to_clipboard" yaml:"copy_to_clipboard"`
	StandupIgnoredTickets []string       `mapstructure:"standup_ignored_tickets" yaml:"standup_ignored_tickets"`
	StandupWebhookURL     string         `mapstructure:"standup_webhook_url" yaml:"standup_webhook_url"`
	DailyTickets          []string       `mapstructure:"daily_tickets" yaml:"daily_tickets"`
	LogToSubtask          bool           `mapstructure:"log_to_subtask" yaml:"log_to_subtask"`
	Schedule              ScheduleConfig `mapstructure:"schedule" yaml:"schedule"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	if dir := os.Getenv("HLP_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/hlp"
	}
	return filepath.Join(home, ".config", "hlp")
}

// GetConfigPath returns the full path to a config file
func GetConfigPath(filename string) string {
	return filepath.Join(GetConfigDir(), filename)
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		Jira: JiraConfig{
			BaseURL: "",
			Email:   "",
		},
		Timesheet: TimesheetConfig{
			JiraURL: "",
		},
		Sheets: SheetsConfig{
			SheetID:     "",
			DailyTabURL: "",
		},
		Dev: DevConfig{
			CommitsFile: "~/Dev/Private/Commits.md",
		},
		Preferences: Preferences{
			AutoDetectContext:     true,
			DefaultDuration:       "1h",
			WorkHoursStart:        "09:00",
			WorkHoursEnd:          "17:00",
			ExpectedHoursPerDay:   "8h",
			RichOutput:            true,
			CopyToClipboard:       true,
			StandupIgnoredTickets: []string{"TDT-2", "TDT-26"},
			StandupWebhookURL:     "",
			LogToSubtask:          false,
			Schedule: ScheduleConfig{
				Slots: []ScheduleSlot{
					{Name: "morning", Time: "08:45"},
					{Name: "afternoon", Time: "15:45"},
				},
				Loop:           true,
				Timezone:       "Europe/Dublin",
				WorkHoursStart: "08:00",
				WorkHoursEnd:   "16:00",
			},
		},
	}
}

// Load reads the configuration from viper
func Load() (*Config, error) {
	cfg := Default()
	if err := viper.Unmarshal(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes the configuration to the config file
func Save(cfg *Config) error {
	configPath := GetConfigPath("config.yaml")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// Set all values in viper
	viper.Set("jira", cfg.Jira)
	viper.Set("timesheet", cfg.Timesheet)
	viper.Set("google_sheets", cfg.Sheets)
	viper.Set("preferences", cfg.Preferences)
	viper.Set("dev", cfg.Dev)

	return viper.WriteConfigAs(configPath)
}
