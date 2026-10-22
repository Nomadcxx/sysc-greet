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

func TestRotatingModelBackgroundsMenuAndLifecycle(t *testing.T) {
	models := []struct{ id, label, credit string }{
		{"jesus-christ-statue", "Jesus Christ Statue", "Jesus Christ Statue / amcgi / CC BY 4.0"},
		{"jason", "Jason", "Jason / noe-3d.at / CC BY-NC 4.0"},
		{"nymphe-der-flora", "Nymphe der Flora", "Nymphe der Flora / noe-3d.at / CC BY-NC 4.0"},
		{"lowe", "Löwe", "Löwe / noe-3d.at / CC BY-NC 4.0"},
		{"roman-gladiator", "Roman Gladiator", "Sculpture of a Roman gladiator / Virtual Museums of Małopolska / CC BY 4.0"},
		{"grabfigur", "Grabfigur", "Grabfigur / noe-3d.at / CC BY-NC 4.0"},
		{"laocoon-and-his-sons", "Laocoon and His Sons", "Laocoon and His Sons / Rigsters / CC BY 4.0"},
		{"napoleon-bust", "Napoleon Bust", "The bust of Napoleon Bonaparte / Virtual Museums of Małopolska / CC BY 4.0"},
		{"brandenburg-gate", "Brandenburg Gate", "Brandenburg Gate / Resinnnn / CC BY 4.0"},
		{"louis-xiv", "Louis XIV", "Louis XIV / HoangHiepVu / CC BY 4.0"},
		{"engel", "Engel", "Engel / noe-3d.at / CC BY-NC 4.0"},
		{"st-george", "St George", "Hl. Georg als Drachentöter / VIMUNE / CC BY-SA 4.0"},
		{"santa-muerte", "Santa Muerte", "Santa Muerte / Mateusz Woliński / CC BY-SA 4.0"},
		{"aspasia", "Aspasia", "Aspasia / noe-3d.at / CC BY-NC 4.0"},
		{"st-nepomuk", "St Nepomuk", "Hl. Nepomuk / noe-3d.at / CC BY-NC 4.0"},
		{"lion-statue", "Lion Statue", "Lion Statue / Thomas Flynn / CC0 1.0"},
		{"f-14-tomcat", "F-14 Tomcat", "F-14 Tomcat - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0"},
		{"f-15e-strike-eagle", "F-15E Strike Eagle", "F-15E Strike Eagle - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0"},
		{"f-22-raptor", "F-22 Raptor", "F-22 Raptor - Fighter Jet - Free / bohmerang / CC BY-NC-SA 4.0"},
		{"fa-18e-super-hornet", "F/A-18E Super Hornet", "F/A-18E Super Hornet (USA) / 42manako / CC BY-NC 4.0"},
	}
	for _, entry := range models {
		t.Run(entry.id, func(t *testing.T) {
			m := initialModel(Config{TestMode: true}, false)
			m.mode = ModeBackgroundsSubmenu
			m.width, m.height, m.currentTheme = 140, 40, "Dracula"
			updated, _ := m.navigateToBackgroundsSubmenu()
			m = updated.(model)
			option := formatCheckbox(entry.label, false)
			if !slices.Contains(m.menuOptions, option) {
				t.Fatalf("background menu omits %q: %v", entry.label, m.menuOptions)
			}
			m.menuIndex = slices.Index(m.menuOptions, option)
			m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
			if m.selectedBackground != entry.id {
				t.Fatalf("selected background = %q, want %q", m.selectedBackground, entry.id)
			}
			if got := m.effectiveFormMode(); got != formSeeThrough {
				t.Fatalf("form mode = %v, want see-through", got)
			}
			if !suppressOuterBorder(entry.id) {
				t.Fatal("rotating model background kept the opaque outer border")
			}

			m.mode = ModeLogin
			if !m.backgroundEffectActive() {
				t.Fatal("selected rotating model background is inactive")
			}
			updated, _ = m.Update(bgTickMsg(time.Now()))
			m = updated.(model)
			view := m.View()
			if view.Layer == nil {
				t.Fatal("rotating model background was not placed behind the login view")
			}
			if rendered := renderRotatingModelLayer(view, m.width, m.height); !strings.Contains(rendered, entry.credit) {
				t.Fatalf("rendered background omits model credit %q", entry.credit)
			}

			updated, _ = m.Update(tea.WindowSizeMsg{Width: 70, Height: 20})
			m = updated.(model)
			if got := lipgloss.Width(renderRotatingModelLayer(m.View(), 70, 20)); got != 70 {
				t.Fatalf("resized background width = %d, want 70", got)
			}
		})
	}
}

func TestRotatingModelBackgroundRecreatesOnThemeChange(t *testing.T) {
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
		t.Fatal("theme selection reused the old rotating model effect")
	}
	if after := m.rotatingModelEffect.Render(); after == before {
		t.Fatal("rotating model rendering did not apply the selected theme")
	}
}
