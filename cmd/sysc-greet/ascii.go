package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/charmbracelet/lipgloss/v2"
)

// ascii.go - ASCII art configuration, loading, rendering, and animations
// Extracted from main.go on 2025-10-11 for better code organization

func loadASCIIConfig(configPath string) (ASCIIConfig, error) {
	var config ASCIIConfig

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var currentVariantLines []string
	inASCII := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(trimmedLine, "#") || trimmedLine == "" {
			continue
		}

		if strings.Contains(trimmedLine, "=") {
			parts := strings.SplitN(trimmedLine, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Check if this is a new ASCII variant (ascii_1, ascii_2, etc.)
			if strings.HasPrefix(key, "ascii_") || key == "ascii" {
				// Save previous variant if exists
				if inASCII && len(currentVariantLines) > 0 {
					variant := strings.Join(currentVariantLines, "\n")
					// Trim trailing newlines for height consistency
					variant = strings.TrimRight(variant, "\n")
					config.ASCIIVariants = append(config.ASCIIVariants, variant)

					// Track max height
					variantLines := strings.Split(variant, "\n")
					height := len(variantLines)
					if height > config.MaxASCIIHeight {
						config.MaxASCIIHeight = height
					}
				}

				// Start new variant
				currentVariantLines = []string{}
				inASCII = true

				if value != "" && value != `"""` {
					currentVariantLines = append(currentVariantLines, value)
				}
			} else {
				// Save any pending ASCII variant before switching to other keys
				if inASCII && len(currentVariantLines) > 0 {
					variant := strings.Join(currentVariantLines, "\n")
					variant = strings.TrimRight(variant, "\n")
					config.ASCIIVariants = append(config.ASCIIVariants, variant)

					variantLines := strings.Split(variant, "\n")
					height := len(variantLines)
					if height > config.MaxASCIIHeight {
						config.MaxASCIIHeight = height
					}

					currentVariantLines = []string{}
					inASCII = false
				}

				// Handle other config keys
				switch key {
				case "name":
					config.Name = value
				case "colors":
					config.Colors = strings.Split(value, ",")
				case "animation_style":
					config.AnimationStyle = value
				case "animation_speed":
					if speed, err := strconv.ParseFloat(value, 64); err == nil {
						config.AnimationSpeed = speed
					}
				case "animation_direction":
					config.AnimationDirection = value
				case "roasts":
					config.Roasts = value
				}
			}
		} else if inASCII {
			if trimmedLine == `"""` {
				// End ASCII section and save variant
				if len(currentVariantLines) > 0 {
					variant := strings.Join(currentVariantLines, "\n")
					variant = strings.TrimRight(variant, "\n")
					config.ASCIIVariants = append(config.ASCIIVariants, variant)

					variantLines := strings.Split(variant, "\n")
					height := len(variantLines)
					if height > config.MaxASCIIHeight {
						config.MaxASCIIHeight = height
					}
				}
				currentVariantLines = []string{}
				inASCII = false
				continue
			}
			currentVariantLines = append(currentVariantLines, line)
		}
	}

	// Save final variant if exists
	if inASCII && len(currentVariantLines) > 0 {
		variant := strings.Join(currentVariantLines, "\n")
		variant = strings.TrimRight(variant, "\n")
		config.ASCIIVariants = append(config.ASCIIVariants, variant)

		variantLines := strings.Split(variant, "\n")
		height := len(variantLines)
		if height > config.MaxASCIIHeight {
			config.MaxASCIIHeight = height
		}
	}

	// Fallback: if no variants found, use old "ascii=" format
	if len(config.ASCIIVariants) == 0 && config.ASCII != "" {
		config.ASCIIVariants = []string{config.ASCII}
		config.MaxASCIIHeight = len(strings.Split(config.ASCII, "\n"))
	}

	// Set defaults for animation if not specified
	if config.AnimationStyle == "" {
		config.AnimationStyle = "gradient"
	}
	if config.AnimationSpeed == 0 {
		config.AnimationSpeed = 1.0
	}
	if config.AnimationDirection == "" {
		config.AnimationDirection = "right"
	}

	return config, nil
}

