package main

import (
	"image"
	"strings"
	"testing"
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
		"Mango greeter compositor",
		"Supported greeter backends: niri (default), cagebreak, sway, mango",
		"Hyprland greeter support deprecated (login sessions unaffected)",
	} {
		if !strings.Contains(notes, text) {
			t.Errorf("release notes omit %q", text)
		}
	}
}
