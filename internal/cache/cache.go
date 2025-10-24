package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nomadcxx/sysc-greet/internal/sessions"
)

const cacheDir = ".cache/sysc-greet"
const sessionFile = "session"
const preferencesFile = "preferences"

// SaveSelectedSession saves the selected session to cache
func SaveSelectedSession(session sessions.Session) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	cachePath := filepath.Join(home, cacheDir)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	filePath := filepath.Join(cachePath, sessionFile)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %v", err)
	}

	return nil
}

// LoadSelectedSession loads the selected session from cache
func LoadSelectedSession() (*sessions.Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	filePath := filepath.Join(home, cacheDir, sessionFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil // No cached session
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %v", err)
	}

	var session sessions.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %v", err)
	}

	return &session, nil
}

// UserPreferences holds cached user preferences
type UserPreferences struct {
	Theme       string `json:"theme"`        // Last selected theme
	Background  string `json:"background"`   // Last selected background animation
	BorderStyle string `json:"border_style"` // Last selected border style
	Session     string `json:"session"`      // Last selected session
	Username    string `json:"username"`     // Last successful username for this session
	ASCIIIndex  int    `json:"ascii_index"`  // Last selected ASCII variant index
}

// SavePreferences saves user preferences to cache
func SavePreferences(prefs UserPreferences) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	cachePath := filepath.Join(home, cacheDir)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	filePath := filepath.Join(cachePath, preferencesFile)
	data, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences file: %v", err)
	}

	return nil
}

// LoadPreferences loads user preferences from cache
func LoadPreferences() (*UserPreferences, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	filePath := filepath.Join(home, cacheDir, preferencesFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil // No cached preferences
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preferences file: %v", err)
	}

	var prefs UserPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %v", err)
	}

	return &prefs, nil
}