// Support multi-variant ASCII with cycling and height normalization
// Get ASCII art for current session
func (m model) getSessionASCII() string {
	if m.selectedSession == nil {
		return ""
	}

	// Extract and map session name to config file
	sessionName := strings.ToLower(strings.Fields(m.selectedSession.Name)[0])

	// Map session names to config file names - CHANGED 2025-10-02 04:45 - Add plasma->kde, xmonad mappings
	var configFileName string
	switch sessionName {
	case "gnome":
		configFileName = "gnome_desktop"
	case "i3":
		configFileName = "i3wm"
	case "bspwm":
		configFileName = "bspwm_manager"
	case "plasma":
		configFileName = "kde"
	case "xmonad":
		configFileName = "xmonad"
	default:
		configFileName = sessionName
	}

	// Try to load ASCII config for this session
	// Fixed path to use sysc-greet instead of bubble-greet
	configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
	asciiConfig, err := loadASCIIConfig(configPath)
	if err != nil {
		// Fallback to session name as text
		return sessionName
	}

	if len(asciiConfig.ASCIIVariants) == 0 {
		// Empty ASCII, return session name
		return sessionName
	}

	// Select current variant based on index
	variantIndex := m.asciiArtIndex
	if variantIndex >= len(asciiConfig.ASCIIVariants) {
		variantIndex = 0
	}
	if variantIndex < 0 {
		variantIndex = 0
	}

	currentASCII := asciiConfig.ASCIIVariants[variantIndex]
	
	// Use print effect if enabled
	if m.selectedBackground == "print" && m.printEffect != nil {
		// Get visible lines from print effect
		visibleLines := m.printEffect.GetVisibleLines()
		currentASCII = strings.Join(visibleLines, "\n")
		if m.config.Debug && len(visibleLines) > 0 {
			logDebug("Print effect rendering %d lines (complete: %v)", len(visibleLines), m.printEffect.IsComplete())
		}
	}
	
	// Use beams effect if enabled
	if m.selectedBackground == "beams" && m.beamsEffect != nil {
		// Beams effect renders with its own colors
		return m.beamsEffect.Render()
	}

	// Use pour effect if enabled
	if m.selectedBackground == "pour" && m.pourEffect != nil {
		// Pour effect renders with its own colors
		return m.pourEffect.Render()
	}

	// CHANGED 2025-10-11 - Removed height normalization padding to ensure consistent 2-line spacing
	// All WM ASCII art now maintains natural height for consistent distance to border elements

	// CHANGED 2025-10-18 19:09 - Apply styling to entire ASCII block instead of per-line to fix width calculation mangling - Problem: Per-line Render() caused JoinVertical(Center) to miscalculate widths
	// Apply static primary color to entire ASCII art block
	style := lipgloss.NewStyle().Foreground(Primary).Background(BgBase)
	return style.Render(currentASCII)
}

// Get color palette for a session type
// CHANGED 2025-09-29 - Added configurable palette support with fallback to defaults
// CHANGED 2025-10-01 - Enhanced animation system with multiple styles
func applyASCIIAnimation(text string, animationOffset float64, palette ColorPalette, config ASCIIConfig) string {
	// Apply animation speed multiplier
	adjustedOffset := animationOffset * config.AnimationSpeed

	switch config.AnimationStyle {
	case "wave":
		return applyWaveAnimation(text, adjustedOffset, palette, config.AnimationDirection)
	case "pulse":
		return applyPulseAnimation(text, adjustedOffset, palette)
	case "rainbow":
		return applyRainbowAnimation(text, adjustedOffset, palette, config.AnimationDirection)
	case "matrix":
		return applyMatrixAnimation(text, adjustedOffset, palette)
	case "typewriter":
		return applyTypewriterAnimation(text, adjustedOffset, palette)
	case "glow":
		return applyGlowAnimation(text, adjustedOffset, palette)
	case "static":
		return applyStaticColors(text, palette)
	case "gradient":
		fallthrough
	default:
		return applySmoothGradient(text, adjustedOffset, palette)
	}
}

