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
