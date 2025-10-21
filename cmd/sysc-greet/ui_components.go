package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// UI Components - Extracted during Phase 3 refactoring
// This file contains reusable UI rendering functions for the greeter interface

// renderMonochromeForm renders the login form with monochrome styling
func (m model) renderMonochromeForm(width int) string {
	monoWhite := lipgloss.Color("#ffffff")
	monoLight := lipgloss.Color("#cccccc")
	// Use BgBase instead of monoDark to prevent bleeding
	monoDark := BgBase

	var sections []string

	// Session selector with monochrome styling
	if len(m.sessions) > 0 {
		// Remove Width() and Align(Center)
		sessionStyle := lipgloss.NewStyle().
			Foreground(monoWhite).
			Background(monoDark).
			Bold(true).
			Padding(0, 1)

		sessionText := fmt.Sprintf("Session: %s (%s)", m.selectedSession.Name, m.selectedSession.Type)
		sections = append(sections, sessionStyle.Render(sessionText))
		sections = append(sections, "")
	}

	// Username input
	usernameStyle := lipgloss.NewStyle().
		Foreground(monoLight).
		Width(width).
		Align(lipgloss.Left)

	m.usernameInput.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(monoWhite).Bold(true)
	m.usernameInput.Styles.Focused.Text = lipgloss.NewStyle().Foreground(monoWhite)
	sections = append(sections, usernameStyle.Render(m.usernameInput.View()))

	// Password input
	passwordStyle := lipgloss.NewStyle().
		Foreground(monoLight).
		Width(width).
		Align(lipgloss.Left)

	m.passwordInput.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(monoWhite).Bold(true)
	m.passwordInput.Styles.Focused.Text = lipgloss.NewStyle().Foreground(monoWhite)
	sections = append(sections, passwordStyle.Render(m.passwordInput.View()))

	// CHANGED 2025-10-05 - Display error message in monochrome style
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)
		sections = append(sections, "")
		sections = append(sections, errorStyle.Render("✗ "+m.errorMessage))
	}

	return strings.Join(sections, "\n")
}

// renderMainForm renders the main login form with session, username/password inputs
func (m model) renderMainForm(width int) string {
	var parts []string

	// Session selection (always visible at top)
	sessionContent := m.renderSessionSelector(width)
	parts = append(parts, sessionContent)

	// Add spacing
	parts = append(parts, "")

	// Current input based on mode
	switch m.mode {
	case ModeLogin:
		usernameLabel := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.getFocusColor(FocusUsername)).
			Width(10).
			Render("Username:")

		// Remove Foreground, use BgBase only
		inputStyle := lipgloss.NewStyle().
			Background(BgBase).
			Padding(0, 1)

		usernameInput := inputStyle.Render(m.usernameInput.View())

		usernameRow := lipgloss.JoinHorizontal(
			lipgloss.Left,
			usernameLabel,
			" ",
			usernameInput,
		)
		parts = append(parts, usernameRow)

	case ModePassword:
		passwordLabel := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.getFocusColor(FocusPassword)).
			Width(10).
			Render("Password:")

		// Remove Foreground, use BgBase only
		inputStyle := lipgloss.NewStyle().
			Background(BgBase).
			Padding(0, 1)

		passwordInput := inputStyle.Render(m.passwordInput.View())

		// Simple password row without spinner
		passwordRow := lipgloss.JoinHorizontal(
			lipgloss.Left,
			passwordLabel,
			" ",
			passwordInput,
		)
		parts = append(parts, passwordRow)

		// CAPS LOCK warning
		if m.capsLockOn && m.focusState == FocusPassword {
			capsLockStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555")).
				Bold(true).
				Align(lipgloss.Center).
				Width(width)
			parts = append(parts, "")
			parts = append(parts, capsLockStyle.Render("⚠ CAPS LOCK ON"))
		}

		// CHANGED 2025-10-05 - Display error message below password in main form
		if m.errorMessage != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555")).
				Bold(true)
			parts = append(parts, "")
			parts = append(parts, errorStyle.Render("✗ "+m.errorMessage))
		}

	case ModeLoading:
		loadingStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(Accent).
			Align(lipgloss.Center).
			Width(width)

		// Show animated spinner
		loadingText := loadingStyle.Render("Authenticating... " + m.spinner.View())
		parts = append(parts, loadingText)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderSessionSelector renders the session selector with dropdown indicator