// Apply rainbow colors with animation using custom palette (lolcat-inspired)
// CHANGED 2025-09-29 - Custom rainbow implementation with configurable palettes
// Replaced lolcat rainbow with smooth gradient
func applySmoothGradient(text string, animationOffset float64, palette ColorPalette) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	// Calculate max line width for consistent gradient
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	if maxWidth == 0 {
		return text
	}

	for lineIndex, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for charIndex, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			// Calculate smooth gradient position (0.0 to 1.0 across the width)
			gradientPos := float64(charIndex) / float64(maxWidth)

			// Add subtle vertical variation for depth
			verticalOffset := float64(lineIndex) * 0.05
			gradientPos += verticalOffset

			// Keep gradient position within bounds
			gradientPos = math.Mod(gradientPos, 1.0)
			if gradientPos < 0 {
				gradientPos += 1.0
			}

			// Interpolate between colors in palette for smooth gradient
			paletteLen := float64(len(palette.Colors))
			colorFloat := gradientPos * (paletteLen - 1)
			colorIndex1 := int(colorFloat)
			colorIndex2 := (colorIndex1 + 1) % len(palette.Colors)

			// Interpolation factor between the two colors
			factor := colorFloat - float64(colorIndex1)

			// Get the two colors to interpolate between
			color1 := palette.Colors[colorIndex1]
			color2 := palette.Colors[colorIndex2]

			// Interpolate RGB values
			interpolatedColor := lipgloss.Color(interpolateColors(color1, color2, factor))

			coloredChar := lipgloss.NewStyle().
				Foreground(interpolatedColor).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

// Added color interpolation for smooth gradients
func interpolateColors(color1, color2 string, factor float64) string {
	// Parse hex colors
	r1, g1, b1 := parseHexColor(color1)
	r2, g2, b2 := parseHexColor(color2)

	// Interpolate each component
	r := uint8(float64(r1)*(1-factor) + float64(r2)*factor)
	g := uint8(float64(g1)*(1-factor) + float64(g2)*factor)
	b := uint8(float64(b1)*(1-factor) + float64(b2)*factor)

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Added hex color parsing helper
func parseHexColor(hex string) (uint8, uint8, uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		// Default to white if invalid
		return 255, 255, 255
	}

	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	return uint8(r), uint8(g), uint8(b)
}

// CHANGED 2025-10-01 - Added sophisticated animation styles
func applyWaveAnimation(text string, animationOffset float64, palette ColorPalette, direction string) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	for lineIndex, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for charIndex, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			// Create wave effect based on direction
			var wavePos float64
			switch direction {
			case "left":
				wavePos = (float64(charIndex) + animationOffset) * 0.2
			case "up":
				wavePos = (float64(lineIndex) + animationOffset) * 0.3
			case "down":
				wavePos = (-float64(lineIndex) + animationOffset) * 0.3
			default: // "right"
				wavePos = (-float64(charIndex) + animationOffset) * 0.2
			}

			// Apply sine wave for smooth transitions
			waveValue := (math.Sin(wavePos) + 1.0) / 2.0 // 0.0 to 1.0

			colorIndex := int(waveValue * float64(len(palette.Colors)-1))
			colorStr := palette.Colors[colorIndex]

			coloredChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorStr)).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

func applyPulseAnimation(text string, animationOffset float64, palette ColorPalette) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	// Global pulse affects all characters
	pulseValue := (math.Sin(animationOffset*0.5) + 1.0) / 2.0 // 0.0 to 1.0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for _, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			// Pulse brightness affects color intensity
			colorIndex := int(pulseValue * float64(len(palette.Colors)-1))
			colorStr := palette.Colors[colorIndex]

			coloredChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorStr)).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

func applyRainbowAnimation(text string, animationOffset float64, palette ColorPalette, direction string) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	for lineIndex, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for charIndex, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			// Rainbow cycle with directional flow
			var rainbowPos float64
			switch direction {
			case "left":
				rainbowPos = float64(charIndex) + animationOffset
			case "up":
				rainbowPos = float64(lineIndex) + animationOffset
			case "down":
				rainbowPos = -float64(lineIndex) + animationOffset
			default: // "right"
				rainbowPos = -float64(charIndex) + animationOffset
			}

			// Cycle through all colors smoothly
			colorFloat := math.Mod(rainbowPos*0.1, float64(len(palette.Colors)))
			colorIndex := int(colorFloat)
			nextIndex := (colorIndex + 1) % len(palette.Colors)
			factor := colorFloat - float64(colorIndex)

			interpolatedColor := lipgloss.Color(interpolateColors(palette.Colors[colorIndex], palette.Colors[nextIndex], factor))

			coloredChar := lipgloss.NewStyle().
				Foreground(interpolatedColor).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

