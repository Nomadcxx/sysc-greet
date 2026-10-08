package main

import (
	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/themes"
	tea "github.com/charmbracelet/bubbletea/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellThemeSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	available := []string{"Dracula", "Blue", "Tokyo Night", "Catppuccin"}
	for body, want := range map[string]string{"dracula\n": "Dracula", "blue": "Blue", "tokyo-night": "Tokyo Night", "catppuccin": "Catppuccin", "unknown": "", "../dracula": "", "": ""} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if got := shellThemeName(path, available); got != want {
			t.Fatalf("%q => %q want %q", body, got, want)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/dev/zero", path); err != nil {
		t.Fatal(err)
	}
	if shellThemeName(path, available) != "" {
		t.Fatal("followed symlink")
	}
}

func TestFollowShellPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	if err := os.WriteFile(path, []byte("blue\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m := model{config: Config{ShellThemeFile: path}, followShell: true, availableThemes: []string{"Dracula", "Nord", "Blue"}}
	if got := m.startupTheme("Nord"); got != "Blue" {
		t.Fatal(got)
	}
	m.config.ThemeName = "dracula"
	if got := m.startupTheme("Nord"); got != "Dracula" {
		t.Fatal(got)
	}
	m.config.ThemeName = ""
	m.followShell = false
	if got := m.startupTheme("Nord"); got != "Nord" {
		t.Fatal(got)
	}
	m.followShell = true
	if err := os.WriteFile(path, []byte("unsupported"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := m.startupTheme("Nord"); got != "Nord" {
		t.Fatal(got)
	}
	t.Setenv("HOME", t.TempDir())
	off := false
	if err := cache.SavePreferences(cache.UserPreferences{Theme: "Nord", FollowShell: &off}); err != nil {
		t.Fatal(err)
	}
	prefs, err := cache.LoadPreferences()
	if err != nil || prefs.FollowShell == nil || *prefs.FollowShell {
		t.Fatal("opt-out not persisted")
	}
}

func TestStartupWallpaperFollowsShell(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	if err := os.WriteFile(path, []byte("dracula"), 0600); err != nil {
		t.Fatal(err)
	}
	prefs := &cache.UserPreferences{Theme: "Nord", Wallpaper: "custom.mp4"}
	got := startupThemePreferences(prefs, Config{ShellThemeFile: path}, []string{"Dracula", "Nord"})
	if got.Theme != "Dracula" || got.Wallpaper != prefs.Wallpaper || prefs.Theme != "Nord" {
		t.Fatal("theme resolution lost custom wallpaper or mutated cache")
	}
}

func TestFollowShellMenu(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme")
	if err := os.WriteFile(path, []byte("blue"), 0600); err != nil {
		t.Fatal(err)
	}
	m := layoutModel(120, 40)
	m.config.TestMode = true
	m.config.ShellThemeFile = path
	m.availableThemes = []string{"Dracula", "Blue"}
	next, _ := m.navigateToThemesSubmenu()
	m = next.(model)
	m.menuIndex = 1
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.followShell || m.currentTheme != "Blue" {
		t.Fatalf("following=%v theme=%s", m.followShell, m.currentTheme)
	}
	next, _ = m.navigateToThemesSubmenu()
	m = next.(model)
	m.menuIndex = 2
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.followShell || m.currentTheme != "Dracula" {
		t.Fatal("explicit theme did not disable following")
	}
}

func TestFollowShellBundledCatalog(t *testing.T) {
	original := themes.CustomThemes
	t.Cleanup(func() { themes.CustomThemes = original })
	themes.CustomThemes = make(map[string]themes.ThemeColors)
	available := themes.ScanCustomThemes([]string{filepath.Join("..", "..", "themes")})
	path := filepath.Join(t.TempDir(), "theme")
	for _, name := range []string{"rose-pine", "kanagawa", "noctalia", "eldritch-abyss", "void", "red", "cyan", "coral", "pink"} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(name+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			m := model{config: Config{ShellThemeFile: path}, followShell: true, availableThemes: available}
			if got := m.startupTheme("Dracula"); !strings.EqualFold(got, name) {
				t.Fatalf("theme=%q, want %q", got, name)
			}
		})
	}
}
