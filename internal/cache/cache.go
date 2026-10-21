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
	FollowShell       *bool  `json:"follow_shell,omitempty"`
	Theme             string `json:"theme"`                         // Last selected theme
	Background        string `json:"background"`                    // Last selected background animation (aquarium, matrix, fire, etc.)
	Wallpaper         string `json:"wallpaper"`                     // Last selected gslapper video wallpaper (separate from background effect)
	BorderStyle       string `json:"border_style"`                  // Last selected border style
	BorderEffect      string `json:"border_effect,omitempty"`       // Last selected border effect (off, shimmer); empty = off
	BorderPaletteMode string `json:"border_palette_mode,omitempty"` // Border palette mode; empty = theme
	FormMode          string `json:"form_mode,omitempty"`           // Last selected form mode (auto, opaque, seethrough, frosted); empty = auto
	Session           string `json:"session"`                       // Last selected session
	Username          string `json:"username"`                      // Last successful username for this session
	ASCIIIndex        int    `json:"ascii_index"`                   // Last selected ASCII variant index
	AnimSpeed         string `json:"anim_speed"`                    // Last selected animation speed (slow, normal, fast)
}

// SavePreferences saves user preferences to cache
func SavePreferences(prefs UserPreferences) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	cachePath := filepath.Join(home, cacheDir)
	if err := os.MkdirAll(cachePath, 0o700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	if info, err := os.Lstat(cachePath); err != nil {
		return fmt.Errorf("failed to inspect cache directory: %w", err)
	} else if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("cache path %q is not a directory", cachePath)
	}
	if err := os.Chmod(cachePath, 0o700); err != nil {
		return fmt.Errorf("failed to secure cache directory: %w", err)
	}

	filePath := filepath.Join(cachePath, preferencesFile)
	data, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	file, err := os.CreateTemp(cachePath, ".preferences-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary preferences file: %w", err)
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("failed to secure temporary preferences file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("failed to write temporary preferences file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("failed to sync temporary preferences file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary preferences file: %w", err)
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("failed to replace preferences file: %w", err)
	}
	dir, err := os.Open(cachePath)
	if err != nil {
		return fmt.Errorf("failed to open cache directory after saving preferences: %w", err)
	}
	if err := dir.Sync(); err != nil {
		dir.Close()
		return fmt.Errorf("failed to sync cache directory after saving preferences: %w", err)
	}
	if err := dir.Close(); err != nil {
		return fmt.Errorf("failed to close cache directory after saving preferences: %w", err)
	}

	return nil
}

// LoadPreferences loads user preferences from cache
func LoadPreferences() (*UserPreferences, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	cachePath := filepath.Join(home, cacheDir)
	filePath := filepath.Join(cachePath, preferencesFile)
	fileInfo, err := os.Lstat(filePath)
	if os.IsNotExist(err) {
		return nil, nil // No cached preferences
	}
	if err != nil {
		return nil, fmt.Errorf("failed to inspect preferences file: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("preferences path %q is not a regular file", filePath)
	}
	dirInfo, err := os.Lstat(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect cache directory: %w", err)
	}
	if !dirInfo.IsDir() || dirInfo.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("cache path %q is not a directory", cachePath)
	}
	if err := os.Chmod(cachePath, 0o700); err != nil {
		return nil, fmt.Errorf("failed to secure cache directory: %w", err)
	}
	if err := os.Chmod(filePath, 0o600); err != nil {
		return nil, fmt.Errorf("failed to secure preferences file: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preferences file: %w", err)
	}

	var prefs UserPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	return &prefs, nil
}
