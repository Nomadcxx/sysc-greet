package ui

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// StripAnsi removes ANSI escape sequences from a string
func StripAnsi(s string) string {
	// Match ANSI escape sequences: ESC[...m
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	// Strip ANSI codes for length calculation
	stripped := StripAnsi(text)
	textLen := utf8.RuneCountInString(stripped)

	if textLen >= width {
		return text
	}

	padding := (width - textLen) / 2
	return strings.Repeat(" ", padding) + text
}

// GetVisibleWidth returns the visible width of a string (excluding ANSI codes)
func GetVisibleWidth(s string) int {
	return utf8.RuneCountInString(StripAnsi(s))
}

// TruncateString truncates a string to a maximum width, preserving ANSI codes
func TruncateString(s string, maxWidth int) string {
	stripped := StripAnsi(s)
	if len(stripped) <= maxWidth {
		return s
	}

	// Simple truncation - could be improved to preserve ANSI codes better
	return s[:maxWidth] + "..."
}

// PadRight pads a string to the right with spaces
func PadRight(s string, width int) string {
	visibleWidth := GetVisibleWidth(s)
	if visibleWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visibleWidth)
}

// PadLeft pads a string to the left with spaces
func PadLeft(s string, width int) string {
	visibleWidth := GetVisibleWidth(s)
	if visibleWidth >= width {
		return s
	}
	return strings.Repeat(" ", width-visibleWidth) + s
}