func applyMatrixAnimation(text string, animationOffset float64, palette ColorPalette) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	// Use green-dominant palette for matrix effect
	matrixPalette := []string{"#00ff00", "#00cc00", "#009900", "#006600"}
	if len(palette.Colors) > 0 {
		matrixPalette = palette.Colors
	}

	for lineIndex, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for charIndex, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			// Random-like effect based on position and time
			seed := float64(charIndex*lineIndex) + animationOffset*0.3
			randomValue := math.Mod(math.Sin(seed)*43758.5453, 1.0)
			if randomValue < 0 {
				randomValue = -randomValue
			}

			colorIndex := int(randomValue * float64(len(matrixPalette)))
			if colorIndex >= len(matrixPalette) {
				colorIndex = len(matrixPalette) - 1
			}

			color := matrixPalette[colorIndex]

			coloredChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(color)).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

func applyTypewriterAnimation(text string, animationOffset float64, palette ColorPalette) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	// Calculate which character should be "highlighted" as being typed
	totalChars := 0
	for _, line := range lines {
		totalChars += len(line)
	}

	currentCharPos := int(animationOffset*0.2) % totalChars
	charCounter := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for _, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				charCounter++
				continue
			}

			var color string
			if charCounter == currentCharPos {
				// Highlight current character being "typed"
				color = palette.Colors[len(palette.Colors)-1] // Use brightest color
			} else if charCounter < currentCharPos {
				// Already typed characters
				color = palette.Colors[0] // Use dimmest color
			} else {
				// Not yet typed
				color = "#333333" // Very dim
			}

			coloredChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(color)).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
			charCounter++
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

func applyGlowAnimation(text string, animationOffset float64, palette ColorPalette) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	// Create glow effect with center-out intensity
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	for lineIndex, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		centerX := float64(maxWidth) / 2.0
		centerY := float64(len(lines)) / 2.0

		for charIndex, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			// Calculate distance from center
			dx := float64(charIndex) - centerX
			dy := float64(lineIndex) - centerY
			distance := math.Sqrt(dx*dx + dy*dy)

			// Create pulsing glow
			glowValue := (math.Sin(animationOffset*0.4-distance*0.2) + 1.0) / 2.0

			colorIndex := int(glowValue * float64(len(palette.Colors)-1))
			colorStr := palette.Colors[colorIndex]

			coloredChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorStr)).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

func applyStaticColors(text string, palette ColorPalette) string {
	lines := strings.Split(text, "\n")
	var coloredLines []string

	// Use first color for static display
	if len(palette.Colors) > 1 {
		// Use middle color if available
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}

		var coloredLine strings.Builder
		for _, char := range line {
			if char == ' ' {
				coloredLine.WriteRune(char)
				continue
			}

			coloredChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff00")).
				Render(string(char))
			coloredLine.WriteString(coloredChar)
		}
		coloredLines = append(coloredLines, coloredLine.String())
	}

	return strings.Join(coloredLines, "\n")
}

// Load configuration from file
// CHANGED 2025-09-29 - Added config file parsing for fonts and palettes
func (m model) getSessionASCIIMonochrome() string {
	if m.selectedSession == nil {
		return ""
	}

	// Extract and map session name to config file
	sessionName := strings.ToLower(strings.Fields(m.selectedSession.Name)[0])

	// Map session names to config file names - CHANGED 2025-10-02 04:46 - Add plasma->kde, xmonad mappings
	var configFileName string
	switch sessionName {
	case "gnome":
		configFileName = "gnome_desktop"
	case "i3":
		configFileName = "i3wm"
	case "bspwm":
		configFileName = "bspwm_manager"
	case "plasma":
		configFileName = "kde"
	case "xmonad":
		configFileName = "xmonad"
	default:
		configFileName = sessionName
	}

	// Try to load ASCII config for this session
	// FIXED 2025-10-15 - Corrected path from bubble-greet to sysc-greet
	configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
	asciiConfig, err := loadASCIIConfig(configPath)
	if err != nil {
		// Fallback to session name as text
		return sessionName
	}

	// Always apply monochrome animation
	// Create monochrome palette
	monoPalette := ColorPalette{
		Name:   "monochrome",
		Colors: []string{"#444444", "#666666", "#888888", "#aaaaaa", "#cccccc", "#ffffff"},
	}

	// Use a subtle monochrome animation
	monoConfig := asciiConfig
	monoConfig.AnimationStyle = "gradient" // Simple gradient for monochrome
	monoConfig.AnimationSpeed = 0.5        // Slower for subtlety

	if asciiConfig.ASCII != "" {
		return applyASCIIAnimation(asciiConfig.ASCII, float64(m.animationFrame)*0.1, monoPalette, monoConfig)
	}

	return asciiConfig.ASCII
}

// Implement actual border style functionality

// Get inner border style based on user selection

