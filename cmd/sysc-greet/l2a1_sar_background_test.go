package main

import (
	"image"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

func TestBackgroundsMenuIncludesL2A1SAR(t *testing.T) {
	got, _ := (model{}).navigateToBackgroundsSubmenu()
	menu := got.(model)
	if !slices.Contains(menu.menuOptions, "[ ] L2A1 SAR") {
		t.Fatalf("background menu = %v, want L2A1 SAR option", menu.menuOptions)
	}
}

func TestL2A1SARBackgroundCanBeSelectedFromMenu(t *testing.T) {
	m := model{mode: ModeBackgroundsSubmenu, config: Config{TestMode: true}}
	updated, _ := m.navigateToBackgroundsSubmenu()
	m = updated.(model)
	m.menuIndex = slices.Index(m.menuOptions, "[ ] L2A1 SAR")
	if m.menuIndex < 0 {
		t.Fatal("L2A1 SAR menu option not found")
	}

	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.selectedBackground != "l2a1-sar" {
		t.Fatalf("selected background = %q, want l2a1-sar", m.selectedBackground)
	}
}

func TestL2A1SARBackgroundInitializesAndRendersCreditInTestMode(t *testing.T) {
	t.Setenv("SYSC_BG", "l2a1-sar")
	m := initialModel(Config{TestMode: true}, false)
	if m.selectedBackground != "l2a1-sar" || !m.backgroundEffectActive() {
		t.Fatalf("test background = %q, active = %v", m.selectedBackground, m.backgroundEffectActive())
	}

	updated, _ := m.Update(bgTickMsg(time.Now()))
	m = updated.(model)
	view := m.View()
	if view.Layer == nil {
		t.Fatal("L2A1 SAR background was not placed behind the login view")
	}
	screen := uv.NewScreenBuffer(m.width, m.height)
	view.Layer.Draw(screen, screen.Bounds())
	if !strings.Contains(screen.String(), "L2A1 SAR / notcplkerry / CC BY 4.0") {
		t.Fatal("rendered background is missing its model credit")
	}
}

func TestL2A1SARBackgroundResizesWithTerminal(t *testing.T) {
	m := initialModel(Config{TestMode: true}, false)
	m.mode = ModeLogin
	m.selectedBackground = "l2a1-sar"
	updated, _ := m.Update(bgTickMsg(time.Now()))
	m = updated.(model)
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = updated.(model)
	view := m.View()
	if view.Layer == nil {
		t.Fatal("L2A1 SAR background layer missing after resize")
	}
	screen := uv.NewScreenBuffer(40, 12)
	view.Layer.Draw(screen, screen.Bounds())
	if !strings.Contains(screen.String(), "L2A1 SAR / notcplkerry / CC BY 4.0") {
		t.Fatal("resized background is missing its model credit")
	}
	bounded, ok := view.Layer.(interface{ Bounds() image.Rectangle })
	if !ok {
		t.Fatal("resized background layer does not expose bounds")
	}
	if got := bounded.Bounds().Dx(); got != 40 {
		t.Fatalf("rendered width after resize = %d, want 40", got)
	}
}
