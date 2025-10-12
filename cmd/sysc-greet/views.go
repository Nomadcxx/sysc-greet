package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// View Rendering Functions - Extracted during Phase 4 refactoring
// This file contains top-level view rendering for different modes (power, menu, release notes)

// renderPowerView renders the power options menu (reboot/shutdown/cancel)
func (m model) renderPowerView(termWidth, termHeight int) string {
	var content []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Danger).
		Align(lipgloss.Center)
	content = append(content, titleStyle.Render("Power Options"))
	content = append(content, "")

	// Power options
	for i, option := range m.powerOptions {
		var style lipgloss.Style
		if i == m.powerIndex {
			// Use BgBase only
			style = lipgloss.NewStyle().
				Bold(true).
				Foreground(Danger).
				Background(BgBase).
				Padding(0, 2).
				Align(lipgloss.Center)
		} else {
			style = lipgloss.NewStyle().
				Foreground(FgSecondary).
				Background(BgBase).
				Padding(0, 2).
				Align(lipgloss.Center)
		}
		content = append(content, style.Render(option))
	}

	// Help
	content = append(content, "")
	helpStyle := lipgloss.NewStyle().Foreground(FgMuted).Align(lipgloss.Center)
	content = append(content, helpStyle.Render("↑↓ Navigate • Enter Select • Esc Cancel"))

	innerContent := lipgloss.JoinVertical(lipgloss.Center, content...)

	// Create bordered power menu
	// Use BgBase explicitly
	powerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Danger).
		Background(BgBase).
		Padding(2, 4)

	powermenu := powerStyle.Render(innerContent)

	// CHANGED 2025-10-06 - Return power menu without Place(), let View() handle centering
	return powermenu
}

// renderMenuView renders the main menu and all submenus (themes, borders, backgrounds, wallpaper)
// CHANGED 2025-09-30 - Added menu view rendering with CRUSH-style ASCII framing
func (m model) renderMenuView(termWidth, termHeight int) string {
	var content []string

	// Dynamic menu titles based on mode
	// Select appropriate title based on current menu mode
	var title string
	switch m.mode {
	case ModeMenu:
		title = "///// Menu //////"
	case ModeThemesSubmenu:
		title = "///// Themes /////"
	case ModeBordersSubmenu:
		title = "///// Borders ////"
	case ModeBackgroundsSubmenu:
		title = "/// Backgrounds ///" // Add backgrounds title
	case ModeWallpaperSubmenu:
		title = "/// Wallpapers ////" // Add wallpapers title
	default:
		title = "///// Menu //////"
	}

	// Will calculate title width after rendering menu items
	content = append(content, "") // Placeholder for title
	content = append(content, "")

	// CHANGED 2025-10-04 - Add pagination for long menus
	maxVisibleItems := 9
	totalItems := len(m.menuOptions)

	// Calculate visible range
	startIdx := 0
	endIdx := totalItems

	if totalItems > maxVisibleItems {
		// Center the selection in the visible window
		startIdx = m.menuIndex - maxVisibleItems/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxVisibleItems
		if endIdx > totalItems {
			endIdx = totalItems
			startIdx = endIdx - maxVisibleItems
			if startIdx < 0 {
				startIdx = 0
			}
		}

		// Show scroll indicator at top if not at beginning
		if startIdx > 0 {
			indicatorStyle := lipgloss.NewStyle().Foreground(FgMuted).Align(lipgloss.Center)
			content = append(content, indicatorStyle.Render("▲ More above ▲"))
		}
	}

	// Menu options (visible window)
	for i := startIdx; i < endIdx; i++ {
		option := m.menuOptions[i]
		// Widened menu to 32
		var style lipgloss.Style
		if i == m.menuIndex {
			// Use BgBase only
			// CHANGED 2025-10-06 - Removed Align()
			style = lipgloss.NewStyle().
				Bold(true).
				Foreground(Accent).
				Background(BgBase).
				Padding(0, 2)
		} else {
			// CHANGED 2025-10-06 - Removed Align()
			style = lipgloss.NewStyle().
				Foreground(FgSecondary).
				Background(BgBase).
				Padding(0, 2)
		}
		content = append(content, style.Render(option))
	}

	// Show scroll indicator at bottom if not at end
	if totalItems > maxVisibleItems && endIdx < totalItems {
		// CHANGED 2025-10-06 - Removed Align(Center)
		indicatorStyle := lipgloss.NewStyle().Foreground(FgMuted)
		content = append(content, indicatorStyle.Render("▼ More below ▼"))
	}

	// Help
	content = append(content, "")
	// CHANGED 2025-10-06 - Removed Align(Center)
	helpStyle := lipgloss.NewStyle().Foreground(FgMuted)
	content = append(content, helpStyle.Render("↑↓ Navigate • Enter Select • Esc Close"))

	// CHANGED 2025-10-06 - Calculate title width from widest rendered content line
	maxWidth := 0
	for i, line := range content {
		if i == 0 || i == 1 { // Skip placeholder title and empty line
			continue
		}
		lineWidth := lipgloss.Width(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	// Render title with calculated width
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent).
		Width(maxWidth).
		Align(lipgloss.Center)
	content[0] = titleStyle.Render(title) // Replace placeholder

	// CHANGED 2025-10-06 - Use Left instead of Center
	innerContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Create bordered menu with ASCII-style framing
	// Use BgBase explicitly
	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Accent).
		Background(BgBase).
		Padding(2, 4)

	menu := menuStyle.Render(innerContent)

	// CHANGED 2025-10-06 - Return menu without Place(), let View() handle centering
	return menu
}

