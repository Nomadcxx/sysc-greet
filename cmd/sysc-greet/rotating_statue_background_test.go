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

func renderRotatingModelLayer(view tea.View, width, height int) string {
	screen := uv.NewScreenBuffer(width, height)
	view.Layer.Draw(screen, screen.Bounds())
	return screen.String()
}

func TestRotatingStatueBackgroundsMenuAndLifecycle(t *testing.T) {
	statues := []struct{ id, label, credit string }{
		{"jesus-christ-statue", "Jesus Christ Statue", "Jesus Christ Statue / amcgi / CC BY 4.0"},
		{"jason", "Jason", "Jason / noe-3d.at / CC BY-NC 4.0"},
		{"nymphe-der-flora", "Nymphe der Flora", "Nymphe der Flora / noe-3d.at / CC BY-NC 4.0"},
		{"lowe", "Löwe", "Löwe / noe-3d.at / CC BY-NC 4.0"},
		{"roman-gladiator", "Roman Gladiator", "Sculpture of a Roman gladiator / Virtual Museums of Małopolska / CC BY 4.0"},
		{"grabfigur", "Grabfigur", "Grabfigur / noe-3d.at / CC BY-NC 4.0"},
	}
	for _, statue := range statues {
		t.Run(statue.id, func(t *testing.T) {
			m := initialModel(Config{TestMode: true}, false)
			m.mode = ModeBackgroundsSubmenu
			m.width, m.height, m.currentTheme = 140, 40, "Dracula"
			updated, _ := m.navigateToBackgroundsSubmenu()
			m = updated.(model)
			option := formatCheckbox(statue.label, false)
			if !slices.Contains(m.menuOptions, option) {
				t.Fatalf("background menu omits %q: %v", statue.label, m.menuOptions)
			}
			m.menuIndex = slices.Index(m.menuOptions, option)
			m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
			if m.selectedBackground != statue.id {
				t.Fatalf("selected background = %q, want %q", m.selectedBackground, statue.id)
			}
			if got := m.effectiveFormMode(); got != formSeeThrough {
				t.Fatalf("form mode = %v, want see-through", got)
			}
			if !suppressOuterBorder(statue.id) {
				t.Fatal("statue background kept the opaque outer border")
			}

			m.mode = ModeLogin
			if !m.backgroundEffectActive() {
				t.Fatal("selected statue background is inactive")
			}
			updated, _ = m.Update(bgTickMsg(time.Now()))
			m = updated.(model)
			view := m.View()
			if view.Layer == nil {
				t.Fatal("statue background was not placed behind the login view")
			}
			if rendered := renderRotatingModelLayer(view, m.width, m.height); !strings.Contains(rendered, statue.credit) {
				t.Fatalf("rendered background omits model credit %q", statue.credit)
			}

			updated, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
			m = updated.(model)
			if got := lipgloss.Width(renderRotatingModelLayer(m.View(), 70, 20)); got != 70 {
				t.Fatalf("resized background width = %d, want 70", got)
			}
		})
	}
}

func TestRotatingStatueBackgroundRecreatesOnThemeChange(t *testing.T) {
	m := initialModel(Config{TestMode: true}, false)
	m.mode = ModeBackgroundsSubmenu
	m.width, m.height, m.currentTheme = 100, 30, "Dracula"
	updated, _ := m.navigateToBackgroundsSubmenu()
	m = updated.(model)
	m.menuIndex = slices.Index(m.menuOptions, formatCheckbox("Jason", false))
	if m.menuIndex < 0 {
		t.Fatal("Jason menu option not found")
	}
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
	oldEffect := m.rotatingModelEffect
	before := oldEffect.Render()

	m.mode = ModeThemesSubmenu
	m.menuOptions = []string{"← Back", "Theme: Nord"}
	m.menuIndex = 1
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.currentTheme != "Nord" {
		t.Fatalf("selected theme = %q, want Nord", m.currentTheme)
	}
	if m.rotatingModelEffect == oldEffect {
		t.Fatal("theme selection reused the old rotating statue effect")
	}
	if after := m.rotatingModelEffect.Render(); after == before {
		t.Fatal("rotating statue rendering did not apply the selected theme")
	}
}
