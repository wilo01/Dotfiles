package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "hlp"
)

// CredentialManager handles secure credential storage
type CredentialManager struct {
	useKeyring bool
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager() *CredentialManager {
	cm := &CredentialManager{useKeyring: true}

	// Test if keyring is available
	if err := keyring.Set(serviceName, "_test", "test"); err != nil {
		cm.useKeyring = false
	} else {
		keyring.Delete(serviceName, "_test")
	}

	return cm
}

// SetJiraToken stores the JIRA API token
func (cm *CredentialManager) SetJiraToken(token string) error {
	return cm.set("jira_token", token)
}

// GetJiraToken retrieves the JIRA API token (legacy - use GetJiraTokenForProfile)
func (cm *CredentialManager) GetJiraToken() (string, error) {
	// First check environment variable
	if token := os.Getenv("JIRA_API_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("HLP_JIRA_TOKEN"); token != "" {
		return token, nil
	}
	return cm.get("jira_token")
}

// GetJiraTokenForProfile retrieves the JIRA API token for a specific profile
// Priority: 1) env vars, 2) MCP .env file, 3) keyring fallback
func (cm *CredentialManager) GetJiraTokenForProfile(profileName string) (string, error) {
	// 1. Global env var override (for scripting)
	if token := os.Getenv("JIRA_API_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("HLP_JIRA_TOKEN"); token != "" {
		return token, nil
	}

	// 2. MCP profile .env file
	if token, err := cm.GetJiraTokenFromMCPProfile(profileName); err == nil && token != "" {
		return token, nil
	}

	// 3. Fallback to keyring/file (for local profile or missing .env)
	return cm.get("jira_token")
}

// GetJiraTokenFromMCPProfile reads token from MCP .env file
func (cm *CredentialManager) GetJiraTokenFromMCPProfile(profileName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	envPath := filepath.Join(homeDir, ".claude", "mcp-servers", "jira-profiles", fmt.Sprintf("jira-%s.env", profileName))

	file, err := os.Open(envPath)
	if err != nil {
		return "", fmt.Errorf("failed to open MCP profile .env: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check for JIRA token variables (Cloud uses JIRA_API_TOKEN, Server/DC uses JIRA_PERSONAL_TOKEN)
		if key == "JIRA_API_TOKEN" || key == "JIRA_PERSONAL_TOKEN" {
			return value, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading MCP profile .env: %w", err)
	}

	return "", errors.New("no JIRA token found in MCP profile .env")
}

// SetTimesheetSession stores timesheet session tokens
func (cm *CredentialManager) SetTimesheetSession(session TimesheetSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return cm.set("timesheet_session", string(data))
}

// GetTimesheetSession retrieves timesheet session tokens
func (cm *CredentialManager) GetTimesheetSession() (*TimesheetSession, error) {
	data, err := cm.get("timesheet_session")
	if err != nil {
		return nil, err
	}

	var session TimesheetSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// TimesheetSession holds session tokens for Timesheet plugin
type TimesheetSession struct {
	TenantSessionToken string `json:"tenant_session_token"`
	AtlassianXSRFToken string `json:"atlassian_xsrf_token"`
	JSESSIONID         string `json:"jsessionid"`
}

// set stores a value in the keyring or fallback file
func (cm *CredentialManager) set(key, value string) error {
	if cm.useKeyring {
		return keyring.Set(serviceName, key, value)
	}
	return cm.setFile(key, value)
}

// get retrieves a value from the keyring or fallback file
func (cm *CredentialManager) get(key string) (string, error) {
	if cm.useKeyring {
		val, err := keyring.Get(serviceName, key)
		if err == nil {
			return val, nil
		}
		// Fall through to file if keyring fails
	}
	return cm.getFile(key)
}

// delete removes a value from the keyring or fallback file
func (cm *CredentialManager) Delete(key string) error {
	if cm.useKeyring {
		return keyring.Delete(serviceName, key)
	}
	return cm.deleteFile(key)
}

// File-based fallback for systems without keyring
type credentialStore struct {
	Credentials map[string]string `json:"credentials"`
}

func (cm *CredentialManager) credentialFilePath() string {
	return GetConfigPath("credentials.json")
}

func (cm *CredentialManager) loadCredentialStore() (*credentialStore, error) {
	path := cm.credentialFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &credentialStore{Credentials: make(map[string]string)}, nil
		}
		return nil, err
	}

	var store credentialStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Credentials == nil {
		store.Credentials = make(map[string]string)
	}
	return &store, nil
}

func (cm *CredentialManager) saveCredentialStore(store *credentialStore) error {
	path := cm.credentialFilePath()
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (cm *CredentialManager) setFile(key, value string) error {
	store, err := cm.loadCredentialStore()
	if err != nil {
		return err
	}
	store.Credentials[key] = value
	return cm.saveCredentialStore(store)
}

func (cm *CredentialManager) getFile(key string) (string, error) {
	store, err := cm.loadCredentialStore()
	if err != nil {
		return "", err
	}
	val, ok := store.Credentials[key]
	if !ok {
		return "", errors.New("credential not found")
	}
	return val, nil
}

func (cm *CredentialManager) deleteFile(key string) error {
	store, err := cm.loadCredentialStore()
	if err != nil {
		return err
	}
	delete(store.Credentials, key)
	return cm.saveCredentialStore(store)
}
