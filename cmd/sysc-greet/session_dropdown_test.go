package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/Nomadcxx/sysc-greet/internal/sessions"
)

func TestSessionDropdownAcrossBorders(t *testing.T) {
	applyTheme("dracula", true)
	available := make([]sessions.Session, 9)
	for i := range available {
		available[i] = sessions.Session{Name: fmt.Sprintf("Session%02d", i), Exec: "session", Type: "Wayland"}
	}
	for _, background := range []string{"none", "sonar"} {
		for _, border := range []string{"classic", "modern", "minimal", "wave", "pulse", "ascii1", "ascii2", "ascii3", "ascii4"} {
			t.Run(border+"/"+background, func(t *testing.T) {
				m := model{
					config: Config{TestMode: true}, mode: ModeLogin,
					sessions: available, selectedSession: &available[0],
					usernameInput: textinput.New(), passwordInput: textinput.New(),
					selectedBorderStyle: border, selectedBackground: background,
				}
				render := func() string { return stripAnsi(m.renderMainView(213, 54)) }
				if strings.Contains(render(), "Session01") {
					t.Fatal("closed dropdown shows an unselected session")
				}
				m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyF2})
				if !m.sessionDropdownOpen || !strings.Contains(render(), "Session01") {
					t.Fatal("F2 must open and render the session dropdown")
				}
				for _, shortcut := range []string{"F1 Menu", "F2 Sessions", "F3 Notes", "F4 Power"} {
					if !strings.Contains(render(), shortcut) {
						t.Errorf("help omits %q", shortcut)
					}
				}
				if strings.Contains(render(), "Session08") || !strings.Contains(render(), "more below") {
					t.Fatal("dropdown must limit the initial list and show a scroll indicator")
				}
				for range len(available) - 1 {
					m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyDown})
				}
				if !strings.Contains(render(), "Session07") || !strings.Contains(render(), "more above") {
					t.Fatal("dropdown must scroll to the last session")
				}
				var cmd tea.Cmd
				m, cmd = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
				if m.sessionDropdownOpen || cmd == nil {
					t.Fatal("Enter must close the dropdown and select a session")
				}
				selected, ok := cmd().(sessionSelectedMsg)
				if !ok || selected.Name != "Session08" {
					t.Fatalf("expected the last session, got %#v", selected)
				}
				if strings.Contains(render(), "Session07") {
					t.Fatal("selected dropdown still shows the list")
				}
				m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyF2})
				m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyF2})
				if m.sessionDropdownOpen {
					t.Fatal("F2 must toggle the dropdown closed")
				}
			})
		}
	}
}

func TestFunctionKeyBindings(t *testing.T) {
	for _, tt := range []struct {
		key  rune
		mode ViewMode
	}{
		{tea.KeyF1, ModeMenu},
		{tea.KeyF3, ModeReleaseNotes},
		{tea.KeyF4, ModePower},
	} {
		t.Run((tea.KeyPressMsg{Code: tt.key}).String(), func(t *testing.T) {
			m := model{
				config: Config{TestMode: true}, sessionDropdownOpen: true,
				usernameInput: textinput.New(), passwordInput: textinput.New(),
			}
			m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tt.key})
			if m.mode != tt.mode || m.sessionDropdownOpen {
				t.Fatalf("mode=%s dropdown=%v; want %s with dropdown closed", m.mode, m.sessionDropdownOpen, tt.mode)
			}
		})
	}
}
