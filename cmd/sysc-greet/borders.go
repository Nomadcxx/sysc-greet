package main

import (
	"image/color"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// borders.go - All border rendering and styling functions
// Extracted from main.go on 2025-10-11 for better code organization

func (m model) renderDualBorderLayout(termWidth, termHeight int) string {
	// Route ASCII-1, ASCII-2 styles
	if m.selectedBorderStyle == "ascii1" {
		return m.renderASCII1BorderLayout(termWidth, termHeight)
	}
	if m.selectedBorderStyle == "ascii2" {
		return m.renderASCII2BorderLayout(termWidth, termHeight)
	}
	// ===== INNER BORDER CONTENT =====
	// Contains: WM ASCII art, session dropdown, username/password fields

	// Calculate inner content area
	// Use reasonable max width like installer
	innerWidth := min(100, termWidth-8) // Reasonable width for content area
	var innerSections []string
	if ticker := m.renderTypewriterLine(innerWidth - 6); ticker != "" {
		innerSections = append(innerSections, ticker)
	}

	// WM/Session ASCII art - prominent display
	// Fix centering, art is already colored
	if m.selectedSession != nil && termHeight >= 26 {
		art := m.getSessionArt(m.selectedSession.Name)
		if art != "" {
			// JoinVertical(Center) handles centering, just add art
			// Art is already colored, JoinVertical(Center) will center each line
			innerSections = append(innerSections, art)
			// Remove automatic spacing and add consistent 2-line spacing
			// innerSections = append(innerSections, "") // spacing
		}
	}

	// Ensure exactly 2 lines of spacing after ASCII art
	if termHeight >= 30 {
		innerSections = append(innerSections, "", "")
	}

	// Main form (session selector, username, password) in bordered box
	// Wrap form in border for left alignment
	// Reverted Width() addition
	// Add fixed width to form content with Place
	formContentWidth := innerWidth - 20
	formContent := m.renderMainForm(formContentWidth)
	// CHANGED 2025-10-06 - Removed Place()
	fixedFormContent := formContent
	formBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderDefault).
		Background(BgBase).
		Padding(1, 2)

	formBorder := formBorderStyle.Render(fixedFormContent)
	innerSections = append(innerSections, formBorder)

	// Simplified border title as first line
	// Create inner border container with user-selected style
	innerBorderColor := m.getInnerBorderColor()
	titleLine := m.renderSessionsHeading(innerWidth)

	// Add title as first element
	contentWithTitle := []string{titleLine}

	// Add spacing
	contentWithTitle = append(contentWithTitle, "")

	contentWithTitle = append(contentWithTitle, innerSections...)

	// Revert to Center, ASCII already has explicit centering
	innerContent := lipgloss.JoinVertical(lipgloss.Center, contentWithTitle...)

	// Reverted: always use normal styling since fire is now inside
	innerPadV, innerPadH := 2, 3
	if termHeight < 30 {
		innerPadV, innerPadH = 1, 2
	}
	innerBorderStyle := lipgloss.NewStyle().
		Border(m.getInnerBorderStyle()).
		BorderForeground(innerBorderColor).
		Width(innerWidth).
		Background(BgBase).
		Padding(innerPadV, innerPadH)

	innerBox := innerBorderStyle.Render(innerContent)

	// Fire at bottom, so no outer border
	// CHANGED 2025-10-08 - Also remove outer border for ascii-rain
	// CHANGED 2025-10-08 - Also remove outer border for ticker
	// Add matrix to backgrounds that remove outer border
	// Add fireworks to backgrounds that remove outer border
	// Add aquarium to backgrounds that remove outer border
	if m.selectedBackground == "fire" || m.selectedBackground == "ascii-rain" || m.selectedBackground == "matrix" || m.selectedBackground == "ticker" || m.selectedBackground == "fire+rain" || m.selectedBackground == "fireworks" || m.selectedBackground == "sonar" || m.selectedBackground == "cracktro" || m.selectedBackground == "aquarium" || m.selectedBackground == "plasma" || m.selectedBackground == "ak-47" || m.selectedBackground == "l2a1-sar" || isRotatingModelBackground(m.selectedBackground) {
		helpText := m.renderMainHelpFit(termWidth - 2)
		helpStyle := lipgloss.NewStyle().
			Foreground(FgMuted).
			Align(lipgloss.Center)

		contentWithHelp := lipgloss.JoinVertical(lipgloss.Center, innerBox, "", helpStyle.Render(helpText))
		// Don't use Place here - View() will handle it
		return contentWithHelp
	}

	// ===== OUTER BORDER CONTENT =====
	// Contains: Inner border + help text at bottom

	// Calculate outer content area
	// CHANGED 2025-10-06 - Reduced margin to make outer border closer to edges
	outerWidth := termWidth - 8 // Small margin from terminal edges

	var outerSections []string

	// Removed bubble-greet title text
	// Title removed per user request

	// Time if enabled
	if m.config.ShowTime {
		// Manual centering without lipgloss Align
		timeStyle := lipgloss.NewStyle().
			Foreground(FgSecondary)
		currentTime := time.Now().Format("15:04:05 Mon Jan 02, 2006")
		centeredTime := centerText(currentTime, outerWidth-4)
		outerSections = append(outerSections, timeStyle.Render(centeredTime))
	}

	outerSections = append(outerSections, "") // spacing

	// CHANGED 2025-10-06 - Removed Place()
	// Just append innerBox without centering - View() will handle positioning
	outerSections = append(outerSections, innerBox)

	// Join all sections with Center to center each section horizontally
	outerContent := lipgloss.JoinVertical(lipgloss.Center, outerSections...)

	// Create outer border container with user-selected style
	outerBorderColor := m.getOuterBorderColor()
	// Remove Align to prevent bleeding with Background
	// Add BgBase back to both
	// CHANGED 2025-10-06 - Calculate both horizontal AND vertical padding to expand outer border to terminal edges
	innerBoxWidth := lipgloss.Width(innerBox)
	innerBoxHeight := lipgloss.Height(outerContent)

	// Outer padding: hug keeps fixed margins around the inner box, fill
	// grows the outer border toward the terminal edges.
	horizontalPadding, verticalPadding := 2, 1
	if m.config.BorderLayout == "fill" {
		horizontalPadding = (termWidth - innerBoxWidth - 10) / 2 // Leave 10 chars total margin (5 per side)
		if horizontalPadding < 1 {
			horizontalPadding = 1
		}
		verticalPadding = (termHeight - innerBoxHeight - 6) / 2 // Leave 6 lines total margin (3 per side) for help text
		if verticalPadding < 0 {
			verticalPadding = 0
		}
	}

	outerBorderStyle := lipgloss.NewStyle().
		Border(m.getOuterBorderStyle()). // User-selected outer border style
		BorderForeground(outerBorderColor).
		Background(BgBase).
		Padding(verticalPadding, horizontalPadding)

	outerBox := outerBorderStyle.Render(outerContent)

	// Move help text BELOW outer border
	// Help text at bottom, below outer border
	helpText := m.renderMainHelpFit(termWidth - 2)
	helpStyle := lipgloss.NewStyle().
		Foreground(FgMuted).
		Align(lipgloss.Center)

	// Join outer box and help text vertically
	contentWithHelp := lipgloss.JoinVertical(lipgloss.Center, outerBox, "", helpStyle.Render(helpText))

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering
	return contentWithHelp
}

