package main

import (
	"regexp"
	"strings"

	"github.com/Nomadcxx/sysc-greet/internal/ui"
)

// Utility Functions - Extracted during Phase 7 refactoring
// This file contains general-purpose utility functions for string manipulation and calculations

// ANSI regex for stripping ANSI escape codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape codes from a string for width calculation
// CHANGED 2025-09-29 - Helper function to strip ANSI codes for width calculation
func stripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// stripAnsi removes ANSI escape codes using the internal ui package
// REFACTORED 2025-10-02 - Moved to internal/ui/utils.go
// This is a wrapper for backward compatibility
func stripAnsi(s string) string {
	return ui.StripAnsi(s)
}

// centerText centers text within a given width using the internal ui package
// REFACTORED 2025-10-02 - Moved to internal/ui/utils.go
// This is a wrapper for backward compatibility
func centerText(text string, width int) string {
	return ui.CenterText(text, width)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractCharsWithAnsi extracts characters from a line while preserving ANSI codes
// Character-by-character line merging
// Extract characters with their ANSI codes attached
func extractCharsWithAnsi(line string) []string {
	var chars []string
	var currentAnsi strings.Builder
	var i int

	for i < len(line) {
		// Check for ANSI escape sequence
		if i < len(line) && line[i] == '\x1b' {
			// Start of ANSI sequence
			start := i
			for i < len(line) && line[i] != 'm' {
				i++
			}
			if i < len(line) {
				i++ // include 'm'
			}
			// Store ANSI code
			currentAnsi.WriteString(line[start:i])
		} else if i < len(line) {
			// Regular character - attach accumulated ANSI and the char
			char := currentAnsi.String() + string(line[i])
			chars = append(chars, char)
			currentAnsi.Reset()
			i++
		}
	}

	return chars
}

// getThemeColorsForBeams returns color palette for beams effect based on theme
func getThemeColorsForBeams(themeName string) ([]string, []string) {
	var beamGradientStops []string
	var finalGradientStops []string

	// Normalize theme name to lowercase for comparison
	themeName = strings.ToLower(themeName)

	switch themeName {
	case "dracula":
		beamGradientStops = []string{"#ffffff", "#8be9fd", "#bd93f9"}
		finalGradientStops = []string{"#6272a4", "#bd93f9", "#f8f8f2"}
	case "gruvbox":
		beamGradientStops = []string{"#ffffff", "#fabd2f", "#fe8019"}
		finalGradientStops = []string{"#504945", "#fabd2f", "#ebdbb2"}
	case "nord":
		beamGradientStops = []string{"#ffffff", "#88c0d0", "#81a1c1"}
		finalGradientStops = []string{"#434c5e", "#88c0d0", "#eceff4"}
	case "tokyo-night":
		beamGradientStops = []string{"#ffffff", "#7dcfff", "#bb9af7"}
		finalGradientStops = []string{"#414868", "#7aa2f7", "#c0caf5"}
	case "catppuccin":
		beamGradientStops = []string{"#ffffff", "#89dceb", "#cba6f7"}
		finalGradientStops = []string{"#45475a", "#cba6f7", "#cdd6f4"}
	case "material":
		beamGradientStops = []string{"#ffffff", "#89ddff", "#bb86fc"}
		finalGradientStops = []string{"#546e7a", "#89ddff", "#eceff1"}
	case "solarized":
		beamGradientStops = []string{"#ffffff", "#2aa198", "#268bd2"}
		finalGradientStops = []string{"#586e75", "#2aa198", "#fdf6e3"}
	case "monochrome":
		beamGradientStops = []string{"#ffffff", "#c0c0c0", "#808080"}
		finalGradientStops = []string{"#3a3a3a", "#9a9a9a", "#ffffff"}
	case "transishardjob":
		beamGradientStops = []string{"#ffffff", "#55cdfc", "#f7a8b8"}
		finalGradientStops = []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	default:
		beamGradientStops = []string{"#ffffff", "#00D1FF", "#8A008A"}
		finalGradientStops = []string{"#4A4A4A", "#00D1FF", "#FFFFFF"}
	}

	return beamGradientStops, finalGradientStops
}

// getThemeColorsForPour returns color palette for pour effect based on theme
func getThemeColorsForPour(themeName string) []string {
	// Normalize theme name to lowercase for comparison
	themeName = strings.ToLower(themeName)

	switch themeName {
	case "dracula":
		return []string{"#ff79c6", "#bd93f9", "#ffffff"}
	case "gruvbox":
		return []string{"#fe8019", "#fabd2f", "#ffffff"}
	case "nord":
		return []string{"#88c0d0", "#81a1c1", "#ffffff"}
	case "tokyo-night":
		return []string{"#9ece6a", "#e0af68", "#ffffff"}
	case "catppuccin":
		return []string{"#cba6f7", "#f5c2e7", "#ffffff"}
	case "material":
		return []string{"#03dac6", "#bb86fc", "#ffffff"}
	case "solarized":
		return []string{"#268bd2", "#2aa198", "#ffffff"}
	case "monochrome":
		return []string{"#808080", "#c0c0c0", "#ffffff"}
	case "transishardjob":
		return []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	default:
		return []string{"#8A008A", "#00D1FF", "#FFFFFF"}
	}
}
