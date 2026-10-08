package main

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

// Created menu.go for modular menu handling

func (m model) navigateToMainMenu() model {
	m.mode = ModeMenu
	m.menuIndex = 0
	m.menuOptions = []string{
		"Close Menu",
		"Themes",
		"Borders",
		"Backgrounds",
		"ASCII Effects",
		"Wallpaper",
	}
	return m
}

// navigateToThemesSubmenu switches to the themes submenu
func (m model) navigateToThemesSubmenu() (tea.Model, tea.Cmd) {
	m.menuOptions = []string{"← Back"}
	mark := "[ ] "
	if m.followShell {
		mark = "[✓] "
	}
	m.menuOptions = append(m.menuOptions, mark+"Follow sysc-shell")
	for _, theme := range m.availableThemes {
		m.menuOptions = append(m.menuOptions, "Theme: "+theme)
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
	rainEnabled := m.selectedBackground == "ascii-rain"
	matrixEnabled := m.selectedBackground == "matrix"
	fireworksEnabled := m.selectedBackground == "fireworks"
	aquariumEnabled := m.selectedBackground == "aquarium"
	sonarEnabled := m.selectedBackground == "sonar"
	cracktroEnabled := m.selectedBackground == "cracktro"
	plasmaEnabled := m.selectedBackground == "plasma"

	m.menuOptions = []string{
		"← Back",
		formatCheckbox("Fire", fireEnabled),
		formatCheckbox("ASCII Rain", rainEnabled),
		formatCheckbox("Matrix", matrixEnabled),
		formatCheckbox("Fireworks", fireworksEnabled),
		formatCheckbox("Aquarium", aquariumEnabled),
		formatCheckbox("Sonar", sonarEnabled),
		formatCheckbox("Cracktro", cracktroEnabled),
		formatCheckbox("Plasma", plasmaEnabled),
		formatSpeedSelector(m.animSpeed),
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

// formatSpeedSelector returns the speed selector row showing current speed
func formatSpeedSelector(speed string) string {
	label := "Normal"
	switch speed {
	case "slow":
		label = "Slow"
	case "fast":
		label = "Fast"
	}
	return "Speed: < " + label + " >"
}

// Removed navigateToVideoWallpapersSubmenu

// navigateToASCIIEffectsSubmenu switches to the ASCII effects submenu
func (m model) navigateToASCIIEffectsSubmenu() (tea.Model, tea.Cmd) {
	// Check which effects are enabled
	typewriterEnabled := m.selectedBackground == "ticker"
	printEnabled := m.selectedBackground == "print"
	beamsEnabled := m.selectedBackground == "beams"
	pourEnabled := m.selectedBackground == "pour"

	m.menuOptions = []string{
		"← Back",
		formatCheckbox("Typewriter", typewriterEnabled),
		formatCheckbox("Print", printEnabled),
		formatCheckbox("Beams", beamsEnabled),
		formatCheckbox("Pour", pourEnabled),
	}
	m.mode = ModeASCIIEffectsSubmenu
	m.menuIndex = 0
	return m, nil
}
