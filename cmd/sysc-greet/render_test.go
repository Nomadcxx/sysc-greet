package main

import (
	"image"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func TestPowerViewLayerCoversFullTerminal(t *testing.T) {
	applyTheme("dracula", true)

	m := model{
		width:        80,
		height:       24,
		mode:         ModePower,
		powerOptions: []string{"Reboot", "Shutdown", "Cancel"},
		powerIndex:   0,
	}

	view := m.View()
	bounded, ok := view.Layer.(interface{ Bounds() image.Rectangle })
	if !ok {
		t.Fatalf("expected power view layer to expose bounds")
	}

	bounds := bounded.Bounds()
	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Dx() != m.width || bounds.Dy() != m.height {
		t.Fatalf("expected power view to cover %dx%d terminal from origin, got %v", m.width, m.height, bounds)
	}
}

func TestReleaseNotesDescribeGreeterSupport(t *testing.T) {
	applyTheme("dracula", true)
	notes := stripAnsi((model{}).renderReleaseNotesView(213, 54))
	for _, text := range []string{
		"Ayu, Amber, Blue, Purple, Green, Orange themes (same names as sysc-shell)",
		"Mango greeter compositor",
		"Supported greeter backends: niri (default), cagebreak, sway, mango",
		"Hyprland greeter support deprecated (login sessions unaffected)",
	} {
		if !strings.Contains(notes, text) {
			t.Errorf("release notes omit %q", text)
		}
	}
}

func TestViewTerminalModes(t *testing.T) {
	for _, altScreen := range []bool{false, true} {
		for _, testMode := range []bool{false, true} {
			for _, mode := range []ViewMode{ModeLogin, ModeMenu, ModePower, ModeReleaseNotes} {
				m := model{width: 80, height: 24, mode: mode, altScreen: altScreen, config: Config{TestMode: testMode}}
				view := m.View()
				if view.AltScreen != altScreen || !view.UniformKeyLayout {
					t.Fatalf("mode %v lost terminal or keyboard settings", mode)
				}
				wantMouse := tea.MouseModeNone
				if altScreen && !testMode {
					wantMouse = tea.MouseModeCellMotion
				}
				if view.MouseMode != wantMouse {
					t.Fatalf("mode %v: mouse mode %v, want %v", mode, view.MouseMode, wantMouse)
				}
			}
		}
	}
}
