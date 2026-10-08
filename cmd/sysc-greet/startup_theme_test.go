package main

import (
	"path/filepath"
	"testing"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/themes"
	"github.com/charmbracelet/lipgloss/v2"
)

func TestInitialModelUsesCLITheme(t *testing.T) {
	originalDir, originalThemes := dataDir, themes.CustomThemes
	dataDir = filepath.Join("..", "..")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SYSC_BG", "")
	t.Cleanup(func() {
		dataDir, themes.CustomThemes = originalDir, originalThemes
		applyTheme("dracula", true)
	})
	if err := cache.SavePreferences(cache.UserPreferences{Theme: "Nord"}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ flag, name, primary string }{
		{"blue", "Blue", "#42a5f5"},
		{" BLUE ", "Blue", "#42a5f5"},
		{"ayu", "Ayu", "#e6b450"},
		{"amber", "Amber", "#ffc107"},
		{"purple", "Purple", "#d0bcff"},
		{"green", "Green", "#4caf50"},
		{"orange", "Orange", "#ff6d00"},
		{"rose-pine", "Rose-Pine", "#ebbcba"},
		{"kanagawa", "Kanagawa", "#76946a"},
		{"noctalia", "Noctalia", "#fff59b"},
		{"eldritch-abyss", "Eldritch-Abyss", "#2dcc82"},
		{"void", "Void", "#ffffff"},
		{"red", "Red", "#f44336"},
		{"cyan", "Cyan", "#00bcd4"},
		{"coral", "Coral", "#ffb4ab"},
		{"pink", "Pink", "#e91e63"},
		{"nord", "Nord", "#81a1c1"},
		{"tokyo-night", "Tokyo Night", "#7aa2f7"},
		{"tokyonight", "Tokyo Night", "#7aa2f7"},
		{"catppuccin-mocha", "Catppuccin", "#cba6f7"},
		{"", "dracula", "#bd93f9"},
		{"does-not-exist", "dracula", "#bd93f9"},
	} {
		t.Run(tc.flag, func(t *testing.T) {
			themes.CustomThemes = make(map[string]themes.ThemeColors)
			m := initialModel(Config{TestMode: true, ThemeName: tc.flag}, false)
			if m.currentTheme != tc.name {
				t.Errorf("active theme=%q, want %q", m.currentTheme, tc.name)
			}
			if Primary != lipgloss.Color(tc.primary) {
				t.Errorf("primary=%v, want %s", Primary, tc.primary)
			}
		})
	}
}

func TestStartupThemePrecedence(t *testing.T) {
	available := []string{"Dracula", "Nord", "Blue", "User Theme", "tokyo-night"}
	for _, tc := range []struct{ requested, cached, want string }{
		{"blue", "Nord", "Blue"},
		{"", "Nord", "Nord"},
		{"", "", "dracula"},
		{"unknown", "Nord", "dracula"},
		{"user theme", "Nord", "User Theme"},
		{"tokyo-night", "Nord", "tokyo-night"},
	} {
		if got := startupThemeName(tc.requested, tc.cached, available); got != tc.want {
			t.Errorf("requested=%q cached=%q: got %q, want %q", tc.requested, tc.cached, got, tc.want)
		}
	}
}
