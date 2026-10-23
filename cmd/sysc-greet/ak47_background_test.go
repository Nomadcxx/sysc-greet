package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

func TestBackgroundsMenuIncludesAK47(t *testing.T) {
	got, _ := (model{}).navigateToBackgroundsSubmenu()
	menu := got.(model)
	if !slices.Contains(menu.menuOptions, "[ ] AK-47") {
		t.Fatalf("background menu = %v, want AK-47 option", menu.menuOptions)
	}
}

func TestAK47BackgroundCanBeSelectedFromMenu(t *testing.T) {
	m := model{mode: ModeBackgroundsSubmenu, config: Config{TestMode: true}}
	updated, _ := m.navigateToBackgroundsSubmenu()
	m = updated.(model)
	m.menuIndex = slices.Index(m.menuOptions, "[ ] AK-47")
	if m.menuIndex < 0 {
		t.Fatal("AK-47 menu option not found")
	}

	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.selectedBackground != "ak-47" {
		t.Fatalf("selected background = %q, want ak-47", m.selectedBackground)
	}
}

func TestAK47BackgroundInitializesAndRendersCreditInTestMode(t *testing.T) {
	t.Setenv("SYSC_BG", "ak-47")
	m := initialModel(Config{TestMode: true}, false)
	if m.selectedBackground != "ak-47" || !m.backgroundEffectActive() {
		t.Fatalf("test background = %q, active = %v", m.selectedBackground, m.backgroundEffectActive())
	}

	updated, _ := m.Update(bgTickMsg(time.Now()))
	m = updated.(model)
	if m.ak47Effect == nil {
		t.Fatal("background tick did not initialize the AK-47 effect")
	}
	firstFrame := m.addAK47Effect("")
	if !strings.Contains(firstFrame, "AK-47 / Lokeig / CC BY-NC 4.0") {
		t.Fatalf("rendered effect is missing its model credit: %q", firstFrame)
	}
	if view := m.View(); view.Layer == nil {
		t.Fatal("AK-47 background was not placed behind the login view")
	}
	updated, _ = m.Update(bgTickMsg(time.Now().Add(50 * time.Millisecond)))
	m = updated.(model)
	if got := m.addAK47Effect(""); got == firstFrame {
		t.Fatal("AK-47 frame did not advance on the background tick")
	}
}

func TestAK47BackgroundResizesWithTerminal(t *testing.T) {
	m := model{
		mode: ModeLogin, width: 80, height: 24,
		selectedBackground: "ak-47", currentTheme: "Dracula",
		ak47Effect: syscgoanimations.NewAK47Effect(80, 24, "Dracula"),
	}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = updated.(model)
	if got := lipgloss.Width(m.addAK47Effect("")); got != 40 {
		t.Fatalf("rendered width after resize = %d, want 40", got)
	}
}

type cellAspectRecorder struct {
	width, height int
}

func (r *cellAspectRecorder) SetCellAspect(width, height int) {
	r.width, r.height = width, height
}

func TestSetCellAspectUsesTerminalPixelGeometry(t *testing.T) {
	r := &cellAspectRecorder{}
	setCellAspect(r, 110, 32, 990, 640)
	if r.width != 9 || r.height != 20 {
		t.Fatalf("cell aspect = %dx%d, want 9x20", r.width, r.height)
	}

	setCellAspect(r, 110, 32, 0, 0)
	if r.width != 9 || r.height != 20 {
		t.Fatalf("missing pixel geometry changed cell aspect to %dx%d", r.width, r.height)
	}
}

func TestThemeChangeRecreatesAK47Background(t *testing.T) {
	m := model{
		config: Config{TestMode: true}, mode: ModeThemesSubmenu,
		menuOptions: []string{"← Back", "Theme: Nord"}, menuIndex: 1,
		currentTheme: "Dracula", selectedBackground: "ak-47",
		width: 80, height: 24,
		ak47Effect: syscgoanimations.NewAK47Effect(80, 24, "Dracula"),
	}
	oldEffect := m.ak47Effect
	oldFrame := oldEffect.Render()
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.currentTheme != "Nord" || m.ak47Effect == oldEffect {
		t.Fatal("theme selection did not recreate the AK-47 effect")
	}
	if got := m.addAK47Effect(""); got == oldFrame {
		t.Fatal("AK-47 rendering did not apply the selected theme")
	}
}

func TestAK47UsesSyscGoNameForTokyoNightTheme(t *testing.T) {
	m := model{width: 80, height: 24, currentTheme: "Tokyo Night"}
	got := m.newAK47Effect().Render()
	want := syscgoanimations.NewAK47Effect(80, 24, "tokyo-night").Render()
	if got != want {
		t.Fatal("Tokyo Night did not use the sysc-Go Tokyo Night palette")
	}
}

func TestAK47BackgroundTicksOnlyOnLoginScreens(t *testing.T) {
	for _, tc := range []struct {
		mode ViewMode
		want bool
	}{
		{mode: ModeLogin, want: true},
		{mode: ModePassword, want: true},
		{mode: ModeMenu, want: false},
	} {
		m := model{mode: tc.mode, selectedBackground: "ak-47"}
		if got := m.backgroundEffectActive(); got != tc.want {
			t.Errorf("backgroundEffectActive() in %s = %v, want %v", tc.mode, got, tc.want)
		}
	}
}
