package main

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func TestMainMenuOrderStable(t *testing.T) {
	want := []string{"Close Menu", "Themes", "Borders", "Backgrounds", "ASCII Effects", "Wallpaper"}
	for _, mode := range []ViewMode{ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeASCIIEffectsSubmenu, ModeWallpaperSubmenu} {
		for _, route := range []struct {
			name string
			key  rune
		}{{"F1", tea.KeyF1}, {"Esc", tea.KeyEscape}, {"Back", tea.KeyEnter}} {
			t.Run(string(mode)+"/"+route.name, func(t *testing.T) {
				m := layoutModel(120, 40)
				m.config.TestMode = true
				m.mode = mode
				m.menuOptions = []string{"first", "second", "third", "← Back"}
				m.menuIndex = 3
				m.usernameInput.SetValue("example")
				m.selectedBackground = "matrix"
				m.selectedWallpaper = "example.mp4"
				m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: route.key})
				if m.mode != ModeMenu || m.menuIndex != 0 || !reflect.DeepEqual(m.menuOptions, want) {
					t.Fatalf("mode=%s index=%d options=%v", m.mode, m.menuIndex, m.menuOptions)
				}
				if m.usernameInput.Value() != "example" || m.selectedBackground != "matrix" || m.selectedWallpaper != "example.mp4" {
					t.Fatal("menu navigation changed input or wallpaper selection")
				}
				// Entering these submenus must not select an effect or wallpaper.
				m.menuIndex = 4
				ascii, _ := m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
				if ascii.mode != ModeASCIIEffectsSubmenu {
					t.Fatalf("fifth item opened %s", ascii.mode)
				}
				m.menuIndex = 5
				wallpaper, _ := m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
				if wallpaper.mode != ModeWallpaperSubmenu || wallpaper.selectedWallpaper != "example.mp4" {
					t.Fatal("sixth item did not open wallpaper submenu without changing selection")
				}
			})
		}
	}
}
