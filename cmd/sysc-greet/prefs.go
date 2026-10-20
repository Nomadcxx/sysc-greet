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

func (m model) currentPreferences() cache.UserPreferences {
	p := cache.UserPreferences{
		FollowShell:  &m.followShell,
		Theme:        m.currentTheme,
		Background:   m.selectedBackground,
		Wallpaper:    m.selectedWallpaper,
		BorderStyle:  m.selectedBorderStyle,
		BorderEffect: m.borderEffect,
		ASCIIIndex:   m.asciiArtIndex,
		AnimSpeed:    m.animSpeed,
		FormMode:     m.formMode,
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
	cache.SavePreferences(m.currentPreferences())
}
