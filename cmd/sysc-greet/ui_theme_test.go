package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

func TestInputThemeIncludesBlurredStyles(t *testing.T) {
	applyTheme("nord", true)
	t.Cleanup(func() { applyTheme("dracula", true) })
	m := model{usernameInput: textinput.New(), passwordInput: textinput.New()}
	m.applyInputTheme()

	for name, input := range map[string]textinput.Model{
		"username": m.usernameInput,
		"password": m.passwordInput,
	} {
		for state, styles := range map[string]textinput.StyleState{
			"focused": input.Styles.Focused,
			"blurred": input.Styles.Blurred,
		} {
			t.Run(name+"/"+state, func(t *testing.T) {
				if !sameColor(styles.Prompt.GetForeground(), Primary) {
					t.Error("prompt does not use the theme primary color")
				}
				if !sameColor(styles.Text.GetForeground(), FgPrimary) {
					t.Error("text does not use the theme foreground color")
				}
				if !sameColor(styles.Placeholder.GetForeground(), FgMuted) {
					t.Error("placeholder does not use the theme muted color")
				}
			})
		}
	}
}

func TestBorderedHelpUsesThemeForeground(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SYSC_BG", "")
	t.Cleanup(func() { applyTheme("dracula", true) })
	m := initialModel(Config{TestMode: true, ThemeName: "Nord"}, false)
	m.width, m.height = 132, 40

	for _, mode := range []ViewMode{ModeLogin, ModePassword, ModeLoading} {
		t.Run(string(mode), func(t *testing.T) {
			m.mode = mode
			canvas, ok := m.borderedLoginCanvas(m.width, m.height)
			if !ok {
				t.Fatal("bordered layout did not render")
			}
			screen := uv.NewScreenBuffer(m.width, m.height)
			canvas.Draw(screen, screen.Bounds())
			text := "F1 Menu"
			if mode == ModeLoading {
				text = "Please wait..."
			}
			for y, line := range strings.Split(screen.Buffer.String(), "\n") {
				if index := strings.Index(line, text); index >= 0 {
					x := lipgloss.Width(line[:index])
					for offset := range len(text) {
						if !sameColor(screen.CellAt(x+offset, y).Style.Fg, FgMuted) {
							t.Fatalf("help text %q does not use the theme muted color", text)
						}
					}
					return
				}
			}
			t.Fatalf("help text %q missing", text)
		})
	}
}

func TestClosingMenuRestoresUsernameFocus(t *testing.T) {
	for _, passwordMode := range []bool{false, true} {
		for _, closeKey := range []struct {
			name string
			code rune
		}{
			{"escape", tea.KeyEscape},
			{"enter", tea.KeyEnter},
		} {
			mode := "username"
			if passwordMode {
				mode = "password"
			}
			t.Run(mode+"/"+closeKey.name, func(t *testing.T) {
				m := model{
					mode: ModeLogin, focusState: FocusUsername,
					usernameInput: textinput.New(), passwordInput: textinput.New(),
				}
				m.usernameInput.SetValue("sample")
				if passwordMode {
					m.mode, m.focusState = ModePassword, FocusPassword
					m.passwordInput.Focus()
				} else {
					m.usernameInput.Focus()
				}
				m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyF1})
				m, cmd := m.handleKeyInput(tea.KeyPressMsg{Code: closeKey.code})
				if m.mode != ModeLogin || m.focusState != FocusUsername || !m.usernameInput.Focused() || m.passwordInput.Focused() {
					t.Fatal("closing the menu must return to a focused username input")
				}
				if cmd == nil {
					t.Error("closing the menu must restart the input cursor blink")
				}
				if m.usernameInput.Value() != "sample" {
					t.Error("closing the menu changed the username")
				}
			})
		}
	}
}
