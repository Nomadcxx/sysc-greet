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
	m.menuOptions = []string{"← Back"}
	if m.borderRegistry != nil {
		for _, style := range m.borderRegistry.styles() {
			m.menuOptions = append(m.menuOptions, formatCheckbox("Style: "+style.Label(), m.selectedBorderStyle == style.ID()))
		}
	} else {
		m.menuOptions = append(m.menuOptions,
			formatCheckbox("Style: Classic", m.selectedBorderStyle == "classic"),
			formatCheckbox("Style: Modern", m.selectedBorderStyle == "modern"),
			formatCheckbox("Style: Minimal", m.selectedBorderStyle == "minimal"),
			formatCheckbox("Style: ASCII-1", m.selectedBorderStyle == "ascii1"),
			formatCheckbox("Style: ASCII-2", m.selectedBorderStyle == "ascii2"),
		)
	}
	style := borderRenderer(nil)
	if m.borderRegistry != nil {
		style = m.borderRegistry.get(m.selectedBorderStyle)
	}
	shimmer := style != nil && style.SupportsEffect("shimmer") && m.borderEffect == "shimmer"
	if style != nil && style.SupportsEffect("shimmer") {
		m.menuOptions = append(m.menuOptions, formatCheckbox("Effect: Shimmer", shimmer))
	}
	m.menuOptions = append(m.menuOptions,
		formatCheckbox("Effect: Off", !shimmer),
		formatCheckbox("Palette: Theme", m.borderPaletteMode == "theme" || m.borderPaletteMode == ""),
		formatCheckbox("Palette: VGA", m.borderPaletteMode == "vga"),
		formatCheckbox("Palette: Terminal", m.borderPaletteMode == "terminal"),
		formatCheckbox("Form: Auto", m.formMode == "auto" || m.formMode == ""),
		formatCheckbox("Form: Opaque", m.formMode == "opaque"),
		formatCheckbox("Form: See-through", m.formMode == "seethrough"),
		formatCheckbox("Form: Frosted", m.formMode == "frosted"),
	)
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
	ak47Enabled := m.selectedBackground == "ak-47"
	l2a1SAREnabled := m.selectedBackground == "l2a1-sar"

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
		formatCheckbox("AK-47", ak47Enabled),
		formatCheckbox("Australian SLR", l2a1SAREnabled),
	}
	for _, background := range rotatingModelBackgrounds {
		m.menuOptions = append(m.menuOptions, formatCheckbox(background.label, m.selectedBackground == background.id))
	}
	m.menuOptions = append(m.menuOptions, formatSpeedSelector(m.animSpeed))
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
