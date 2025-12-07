package config

import (
	"encoding/json"
	"errors"
	"os"

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

// GetJiraToken retrieves the JIRA API token
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
