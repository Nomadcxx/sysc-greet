package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// Theme Management - Extracted during Phase 6 refactoring
// This file contains theme application, wallpaper management, and animation color helpers

// applyTheme sets the color scheme for the entire application based on theme name
// CHANGED 2025-10-01 - Theme support with proper color palettes
// CHANGED 2025-10-11 - Added testMode parameter
func applyTheme(themeName string, testMode bool) {
	switch strings.ToLower(themeName) {
	case "gruvbox":
		// Gruvbox Dark theme
		// All backgrounds same to prevent bleed
		BgBase = lipgloss.Color("#282828")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#fe8019")
		Secondary = lipgloss.Color("#8ec07c")
		Accent = lipgloss.Color("#fabd2f")
		FgPrimary = lipgloss.Color("#ebdbb2")
		FgSecondary = lipgloss.Color("#d5c4a1")
		FgMuted = lipgloss.Color("#bdae93")

	case "material":
		// Material Dark theme
		BgBase = lipgloss.Color("#263238")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#80cbc4")
		Secondary = lipgloss.Color("#64b5f6")
		Accent = lipgloss.Color("#ffab40")
		FgPrimary = lipgloss.Color("#eceff1")
		FgSecondary = lipgloss.Color("#cfd8dc")
		FgMuted = lipgloss.Color("#90a4ae")

	case "nord":
		// Nord theme
		BgBase = lipgloss.Color("#2e3440")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#81a1c1")
		Secondary = lipgloss.Color("#88c0d0")
		Accent = lipgloss.Color("#8fbcbb")
		FgPrimary = lipgloss.Color("#eceff4")
		FgSecondary = lipgloss.Color("#e5e9f0")
		FgMuted = lipgloss.Color("#d8dee9")

	case "dracula":
		// Dracula theme
		// All backgrounds same to prevent bleed
		BgBase = lipgloss.Color("#282a36")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#bd93f9")
		Secondary = lipgloss.Color("#8be9fd")
		Accent = lipgloss.Color("#50fa7b")
		FgPrimary = lipgloss.Color("#f8f8f2")
		FgSecondary = lipgloss.Color("#f1f2f6")
		FgMuted = lipgloss.Color("#6272a4")

	case "catppuccin":
		// Catppuccin Mocha theme
		BgBase = lipgloss.Color("#1e1e2e")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#cba6f7")
		Secondary = lipgloss.Color("#89b4fa")
		Accent = lipgloss.Color("#a6e3a1")
		FgPrimary = lipgloss.Color("#cdd6f4")
		FgSecondary = lipgloss.Color("#bac2de")
		FgMuted = lipgloss.Color("#a6adc8")

	case "tokyo night":
		// Tokyo Night theme
		BgBase = lipgloss.Color("#1a1b26")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#7aa2f7")
		Secondary = lipgloss.Color("#bb9af7")
		Accent = lipgloss.Color("#9ece6a")
		FgPrimary = lipgloss.Color("#c0caf5")
		FgSecondary = lipgloss.Color("#a9b1d6")
		FgMuted = lipgloss.Color("#565f89")

	case "solarized":
		// Solarized Dark theme
		BgBase = lipgloss.Color("#002b36")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#268bd2")
		Secondary = lipgloss.Color("#2aa198")
		Accent = lipgloss.Color("#859900")
		FgPrimary = lipgloss.Color("#fdf6e3")
		FgSecondary = lipgloss.Color("#eee8d5")
		FgMuted = lipgloss.Color("#93a1a1")

	case "monochrome":
		// Monochrome theme (black/white/gray)
		BgBase = lipgloss.Color("#1a1a1a") // Dark background
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#ffffff")     // White primary
		Secondary = lipgloss.Color("#cccccc")   // Light gray
		Accent = lipgloss.Color("#888888")      // Medium gray
		FgPrimary = lipgloss.Color("#ffffff")   // White text
		FgSecondary = lipgloss.Color("#cccccc") // Light gray text
		FgMuted = lipgloss.Color("#666666")     // Dark gray muted

	case "transishardjob":
		// TransIsHardJob - Transgender flag colors theme
		BgBase = lipgloss.Color("#1a1a1a")      // Dark background
		BgElevated = BgBase                     // Elevated surface
		BgSubtle = BgBase                       // Subtle background
		Primary = lipgloss.Color("#5BCEFA")     // Trans flag light blue
		Secondary = lipgloss.Color("#F5A9B8")   // Trans flag pink
		Accent = lipgloss.Color("#FFFFFF")      // Trans flag white
		FgPrimary = lipgloss.Color("#FFFFFF")   // White text
		FgSecondary = lipgloss.Color("#F5A9B8") // Pink text
		FgMuted = lipgloss.Color("#5BCEFA")     // Light blue muted

	case "rama":
		// RAMA theme - Inspired by RAMA keyboard aesthetics
		BgBase = lipgloss.Color("#2b2d42")      // Space cadet
		BgElevated = BgBase                     // Keep consistent
		BgSubtle = BgBase                       // Keep consistent
		Primary = lipgloss.Color("#ef233c")     // Red Pantone
		Secondary = lipgloss.Color("#d90429")   // Fire engine red
		Accent = lipgloss.Color("#8d99ae")      // Cool gray
		FgPrimary = lipgloss.Color("#edf2f4")   // Anti-flash white
		FgSecondary = lipgloss.Color("#8d99ae") // Cool gray
		FgMuted = lipgloss.Color("#8d99ae")     // Cool gray

	case "dark":
		// DARK theme - True black and true white minimalism
		BgBase = lipgloss.Color("#000000")      // True black
		BgElevated = BgBase                     // Keep pure black
		BgSubtle = BgBase                       // Keep pure black
		Primary = lipgloss.Color("#ffffff")     // True white
		Secondary = lipgloss.Color("#ffffff")   // True white
		Accent = lipgloss.Color("#808080")      // Mid gray accent
		FgPrimary = lipgloss.Color("#ffffff")   // True white
		FgSecondary = lipgloss.Color("#cccccc") // Light gray
		FgMuted = lipgloss.Color("#666666")     // Dark gray

	default: // "default"
		// Original Crush-inspired theme
		BgBase = lipgloss.Color("#1a1a1a")
		BgElevated = BgBase
		BgSubtle = BgBase
		Primary = lipgloss.Color("#8b5cf6")
		Secondary = lipgloss.Color("#06b6d4")
		Accent = lipgloss.Color("#10b981")
		FgPrimary = lipgloss.Color("#f8fafc")
		FgSecondary = lipgloss.Color("#cbd5e1")
		FgMuted = lipgloss.Color("#94a3b8")
	}

	// Update border colors based on new primary
	BorderFocus = Primary

	// CHANGED 2025-10-10 - Set theme-aware wallpaper via swww
	setThemeWallpaper(themeName, testMode)
}