func (m model) renderSessionsHeading(innerWidth int) string {
	width := max(1, innerWidth-6)
	style := lipgloss.NewStyle().Foreground(m.getInnerBorderColor()).Bold(true).Width(width).Align(lipgloss.Center)
	dashCount := max(4, min(30, (width-22)/2))
	if m.selectedBorderStyle == "minimal" {
		dashCount = 4
	}
	dashes := strings.Repeat("─", dashCount)
	slashes := strings.Repeat("/", 7)
	return style.Render(dashes + slashes + "SESSIONS" + slashes + dashes)
}

func (m model) renderTypewriterLine(width int) string {
	if m.selectedBackground != "ticker" || m.typewriterTicker == nil {
		return ""
	}
	width = max(1, width)
	return lipgloss.NewStyle().Foreground(m.getInnerBorderColor()).Bold(true).Width(width).Align(lipgloss.Center).
		Render(m.typewriterTicker.GetTypewriterText(width))
}

// ASCII-1: Just a border style, uses current theme colors
func (m model) renderASCII1BorderLayout(termWidth, termHeight int) string {
	// Custom ASCII art border using block characters
	asciiBorder := lipgloss.Border{
		Top:         "▀",
		Bottom:      "▄",
		Left:        "█",
		Right:       "█",
		TopLeft:     "█",
		TopRight:    "█",
		BottomLeft:  "█",
		BottomRight: "█",
	}

	// THE GOODS container style - uses theme colors
	goodsWidth := min(100, termWidth-4)
	if goodsWidth < 40 {
		goodsWidth = 40
	}
	goodsPadV, goodsPadH := 2, 4
	if termHeight < 30 {
		goodsPadV, goodsPadH = 1, 2
	}
	goodsContentWidth := goodsWidth - 2*goodsPadH

	goodsStyle := lipgloss.NewStyle().
		Border(asciiBorder).
		BorderForeground(BorderDefault). // Use theme border color
		Padding(goodsPadV, goodsPadH).
		Background(BgBase)

	// Build THE GOODS content
	var sections []string

	// WM/Session ASCII art - use theme colors
	if m.selectedSession != nil && termHeight >= 26 {
		art := m.getSessionASCII() // Use normal colored ASCII, not monochrome
		if art != "" {
			// Center ASCII art within border
			// Center the ASCII art within the available width
			artStyle := lipgloss.NewStyle().
				Width(goodsContentWidth).
				Align(lipgloss.Center)
			sections = append(sections, artStyle.Render(art))
			// Control spacing explicitly - remove old spacing and add exactly 2 lines
			// sections = append(sections, "") // Remove old spacing
		}
	}

	// Ensure exactly 2 lines of spacing after ASCII art
	if termHeight >= 30 {
		sections = append(sections, "", "")
	}

	// Session selector - use theme colors
	if len(m.sessions) > 0 && m.selectedSession != nil {
		sections = append(sections, lipgloss.PlaceHorizontal(goodsContentWidth, lipgloss.Left, m.renderSessionSelector(goodsContentWidth)))
		sections = append(sections, "")
	}

	// Username and password inputs with labels - CHANGED 2025-10-02 04:05 - Add labels for ASCII-1
	usernameLabel := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Render("Username:")
	usernameRow := lipgloss.JoinHorizontal(lipgloss.Left, usernameLabel, " ", m.usernameInput.View())
	sections = append(sections, usernameRow)
	sections = append(sections, "")

	passwordLabel := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Render("Password:")
	passwordRow := lipgloss.JoinHorizontal(lipgloss.Left, passwordLabel, " ", m.passwordInput.View())
	sections = append(sections, passwordRow)

	// CHANGED 2025-10-05 - Display error message below password field
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")). // Red color
			Bold(true)
		sections = append(sections, "")
		sections = append(sections, errorStyle.Render("✗ "+m.errorMessage))
	}

	if row := m.renderAmbientRow(goodsContentWidth); row != "" {
		sections = append(sections, "", row)
	}

	// Join THE GOODS
	goodsContent := strings.Join(sections, "\n")

	// Wrap THE GOODS in ASCII border
	borderedGoods := goodsStyle.Render(goodsContent)

	// Help text
	// CHANGED 2025-10-06 - Removed Width(termWidth)
	helpText := m.renderMainHelpFit(termWidth - 2)
	helpStyle := lipgloss.NewStyle().
		Foreground(FgMuted) // Theme muted color

	// Final layout
	finalContent := lipgloss.JoinVertical(
		lipgloss.Center,
		borderedGoods,
		"",
		helpStyle.Render(helpText),
	)

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering
	return finalContent
}

