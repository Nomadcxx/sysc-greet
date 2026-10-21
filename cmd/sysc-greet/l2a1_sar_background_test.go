package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

func renderLayer(view tea.View, width, height int) string {
	screen := uv.NewScreenBuffer(width, height)
	view.Layer.Draw(screen, screen.Bounds())
	return screen.String()
}

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
	if !strings.Contains(renderLayer(view, m.width, m.height), "L2A1 SAR / notcplkerry / CC BY 4.0") {
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
	rendered := renderLayer(view, 40, 12)
	if !strings.Contains(rendered, "L2A1 SAR / notcplkerry / CC BY 4.0") {
		t.Fatal("resized background is missing its model credit")
	}
	if got := lipgloss.Width(rendered); got != 40 {
		t.Fatalf("rendered width after resize = %d, want 40", got)
	}
}