// setThemeWallpaper sets a theme-specific wallpaper using swww daemon
// CHANGED 2025-10-10 - Set wallpaper for current theme using swww
func setThemeWallpaper(themeName string, testMode bool) {
	// CHANGED 2025-10-11 - Never run swww in test mode to avoid disrupting user's wallpapers during debugging
	if testMode {
		return
	}

	// Check if swww is available
	if _, err := exec.LookPath("swww"); err != nil {
		// swww not installed, skip silently
		return
	}

	// Normalize theme name for filename
	themeFile := strings.ToLower(strings.ReplaceAll(themeName, " ", "-"))
	wallpaperPath := fmt.Sprintf("/usr/share/sysc-greet/wallpapers/sysc-greet-%s.png", themeFile)

	// Check if wallpaper exists
	if _, err := os.Stat(wallpaperPath); err != nil {
		return
	}

	// Set wallpaper on all outputs using swww
	// Use goroutine to avoid blocking the UI
	go func() {
		// First ensure swww-daemon is running (redirect all output to avoid UI pollution)
		daemonCmd := exec.Command("swww-daemon")
		daemonCmd.Stdout = nil
		daemonCmd.Stderr = nil
		_ = daemonCmd.Start() // Ignore error - daemon may already be running

		// Give daemon a moment to start if it wasn't running
		time.Sleep(100 * time.Millisecond)

		// Set wallpaper on all monitors (redirect all output)
		cmd := exec.Command("swww", "img", wallpaperPath, "--transition-type", "fade", "--transition-duration", "0.5")
		cmd.Stdout = nil
		cmd.Stderr = nil
		_ = cmd.Run()
	}()
}

// getAnimatedColor cycles through primary brand colors for animations
func (m model) getAnimatedColor() color.Color {
	colors := []color.Color{Primary, Secondary, Accent}
	index := (m.animationFrame / 20) % len(colors)
	return colors[index]
}

// getAnimatedBorderColor cycles through border colors for animated borders
func (m model) getAnimatedBorderColor() color.Color {
	colors := []color.Color{BorderDefault, Primary, Secondary}
	index := (m.borderFrame / 5) % len(colors)
	return colors[index]
}

// getFocusColor returns the appropriate color based on focus state
func (m model) getFocusColor(target FocusState) color.Color {
	if m.focusState == target {
		return Primary
	}
	return FgSecondary
}
