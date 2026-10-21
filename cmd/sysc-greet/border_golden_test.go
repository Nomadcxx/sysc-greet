package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
)

var updateBorderGoldens = flag.Bool("update", false, "update border golden renders")

func TestBorderGolden(t *testing.T) {
	useBorderGoldenTheme(t)

	registry := newBorderRegistry("")
	styles := []string{"classic", "modern", "minimal", "ascii1", "ascii2", "nfo01", "nfo02", "nfo03", "braille-hairline"}
	sizes := [][2]int{{80, 24}, {132, 40}, {200, 55}, {480, 135}}
	modes := []string{"opaque", "seethrough"}
	effects := []struct {
		name    string
		value   string
		elapsed time.Duration
	}{
		{name: "off", value: "off"},
		{name: "shimmer_0", value: "shimmer"},
		{name: "shimmer_500ms", value: "shimmer", elapsed: 500 * time.Millisecond},
		{name: "shimmer_3s", value: "shimmer", elapsed: 3 * time.Second},
	}

	for _, style := range styles {
		renderer := registry.get(style)
		if renderer == nil {
			t.Fatalf("registry omitted %s", style)
		}
		for _, size := range sizes {
			for _, mode := range modes {
				for _, effect := range effects {
					if effect.value != "off" && !renderer.SupportsEffect(effect.value) {
						continue
					}
					name := fmt.Sprintf("%s_%dx%d_%s_%s.golden", style, size[0], size[1], mode, effect.name)
					t.Run(name, func(t *testing.T) {
						got := renderBorderGolden(t, registry, style, size[0], size[1], mode, effect.value, effect.elapsed, "none")
						checkBorderGolden(t, name, got)
					})
				}
			}
		}
	}

	for _, style := range styles {
		for _, size := range [][2]int{{132, 40}, {200, 55}} {
			for _, mode := range []string{"seethrough", "frosted"} {
				name := fmt.Sprintf("%s_ak-47_%dx%d_%s_off.golden", style, size[0], size[1], mode)
				t.Run(name, func(t *testing.T) {
					got := renderBorderGolden(t, registry, style, size[0], size[1], mode, "off", 0, "ak-47")
					checkBorderGolden(t, name, got)
				})
			}
		}
	}
}

func TestBorderGoldenCompactFallback(t *testing.T) {
	useBorderGoldenTheme(t)

	m := layoutModel(60, 20)
	m.selectedSession = nil
	m.sessions = nil
	m.selectedBackground = "none"
	base := m.renderMainView(60, 20)
	checkBorderGolden(t, "compact_60x20.golden", base)

	for _, style := range []string{"classic", "modern", "minimal", "ascii1", "ascii2", "nfo01", "nfo02", "nfo03", "braille-hairline"} {
		m.selectedBorderStyle = style
		m.borderEffect = "shimmer"
		m.formMode = "seethrough"
		if got := m.renderMainView(60, 20); got != base {
			t.Errorf("compact fallback changed for %s", style)
		}
	}
}

func renderBorderGolden(t *testing.T, registry *borderRegistry, style string, width, height int, mode, effect string, elapsed time.Duration, background string) string {
	t.Helper()
	m := layoutModel(width, height)
	m.selectedSession = nil
	m.sessions = nil
	m.usernameInput.Blur()
	m.passwordInput.Blur()
	m.selectedBorderStyle = style
	m.borderRegistry = registry
	m.selectedBackground = background
	m.borderPaletteMode = "theme"
	m.borderEffect = effect
	m.formMode = mode
	m.config = Config{BorderLayout: "hug", FormHalo: 1, FormDim: 0.55}
	m.formLayerCache = &layerCache{}
	if effect == "shimmer" {
		m.idleTimer = time.Now().Add(-10*time.Second - elapsed - 25*time.Millisecond)
	}
	if background == "ak-47" {
		m.ak47Effect = m.newAK47Effect()
	}
	view := m.View()
	canvas, ok := view.Layer.(*lipgloss.Canvas)
	if !ok {
		t.Fatalf("View returned %T instead of the registered border canvas", view.Layer)
	}
	return canvas.Render()
}

func checkBorderGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "borders", name)
	if *updateBorderGoldens {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run go test ./cmd/sysc-greet -run TestBorderGolden -update): %v", path, err)
	}
	if !bytes.Equal([]byte(got), want) {
		t.Errorf("render differs from %s", path)
	}
}

func useBorderGoldenTheme(t *testing.T) {
	t.Helper()
	oldProfile := colorProfile
	oldColors := [15]color.Color{BgBase, BgElevated, BgSubtle, BgActive, Primary, Secondary, Accent, Warning, Danger,
		FgPrimary, FgSecondary, FgMuted, FgSubtle, BorderDefault, BorderFocus}
	colorProfile = colorprofile.TrueColor
	applyTheme("dracula", true)
	t.Cleanup(func() {
		colorProfile = oldProfile
		BgBase, BgElevated, BgSubtle, BgActive = oldColors[0], oldColors[1], oldColors[2], oldColors[3]
		Primary, Secondary, Accent, Warning, Danger = oldColors[4], oldColors[5], oldColors[6], oldColors[7], oldColors[8]
		FgPrimary, FgSecondary, FgMuted, FgSubtle = oldColors[9], oldColors[10], oldColors[11], oldColors[12]
		BorderDefault, BorderFocus = oldColors[13], oldColors[14]
	})
}
