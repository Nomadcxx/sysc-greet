package main

import (
	"testing"

	"github.com/charmbracelet/bubbles/v2/textinput"
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