// renderReleaseNotesView renders the F3 release notes popup
// Added F5 release notes view rendering function
// Updated with NOTES_POPUP.txt format
func (m model) renderReleaseNotesView(termWidth, termHeight int) string {
	// Rewrite to match NOTES popup format

	// NOTES ASCII header (from NOTES_POPUP.txt template)
	notesHeader := `
                   ▀▀▀▀▀▀█▄ ▄█▀▀▀▀██ ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀ ▄█▀▀▀▀▀▀
                   █▓    █▓ █▒    █▒    █▓   ▓█▀▀▀▀  ▀▀▀▀▀█▓
▀▀ ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀    ▀▀▀▀▀▀▀▀▀▀     █▒   ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀ ▀▀
                                        ▀`

	separator := "· ─ ─·-─────────────────────────────────────────────────────────────────-·─ ─ ·"

	features := []string{
		"Features:",
		"  • 9 Themes: Dracula, Gruvbox, Material, Nord, Tokyo Night, Catppuccin, Solarized, Monochrome, TransIsHardJob",
		"  • Background Animations: Fire (DOOM-style), Rain, Matrix",
		"  • 7 Border Styles: Classic, Modern, Minimal, ASCII-1, ASCII-2, Wave, Pulse",
		"  • Preference Caching: Remembers theme, background, border, and session",
		"  • Key Bindings: F2 Menu • F3 Sessions • F4 Power • F5 Release Notes",
		"",
	}

	thanks := []string{
		"Thanks:",
		"  • Original tuigreet by apognu",
		"  • Bubble Tea framework by Charm",
		"  • Lipgloss styling library by Charm",
		"  • Go community",
	}

	// Define width first, then build content
	popupWidth := min(100, termWidth-8)

	// Build content
	var contentLines []string

	// Hardcode ASCII center position, left-align all text
	// Add header with FIXED manual centering - position ASCII block in center of contentWidth
	headerLines := strings.Split(notesHeader, "\n")

	// Hardcoded left padding to center the ASCII block (contentWidth is ~94, ASCII is ~85 chars, so pad by ~5)
	asciiLeftPad := 5

	// Add each line with fixed left padding
	for _, line := range headerLines {
		trimmed := strings.TrimRight(line, " ")
		if trimmed != "" {
			paddedLine := strings.Repeat(" ", asciiLeftPad) + trimmed
			contentLines = append(contentLines, lipgloss.NewStyle().Foreground(Primary).Render(paddedLine))
		}
	}

	// Separator (left-aligned, no centering)
	contentLines = append(contentLines, lipgloss.NewStyle().Foreground(FgMuted).Render(separator))

	// Add features (left-aligned)
	for _, line := range features {
		if strings.HasPrefix(line, "Features:") {
			contentLines = append(contentLines, lipgloss.NewStyle().Bold(true).Foreground(Accent).Render(line))
		} else {
			contentLines = append(contentLines, lipgloss.NewStyle().Foreground(FgPrimary).Render(line))
		}
	}

	// Add thanks (left-aligned)
	for _, line := range thanks {
		if strings.HasPrefix(line, "Thanks:") {
			contentLines = append(contentLines, lipgloss.NewStyle().Bold(true).Foreground(Accent).Render(line))
		} else {
			contentLines = append(contentLines, lipgloss.NewStyle().Foreground(FgPrimary).Render(line))
		}
	}

	contentLines = append(contentLines, "")

	// Bottom separator (left-aligned)
	contentLines = append(contentLines, lipgloss.NewStyle().Foreground(FgMuted).Render(separator))

	// Join all content
	innerContent := strings.Join(contentLines, "\n")

	// Remove global Align, center individual elements instead
	// Create bordered box (matching Menu/Power popup style)

	notesStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Background(BgBase).
		Padding(2, 3).
		Width(popupWidth)

	notesBox := notesStyle.Render(innerContent)

	// CHANGED 2025-10-06 - Return notes without Place(), let View() handle centering
	return notesBox
}