func (m model) getInnerBorderStyle() lipgloss.Border {
	switch m.selectedBorderStyle {
	case "classic":
		return lipgloss.RoundedBorder()
	case "modern":
		return lipgloss.ThickBorder()
	case "minimal":
		return lipgloss.Border{
			Top:    " ",
			Bottom: " ",
			Left:   " ",
			Right:  " ",
		} // Use single space border for truly minimal look
	default:
		return lipgloss.RoundedBorder() // Default
	}
}

// Get outer border style based on user selection
func (m model) getOuterBorderStyle() lipgloss.Border {
	switch m.selectedBorderStyle {
	case "classic":
		return lipgloss.DoubleBorder()
	case "modern":
		return lipgloss.HiddenBorder() // Hide outer border for clean modern look
	case "minimal":
		return lipgloss.HiddenBorder() // Hide outer for clean minimal look
	default:
		return lipgloss.DoubleBorder() // Default
	}
}

// Get inner border color
func (m model) getInnerBorderColor() color.Color {
	return Primary
}

// Get outer border color
func (m model) getOuterBorderColor() color.Color {
	return FgSubtle
}

func (m model) renderASCII2BorderLayout(termWidth, termHeight int) string {
	// Complete rewrite to match ASCII_TEMPLATE.png reference
	// Fancy gradient border matching the reference template with proper wide spacing

	// Calculate border dynamically based on ASCII art width
	// Build content section FIRST to determine required width
	var contentLines []string

	// Split ASCII art into lines to prevent border corruption
	// Enforce mandatory 2-line gap between ASCII and input fields
	// WM/Session ASCII art
	if m.selectedSession != nil && termHeight >= 26 {
		art := m.getSessionASCII()
		if art != "" {
			// Split multi-line ASCII art into separate lines
			artLines := strings.Split(art, "\n")
			for _, line := range artLines {
				contentLines = append(contentLines, line)
			}
			// MANDATORY 2-line gap after ASCII art
			if termHeight >= 30 {
				contentLines = append(contentLines, "")
				contentLines = append(contentLines, "")
			}
		}
	}

	// Session display
	if len(m.sessions) > 0 && m.selectedSession != nil {
		contentLines = append(contentLines, strings.Split(m.renderSessionSelector(termWidth-20), "\n")...)
		contentLines = append(contentLines, "")
	}

	// Username input
	usernameLabel := lipgloss.NewStyle().Foreground(Primary).Bold(true).Render("Username:")
	usernameRow := lipgloss.JoinHorizontal(lipgloss.Left, usernameLabel, " ", m.usernameInput.View())
	contentLines = append(contentLines, usernameRow)
	contentLines = append(contentLines, "")

	// Password input
	passwordLabel := lipgloss.NewStyle().Foreground(Primary).Bold(true).Render("Password:")
	passwordRow := lipgloss.JoinHorizontal(lipgloss.Left, passwordLabel, " ", m.passwordInput.View())
	contentLines = append(contentLines, passwordRow)
	if row := m.renderAmbientRow(min(72, termWidth-36)); row != "" {
		contentLines = append(contentLines, "", row)
	}

	// Calculate border width based on actual content
	// Find maximum content width
	maxContentWidth := 0
	for _, line := range contentLines {
		width := lipgloss.Width(line)
		if width > maxContentWidth {
			maxContentWidth = width
		}
	}

	// Set border width with padding, but cap at reasonable max
	innerPadding := 8
	borderWidth := maxContentWidth + (innerPadding * 2)
	if borderWidth < 80 {
		borderWidth = 80 // Minimum width
	}
	maxBorderWidth := min(120, termWidth-20)
	if maxBorderWidth < 20 {
		maxBorderWidth = 20
	}
	if borderWidth > maxBorderWidth {
		borderWidth = maxBorderWidth
	}

	// Now render borders and content
	var lines []string

	// Recreate top border matching template corners
	// Top decorations - stepped corner fade matching template
	// Line 1: Top edge with corner blocks
	topLine1 := "▄▄▄▄" + strings.Repeat("█", borderWidth-8) + "▄▄▄▄"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(topLine1))

	// Line 2: Corner step inward
	topLine2 := "▄▀▀" + strings.Repeat(" ", borderWidth-6) + "▀▀▄"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(topLine2))

	// Line 3: Inner corner fade
	topLine3 := "█▀" + strings.Repeat(" ", borderWidth-4) + "▀█"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(topLine3))

	// Broken border design - gradient only at top, clean middle, gradient at bottom
	// Top gradient fade: ▓ → ▒ → ░ (only 2 lines for shorter height)
	gradientChars := []string{"▓", "▒"}
	gradientColors := []color.Color{Secondary, Accent}

	for i, char := range gradientChars {
		gradLine := char + strings.Repeat(" ", borderWidth-2) + char
		lines = append(lines, lipgloss.NewStyle().Foreground(gradientColors[i]).Render(gradLine))
	}

	// Clean content area with NO side borders
	// Main content area - NO side borders, just centered content with empty space
	for _, contentLine := range contentLines {
		visibleWidth := lipgloss.Width(contentLine)
		leftPad := (borderWidth - visibleWidth) / 2
		if leftPad < 0 {
			leftPad = 0
		}

		centeredContent := strings.Repeat(" ", leftPad) + contentLine
		rightPad := borderWidth - lipgloss.Width(centeredContent)
		if rightPad > 0 {
			centeredContent += strings.Repeat(" ", rightPad)
		}

		// No side border characters, just content in space
		lines = append(lines, centeredContent)
	}

	// Bottom gradient fade (reverse): ▒ → ▓ (only 2 lines for shorter height)
	for i := len(gradientChars) - 1; i >= 0; i-- {
		gradLine := gradientChars[i] + strings.Repeat(" ", borderWidth-2) + gradientChars[i]
		lines = append(lines, lipgloss.NewStyle().Foreground(gradientColors[i]).Render(gradLine))
	}

	// Bottom decorations matching template
	// Bottom corner fade (mirroring top)
	bottomLine3 := "█▄" + strings.Repeat(" ", borderWidth-4) + "▄█"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(bottomLine3))

	bottomLine2 := "▀▄▄" + strings.Repeat(" ", borderWidth-6) + "▄▄▀"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(bottomLine2))

	bottomLine1 := "▀▀▀▀" + strings.Repeat("█", borderWidth-8) + "▀▀▀▀"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(bottomLine1))

	// Add help text below border
	// Build bordered content
	borderedContent := strings.Join(lines, "\n")

	// Add help text below border
	// CHANGED 2025-10-06 - Removed Width(termWidth)
	helpText := m.renderMainHelpFit(termWidth - 2)
	helpStyle := lipgloss.NewStyle().
		Foreground(FgMuted)

	// Join border and help text vertically
	contentWithHelp := lipgloss.JoinVertical(lipgloss.Center, borderedContent, "", helpStyle.Render(helpText))

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering
	return contentWithHelp
}
