package main

import (
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	"github.com/Nomadcxx/sysc-greet/internal/themes"
)

func TestThemeChangeRefreshesTextEffectsInTestMode(t *testing.T) {
	originalDir, originalThemes := dataDir, themes.CustomThemes
	dataDir = filepath.Join("..", "..")
	themes.CustomThemes = make(map[string]themes.ThemeColors)
	themes.ScanCustomThemes([]string{filepath.Join(dataDir, "themes")})
	t.Cleanup(func() {
		dataDir, themes.CustomThemes = originalDir, originalThemes
		applyTheme("dracula", true)
	})
	for _, theme := range []string{"Nord", "Blue"} {
		for _, effect := range []string{"beams", "pour"} {
			t.Run(theme+"/"+effect, func(t *testing.T) {
				m := model{
					config: Config{TestMode: true}, mode: ModeThemesSubmenu,
					selectedSession: &sessions.Session{Name: "Niri", Type: "Wayland"},
					usernameInput:   textinput.New(), passwordInput: textinput.New(),
					menuOptions:        []string{"Theme: " + theme},
					selectedBackground: effect,
					beamsEffect:        animations.NewBeamsTextEffect(animations.BeamsTextConfig{Width: 1, Height: 1, Text: "X"}),
					pourEffect:         animations.NewPourEffect(animations.PourConfig{Width: 1, Height: 1, Text: "X", FinalGradientStops: []string{"#ffffff"}}),
				}
				oldBeams, oldPour := m.beamsEffect, m.pourEffect
				m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
				if m.currentTheme != theme || m.mode != ModeLogin {
					t.Fatal("theme selection did not return to login")
				}
				if effect == "beams" && m.beamsEffect == oldBeams {
					t.Fatal("beams retained the previous theme's effect")
				}
				if effect == "pour" && m.pourEffect == oldPour {
					t.Fatal("pour retained the previous theme's effect")
				}
			})
		}
	}
}