func (m model) getSessionArt(sessionName string) string {
	// Use pre-made ASCII from config files
	return m.getSessionASCII()
}


// resetPrintEffectForSession resets the print effect with the specified session's ASCII
func (m *model) resetPrintEffectForSession(sessionName string) {
if m.selectedBackground != "print" || m.printEffect == nil {
return
}

// Map session names to config file names
sessionLower := strings.ToLower(strings.Fields(sessionName)[0])
var configFileName string
switch sessionLower {
case "gnome":
configFileName = "gnome_desktop"
case "i3":
configFileName = "i3wm"
case "bspwm":
configFileName = "bspwm_manager"
case "plasma":
configFileName = "kde"
case "xmonad":
configFileName = "xmonad"
default:
configFileName = sessionLower
}

configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
variantIndex := m.asciiArtIndex
if variantIndex >= len(asciiConfig.ASCIIVariants) {
variantIndex = 0
}
ascii := asciiConfig.ASCIIVariants[variantIndex]
m.printEffect.Reset(ascii)
}
}

// resetPourEffectForSession resets the pour effect with the specified session's ASCII
func (m *model) resetPourEffectForSession(sessionName string) {
	if m.selectedBackground != "pour" || m.pourEffect == nil {
		return
	}

	sessionLower := strings.ToLower(strings.Fields(sessionName)[0])
	var configFileName string
	switch sessionLower {
	case "gnome":
		configFileName = "gnome_desktop"
	case "i3":
		configFileName = "i3wm"
	case "bspwm":
		configFileName = "bspwm_manager"
	case "plasma":
		configFileName = "kde"
	case "xmonad":
		configFileName = "xmonad"
	default:
		configFileName = sessionLower
	}

	configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
	if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
		variantIndex := m.asciiArtIndex
		if variantIndex >= len(asciiConfig.ASCIIVariants) {
			variantIndex = 0
		}
		ascii := asciiConfig.ASCIIVariants[variantIndex]

		// Recalculate dimensions for the new ASCII
		lines := strings.Split(ascii, "\n")
		asciiHeight := len(lines)
		asciiWidth := 0
		for _, line := range lines {
			if len([]rune(line)) > asciiWidth {
				asciiWidth = len([]rune(line))
			}
		}

		// Get current theme colors
		pourColors := getThemeColorsForPour(m.currentTheme)

		// Reinitialize pour effect completely with new dimensions and colors
		m.pourEffect = animations.NewPourEffect(animations.PourConfig{
			Width:                  asciiWidth,
			Height:                 asciiHeight,
			Text:                   ascii,
			PourDirection:          "down",
			PourSpeed:              1,
			MovementSpeed:          0.05,
			Gap:                    2,
			StartingColor:          "#ffffff",
			FinalGradientStops:     pourColors,
			FinalGradientSteps:     12,
			FinalGradientFrames:    5,
			FinalGradientDirection: "horizontal",
		})
	}
}

// resetBeamsEffectForSession resets the beams effect with the specified session's ASCII
func (m *model) resetBeamsEffectForSession(sessionName string) {
if m.selectedBackground != "beams" || m.beamsEffect == nil {
return
}

sessionLower := strings.ToLower(strings.Fields(sessionName)[0])
var configFileName string
switch sessionLower {
case "gnome":
configFileName = "gnome_desktop"
case "i3":
configFileName = "i3wm"
case "bspwm":
configFileName = "bspwm_manager"
case "plasma":
configFileName = "kde"
case "xmonad":
configFileName = "xmonad"
default:
configFileName = sessionLower
}

configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
variantIndex := m.asciiArtIndex
if variantIndex >= len(asciiConfig.ASCIIVariants) {
variantIndex = 0
}
ascii := asciiConfig.ASCIIVariants[variantIndex]

// Recalculate dimensions for the new ASCII
lines := strings.Split(ascii, "\n")
asciiHeight := len(lines)
asciiWidth := 0
for _, line := range lines {
if len([]rune(line)) > asciiWidth {
asciiWidth = len([]rune(line))
}
}

// Get current theme colors
beamColors, finalColors := getThemeColorsForBeams(m.currentTheme)

// Reinitialize beams effect completely with new dimensions and colors
m.beamsEffect = animations.NewBeamsTextEffect(animations.BeamsTextConfig{
Width:               asciiWidth,
Height:              asciiHeight,
Text:                ascii,
BeamGradientStops:   beamColors,
FinalGradientStops:  finalColors,
})
}
}
