package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	UsernameInput textinput.Model
	PasswordInput textinput.Model
	Mode          string // "username" or "password"
	TestMode      bool
}

func NewMainModel(testMode bool) MainModel {
	ti := textinput.New()
	ti.Placeholder = "Username"
	ti.Focus()

	return MainModel{
		UsernameInput: ti,
		PasswordInput: textinput.New(),
		Mode:          "username",
		TestMode:      testMode,
	}
}

func (m MainModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.Mode == "username" {
				m.Mode = "password"
				m.UsernameInput.Blur()
				m.PasswordInput.Focus()
				return m, textinput.Blink
			} else {
				// Simulate auth
				if m.TestMode {
					// In test mode, just quit without actual auth
				}
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	if m.Mode == "username" {
		m.UsernameInput, cmd = m.UsernameInput.Update(msg)
	} else {
		m.PasswordInput, cmd = m.PasswordInput.Update(msg)
	}
	return m, cmd
}

func (m MainModel) View() string {
	var b strings.Builder

	b.WriteString("sysc-greet\n\n") // FIXED 2025-10-15 - Corrected title from bubble-greet to sysc-greet

	if m.Mode == "username" {
		b.WriteString("Username: " + m.UsernameInput.View() + "\n")
	} else {
		b.WriteString("Password: " + m.PasswordInput.View() + "\n")
	}

	b.WriteString("\nPress Enter to submit, Ctrl+C to quit.\n")

	return b.String()
}
