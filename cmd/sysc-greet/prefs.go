package main

import "github.com/Nomadcxx/sysc-greet/internal/cache"

func migrateBorderPrefs(style, effect string) (string, string) {
	switch style {
	case "ascii3", "ascii4":
		style = "ascii2"
	case "wave":
		style = "classic"
		if effect == "" {
			effect = "shimmer"
		}
	case "pulse":
		style = "modern"
		if effect == "" {
			effect = "shimmer"
		}
	}
	if effect == "" {
		effect = "off"
	}
	return style, effect
}

func preferredFormMode(flagMode, savedMode string, flagSet bool) string {
	if flagSet {
		return flagMode
	}
	if validFormMode(savedMode) {
		return savedMode
	}
	return flagMode
}

func (m model) currentPreferences() cache.UserPreferences {
	p := cache.UserPreferences{
		FollowShell:       &m.followShell,
		Theme:             m.currentTheme,
		Background:        m.selectedBackground,
		Wallpaper:         m.selectedWallpaper,
		BorderStyle:       m.selectedBorderStyle,
		BorderEffect:      m.borderEffect,
		BorderPaletteMode: m.borderPaletteMode,
		ASCIIIndex:        m.asciiArtIndex,
		AnimSpeed:         m.animSpeed,
		FormMode:          m.formMode,
	}
	if m.selectedSession != nil {
		p.Session = m.selectedSession.Name
	}
	if m.config.RememberUsername {
		p.Username = m.usernameInput.Value()
	}
	return p
}

func (m model) savePreferences() {
	if m.config.TestMode {
		return
	}
	m.persistPreferences(m.currentPreferences())
}

func (m model) saveSuccessfulLogin(username, session string) {
	if m.config.TestMode {
		return
	}
	p := m.currentPreferences()
	p.Session = session
	p.Username = ""
	if m.config.RememberUsername {
		p.Username = username
	}
	m.persistPreferences(p)
}

func (m model) persistPreferences(p cache.UserPreferences) {
	if m.config.TestMode {
		return
	}
	if err := cache.SavePreferences(p); err != nil && m.config.Debug {
		logDebug("Failed to save preferences: %v", err)
	}
}
