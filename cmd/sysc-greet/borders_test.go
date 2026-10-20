package main

import (
	"testing"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
)

func TestBorderPrefsMigration(t *testing.T) {
	cases := []struct {
		style, effect, wantStyle, wantEffect string
	}{
		{"ascii3", "", "ascii2", "off"},
		{"ascii4", "", "ascii2", "off"},
		{"wave", "", "classic", "shimmer"},
		{"pulse", "", "modern", "shimmer"},
		{"wave", "off", "classic", "off"},
		{"classic", "shimmer", "classic", "shimmer"},
		{"", "", "", "off"},
	}
	for _, c := range cases {
		style, effect := migrateBorderPrefs(c.style, c.effect)
		if style != c.wantStyle || effect != c.wantEffect {
			t.Errorf("migrateBorderPrefs(%q, %q) = (%q, %q), want (%q, %q)",
				c.style, c.effect, style, effect, c.wantStyle, c.wantEffect)
		}
	}
}

func TestMigratedBorderStyleRendersWithoutPanic(t *testing.T) {
	for _, size := range [][2]int{{40, 12}, {60, 20}, {80, 24}} {
		for _, removed := range []string{"ascii3", "ascii4"} {
			t.Run(removed, func(t *testing.T) {
				m := layoutModel(size[0], size[1])
				m.selectedBorderStyle, m.borderEffect = migrateBorderPrefs(removed, "")
				if m.selectedBorderStyle != "ascii2" {
					t.Fatalf("style = %q, want ascii2", m.selectedBorderStyle)
				}
				view := m.renderMainView(m.width, m.height)
				if view == "" {
					t.Fatal("empty render")
				}
			})
		}
	}
}

func TestBorderEffectOffIsByteIdentical(t *testing.T) {
	for _, style := range []string{"classic", "modern", "minimal", "ascii1", "ascii2"} {
		m := layoutModel(80, 24)
		m.selectedBorderStyle = style
		m.borderEffect = "off"
		m.animationFrame = 0
		a := m.renderMainView(m.width, m.height)
		m.animationFrame = 777
		b := m.renderMainView(m.width, m.height)
		if a != b {
			t.Fatalf("style %s: effect off is not byte-identical across frames", style)
		}
	}
}

func TestBorderPrefsRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := layoutModel(80, 24)
	m.borderEffect = "shimmer"
	m.formMode = "frosted"
	if err := cache.SavePreferences(m.currentPreferences()); err != nil {
		t.Fatal(err)
	}
	prefs, err := cache.LoadPreferences()
	if err != nil {
		t.Fatal(err)
	}
	if prefs.BorderEffect != "shimmer" {
		t.Errorf("BorderEffect = %q, want shimmer", prefs.BorderEffect)
	}
	if prefs.FormMode != "frosted" {
		t.Errorf("FormMode = %q, want frosted", prefs.FormMode)
	}
}
