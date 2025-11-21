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
	case "rama":
		beamGradientStops = []string{"#edf2f4", "#ef233c", "#d90429"}
		finalGradientStops = []string{"#8d99ae", "#ef233c", "#edf2f4"}
	case "eldritch":
		beamGradientStops = []string{"#ebfafa", "#37f499", "#04d1f9"}
		finalGradientStops = []string{"#7081d0", "#a48cf2", "#ebfafa"}
	case "dark":
		beamGradientStops = []string{"#ffffff", "#cccccc", "#999999"}
		finalGradientStops = []string{"#666666", "#cccccc", "#ffffff"}
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
	case "rama":
		return []string{"#ef233c", "#d90429", "#edf2f4"}
	case "eldritch":
		return []string{"#37f499", "#04d1f9", "#ebfafa"}
	case "dark":
		return []string{"#ffffff", "#cccccc", "#999999"}
	default:
		return []string{"#8A008A", "#00D1FF", "#FFFFFF"}
	}
}

// getThemeColorsForAquarium returns color palette for aquarium effect based on theme
func getThemeColorsForAquarium(themeName string) (fishColors, waterColors, seaweedColors []string, bubbleColor, diverColor, boatColor, mermaidColor, anchorColor string) {
	// Normalize theme name to lowercase for comparison
	themeName = strings.ToLower(themeName)

	switch themeName {
	case "dracula":
		fishColors = []string{"#ff79c6", "#bd93f9", "#8be9fd", "#50fa7b", "#ffb86c"}
		waterColors = []string{"#6272a4", "#c2b280"}
		seaweedColors = []string{"#44475a", "#50fa7b", "#8be9fd"}
		bubbleColor = "#8be9fd"
		diverColor = "#f8f8f2"
		boatColor = "#ffb86c"
		mermaidColor = "#ff79c6"
		anchorColor = "#6272a4"
	case "gruvbox":
		fishColors = []string{"#fe8019", "#fabd2f", "#b8bb26", "#83a598", "#d3869b"}
		waterColors = []string{"#458588", "#d79921"}
		seaweedColors = []string{"#3c3836", "#98971a", "#b8bb26"}
		bubbleColor = "#83a598"
		diverColor = "#ebdbb2"
		boatColor = "#fabd2f"
		mermaidColor = "#d3869b"
		anchorColor = "#504945"
	case "nord":
		fishColors = []string{"#88c0d0", "#81a1c1", "#5e81ac", "#8fbcbb", "#b48ead"}
		waterColors = []string{"#5e81ac", "#d08770"}
		seaweedColors = []string{"#2e3440", "#a3be8c", "#8fbcbb"}
		bubbleColor = "#88c0d0"
		diverColor = "#eceff4"
		boatColor = "#d08770"
		mermaidColor = "#b48ead"
		anchorColor = "#4c566a"
	case "tokyo-night":
		fishColors = []string{"#7aa2f7", "#bb9af7", "#7dcfff", "#9ece6a", "#f7768e"}
		waterColors = []string{"#7aa2f7", "#e0af68"}
		seaweedColors = []string{"#1a1b26", "#9ece6a", "#7dcfff"}
		bubbleColor = "#7dcfff"
		diverColor = "#c0caf5"
		boatColor = "#e0af68"
		mermaidColor = "#bb9af7"
		anchorColor = "#414868"
	case "catppuccin":
		fishColors = []string{"#f5c2e7", "#cba6f7", "#89dceb", "#a6e3a1", "#fab387"}
		waterColors = []string{"#89b4fa", "#f9e2af"}
		seaweedColors = []string{"#1e1e2e", "#a6e3a1", "#94e2d5"}
		bubbleColor = "#89dceb"
		diverColor = "#cdd6f4"
		boatColor = "#fab387"
		mermaidColor = "#f5c2e7"
		anchorColor = "#45475a"
	case "material":
		fishColors = []string{"#82aaff", "#c792ea", "#89ddff", "#c3e88d", "#f78c6c"}
		waterColors = []string{"#82aaff", "#ffcb6b"}
		seaweedColors = []string{"#263238", "#c3e88d", "#89ddff"}
		bubbleColor = "#89ddff"
		diverColor = "#eceff1"
		boatColor = "#ffcb6b"
		mermaidColor = "#c792ea"
		anchorColor = "#37474f"
	case "solarized":
		fishColors = []string{"#268bd2", "#2aa198", "#859900", "#cb4b16", "#6c71c4"}
		waterColors = []string{"#268bd2", "#b58900"}
		seaweedColors = []string{"#002b36", "#859900", "#2aa198"}
		bubbleColor = "#2aa198"
		diverColor = "#fdf6e3"
		boatColor = "#cb4b16"
		mermaidColor = "#d33682"
		anchorColor = "#073642"
	case "monochrome":
		fishColors = []string{"#9a9a9a", "#bababa", "#dadada", "#c0c0c0", "#808080"}
		waterColors = []string{"#5a5a5a", "#8a8a8a"}
		seaweedColors = []string{"#1a1a1a", "#5a5a5a", "#7a7a7a"}
		bubbleColor = "#c0c0c0"
		diverColor = "#ffffff"
		boatColor = "#9a9a9a"
		mermaidColor = "#bababa"
		anchorColor = "#3a3a3a"
	case "transishardjob":
		fishColors = []string{"#55cdfc", "#f7a8b8", "#ffffff", "#f7a8b8", "#55cdfc"}
		waterColors = []string{"#55cdfc", "#f7a8b8"}
		seaweedColors = []string{"#1a1a1a", "#55cdfc", "#f7a8b8"}
		bubbleColor = "#ffffff"
		diverColor = "#ffffff"
		boatColor = "#f7a8b8"
		mermaidColor = "#f7a8b8"
		anchorColor = "#1a1a1a"
	case "rama":
		fishColors = []string{"#ef233c", "#d90429", "#8d99ae", "#edf2f4"}
		waterColors = []string{"#2b2d42", "#8d99ae"}
		seaweedColors = []string{"#2b2d42", "#8d99ae", "#ef233c"}
		bubbleColor = "#edf2f4"
		diverColor = "#edf2f4"
		boatColor = "#d90429"
		mermaidColor = "#ef233c"
		anchorColor = "#8d99ae"
	case "eldritch":
		fishColors = []string{"#37f499", "#04d1f9", "#a48cf2", "#f265b5", "#f7c67f"}
		waterColors = []string{"#7081d0", "#f7c67f"}
		seaweedColors = []string{"#44475a", "#50fa7b", "#10A1BD"}
		bubbleColor = "#10A1BD"
		diverColor = "#ebfafa"
		boatColor = "#f7c67f"
		mermaidColor = "#f265b5"
		anchorColor = "#3b4261"
	case "dark":
		fishColors = []string{"#ffffff", "#cccccc", "#999999", "#666666"}
		waterColors = []string{"#000000", "#333333"}
		seaweedColors = []string{"#000000", "#666666", "#999999"}
		bubbleColor = "#ffffff"
		diverColor = "#ffffff"
		boatColor = "#cccccc"
		mermaidColor = "#ffffff"
		anchorColor = "#666666"
	default:
		fishColors = []string{"#00D1FF", "#8A008A", "#FF00FF", "#00FFFF", "#FFFF00"}
		waterColors = []string{"#0066CC", "#FFD700"}
		seaweedColors = []string{"#1a1a1a", "#00FF00", "#00D1FF"}
		bubbleColor = "#00FFFF"
		diverColor = "#FFFFFF"
		boatColor = "#FFD700"
		mermaidColor = "#FF00FF"
		anchorColor = "#4A4A4A"
	}

	return
}
