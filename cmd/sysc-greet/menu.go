package main

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

// Created menu.go for modular menu handling

// navigateToThemesSubmenu switches to the themes submenu
func (m model) navigateToThemesSubmenu() (tea.Model, tea.Cmd) {
	m.menuOptions = []string{
		"← Back",
		"Theme: Nord",
		"Theme: Gruvbox",
		"Theme: Material",
		"Theme: Dracula",
		"Theme: Catppuccin",
		"Theme: Tokyo Night",
		"Theme: Solarized",
		"Theme: Monochrome",
		"Theme: TransIsHardJob",
		"Theme: Default",
	}

	m.mode = ModeThemesSubmenu
	m.menuIndex = 0
	return m, nil
}

// navigateToBordersSubmenu switches to the borders submenu
func (m model) navigateToBordersSubmenu() (tea.Model, tea.Cmd) {
	m.menuOptions = []string{
		"← Back",
		"Style: Classic",
		"Style: Modern",
		"Style: Minimal",
		"Style: ASCII-1",
		"Style: ASCII-2",
		"Style: ASCII-3",
		"Style: ASCII-4",
		"Animation: Wave",
		"Animation: Pulse",
		"Animation: Off",
	}
	m.mode = ModeBordersSubmenu
	m.menuIndex = 0
	return m, nil
}

// navigateToBackgroundsSubmenu switches to the backgrounds submenu
// CHANGED 2025-10-04 - Show checkbox status for enabled backgrounds
func (m model) navigateToBackgroundsSubmenu() (tea.Model, tea.Cmd) {
	// Build menu with checkbox indicators
	fireEnabled := m.selectedBackground == "fire" || m.enableFire
	rainEnabled := m.selectedBackground == "ascii-rain" // CHANGED 2025-10-08 - Add ascii-rain option
	matrixEnabled := m.selectedBackground == "matrix"   // Add matrix option

	m.menuOptions = []string{
		"← Back",
		formatCheckbox("Fire", fireEnabled),
		formatCheckbox("ASCII Rain", rainEnabled), // CHANGED 2025-10-08 - Add ascii-rain option
		formatCheckbox("Matrix", matrixEnabled),   // Add matrix option
	}
	m.mode = ModeBackgroundsSubmenu
	m.menuIndex = 0
	return m, nil
}

// formatCheckbox returns a string with checkbox indicator
func formatCheckbox(label string, checked bool) string {
	if checked {
		return "[✓] " + label
	}
	return "[ ] " + label
}

// Removed navigateToVideoWallpapersSubmenu
