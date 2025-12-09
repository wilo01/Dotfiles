package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// JiraProfile represents a JIRA connection profile
type JiraProfile struct {
	Name      string   `yaml:"name,omitempty"`
	BaseURL   string   `yaml:"base_url"`
	Email     string   `yaml:"email"`
	Protected bool     `yaml:"protected"` // Requires confirmation for writes
	Projects  []string `yaml:"projects,omitempty"`
}

// IsLocal returns true if this profile connects to a local JIRA instance
// Used to enable LOCAL-only features like auto-creating tickets and separate CSV
func (p *JiraProfile) IsLocal() bool {
	if p == nil {
		return false
	}
	return strings.HasPrefix(p.BaseURL, "http://localhost:")
}

// JiraProfiles holds all profiles and the active one
type JiraProfiles struct {
	ActiveProfile string                 `yaml:"active_profile"`
	Profiles      map[string]JiraProfile `yaml:"profiles"`
}

// DefaultProfiles returns the default profile configuration
func DefaultProfiles() *JiraProfiles {
	return &JiraProfiles{
		ActiveProfile: "local",
		Profiles: map[string]JiraProfile{
			"local": {
				Name:      "local",
				BaseURL:   "http://localhost:8080",
				Email:     "",
				Protected: false,
			},
		},
	}
}

// GetActiveProfile returns the currently active profile
func GetActiveProfile() (*JiraProfile, error) {
	profiles, err := LoadProfiles()
	if err != nil {
		return nil, err
	}

	profileName := profiles.ActiveProfile
	if profileName == "" {
		profileName = "local"
	}

	profile, ok := profiles.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found", profileName)
	}

	profile.Name = profileName
	return &profile, nil
}

// SetActiveProfile changes the active profile
func SetActiveProfile(name string) error {
	profiles, err := LoadProfiles()
	if err != nil {
		return err
	}

	if _, ok := profiles.Profiles[name]; !ok {
		return fmt.Errorf("profile '%s' not found. Available: %v", name, getProfileNames(profiles))
	}

	profiles.ActiveProfile = name

	// Update viper and save
	viper.Set("jira.active_profile", name)

	// Also update the legacy fields for backwards compatibility
	profile := profiles.Profiles[name]
	viper.Set("jira.base_url", profile.BaseURL)
	viper.Set("jira.email", profile.Email)

	return viper.WriteConfig()
}

// LoadProfiles loads profiles from config
func LoadProfiles() (*JiraProfiles, error) {
	profiles := DefaultProfiles()

	// Try to load from viper
	if viper.IsSet("jira.profiles") {
		profilesMap := viper.GetStringMap("jira.profiles")
		profiles.Profiles = make(map[string]JiraProfile)

		for name, v := range profilesMap {
			if pm, ok := v.(map[string]interface{}); ok {
				profile := JiraProfile{Name: name}
				if url, ok := pm["base_url"].(string); ok {
					profile.BaseURL = url
				}
				if email, ok := pm["email"].(string); ok {
					profile.Email = email
				}
				if protected, ok := pm["protected"].(bool); ok {
					profile.Protected = protected
				}
				if projects, ok := pm["projects"].([]interface{}); ok {
					for _, p := range projects {
						if ps, ok := p.(string); ok {
							profile.Projects = append(profile.Projects, ps)
						}
					}
				}
				profiles.Profiles[name] = profile
			}
		}
	} else {
		// Backwards compatibility: use legacy config
		baseURL := viper.GetString("jira.base_url")
		email := viper.GetString("jira.email")
		if baseURL != "" {
			profiles.Profiles["local"] = JiraProfile{
				Name:      "local",
				BaseURL:   baseURL,
				Email:     email,
				Protected: false,
			}
		}
	}

	if viper.IsSet("jira.active_profile") {
		profiles.ActiveProfile = viper.GetString("jira.active_profile")
	}

	return profiles, nil
}

// SaveProfile saves or updates a profile
func SaveProfile(name string, profile JiraProfile) error {
	profiles, err := LoadProfiles()
	if err != nil {
		return err
	}

	profile.Name = name
	profiles.Profiles[name] = profile

	// Convert to viper-compatible format
	profilesMap := make(map[string]interface{})
	for n, p := range profiles.Profiles {
		profilesMap[n] = map[string]interface{}{
			"base_url":  p.BaseURL,
			"email":     p.Email,
			"protected": p.Protected,
			"projects":  p.Projects,
		}
	}

	viper.Set("jira.profiles", profilesMap)
	viper.Set("jira.active_profile", profiles.ActiveProfile)

	return viper.WriteConfig()
}

// ListProfiles returns all available profile names
func ListProfiles() ([]string, error) {
	profiles, err := LoadProfiles()
	if err != nil {
		return nil, err
	}
	return getProfileNames(profiles), nil
}

// GetProfilesFilePath returns the path to profiles config
func GetProfilesFilePath() string {
	return filepath.Join(GetConfigDir(), "profiles.yaml")
}

// InitializeProfiles creates default profiles if they don't exist
func InitializeProfiles() error {
	configPath := GetConfigPath("config.yaml")

	// Check if profiles already exist in config
	if viper.IsSet("jira.profiles") {
		return nil
	}

	// Create default profiles
	profiles := DefaultProfiles()

	// Migrate existing config
	baseURL := viper.GetString("jira.base_url")
	email := viper.GetString("jira.email")

	if baseURL != "" {
		// Determine if existing config is localhost or production
		isLocal := baseURL == "http://localhost:8080" || baseURL == "http://localhost:2990/jira"

		if isLocal {
			profiles.Profiles["local"] = JiraProfile{
				Name:      "local",
				BaseURL:   baseURL,
				Email:     email,
				Protected: false,
			}
		} else {
			// Existing config is production, keep as-is and add local
			profiles.Profiles["production"] = JiraProfile{
				Name:      "production",
				BaseURL:   baseURL,
				Email:     email,
				Protected: true,
			}
			profiles.ActiveProfile = "production"
		}
	}

	// Save profiles
	profilesMap := make(map[string]interface{})
	for n, p := range profiles.Profiles {
		profilesMap[n] = map[string]interface{}{
			"base_url":  p.BaseURL,
			"email":     p.Email,
			"protected": p.Protected,
			"projects":  p.Projects,
		}
	}

	viper.Set("jira.profiles", profilesMap)
	viper.Set("jira.active_profile", profiles.ActiveProfile)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	return viper.WriteConfig()
}

// ExportProfilesYAML exports profiles as YAML for display
func ExportProfilesYAML() (string, error) {
	profiles, err := LoadProfiles()
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(profiles)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getProfileNames(profiles *JiraProfiles) []string {
	names := make([]string, 0, len(profiles.Profiles))
	for name := range profiles.Profiles {
		names = append(names, name)
	}
	return names
}
