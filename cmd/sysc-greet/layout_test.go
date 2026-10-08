package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nomadcxx/sysc-greet/internal/sessions"
)

func layoutModel(width, height int) model {
	username, password := textinput.New(), textinput.New()
	username.Prompt, password.Prompt = "", ""
	username.SetValue("nomadx")
	password.EchoMode = textinput.EchoPassword
	password.SetValue("never-render-this-password")
	username.Focus()
	session := sessions.Session{Name: "XFCE", Type: "Wayland"}
	return model{width: width, height: height, mode: ModeLogin, usernameInput: username, passwordInput: password, spinner: spinner.New(), selectedSession: &session, sessions: []sessions.Session{session}, selectedBorderStyle: "classic", selectedBackground: "none", focusState: FocusUsername, currentTheme: "dracula"}
}

func TestCompactAuthenticationFits(t *testing.T) {
	for _, size := range [][2]int{{40, 16}, {60, 20}, {80, 24}, {120, 32}} {
		for _, border := range []string{"classic", "modern", "minimal", "ascii1", "ascii2", "ascii3", "ascii4"} {
			for _, mode := range []ViewMode{ModeLogin, ModePassword, ModeLoading} {
				t.Run(fmt.Sprintf("%dx%d/%s/%s", size[0], size[1], border, mode), func(t *testing.T) {
					m := layoutModel(size[0], size[1])
					m.mode = mode
					m.selectedBorderStyle = border
					if mode == ModePassword {
						m.focusState = FocusPassword
						m.passwordInput.Focus()
						m.usernameInput.Blur()
						m.errorMessage = "Authentication failed: please check your credentials"
						m.failedAttempts = 3
						m.capsLockOn = true
					}
					view := m.renderMainView(m.width, m.height)
					if w, h := lipgloss.Width(view), lipgloss.Height(view); w > m.width || h > m.height {
						t.Fatalf("rendered %dx%d exceeds %dx%d", w, h, m.width, m.height)
					}
					plain := ansi.Strip(view)
					required := []string{"Session:"}
					switch mode {
					case ModeLogin:
						required = append(required, "Username:", "Enter")
					case ModePassword:
						required = append(required, "Password:", "Authentication failed", "Failed attempts: 3", "CAPS LOCK ON", "Enter")
					case ModeLoading:
						required = append(required, "Authenticating")
					}
					for _, text := range required {
						if !strings.Contains(plain, text) {
							t.Errorf("missing %q in %s", text, plain)
						}
					}
					if strings.Contains(plain, "never-render-this-password") {
						t.Fatal("password rendered in clear text")
					}
				})
			}
		}
	}
}

func TestCompactLongInputKeepsCursorTail(t *testing.T) {
	m := layoutModel(80, 24)
	m.usernameInput.SetValue(strings.Repeat("u", 200) + "TAIL")
	m.usernameInput.CursorEnd()
	m.selectedSession.Name = "XFCE " + strings.Repeat("界", 90)
	view := m.renderMainView(m.width, m.height)
	if lipgloss.Width(view) > m.width || lipgloss.Height(view) > m.height {
		t.Fatalf("long input overflow: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	if !strings.Contains(ansi.Strip(view), "TAIL") {
		t.Fatal("focused input tail clipped")
	}
	if m.usernameInput.Value() != strings.Repeat("u", 200)+"TAIL" {
		t.Fatal("rendering modified input")
	}
}

func TestCompactSessionDropdownFits(t *testing.T) {
	m := layoutModel(80, 24)
	m.mode = ModePassword
	m.focusState = FocusSession
	m.sessionDropdownOpen = true
	m.errorMessage = "Authentication failed"
	m.failedAttempts = 3
	for i := 1; i < 16; i++ {
		m.sessions = append(m.sessions, sessions.Session{Name: fmt.Sprintf("Session %02d", i), Type: "Wayland"})
	}
	m.sessionIndex = 7
	m.selectedSession = &m.sessions[7]
	view := m.renderMainView(m.width, m.height)
	if lipgloss.Width(view) > m.width || lipgloss.Height(view) > m.height {
		t.Fatalf("dropdown overflow: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	for _, text := range []string{"Session 07", "Password:", "Authentication failed", "Enter", "Esc"} {
		if !strings.Contains(ansi.Strip(view), text) {
			t.Errorf("missing %q", text)
		}
	}
}

func TestCompactKeepsFittingDecoration(t *testing.T) {
	m := layoutModel(240, 100)
	normal := m.renderDualBorderLayout(m.width, m.height)
	if lipgloss.Width(normal) > m.width || lipgloss.Height(normal) > m.height {
		t.Fatal("test's decorated layout does not fit")
	}
	if m.renderMainView(m.width, m.height) != normal {
		t.Fatal("fitting decoration changed")
	}
}