func (m model) renderSessionSelector(width int) string {
	// Session label with focus indication
	sessionLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.getFocusColor(FocusSession)).
		Width(10).
		Render("Session:")

	// Current session display
	var sessionDisplay string
	if m.selectedSession != nil {
		sessionText := fmt.Sprintf("%s (%s)", m.selectedSession.Name, m.selectedSession.Type)

		borderColor := BorderDefault
		if m.focusState == FocusSession {
			borderColor = BorderFocus
		}

		// Use Inline(true) to prevent border from adding height
		sessionStyle := lipgloss.NewStyle().
			Foreground(FgPrimary).
			Background(BgBase).
			Padding(0, 1).
			Bold(true).
			Inline(true) // Force single-line rendering, ignore border height

		if m.sessionDropdownOpen || m.focusState == FocusSession {
			sessionStyle = sessionStyle.
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(borderColor)
		}

		sessionDisplay = sessionStyle.Render(sessionText)
	} else {
		sessionDisplay = lipgloss.NewStyle().
			Foreground(FgMuted).
			Width(width - 14).
			Render("No session selected")
	}

	// Dropdown indicator
	dropdownIndicator := "▼"
	if m.sessionDropdownOpen {
		dropdownIndicator = "▲"
	}
	indicatorStyle := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Render(dropdownIndicator)

	sessionRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		sessionLabel,
		" ",
		sessionDisplay,
		" ",
		indicatorStyle,
	)

	// Minimal spacing, show dropdown when open
	if m.sessionDropdownOpen && len(m.sessions) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, sessionRow, m.renderSessionDropdown(width))
	}

	// When closed, just return session row (no gap)
	return sessionRow
}

// renderSessionDropdown renders the dropdown list of available sessions
func (m model) renderSessionDropdown(width int) string {
	maxDropdownHeight := 8
	dropdownContent := make([]string, 0, min(len(m.sessions), maxDropdownHeight))

	start := 0
	end := len(m.sessions)

	// Scroll logic if too many sessions
	if len(m.sessions) > maxDropdownHeight {
		if m.sessionIndex >= maxDropdownHeight/2 {
			start = m.sessionIndex - maxDropdownHeight/2
			end = start + maxDropdownHeight
			if end > len(m.sessions) {
				end = len(m.sessions)
				start = end - maxDropdownHeight
			}
		} else {
			end = maxDropdownHeight
		}
	}

	for i := start; i < end; i++ {
		session := m.sessions[i]
		sessionText := fmt.Sprintf("%s (%s)", session.Name, session.Type)

		var sessionStyle lipgloss.Style
		if i == m.sessionIndex {
			// Use BgBase only
			sessionStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Background(BgBase).
				Bold(true).
				Padding(0, 1)
		} else {
			// Use BgBase only
			sessionStyle = lipgloss.NewStyle().
				Foreground(FgSecondary).
				Background(BgBase).
				Padding(0, 1)
		}

		dropdownContent = append(dropdownContent, sessionStyle.Render(sessionText))
	}

	// Add scroll indicators if needed
	if start > 0 {
		scrollUp := lipgloss.NewStyle().Foreground(FgMuted).Render("  ↑ more above")
		dropdownContent = append([]string{scrollUp}, dropdownContent...)
	}
	if end < len(m.sessions) {
		scrollDown := lipgloss.NewStyle().Foreground(FgMuted).Render("  ↓ more below")
		dropdownContent = append(dropdownContent, scrollDown)
	}

	dropdown := lipgloss.JoinVertical(lipgloss.Left, dropdownContent...)

	// Add border to dropdown
	// Remove Width() to prevent background bleeding
	// Use BgBase explicitly
	// Add left margin to align with session text
	dropdownStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderFocus).
		Background(BgBase).
		Padding(0, 1).
		MarginLeft(11) // "Session:" label is 10 chars wide + 1 space

	return dropdownStyle.Render(dropdown)
}

// renderMainHelp renders the help text at the bottom of the screen
func (m model) renderMainHelp() string {
	switch m.mode {
	case ModeLogin, ModePassword:
		// Reorder function keys to F1-F4 logical sequence
		// Add Page Up/Down navigation for ASCII variants
		if m.sessionDropdownOpen {
			return "↑↓ Navigate • ⇞⇟ ASCII • Enter Select • Esc Close • Tab Focus • F1 Menu • F2 Sessions • F3 Notes • F4 Power"
		}
		return "Tab Focus • ↑↓ Sessions • ⇞⇟ ASCII • F1 Menu • F2 Sessions • F3 Notes • F4 Power • Enter Continue"
	case ModeLoading:
		return "Please wait..."
	default:
		return "Ctrl+C Quit"
	}
}
