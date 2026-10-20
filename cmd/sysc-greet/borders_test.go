package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

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

func TestDecorativeFitsAt80x24(t *testing.T) {
	for _, style := range []string{"classic", "modern", "minimal", "ascii1", "ascii2"} {
		t.Run(style, func(t *testing.T) {
			m := layoutModel(80, 24)
			m.selectedBorderStyle = style
			deco := m.renderDualBorderLayout(80, 24)
			if w, h := lipgloss.Width(deco), lipgloss.Height(deco); w > 80 || h > 24 {
				t.Fatalf("decorative %dx%d exceeds 80x24", w, h)
			}
			view := m.renderMainView(80, 24)
			if view != deco {
				t.Fatal("renderMainView used the fallback, not the decorative layout")
			}
		})
	}
}

func TestFrameRowsEqualWidth(t *testing.T) {
	for _, style := range []string{"classic", "modern", "minimal", "ascii1", "ascii2"} {
		for _, size := range [][2]int{{80, 24}, {132, 40}, {200, 55}} {
			m := layoutModel(size[0], size[1])
			m.selectedBorderStyle = style
			deco := m.renderDualBorderLayout(size[0], size[1])
			widths := map[int]bool{}
			for _, line := range strings.Split(deco, "\n") {
				widths[ansi.StringWidth(line)] = true
			}
			if len(widths) != 1 {
				t.Fatalf("%s at %dx%d: row widths %v", style, size[0], size[1], widths)
			}
		}
	}
}

func TestRenderMainHelpFit(t *testing.T) {
	m := layoutModel(80, 24)
	for w := 20; w <= 140; w++ {
		help := m.renderMainHelpFit(w)
		if got := ansi.StringWidth(help); got > w {
			t.Fatalf("help at width %d is %d wide: %q", w, got, help)
		}
		if w >= 30 {
			if !strings.Contains(help, "Enter") || !strings.Contains(help, "F1") {
				t.Fatalf("help at width %d missing Enter/F1: %q", w, help)
			}
		}
	}
	if got := m.renderMainHelpFit(78); got != "Tab Focus • ↑↓ Sessions • F1 Menu • F2 Sessions • F4 Power • Enter Continue" {
		t.Fatalf("78-col help = %q", got)
	}
}

func TestAscii1ResponsiveWidths(t *testing.T) {
	for w := 60; w <= 200; w += 10 {
		m := layoutModel(w, 32)
		m.selectedBorderStyle = "ascii1"
		deco := m.renderDualBorderLayout(w, 32)
		if got := lipgloss.Width(deco); got > w {
			t.Fatalf("ascii1 at %d cols renders %d wide", w, got)
		}
	}
}

func FuzzBorderRender(f *testing.F) {
	for _, seed := range [][2]int{{1, 1}, {40, 12}, {60, 20}, {80, 24}, {132, 40}, {200, 55}, {600, 300}} {
		f.Add(seed[0], seed[1])
	}
	f.Fuzz(func(t *testing.T, w, h int) {
		if w < 1 || w > 600 || h < 1 || h > 300 {
			t.Skip()
		}
		for _, style := range []string{"classic", "modern", "minimal", "ascii1", "ascii2"} {
			m := layoutModel(w, h)
			m.selectedBorderStyle = style
			_ = m.renderDualBorderLayout(w, h)
			_ = m.renderMainView(w, h)
		}
	})
}
