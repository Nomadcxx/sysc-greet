package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/ipc"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	themesOld "github.com/Nomadcxx/sysc-greet/internal/themes"
	"github.com/Nomadcxx/sysc-greet/internal/ui"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mbndr/figlet4go"
)

// CHANGED 2025-10-06 - Add debug logging to file - Problem: Need persistent logs to debug greeter issues
var debugLog *log.Logger

func initDebugLog() {
	logFile, err := os.OpenFile("/tmp/sysc-greet-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to stderr if can't open log file
		debugLog = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
		return
	}
	debugLog = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func logDebug(format string, args ...interface{}) {
	if debugLog != nil {
		debugLog.Printf(format, args...)
	}
}

// CHANGED 2025-10-02 01:45 - TTY-safe colors with profile detection - Problem: Hex colors fail on TTY
var (
	// Detect color profile once at startup
	colorProfile = colorprofile.Detect(os.Stdout, os.Environ())
	complete     = lipgloss.Complete(colorProfile)

	// Backgrounds - using Complete() for TTY compatibility
	BgBase     color.Color
	BgElevated color.Color
	BgSubtle   color.Color
	BgActive   color.Color

	// Primary brand colors
	Primary   color.Color
	Secondary color.Color
	Accent    color.Color
	Warning   color.Color
	Danger    color.Color

	// Text colors
	FgPrimary   color.Color
	FgSecondary color.Color
	FgMuted     color.Color
	FgSubtle    color.Color

	// Border colors
	BorderDefault color.Color
	BorderFocus   color.Color
)

func init() {
	// CHANGED 2025-10-02 01:45 - Initialize colors with TTY fallbacks
	// Dark background - fallback to black on TTY
	BgBase = complete(
		lipgloss.Color("0"),       // ANSI black
		lipgloss.Color("235"),     // ANSI256 dark gray
		lipgloss.Color("#1a1a1a"), // TrueColor charcoal
	)
	BgElevated = BgBase
	BgSubtle = BgBase
	BgActive = BgBase

	// Primary violet - fallback to magenta on TTY
	Primary = complete(
		lipgloss.Color("5"),       // ANSI magenta
		lipgloss.Color("141"),     // ANSI256 purple
		lipgloss.Color("#8b5cf6"), // TrueColor violet
	)

	// Secondary cyan
	Secondary = complete(
		lipgloss.Color("6"),       // ANSI cyan
		lipgloss.Color("45"),      // ANSI256 cyan
		lipgloss.Color("#06b6d4"), // TrueColor cyan
	)

	// Accent green
	Accent = complete(
		lipgloss.Color("2"),       // ANSI green
		lipgloss.Color("42"),      // ANSI256 green
		lipgloss.Color("#10b981"), // TrueColor emerald
	)

	// Warning amber
	Warning = complete(
		lipgloss.Color("3"),       // ANSI yellow
		lipgloss.Color("214"),     // ANSI256 orange
		lipgloss.Color("#f59e0b"), // TrueColor amber
	)

	// Danger red
	Danger = complete(
		lipgloss.Color("1"),       // ANSI red
		lipgloss.Color("196"),     // ANSI256 red
		lipgloss.Color("#ef4444"), // TrueColor red
	)

	// Primary text - white
	FgPrimary = complete(
		lipgloss.Color("7"),       // ANSI white
		lipgloss.Color("255"),     // ANSI256 white
		lipgloss.Color("#f8fafc"), // TrueColor white
	)

	// Secondary text - light gray
	FgSecondary = complete(
		lipgloss.Color("7"),       // ANSI white
		lipgloss.Color("252"),     // ANSI256 light gray
		lipgloss.Color("#cbd5e1"), // TrueColor light gray
	)

	// Muted text - gray
	FgMuted = complete(
		lipgloss.Color("8"),       // ANSI bright black
		lipgloss.Color("244"),     // ANSI256 gray
		lipgloss.Color("#94a3b8"), // TrueColor gray
	)

	// Subtle text - dark gray
	FgSubtle = complete(
		lipgloss.Color("8"),       // ANSI bright black
		lipgloss.Color("240"),     // ANSI256 dark gray
		lipgloss.Color("#64748b"), // TrueColor dark gray
	)

	// Border default - dark gray
	BorderDefault = complete(
		lipgloss.Color("8"),       // ANSI bright black
		lipgloss.Color("238"),     // ANSI256 dark gray
		lipgloss.Color("#374151"), // TrueColor gray
	)

	BorderFocus = Primary
}

// CHANGED 2025-10-01 - Theme support with proper color palettes - Problem: User wants themes to actually work
func applyTheme(themeName string) {
	switch strings.ToLower(themeName) {
	case "gruvbox":
		// Gruvbox Dark theme
		// CHANGED 2025-10-01 21:55 - All backgrounds same to prevent bleed
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
		// CHANGED 2025-10-01 21:55 - All backgrounds same to prevent bleed - Problem: Different bg colors cause visible difference
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
		// CHANGED 2025-10-02 03:48 - Monochrome theme (black/white/gray)
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

	// CHANGED 2025-10-10 - Set theme-aware wallpaper via swww - Problem: Multi-monitor needs themed backgrounds
	setThemeWallpaper(themeName)
}

// CHANGED 2025-10-10 - Set wallpaper for current theme using swww - Problem: Need themed backgrounds on all monitors
func setThemeWallpaper(themeName string) {
	// Only set wallpaper in production mode (not test mode)
	if debugLog != nil {
		logDebug("Attempting to set wallpaper for theme: %s", themeName)
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
		if debugLog != nil {
			logDebug("Wallpaper not found: %s", wallpaperPath)
		}
		return
	}

	// Set wallpaper on all outputs using swww
	// Use goroutine to avoid blocking the UI
	go func() {
		// First ensure swww-daemon is running
		daemonCmd := exec.Command("swww-daemon")
		_ = daemonCmd.Start() // Ignore error - daemon may already be running

		// Give daemon a moment to start if it wasn't running
		time.Sleep(100 * time.Millisecond)

		// Set wallpaper on all monitors
		cmd := exec.Command("swww", "img", wallpaperPath, "--transition-type", "fade", "--transition-duration", "0.5")
		if err := cmd.Run(); err != nil {
			if debugLog != nil {
				logDebug("Failed to set wallpaper: %v", err)
			}
		} else if debugLog != nil {
			logDebug("Successfully set wallpaper: %s", wallpaperPath)
		}
	}()
}

// REFACTORED 2025-10-02 - Moved to internal/ui/utils.go
func centerText(text string, width int) string {
	return ui.CenterText(text, width)
}

// Color palette definitions for different WMs/sessions
// CHANGED 2025-09-29 - Added custom color palettes for different session types
type ColorPalette struct {
	Name   string
	Colors []string // Hex colors for the rainbow effect
}

// CHANGED 2025-10-02 05:30 - Fire effect implementation (PSX DOOM algorithm) - Problem: User wants fire background with theme color support

var sessionPalettes = map[string]ColorPalette{
	"GNOME": {
		Name:   "GNOME Blue",
		Colors: []string{"#4285f4", "#34a853", "#fbbc05", "#ea4335", "#9c27b0", "#ff9800"},
	},
	"KDE": {
		Name:   "KDE Plasma",
		Colors: []string{"#3daee9", "#1cdc9a", "#f67400", "#da4453", "#8e44ad", "#f39c12"},
	},
	"Hyprland": {
		Name:   "Hyprland Neon",
		Colors: []string{"#89b4fa", "#a6e3a1", "#f9e2af", "#fab387", "#f38ba8", "#cba6f7"},
	},
	"Sway": {
		Name:   "Sway Minimal",
		Colors: []string{"#458588", "#98971a", "#d79921", "#cc241d", "#b16286", "#689d6a"},
	},
	"i3": {
		Name:   "i3 Classic",
		Colors: []string{"#458588", "#98971a", "#d79921", "#cc241d", "#b16286", "#689d6a"},
	},
	"Xfce": {
		Name:   "Xfce Fresh",
		Colors: []string{"#4e9a06", "#f57900", "#cc0000", "#75507b", "#3465a4", "#c4a000"},
	},
	"default": {
		Name:   "Glamorous",
		Colors: []string{"#8b5cf6", "#06b6d4", "#10b981", "#f59e0b", "#ef4444", "#ec4899"},
	},
}

// CHANGED 2025-10-01 14:50 - Helper function to strip ANSI codes for width calculation - Problem: ASCII borders showing literal ANSI codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// ASCII art generator with proper Unicode block character support
// CHANGED 2025-09-29 - Fixed Unicode block character handling issue in figlet4go
// CHANGED 2025-09-30 15:21 - Removed old session art generation - Problem: Now using pre-made ASCII from config files

// CHANGED 2025-09-30 - Use real figlet binary instead of broken custom parser
// Fallback to figlet4go
func renderWithFiglet4goFallback(text, fontPath string, debug bool) (string, error) {
	ascii := figlet4go.NewAsciiRender()
	ascii.LoadFont(fontPath) // Ignore errors, use default if needed
	return ascii.Render(text)
}

// Parse figlet font file directly with proper Unicode support
// CHANGED 2025-09-29 - Core fix for Unicode block character rendering + encoding
// Parse figlet font file directly with proper Unicode support
// CHANGED 2025-09-30 15:18 - Added ASCII config loading system - Problem: User wants pre-made ASCII from .conf files instead of generated ones

// CHANGED 2025-10-01 - Enhanced ASCIIConfig with animation controls - Problem: User requested animation controls in .conf files and menu
// CHANGED 2025-10-07 19:00 - Added multi-ASCII variant support - Problem: User wants to cycle through multiple ASCII arts per session
type ASCIIConfig struct {
	Name               string
	ASCII              string   // DEPRECATED: Use ASCIIVariants instead
	ASCIIVariants      []string // Support multiple ASCII art variants (ascii_1, ascii_2, etc.)
	MaxASCIIHeight     int      // Track max height across all variants for normalization
	Colors             []string
	AnimationStyle     string  // "gradient", "wave", "pulse", "rainbow", "matrix", "typewriter", "glow", "static"
	AnimationSpeed     float64 // 0.1 (slow) to 2.0 (fast), default 1.0
	AnimationDirection string  // "left", "right", "up", "down", "center-out", "random"
}

// CHANGED 2025-10-07 19:00 - Parse multiple ASCII variants (ascii_1, ascii_2, etc.) - Problem: Support cycling through variants
// Load ASCII configuration from file
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
					// CHANGED 2025-10-07 19:00 - Trim trailing newlines for height consistency
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
				}
			}
		} else if inASCII {
			if trimmedLine == `"""` {
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

// CHANGED 2025-10-07 19:10 - Support multi-variant ASCII with cycling and height normalization - Problem: User wants Page Up/Down to cycle variants
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
	// CHANGED 2025-10-07 19:10 - Fixed path to use sysc-greet instead of bubble-greet - Problem: Wrong project name
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

	// CHANGED 2025-10-07 19:10 - Height normalization: pad smaller variants to match max height - Problem: Borders jump when cycling
	maxHeight := asciiConfig.MaxASCIIHeight
	currentHeight := len(strings.Split(currentASCII, "\n"))

	if maxHeight > currentHeight {
		paddingNeeded := maxHeight - currentHeight
		for i := 0; i < paddingNeeded; i++ {
			currentASCII += "\n"
		}
	}

	// CHANGED 2025-10-01 15:25 - Disable animations, use static primary color - Problem: User reports animations make ASCII look fucked up
	// Apply static primary color to ASCII art
	lines := strings.Split(currentASCII, "\n")
	var coloredLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			coloredLines = append(coloredLines, line)
			continue
		}
		// CHANGED 2025-10-02 00:05 - Add Background to prevent bleeding through - Problem: ASCII needs explicit background
		// CHANGED 2025-10-02 11:46 - Reverted: fire is inside now, always use background - Problem: Fire is inner content, not background
		style := lipgloss.NewStyle().Foreground(Primary).Background(BgBase)
		coloredLines = append(coloredLines, style.Render(line))
	}

	return strings.Join(coloredLines, "\n")
}

// Get color palette for a session type
// CHANGED 2025-09-29 - Added configurable palette support with fallback to defaults
// CHANGED 2025-10-01 - Enhanced animation system with multiple styles - Problem: User requested sophisticated animation controls
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
// CHANGED 2025-09-30 14:25 - Replaced lolcat rainbow with smooth gradient - Problem: User requested removal of "gay and lame" rainbow animation
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

// CHANGED 2025-09-30 14:25 - Added color interpolation for smooth gradients - Problem: Needed smooth color transitions instead of discrete color cycling
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

// CHANGED 2025-09-30 14:25 - Added hex color parsing helper - Problem: Needed to parse hex colors for interpolation
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

// CHANGED 2025-10-01 - Added sophisticated animation styles - Problem: User requested multiple animation types with directional control
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
func loadConfig(configPath string) (Config, error) {
	config := Config{
		FontPath: "/usr/share/bubble-greet/fonts/dos_rebel.flf", // CHANGED 2025-10-02 01:35 - Absolute path
		Palettes: make(map[string]ColorPalette),
	}

	file, err := os.Open(configPath)
	if err != nil {
		// If config file doesn't exist, use defaults
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key = value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "font_path" {
			config.FontPath = value
		} else {
			// Parse color palette: session_type = color1,color2,color3,...
			colors := strings.Split(value, ",")
			for i, color := range colors {
				colors[i] = strings.TrimSpace(color)
			}

			config.Palettes[key] = ColorPalette{
				Name:   fmt.Sprintf("%s theme", strings.Title(key)),
				Colors: colors,
			}
		}
	}

	return config, scanner.Err()
}

type Config struct {
	TestMode  bool
	Debug     bool
	Greeting  string
	ShowTime  bool
	ShowIssue bool
	Width     int
	ThemeName string
	// CHANGED 2025-09-29 - Added font and palette configuration
	FontPath string                  // Path to figlet font file
	Palettes map[string]ColorPalette // Custom color palettes per session type
}

type ViewMode string

const (
	ModeLogin    ViewMode = "login"
	ModePassword ViewMode = "password"
	ModeLoading  ViewMode = "loading"
	ModePower    ViewMode = "power"
	ModeMenu     ViewMode = "menu"
	// CHANGED 2025-09-30 14:40 - Added new menu modes for structured menu system - Problem: Need hierarchical menu with submenus
	ModeThemesSubmenu      ViewMode = "themes_submenu"
	ModeBordersSubmenu     ViewMode = "borders_submenu"
	ModeBackgroundsSubmenu ViewMode = "backgrounds_submenu"
	ModeWallpaperSubmenu ViewMode = "wallpaper_submenu" // CHANGED 2025-10-03 - Add wallpaper submenu for gslapper videos
	// CHANGED 2025-10-01 - Added release notes mode - Problem: User requested F5 release notes functionality
	ModeReleaseNotes ViewMode = "release_notes"
	// CHANGED 2025-10-10 - Added screensaver mode - Problem: Need screensaver with idle timeout
	ModeScreensaver ViewMode = "screensaver"
)

type FocusState int

const (
	FocusSession FocusState = iota
	FocusUsername
	FocusPassword
)

type model struct {
	usernameInput   textinput.Model
	passwordInput   textinput.Model
	spinner         spinner.Model
	sessions        []sessions.Session
	selectedSession *sessions.Session
	sessionIndex    int
	ipcClient       *ipc.Client
	theme           themesOld.Theme
	mode            ViewMode
	config          Config
	issueContent    string
	startTime       time.Time

	// Terminal dimensions
	width  int
	height int

	// Power menu
	powerOptions []string
	powerIndex   int

	// Session dropdown
	sessionDropdownOpen bool

	// Menu system
	menuOptions []string
	menuIndex   int
	// CHANGED 2025-09-30 15:10 - Added fields for functional menu system - Problem: Need to track all user preferences and apply them
	customASCIIText        string
	selectedBorderStyle    string
	selectedBackground     string
	currentTheme           string
	borderAnimationEnabled bool
	selectedFont           string
	// CHANGED 2025-10-01 - Added animation control fields - Problem: User requested animation controls in menu
	selectedAnimationStyle     string
	selectedAnimationSpeed     float64
	selectedAnimationDirection string
	animationStyleOptions      []string
	animationDirectionOptions  []string

	// Focus management
	focusState FocusState

	// Animation state
	animationFrame int
	pulseColor     int
	borderFrame    int

	// CHANGED 2025-10-02 05:35 - Fire effect instance - Problem: Need to maintain fire state across frames
	fireEffect     *animations.FireEffect
	lastFireWidth  int
	lastFireHeight int

	// CHANGED 2025-10-08 - Rain effect instance - Problem: Need to maintain rain state across frames
	rainEffect     *animations.RainEffect
	lastRainWidth  int
	lastRainHeight int

	// Matrix effect instance
	matrixEffect     *animations.MatrixEffect
	lastMatrixWidth  int
	lastMatrixHeight int

	// CHANGED 2025-10-04 - Separate flags for multiple backgrounds - Problem: User wants Fire + Rain/Matrix enabled simultaneously
	enableFire bool

	// CHANGED 2025-10-05 - Add error message for authentication failures - Problem: BUG #4 - Greeter exits on auth failure
	errorMessage string

	// CHANGED 2025-10-10 - Screensaver fields - Problem: Need screensaver mode with idle timeout
	idleTimer       time.Time // Time when idle started
	screensaverTime time.Time // Current time for screensaver display

	// CHANGED 2025-10-07 19:05 - ASCII navigation fields for multi-variant support - Problem: User wants Page Up/Down to cycle ASCII variants
	asciiArtIndex      int         // Current variant index (0-indexed)
	asciiArtCount      int         // Total variants available
	asciiMaxHeight     int         // Max height for normalization
	currentASCIIConfig ASCIIConfig // Cached config for current session
}

type sessionSelectedMsg sessions.Session
type powerSelectedMsg string
type tickMsg time.Time

func doTick() tea.Cmd {
	// CHANGED 2025-10-04 - Reduced tick interval to 30ms for smoother ticker animation - Problem: Ticker speed bottlenecked by UI refresh rate
	return tea.Tick(time.Millisecond*30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel(config Config) model {
	// Setup username input with proper styling
	ti := textinput.New()
	ti.Prompt = ""      // CHANGED 2025-10-02 04:02 - Remove prompt, will be added by layout - Problem: Duplicate "Username:" in ASCII-1
	ti.Placeholder = "" // CHANGED 2025-10-02 04:08 - Remove placeholder - Problem: Shows italic "E" in empty field
	// CHANGED 2025-10-01 23:00 - Updated for textinput v2 API - Problem: v2 uses Focused/Blurred StyleState
	ti.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	ti.Styles.Focused.Text = lipgloss.NewStyle().Foreground(FgPrimary)
	ti.Styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(FgMuted).Italic(true)

	// Setup password input
	pi := textinput.New()
	pi.Prompt = ""      // CHANGED 2025-10-02 04:02 - Remove prompt, will be added by layout - Problem: Duplicate "Password:" in ASCII-1
	pi.Placeholder = "" // CHANGED 2025-10-02 04:08 - Remove placeholder - Problem: Shows italic "E" in empty field
	pi.EchoMode = textinput.EchoPassword
	// CHANGED 2025-10-01 23:00 - Updated for textinput v2 API - Problem: v2 uses Focused/Blurred StyleState
	pi.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	pi.Styles.Focused.Text = lipgloss.NewStyle().Foreground(FgPrimary)
	pi.Styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(FgMuted).Italic(true)

	// Load sessions
	sess, _ := sessions.LoadSessions()
	if config.TestMode && len(sess) == 0 {
		sess = []sessions.Session{
			{Name: "GNOME", Exec: "gnome-session", Type: "X11"},
			{Name: "KDE Plasma", Exec: "startplasma-x11", Type: "X11"},
			{Name: "Sway", Exec: "sway", Type: "Wayland"},
			{Name: "Hyprland", Exec: "Hyprland", Type: "Wayland"},
			{Name: "i3", Exec: "i3", Type: "X11"},
			{Name: "Xfce Session", Exec: "startxfce4", Type: "X11"},
		}
	}
	if config.Debug {
		logDebug(" Loaded %d sessions", len(sess))
		for _, s := range sess {
			fmt.Printf("  - %s (%s)\n", s.Name, s.Type)
		}
	}

	// Setup animated spinner
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(Primary)

	var ipcClient *ipc.Client
	var selectedSession *sessions.Session
	var sessionIndex int

	if !config.TestMode {
		// CHANGED 2025-10-05 - Proper IPC client error handling - Problem: Silent failure causes nil pointer panic
		logDebug("Attempting to create IPC client...")
		client, err := ipc.NewClient()
		if err != nil {
			// CRITICAL: If IPC fails, we cannot authenticate with greetd
			// Log the error and exit rather than continue with nil client
			logDebug("FATAL: IPC client creation failed: %v", err)
			fmt.Fprintf(os.Stderr, "FATAL: Failed to create IPC client: %v\n", err)
			fmt.Fprintf(os.Stderr, "GREETD_SOCK environment variable: %s\n", os.Getenv("GREETD_SOCK"))
			fmt.Fprintf(os.Stderr, "This greeter must be run by greetd with GREETD_SOCK set.\n")
			os.Exit(1)
		}
		ipcClient = client
		logDebug("IPC client created successfully")

		// Load cached session and find its index
		cached, err := cache.LoadSelectedSession()
		if err != nil && config.Debug {
			logDebug(" Failed to load cached session: %v", err)
		} else if cached != nil {
			selectedSession = cached
			// Find the index of the cached session
			for i, s := range sess {
				if s.Name == cached.Name && s.Type == cached.Type {
					sessionIndex = i
					break
				}
			}
		}
	}

	// Default to first session if none selected
	if selectedSession == nil && len(sess) > 0 {
		selectedSession = &sess[0]
		sessionIndex = 0
	}

	// Load issue file content if requested
	var issueContent string
	if config.ShowIssue {
		if content, err := ioutil.ReadFile("/etc/issue"); err == nil {
			issueContent = strings.TrimSpace(string(content))
		}
	}

	// Load themes from directory
	themesDir := "themes"
	loadedThemes, err := themesOld.LoadThemesFromDir(themesDir)
	if err != nil && config.Debug {
		logDebug(" Failed to load themes: %v", err)
	}

	// Use specified theme if available, otherwise default
	currentTheme := themesOld.DefaultTheme
	if config.ThemeName != "" {
		if theme, ok := loadedThemes[config.ThemeName]; ok {
			currentTheme = theme
		}
	} else if theme, ok := loadedThemes["gnome"]; ok {
		currentTheme = theme
	}

	// Set initial focus
	ti.Focus()

	// CHANGED 2025-10-01 15:01 - Apply Dracula theme at initialization - Problem: User wants Dracula as default
	applyTheme("dracula")

	m := model{
		usernameInput:       ti,
		passwordInput:       pi,
		spinner:             sp,
		sessions:            sess,
		selectedSession:     selectedSession,
		sessionIndex:        sessionIndex,
		ipcClient:           ipcClient,
		theme:               currentTheme,
		mode:                ModeLogin,
		config:              config,
		issueContent:        issueContent,
		startTime:           time.Now(),
		width:               80,
		height:              24,
		powerOptions:        []string{"Reboot", "Shutdown", "Cancel"},
		powerIndex:          0,
		sessionDropdownOpen: false,
		focusState:          FocusUsername,
		animationFrame:      0,
		pulseColor:          0,
		borderFrame:         0,
		// CHANGED 2025-09-30 15:38 - Initialize default border and background settings - Problem: New fields need default values
		// CHANGED 2025-10-01 15:00 - Set Dracula as default theme and disable border animation - Problem: User wants static borders and Dracula theme
		selectedBorderStyle:    "classic",
		selectedBackground:     "none",
		currentTheme:           "dracula",
		borderAnimationEnabled: false,
		selectedFont:           "/usr/share/bubble-greet/fonts/dos_rebel.flf", // CHANGED 2025-10-02 01:35 - Absolute path
		customASCIIText:        "",
		// CHANGED 2025-10-01 - Initialize animation control defaults - Problem: New animation fields need default values
		selectedAnimationStyle:     "gradient",
		selectedAnimationSpeed:     1.0,
		selectedAnimationDirection: "right",
		animationStyleOptions:      []string{"gradient", "wave", "pulse", "rainbow", "matrix", "typewriter", "glow", "static"},
		animationDirectionOptions:  []string{"right", "left", "up", "down", "center-out"},
		// CHANGED 2025-10-02 05:40 - Initialize fire effect with default size - Problem: Fire needs initialization
		fireEffect: animations.NewFireEffect(80, 30, animations.GetDefaultFirePalette()),
		// CHANGED 2025-10-08 - Initialize rain effect with default size - Problem: Rain needs initialization
		rainEffect: animations.NewRainEffect(80, 30, animations.GetRainPalette("default")),
		// Initialize matrix effect with default size
		matrixEffect: animations.NewMatrixEffect(80, 30, animations.GetMatrixPalette("default")),
	}

	// CHANGED 2025-10-03 - Load cached preferences including session - Problem: Need to persist user preferences across sessions
	// CHANGED 2025-10-03 - Skip cache in test mode - Problem: Need fresh start when testing to avoid broken themes/borders
	if !m.config.TestMode {
		if prefs, err := cache.LoadPreferences(); err == nil && prefs != nil {
			if prefs.Theme != "" {
				m.currentTheme = prefs.Theme
				applyTheme(prefs.Theme)
			}
			if prefs.Background != "" {
				m.selectedBackground = prefs.Background
			}
			if prefs.BorderStyle != "" {
				m.selectedBorderStyle = prefs.BorderStyle
			}
			if prefs.Session != "" {
				// Find matching session in m.sessions
				for i, s := range m.sessions {
					if s.Name == prefs.Session {
						m.selectedSession = &m.sessions[i]
						m.sessionIndex = i
						break
					}
				}
			}
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick, doTick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		logDebug("Terminal resized: %dx%d", msg.Width, msg.Height)
		return m, nil

	case tickMsg:
		m.animationFrame++
		m.pulseColor = (m.pulseColor + 1) % 100
		m.borderFrame = (m.borderFrame + 1) % 20

		// CHANGED 2025-10-10 - Update screensaver time and check for activation - Problem: Need screensaver mode
		m.screensaverTime = time.Time(msg)

		// Check for screensaver activation using configurable timeout
		if m.mode == ModeLogin || m.mode == ModePassword {
			ssConfig := loadScreensaverConfig()
			idleDuration := time.Since(m.idleTimer)
			if idleDuration >= time.Duration(ssConfig.IdleTimeout)*time.Minute && m.mode != ModeScreensaver {
				m.mode = ModeScreensaver
			}
		}

		// CHANGED 2025-10-04 - Update fire when enableFire is true - Problem: Support Fire + Rain/Matrix combination
		if (m.enableFire || m.selectedBackground == "fire" || m.selectedBackground == "fire+rain") && m.fireEffect != nil {
			m.fireEffect.Update(m.animationFrame)
		}

		// CHANGED 2025-10-08 - Update rain when ascii-rain is selected - Problem: Need to update rain effect
		if m.selectedBackground == "ascii-rain" && m.rainEffect != nil {
			m.rainEffect.Update(m.animationFrame)
		}

		// Update matrix when matrix background is selected
		if m.selectedBackground == "matrix" && m.matrixEffect != nil {
			m.matrixEffect.Update(m.animationFrame)
		}

		cmds = append(cmds, doTick())

	case sessionSelectedMsg:
		session := sessions.Session(msg)
		m.selectedSession = &session
		// Update session index
		for i, s := range m.sessions {
			if s.Name == session.Name && s.Type == session.Type {
				m.sessionIndex = i
				break
			}
		}
		m.sessionDropdownOpen = false
		if m.config.Debug {
			logDebug(" Selected session: %s", session.Name)
		}
		if m.config.TestMode {
			fmt.Println("Test mode: Selected session:", session.Name)
			return m, tea.Quit
		} else {
			// Save to cache
			if err := cache.SaveSelectedSession(session); err != nil && m.config.Debug {
				logDebug(" Failed to save session: %v", err)
			}
			// CHANGED 2025-10-03 - Save session preference - Problem: Session selection not persisted in preferences
			// CHANGED 2025-10-03 - Skip saving in test mode - Problem: Don't persist during testing
			if !m.config.TestMode {
				cache.SavePreferences(cache.UserPreferences{
					Theme:       m.currentTheme,
					Background:  m.selectedBackground,
					BorderStyle: m.selectedBorderStyle,
					Session:     session.Name,
				})
			}
			return m, tea.Batch(cmds...)
		}

	case powerSelectedMsg:
		action := string(msg)
		switch action {
		case "Reboot":
			if m.config.TestMode {
				fmt.Println("Test mode: Would reboot system")
				return m, tea.Quit
			}
			fmt.Println("Rebooting...")
			return m, tea.Quit
		case "Shutdown":
			if m.config.TestMode {
				fmt.Println("Test mode: Would shutdown system")
				return m, tea.Quit
			}
			fmt.Println("Shutting down...")
			return m, tea.Quit
		case "Cancel":
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			cmds = append(cmds, textinput.Blink)
		}

	case string:
		if msg == "success" {
			// CHANGED 2025-10-09 21:00 - Removed delay workaround - Problem: Proper fix is in IPC layer
			// Now we properly wait for greetd's success response in StartSession() before returning
			// This ensures greetd has finished session initialization regardless of hardware speed
			fmt.Println("Session started successfully")
			return m, tea.Quit
		} else {
			// CHANGED 2025-10-05 - Store error message and return to password mode - Problem: BUG #4 - Greeter was exiting on non-success messages
			m.errorMessage = msg
			m.mode = ModePassword
			m.passwordInput.SetValue("") // Clear password field
			m.passwordInput.Focus()
			m.focusState = FocusPassword
			return m, textinput.Blink
		}
	case error:
		// CHANGED 2025-10-05 - Store error message and return to password mode - Problem: BUG #4 - Greeter was exiting on auth errors
		m.errorMessage = msg.Error()
		m.mode = ModePassword
		m.passwordInput.SetValue("") // Clear password field
		m.passwordInput.Focus()
		m.focusState = FocusPassword
		return m, textinput.Blink

	case tea.KeyMsg:
		newModel, cmd := m.handleKeyInput(msg)
		m = newModel
		cmds = append(cmds, cmd)
	}

	// Update components based on current mode and focus
	switch m.mode {
	case ModeLogin:
		if m.focusState == FocusUsername {
			var cmd tea.Cmd
			m.usernameInput, cmd = m.usernameInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	case ModePassword:
		if m.focusState == FocusPassword {
			var cmd tea.Cmd
			m.passwordInput, cmd = m.passwordInput.Update(msg)
			cmds = append(cmds, cmd)
			// CHANGED 2025-10-05 - Clear error message when user starts typing - Problem: Error should disappear when retry begins
			if m.errorMessage != "" && len(m.passwordInput.Value()) > 0 {
				m.errorMessage = ""
			}
		}
	case ModeLoading:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleKeyInput(msg tea.KeyMsg) (model, tea.Cmd) {
	// CHANGED 2025-10-01 23:00 - Updated for tea.KeyMsg v2 API - Problem: Type and Runes fields no longer exist
	if m.config.Debug {
		keyStr := msg.String()
		fmt.Fprintf(os.Stderr, "KEY DEBUG: String='%s'\n", keyStr)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		// CHANGED 2025-10-10 - Disable Ctrl+C in production mode - Problem: Ctrl+C kills kitty terminal hosting the greeter
		// Only allow Ctrl+C/Q to quit in test mode (when ipcClient is nil)
		if m.ipcClient == nil {
			// Test mode - allow quit
			return m, tea.Quit
		}
		// Production mode - ignore Ctrl+C/Q (security measure)
		return m, nil

	case "f1":
		// CHANGED 2025-10-03 17:15 - Remapped F1 to Menu - Problem: User wanted F1=Menu, F2=Sessions, F3=Notes, F4=Power
		// Main menu
		if m.mode == ModeLogin || m.mode == ModePassword {
			m.mode = ModeMenu
			m.menuIndex = 0

			// Build new structured menu
			m.menuOptions = []string{
				"Close Menu",
				"Themes",
				"Borders",
				"Backgrounds",
				"Wallpaper",
			}
			return m, nil
		}

	case "f2":
		// CHANGED 2025-10-03 17:15 - Remapped F2 to Sessions - Problem: User wanted F1=Menu, F2=Sessions, F3=Notes, F4=Power
		// Toggle session dropdown
		if m.mode == ModeLogin || m.mode == ModePassword {
			m.sessionDropdownOpen = !m.sessionDropdownOpen
			return m, nil
		}

	case "f3":
		// CHANGED 2025-10-03 17:15 - Remapped F3 to Notes - Problem: User wanted F1=Menu, F2=Sessions, F3=Notes, F4=Power
		// Release notes popup
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.config.Debug {
				fmt.Println("Debug: Opening release notes")
			}
			m.mode = ModeReleaseNotes
			m.usernameInput.Blur()
			m.passwordInput.Blur()
			return m, nil
		}

	case "f4":
		// CHANGED 2025-10-03 17:15 - F4 remains Power - Problem: User wanted F1=Menu, F2=Sessions, F3=Notes, F4=Power
		// Power menu
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.config.Debug {
				fmt.Println("Debug: Opening power menu")
			}
			m.mode = ModePower
			m.usernameInput.Blur()
			m.passwordInput.Blur()
			return m, nil
		}

	case "tab":
		// Cycle focus through form elements
		if m.mode == ModeLogin {
			switch m.focusState {
			case FocusSession:
				m.focusState = FocusUsername
				m.usernameInput.Focus()
			case FocusUsername:
				m.focusState = FocusSession
				m.usernameInput.Blur()
			}
			return m, textinput.Blink
		} else if m.mode == ModePassword {
			switch m.focusState {
			case FocusSession:
				m.focusState = FocusPassword
				m.passwordInput.Focus()
			case FocusPassword:
				m.focusState = FocusSession
				m.passwordInput.Blur()
			}
			return m, textinput.Blink
		}

	case "esc":
		switch m.mode {
		case ModePower:
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			return m, textinput.Blink
		case ModeMenu:
			// CHANGED 2025-09-30 - Add escape from menu
			m.mode = ModeLogin
			return m, nil
		// CHANGED 2025-09-30 14:50 - Add escape handling for submenus - Problem: Need to navigate back from submenus to main menu
		case ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeWallpaperSubmenu:
			// Go back to main menu
			m.mode = ModeMenu
			m.menuOptions = []string{
				"Close Menu",
				"Themes",
				"Borders",
				"Backgrounds",
				"Wallpaper",
			}
			m.menuIndex = 0
			return m, nil
		// CHANGED 2025-10-01 14:18 - Add escape handling for release notes - Problem: User needs to return from F5 release notes view
		case ModeReleaseNotes:
			// Return to login mode
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			return m, textinput.Blink
		default:
			if m.sessionDropdownOpen {
				m.sessionDropdownOpen = false
				return m, nil
			}
		}

	case "up", "k":
		if m.sessionDropdownOpen {
			if m.sessionIndex > 0 {
				m.sessionIndex--
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		} else if m.mode == ModePower {
			if m.powerIndex > 0 {
				m.powerIndex--
			}
			return m, nil
		} else if m.mode == ModeMenu || m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu {
			// CHANGED 2025-10-03 17:55 - Removed ModeVideoWallpapersSubmenu from navigation - Problem: Dead code cleanup
			if m.menuIndex > 0 {
				m.menuIndex--
			}
			return m, nil
		} else if m.focusState == FocusSession {
			// Navigate sessions when session selector is focused
			if m.sessionIndex > 0 {
				m.sessionIndex--
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		}

	case "down", "j":
		if m.sessionDropdownOpen {
			if m.sessionIndex < len(m.sessions)-1 {
				m.sessionIndex++
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		} else if m.mode == ModePower {
			if m.powerIndex < len(m.powerOptions)-1 {
				m.powerIndex++
			}
		} else if m.mode == ModeMenu || m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu {
			// CHANGED 2025-10-03 17:55 - Removed ModeVideoWallpapersSubmenu from navigation - Problem: Dead code cleanup
			if m.menuIndex < len(m.menuOptions)-1 {
				m.menuIndex++
			}
			return m, nil
		} else if m.focusState == FocusSession {
			// Navigate sessions when session selector is focused
			if m.sessionIndex < len(m.sessions)-1 {
				m.sessionIndex++
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		}

	// CHANGED 2025-10-07 19:15 - Add Page Up/Down handlers for ASCII variant cycling - Problem: User wants to cycle through ASCII art variants
	case "pgup", "page up":
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.selectedSession != nil {
				// Load config to get variant count
				sessionName := strings.ToLower(strings.Fields(m.selectedSession.Name)[0])
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

				configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
				if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
					m.asciiArtCount = len(asciiConfig.ASCIIVariants)
					m.asciiMaxHeight = asciiConfig.MaxASCIIHeight

					// Cycle backwards (decrement index with wraparound)
					m.asciiArtIndex--
					if m.asciiArtIndex < 0 {
						m.asciiArtIndex = m.asciiArtCount - 1
					}
				}
			}
			return m, nil
		}

	case "pgdn", "page down":
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.selectedSession != nil {
				// Load config to get variant count
				sessionName := strings.ToLower(strings.Fields(m.selectedSession.Name)[0])
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

				configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
				if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
					m.asciiArtCount = len(asciiConfig.ASCIIVariants)
					m.asciiMaxHeight = asciiConfig.MaxASCIIHeight

					// Cycle forwards (increment index with wraparound)
					m.asciiArtIndex++
					if m.asciiArtIndex >= m.asciiArtCount {
						m.asciiArtIndex = 0
					}
				}
			}
			return m, nil
		}

	case "enter":
		if m.sessionDropdownOpen {
			// Select current session from dropdown
			session := m.sessions[m.sessionIndex]
			m.sessionDropdownOpen = false
			return m, func() tea.Msg { return sessionSelectedMsg(session) }
		}

		// CHANGED 2025-09-30 14:52 - Add submenu selection handling - Problem: Handle main menu and all submenu selections
		if m.mode == ModeMenu {
			selectedOption := m.menuOptions[m.menuIndex]
			switch selectedOption {
			case "Close Menu":
				m.mode = ModeLogin
				return m, nil
			case "Themes":
				newModel, cmd := m.navigateToThemesSubmenu()
				return newModel.(model), cmd
			case "Borders":
				newModel, cmd := m.navigateToBordersSubmenu()
				return newModel.(model), cmd
			case "Backgrounds":
				newModel, cmd := m.navigateToBackgroundsSubmenu()
				return newModel.(model), cmd
			case "Wallpaper":
				newModel, cmd := m.navigateToWallpaperSubmenu()
				return newModel.(model), cmd
			}
			return m, nil
		}

		// Handle submenu selections
		if m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu {
			selectedOption := m.menuOptions[m.menuIndex]

			// Handle "← Back" option for all submenus
			if selectedOption == "← Back" {
				m.mode = ModeMenu
				m.menuOptions = []string{
					"Close Menu",
					"Themes",
					"Borders",
					"Backgrounds",
					"Wallpaper",
				}
				m.menuIndex = 0
				return m, nil
			}

			// CHANGED 2025-09-30 15:12 - Implement actual submenu functionality - Problem: Submenus were just placeholders doing nothing
			switch m.mode {
			case ModeThemesSubmenu:
				// Parse theme selection and apply it
				if strings.HasPrefix(selectedOption, "Theme: ") {
					themeName := strings.TrimPrefix(selectedOption, "Theme: ")
					m.currentTheme = themeName
					// Apply theme immediately
					applyTheme(themeName)
					// CHANGED 2025-10-03 - Save theme preference - Problem: Theme selection not persisted
					// CHANGED 2025-10-03 - Skip saving in test mode - Problem: Don't persist during testing
					if !m.config.TestMode {
						sessionName := ""
						if m.selectedSession != nil {
							sessionName = m.selectedSession.Name
						}
						cache.SavePreferences(cache.UserPreferences{
							Theme:       m.currentTheme,
							Background:  m.selectedBackground,
							BorderStyle: m.selectedBorderStyle,
							Session:     sessionName,
						})
					}
				}
				m.mode = ModeLogin
				return m, nil

			case ModeBordersSubmenu:
				// CHANGED 2025-10-01 15:29 - Restored ASCII border handling - Problem: User wants ASCII borders back
				switch selectedOption {
				case "Style: Classic":
					m.selectedBorderStyle = "classic"
				case "Style: Modern":
					m.selectedBorderStyle = "modern"
				case "Style: Minimal":
					m.selectedBorderStyle = "minimal"
				case "Style: ASCII-1":
					m.selectedBorderStyle = "ascii1"
				case "Style: ASCII-2":
					m.selectedBorderStyle = "ascii2"
				case "Animation: Wave":
					m.borderAnimationEnabled = true
					m.selectedBorderStyle = "wave"
				case "Animation: Pulse":
					m.borderAnimationEnabled = true
					m.selectedBorderStyle = "pulse"
				case "Animation: Off":
					m.borderAnimationEnabled = false
				}
				// CHANGED 2025-10-03 - Save border preference - Problem: Border selection not persisted
				// CHANGED 2025-10-03 - Skip saving in test mode - Problem: Don't persist during testing
				if !m.config.TestMode {
					sessionName := ""
					if m.selectedSession != nil {
						sessionName = m.selectedSession.Name
					}
					cache.SavePreferences(cache.UserPreferences{
						Theme:       m.currentTheme,
						Background:  m.selectedBackground,
						BorderStyle: m.selectedBorderStyle,
						Session:     sessionName,
					})
				}
				m.mode = ModeLogin
				return m, nil

			case ModeBackgroundsSubmenu:
				// CHANGED 2025-10-04 - Toggle backgrounds instead of replacing - Problem: User wants Fire + Rain/Matrix enabled together
				// Strip checkbox prefix to get actual option name
				optionName := strings.TrimPrefix(selectedOption, "[✓] ")
				optionName = strings.TrimPrefix(optionName, "[ ] ")

				switch optionName {
				case "Fire":
					m.enableFire = !m.enableFire
				case "ASCII Rain": // CHANGED 2025-10-08 - Add ascii-rain option
					// Rain is exclusive - disable others
					m.enableFire = false
					if m.selectedBackground != "ascii-rain" {
						m.selectedBackground = "ascii-rain"
					} else {
						m.selectedBackground = "none"
					}
				case "Matrix": // Add matrix option
					// Matrix is exclusive - disable others
					m.enableFire = false
					if m.selectedBackground != "matrix" {
						m.selectedBackground = "matrix"
					} else {
						m.selectedBackground = "none"
					}
				}

				// Update selectedBackground based on enabled flags
				// Priority: Fire > Matrix > ASCII Rain > none
				if m.enableFire {
					m.selectedBackground = "fire"
				} else if m.selectedBackground != "pattern" && m.selectedBackground != "ascii-rain" && m.selectedBackground != "matrix" {
					m.selectedBackground = "none"
				}
				// CHANGED 2025-10-03 - Save background preference - Problem: Background selection not persisted
				// CHANGED 2025-10-03 - Skip saving in test mode - Problem: Don't persist during testing
				if !m.config.TestMode {
					sessionName := ""
					if m.selectedSession != nil {
						sessionName = m.selectedSession.Name
					}
					cache.SavePreferences(cache.UserPreferences{
						Theme:       m.currentTheme,
						Background:  m.selectedBackground,
						BorderStyle: m.selectedBorderStyle,
						Session:     sessionName,
					})
				}
				// CHANGED 2025-10-06 - Refresh menu to update checkboxes - Problem: Checkboxes not updating after toggle
				newModel, cmd := m.navigateToBackgroundsSubmenu()
				return newModel.(model), cmd
			case ModeWallpaperSubmenu:
				// CHANGED 2025-10-03 17:35 - Use modular wallpaper handler - Problem: Keep main.go clean
				newModel, cmd := m.handleWallpaperSelection(selectedOption)
				return newModel.(model), cmd
			}
			return m, nil
		}

		switch m.mode {
		case ModeLogin:
			if m.focusState == FocusSession {
				// Enter from session focus goes to username
				m.focusState = FocusUsername
				m.usernameInput.Focus()
				return m, textinput.Blink
			} else {
				// Enter from username goes to password
				if m.config.Debug {
					fmt.Println("Debug: Switching to password mode")
				}
				m.mode = ModePassword
				m.focusState = FocusPassword
				m.usernameInput.Blur()
				m.passwordInput.Focus()
				return m, textinput.Blink
			}

		case ModePassword:
			if m.focusState == FocusSession {
				// Enter from session focus goes to password input
				m.focusState = FocusPassword
				m.passwordInput.Focus()
				return m, textinput.Blink
			} else {
				// Enter from password submits
				username := m.usernameInput.Value()
				password := m.passwordInput.Value()
				if m.config.Debug {
					logDebug(" Username: %s, Password: %s", username, password)
				}
				if m.config.TestMode {
					fmt.Println("Test mode: Auth successful")
					return m, tea.Quit
				} else {
					if m.ipcClient == nil {
						fmt.Println("Error: No IPC client available")
						return m, tea.Quit
					}
					m.mode = ModeLoading
					return m, m.authenticate(username, password)
				}
			}

		case ModePower:
			if m.powerIndex < len(m.powerOptions) {
				option := m.powerOptions[m.powerIndex]
				return m, func() tea.Msg { return powerSelectedMsg(option) }
			}
		}
	}

	return m, nil
}

// CHANGED 2025-10-01 22:45 - Return tea.View with BackgroundColor set - Problem: Terminal background shows through, causing color bleeding
func (m model) View() tea.View {
	// Use full terminal dimensions
	termWidth := m.width
	termHeight := m.height
	if termWidth == 0 {
		termWidth = 80
	}
	if termHeight == 0 {
		termHeight = 24
	}

	var content string
	switch m.mode {
	case ModePower:
		// CHANGED 2025-10-03 19:00 - Fixed missing power menu rendering - Problem: F4 power menu showed blank screen
		content = m.renderPowerView(termWidth, termHeight)
	case ModeMenu, ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeWallpaperSubmenu:
		// CHANGED 2025-10-03 18:00 - Removed ModeVideoWallpapersSubmenu from rendering - Problem: Dead code cleanup
		content = m.renderMenuView(termWidth, termHeight)
	case ModeReleaseNotes:
		// CHANGED 2025-10-01 14:15 - Added F5 release notes view rendering - Problem: User requested F5 release notes functionality
		content = m.renderReleaseNotesView(termWidth, termHeight)
	case ModeScreensaver:
		// CHANGED 2025-10-10 - Added screensaver rendering - Problem: Need screensaver mode
		content = renderScreensaverView(m, termWidth, termHeight)
	default:
		content = m.renderMainView(termWidth, termHeight)
	}

	var view tea.View

	// Check if fire background is enabled
	// CHANGED 2025-10-06 - Removed wallpaper check - Problem: Wallpapers shouldn't enable fire
	// CHANGED 2025-10-06 - Only show fire on main login screen, not in menus - Problem: Fire hidden behind menu content
	// CHANGED 2025-10-08 - Add ascii-rain background support - Problem: Need to handle ascii-rain background
	hasFireBackground := (m.enableFire || m.selectedBackground == "fire" || m.selectedBackground == "fire+rain") && m.mode == ModeLogin
	hasRainBackground := (m.selectedBackground == "ascii-rain") && m.mode == ModeLogin

	if hasFireBackground {
		// CHANGED 2025-10-06 - Use multi-layer approach: fire at bottom, centered UI on top - Problem: Need to center UI while keeping fire at bottom
		fireHeight := (termHeight * 2) / 5 // Bottom 40% of terminal
		fireY := termHeight - fireHeight

		// Render fire
		backgroundContent := m.addFireEffect("", termWidth, fireHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: fire at bottom, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(fireY),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	} else if hasRainBackground {
		// CHANGED 2025-10-08 - Render ascii-rain as full background - Problem: Need to handle ascii-rain background
		// Render rain as full background
		backgroundContent := m.addAsciiRain("", termWidth, termHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: rain as background, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(0),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	} else if m.selectedBackground == "matrix" && m.mode == ModeLogin {
		// Render matrix as full background
		backgroundContent := m.addMatrixEffect("", termWidth, termHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: matrix as background, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(0),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	}

	// CHANGED 2025-10-06 - Use X/Y positioning instead of Place() to avoid ghosting - Problem: lipgloss.Place() causes ghosting in fullscreen kitty
	// Calculate center position manually (CRUSH approach)
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)
	x := (termWidth - contentWidth) / 2
	y := (termHeight - contentHeight) / 2

	// CHANGED 2025-10-09 20:40 - Removed ticker fullscreen check - Problem: Ticker deleted, no longer needed
	// Use layer X/Y positioning instead of Place()
	view.Layer = lipgloss.NewCanvas(lipgloss.NewLayer(content).X(x).Y(y))
	view.BackgroundColor = BgBase
	return view
}

// CHANGED 2025-10-06 - Ensure content fills entire terminal to prevent ghosting
// Problem: In fullscreen kitty, ANSI codes mess with Bubble Tea's diff renderer cell counting
func ensureFullTerminalCoverage(content string, termWidth, termHeight int) string {
	lines := strings.Split(content, "\n")

	// Pad lines to exact terminal width with plain spaces (no ANSI styling)
	// CRITICAL: Use PLAIN spaces, not lipgloss.Render(), to avoid ANSI code length variations
	for i := range lines {
		// Strip ANSI to get actual visible width
		visibleWidth := len([]rune(stripAnsi(lines[i])))
		if visibleWidth < termWidth {
			// Use plain spaces - Bubble Tea renderer will fill with background color
			lines[i] += strings.Repeat(" ", termWidth-visibleWidth)
		}
	}

	// Create full-width empty line with plain spaces
	emptyLine := strings.Repeat(" ", termWidth)

	// Ensure we have exactly termHeight lines
	for len(lines) < termHeight {
		lines = append(lines, emptyLine)
	}

	// Trim to exactly termHeight lines
	if len(lines) > termHeight {
		lines = lines[:termHeight]
	}

	return strings.Join(lines, "\n")
}

// CHANGED 2025-10-01 - Replaced WM-named themes with common themes - Problem: User wants gruvbox, material, nord etc instead of WM names
// CHANGED 2025-10-03 17:40 - Moved navigation functions to menu.go and wallpaper.go - Problem: Keep main.go clean and modular

// CHANGED 2025-09-30 15:25 - Complete dual border redesign - Problem: User wants dual border layout with help text in outer border
func (m model) renderMainView(termWidth, termHeight int) string {
	return m.renderDualBorderLayout(termWidth, termHeight)
}

// New dual border layout system
func (m model) renderDualBorderLayout(termWidth, termHeight int) string {
	// CHANGED 2025-10-02 03:42 - Route ASCII-1 and ASCII-2 styles
	if m.selectedBorderStyle == "ascii1" {
		return m.renderASCII1BorderLayout(termWidth, termHeight)
	}
	if m.selectedBorderStyle == "ascii2" {
		return m.renderASCII2BorderLayout(termWidth, termHeight)
	}
	// ===== INNER BORDER CONTENT =====
	// Contains: WM ASCII art, session dropdown, username/password fields

	// Calculate inner content area
	// CHANGED 2025-10-01 19:10 - Use reasonable max width like installer - Problem: termWidth-8 too large, pushes ASCII outside border
	innerWidth := min(100, termWidth-8) // Reasonable width for content area
	var innerSections []string

	// WM/Session ASCII art - prominent display
	// CHANGED 2025-10-01 15:27 - Fix centering, art is already colored - Problem: ASCII aligned right and double-styled
	if m.selectedSession != nil {
		art := m.getSessionArt(m.selectedSession.Name)
		if art != "" {
			// CHANGED 2025-10-01 19:30 - JoinVertical(Center) handles centering, just add art - Problem: Width() was causing bleeding
			// Art is already colored, JoinVertical(Center) will center each line
			innerSections = append(innerSections, art)
			// Remove automatic spacing and add consistent 2-line spacing
			// innerSections = append(innerSections, "") // spacing
		}
	}

	// Ensure exactly 2 lines of spacing after ASCII art
	innerSections = append(innerSections, "", "")

	// Main form (session selector, username, password) in bordered box
	// CHANGED 2025-10-01 19:35 - Wrap form in border for left alignment - Problem: JoinVertical(Center) centers form fields
	// CHANGED 2025-10-01 20:05 - Reverted Width() addition - Problem: Width broke entire middle section alignment
	// CHANGED 2025-10-01 20:40 - Add fixed width to form content with Place - Problem: Variable content width breaks border alignment
	formContentWidth := innerWidth - 20
	formContent := m.renderMainForm(formContentWidth)
	// CHANGED 2025-10-06 - Removed Place() - Problem: Place() causes ghosting and duplicate borders
	fixedFormContent := formContent
	formBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderDefault).
		Background(BgBase).
		Padding(1, 2)

	formBorder := formBorderStyle.Render(fixedFormContent)
	innerSections = append(innerSections, formBorder)

	// CHANGED 2025-10-01 15:40 - Simplified border title as first line - Problem: String manipulation breaking border rendering
	// Create inner border container with user-selected style
	innerBorderColor := m.getInnerBorderColor()

	// Standard SESSIONS title
	var titleLine string
	titleText := "SESSIONS"
	slashes := strings.Repeat("/", 7)

	// Use shorter dashes for minimal border style
	dashCount := 30
	if m.selectedBorderStyle == "minimal" {
		dashCount = 4
	}
	dashes := strings.Repeat("─", dashCount)

	// CHANGED 2025-10-06 - Re-added Align() for title centering WITHIN fixed width - Problem: Title needs to be centered in inner border
	// NOTE: This Align() is safe because it's within a fixed Width() container, not used with Place()
	titleLine = lipgloss.NewStyle().
		Foreground(innerBorderColor).
		Bold(true).
		Width(innerWidth - 6).
		Align(lipgloss.Center).
		Render(dashes + slashes + titleText + slashes + dashes)

	// Add title as first element
	contentWithTitle := []string{titleLine}

	// Add spacing
	contentWithTitle = append(contentWithTitle, "")

	contentWithTitle = append(contentWithTitle, innerSections...)

	// CHANGED 2025-10-01 19:30 - Revert to Center, ASCII already has explicit centering - Problem: Left join pushed ASCII right
	innerContent := lipgloss.JoinVertical(lipgloss.Center, contentWithTitle...)

	// CHANGED 2025-10-02 11:42 - Reverted: always use normal styling since fire is now inside - Problem: Fire is inner content now, not background
	innerBorderStyle := lipgloss.NewStyle().
		Border(m.getInnerBorderStyle()).
		BorderForeground(innerBorderColor).
		Width(innerWidth).
		Background(BgBase).
		Padding(2, 3)

	innerBox := innerBorderStyle.Render(innerContent)

	// CHANGED 2025-10-02 11:49 - Fire at bottom, so no outer border - Problem: Fire is fullscreen at bottom, no room for outer border
	// CHANGED 2025-10-08 - Also remove outer border for ascii-rain - Problem: Ascii-rain needs fullscreen rendering
	// CHANGED 2025-10-08 - Also remove outer border for ticker - Problem: Ticker needs fullscreen rendering like fire and rain
	// Add matrix to backgrounds that remove outer border
	if m.selectedBackground == "fire" || m.selectedBackground == "ascii-rain" || m.selectedBackground == "matrix" || m.selectedBackground == "ticker" || m.selectedBackground == "fire+rain" {
		helpText := m.renderMainHelp()
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
	// CHANGED 2025-10-06 - Reduced margin to make outer border closer to edges - Problem: Outer border too small, too far from edges
	outerWidth := termWidth - 8 // Small margin from terminal edges

	var outerSections []string

	// CHANGED 2025-10-03 16:35 - Removed bubble-greet title text - Problem: User requested removal of green title
	// Title removed per user request

	// Time if enabled
	if m.config.ShowTime {
		// CHANGED 2025-10-01 20:25 - Manual centering without lipgloss Align - Problem: lipgloss Align causes bleeding and doesn't center properly
		timeStyle := lipgloss.NewStyle().
			Foreground(FgSecondary)
		currentTime := time.Now().Format("15:04:05 Mon Jan 02, 2006")
		centeredTime := centerText(currentTime, outerWidth-4)
		outerSections = append(outerSections, timeStyle.Render(centeredTime))
	}

	outerSections = append(outerSections, "") // spacing

	// CHANGED 2025-10-06 - Removed Place() - Problem: Place() causes ghosting and duplicate borders
	// Just append innerBox without centering - View() will handle positioning
	outerSections = append(outerSections, innerBox)

	// Join all sections with Center to center each section horizontally
	outerContent := lipgloss.JoinVertical(lipgloss.Center, outerSections...)

	// Create outer border container with user-selected style
	outerBorderColor := m.getOuterBorderColor()
	// CHANGED 2025-10-01 19:50 - Remove Align to prevent bleeding with Background - Problem: Background+Align causes bleed around centered text
	// CHANGED 2025-10-01 22:00 - Add BgBase back to both - Problem: Terminal default background shows through differently
	// CHANGED 2025-10-06 - Calculate both horizontal AND vertical padding to expand outer border to terminal edges
	innerBoxWidth := lipgloss.Width(innerBox)
	innerBoxHeight := lipgloss.Height(outerContent)

	// Calculate padding to make outer border reach near terminal edges
	horizontalPadding := (termWidth - innerBoxWidth - 10) / 2 // Leave 10 chars total margin (5 per side)
	if horizontalPadding < 2 {
		horizontalPadding = 2 // Minimum padding
	}

	// CHANGED 2025-10-06 - Increased vertical padding to push border closer to edges, leaving room for help text at bottom
	verticalPadding := (termHeight - innerBoxHeight - 6) / 2 // Leave 6 lines total margin (3 per side) for help text
	if verticalPadding < 2 {
		verticalPadding = 2 // Minimum padding
	}

	outerBorderStyle := lipgloss.NewStyle().
		Border(m.getOuterBorderStyle()). // User-selected outer border style
		BorderForeground(outerBorderColor).
		Background(BgBase).
		Padding(verticalPadding, horizontalPadding)

	outerBox := outerBorderStyle.Render(outerContent)

	// CHANGED 2025-10-01 19:25 - Move help text BELOW outer border - Problem: Help text was inside border
	// Help text at bottom, below outer border
	helpText := m.renderMainHelp()
	helpStyle := lipgloss.NewStyle().
		Foreground(FgMuted).
		Align(lipgloss.Center)

	// Join outer box and help text vertically
	contentWithHelp := lipgloss.JoinVertical(lipgloss.Center, outerBox, "", helpStyle.Render(helpText))

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering - Problem: Place() creates uncolored padding causing ghosting
	return contentWithHelp
}

// CHANGED 2025-10-02 03:45 - ASCII-1: Just a border style, uses current theme colors
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
	goodsWidth := 100

	// CHANGED 2025-10-06 - Reduced vertical padding - Problem: Too much vertical space
	// Use fixed smaller vertical padding instead of calculated
	goodsStyle := lipgloss.NewStyle().
		Border(asciiBorder).
		BorderForeground(BorderDefault). // Use theme border color
		Padding(2, 4).                   // Fixed vertical and horizontal padding
		Background(BgBase)

	// Build THE GOODS content
	var sections []string

	// WM/Session ASCII art - use theme colors
	if m.selectedSession != nil {
		art := m.getSessionASCII() // Use normal colored ASCII, not monochrome
		if art != "" {
			// CHANGED 2025-10-03 16:25 - Center ASCII art within border - Problem: ASCII art was left-aligned in ASCII-1 border
			// Center the ASCII art within the available width
			artStyle := lipgloss.NewStyle().
				Width(goodsWidth - 8).
				Align(lipgloss.Center)
			sections = append(sections, artStyle.Render(art))
			// Control spacing explicitly - remove old spacing and add exactly 2 lines
			// sections = append(sections, "") // Remove old spacing
		}
	}
	
	// Ensure exactly 2 lines of spacing after ASCII art
	sections = append(sections, "", "")

	// Session selector - use theme colors
	if len(m.sessions) > 0 && m.selectedSession != nil {
		sessionStyle := lipgloss.NewStyle().
			Foreground(Primary). // Theme primary color
			Background(BgBase).
			Bold(true).
			Width(goodsWidth - 8).
			Align(lipgloss.Center)

		sessionText := fmt.Sprintf("[ %s (%s) ]", m.selectedSession.Name, m.selectedSession.Type)
		sections = append(sections, sessionStyle.Render(sessionText))
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

	// CHANGED 2025-10-05 - Display error message below password field - Problem: BUG #4 - User needs to see auth errors
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")). // Red color
			Bold(true)
		sections = append(sections, "")
		sections = append(sections, errorStyle.Render("✗ "+m.errorMessage))
	}

	// Join THE GOODS
	goodsContent := strings.Join(sections, "\n")

	// Wrap THE GOODS in ASCII border
	borderedGoods := goodsStyle.Render(goodsContent)

	// Help text
	// CHANGED 2025-10-06 - Removed Width(termWidth) - Problem: Full-width help text stretches border to screen edges
	helpText := "F2=Menu | F3=Sessions | F4=Power | F5=Release Notes | Enter=Login | ESC=Back"
	helpStyle := lipgloss.NewStyle().
		Foreground(FgMuted) // Theme muted color

	// Final layout
	finalContent := lipgloss.JoinVertical(
		lipgloss.Center,
		borderedGoods,
		"",
		helpStyle.Render(helpText),
	)

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering - Problem: Place() creates uncolored padding
	return finalContent
}

// Fallback ASCII border if file not found
func (m model) renderASCIIBorderFallback(termWidth, termHeight int) string {
	// Simple ASCII box as fallback
	monoMedium := lipgloss.Color("#888888")

	content := "┌" + strings.Repeat("─", 60) + "┐\n"
	for i := 0; i < 20; i++ {
		content += "│" + strings.Repeat(" ", 60) + "│\n"
	}
	content += "└" + strings.Repeat("─", 60) + "┘"

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering - Problem: Place() creates uncolored padding
	style := lipgloss.NewStyle().Foreground(monoMedium)
	return style.Render(content)
}

// Render form with monochrome colors
func (m model) renderMonochromeForm(width int) string {
	monoWhite := lipgloss.Color("#ffffff")
	monoLight := lipgloss.Color("#cccccc")
	// CHANGED 2025-10-01 23:50 - Use BgBase instead of monoDark to prevent bleeding - Problem: Different background color causes visible bleed
	monoDark := BgBase

	var sections []string

	// Session selector with monochrome styling
	if len(m.sessions) > 0 {
		// CHANGED 2025-10-01 19:15 - Remove Width() and Align(Center) - Problem: Should be left-aligned like other fields
		sessionStyle := lipgloss.NewStyle().
			Foreground(monoWhite).
			Background(monoDark).
			Bold(true).
			Padding(0, 1)

		sessionText := fmt.Sprintf("Session: %s (%s)", m.selectedSession.Name, m.selectedSession.Type)
		sections = append(sections, sessionStyle.Render(sessionText))
		sections = append(sections, "")
	}

	// Username input
	usernameStyle := lipgloss.NewStyle().
		Foreground(monoLight).
		Width(width).
		Align(lipgloss.Left)

	m.usernameInput.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(monoWhite).Bold(true)
	m.usernameInput.Styles.Focused.Text = lipgloss.NewStyle().Foreground(monoWhite)
	sections = append(sections, usernameStyle.Render(m.usernameInput.View()))

	// Password input
	passwordStyle := lipgloss.NewStyle().
		Foreground(monoLight).
		Width(width).
		Align(lipgloss.Left)

	m.passwordInput.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(monoWhite).Bold(true)
	m.passwordInput.Styles.Focused.Text = lipgloss.NewStyle().Foreground(monoWhite)
	sections = append(sections, passwordStyle.Render(m.passwordInput.View()))

	// CHANGED 2025-10-05 - Display error message in monochrome style - Problem: BUG #4 - Error display for all themes
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)
		sections = append(sections, "")
		sections = append(sections, errorStyle.Render("✗ "+m.errorMessage))
	}

	return strings.Join(sections, "\n")
}

// CHANGED 2025-10-01 - Monochrome ASCII art for ASCII border style - Problem: ASCII border needs monochrome theme
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
	configPath := fmt.Sprintf("/usr/share/bubble-greet/ascii_configs/%s.conf", configFileName)
	asciiConfig, err := loadASCIIConfig(configPath)
	if err != nil {
		// Fallback to session name as text
		return sessionName
	}

	// CHANGED 2025-10-01 14:47 - Always apply monochrome animation - Problem: ASCII not displaying without colors in config
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

// CHANGED 2025-09-30 15:30 - Implement actual border style functionality - Problem: User's border selections didn't do anything

// Get inner border style based on user selection
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
		} // CHANGED 2025-10-03 16:20 - Use single space border for truly minimal look - Problem: NormalBorder looked identical to Modern
	case "ascii":
		return lipgloss.HiddenBorder() // ASCII borders handle their own rendering
	case "wave":
		return lipgloss.Border{
			Top:         "~",
			Bottom:      "~",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "╰",
			BottomRight: "╯",
		} // CHANGED 2025-10-03 16:20 - Use wavy characters for wave border - Problem: Looked identical to pulse
	case "pulse":
		return lipgloss.DoubleBorder() // CHANGED 2025-10-03 16:20 - Use double border for pulse - Problem: Looked identical to wave
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
		return lipgloss.ThickBorder() // CHANGED 2025-10-03 16:20 - Use thick outer for modern double-border look - Problem: Hidden outer made modern look minimal
	case "minimal":
		return lipgloss.HiddenBorder() // CHANGED 2025-10-03 16:20 - Hide outer for clean minimal look - Problem: Was using empty Border{}
	case "ascii":
		return lipgloss.HiddenBorder() // ASCII style uses only custom border
	case "wave":
		return lipgloss.RoundedBorder() // CHANGED 2025-10-03 16:20 - Rounded outer for wave - Problem: DoubleBorder looked identical to pulse
	case "pulse":
		return lipgloss.ThickBorder() // CHANGED 2025-10-03 16:20 - Thick outer for pulse - Problem: DoubleBorder looked identical to wave
	default:
		return lipgloss.DoubleBorder() // Default
	}
}

// Get inner border color with animation support
func (m model) getInnerBorderColor() color.Color {
	if !m.borderAnimationEnabled {
		// Static color based on current theme
		return Primary
	}

	switch m.selectedBorderStyle {
	case "wave":
		// CHANGED 2025-10-03 16:20 - Wave cycles through all theme colors smoothly - Problem: Looked identical to pulse
		// Wave animation - smooth color transitions through full palette
		colors := []color.Color{Primary, Secondary, Accent, Warning}
		return colors[(m.animationFrame/2)%len(colors)]
	case "pulse":
		// CHANGED 2025-10-03 16:20 - Pulse alternates between bright and dim - Problem: Looked identical to wave
		// Pulse animation - brightness oscillation (bright/dim/bright/dim)
		if m.animationFrame%8 < 4 {
			return Primary // Bright phase
		}
		return FgMuted // Dim phase
	default:
		// Default animated border
		return m.getAnimatedBorderColor()
	}
}

// Get outer border color with animation support
func (m model) getOuterBorderColor() color.Color {
	if !m.borderAnimationEnabled {
		// Static muted color for outer border
		return FgSubtle
	}

	switch m.selectedBorderStyle {
	case "wave":
		// CHANGED 2025-10-03 16:20 - Complementary wave offset from inner - Problem: Needed distinct outer animation
		// Complementary wave for outer border (offset from inner)
		colors := []color.Color{Secondary, Accent, Warning, Primary}
		return colors[(m.animationFrame/2+2)%len(colors)] // Offset by 2 for complementary effect
	case "pulse":
		// CHANGED 2025-10-03 16:20 - Outer stays subtle during pulse - Problem: Needed to complement inner pulse
		// Subtle static color for outer border during pulse
		return FgSecondary // Keep outer border constant while inner pulses
	default:
		// Default secondary animation
		colors := []color.Color{FgSubtle, FgSecondary, Primary}
		return colors[m.animationFrame%len(colors)]
	}
}

// CHANGED 2025-09-30 15:35 - Implement background animations - Problem: User wants Matrix rain, particles, etc. in background
func (m model) applyBackgroundAnimation(content string, width, height int) string {
	switch m.selectedBackground {
	case "fire": // CHANGED 2025-10-02 06:05 - Add fire effect rendering - Problem: Fire needs to be rendered as background
		return m.addFireEffect(content, width, height)
	case "matrix":
		return m.addMatrixEffect(content, width, height)
	case "ascii-rain": // CHANGED 2025-10-08 - Add ascii rain effect - Problem: User wants ascii rain background
		return m.addAsciiRain(content, width, height)
	case "none":
		fallthrough
	default:
		return content
	}
}

// CHANGED 2025-10-02 06:25 - Matrix rain with theme color support - Problem: Hardcoded green doesn't respect themes
func (m model) addMatrixRain(content string, width, height int) string {
	// Simple matrix rain simulation
	matrixChars := []rune{'0', '1', '░', '▒', '▓', '█'}
	frame := m.animationFrame

	// Get matrix colors from theme
	matrixColors := animations.GetMatrixPalette(m.currentTheme)

	lines := strings.Split(content, "\n")
	for i := 0; i < height-len(lines); i++ {
		var rainLine strings.Builder
		for j := 0; j < width; j++ {
			if (frame+i+j*3)%20 == 0 {
				char := matrixChars[(frame+j)%len(matrixChars)]
				// Vary color intensity based on position
				colorIndex := (frame + i) % len(matrixColors)
				coloredChar := lipgloss.NewStyle().
					Foreground(lipgloss.Color(matrixColors[colorIndex])).
					Render(string(char))
				rainLine.WriteString(coloredChar)
			} else {
				rainLine.WriteString(" ")
			}
		}
		lines = append(lines, rainLine.String())
	}
	return strings.Join(lines, "\n")
}

// CHANGED 2025-10-02 06:30 - Particle field with theme color support - Problem: Hardcoded Catppuccin colors don't respect themes

// CHANGED 2025-10-02 06:10 - Fire effect background rendering - Problem: Fire needs proper integration with content overlay
func (m model) addFireEffect(content string, width, height int) string {
	if m.fireEffect == nil {
		return content
	}

	// CHANGED 2025-10-06 - Only resize if dimensions actually changed - Problem: Calling Resize() every frame causes ghosting/flickering
	// Store last dimensions to avoid unnecessary reinits
	if m.lastFireWidth != width || m.lastFireHeight != height {
		m.fireEffect.Resize(width, height)
		m.lastFireWidth = width
		m.lastFireHeight = height
	}

	// Update palette from current theme
	m.fireEffect.UpdatePalette(animations.GetFirePalette(m.currentTheme))

	// Update fire simulation
	m.fireEffect.Update(m.animationFrame)

	// Render fire background
	fireBackground := m.fireEffect.Render()

	// For now, just return fire (content will be overlaid by main rendering)
	return fireBackground
}

// CHANGED 2025-10-08 - Rain effect background rendering - Problem: Ascii rain needs proper integration with content overlay
func (m model) addAsciiRain(content string, width, height int) string {
	if m.rainEffect == nil {
		return content
	}

	// CHANGED 2025-10-08 - Only resize if dimensions actually changed - Problem: Calling Resize() every frame causes ghosting/flickering
	// Store last dimensions to avoid unnecessary reinits
	if m.lastRainWidth != width || m.lastRainHeight != height {
		m.rainEffect.Resize(width, height)
		m.lastRainWidth = width
		m.lastRainHeight = height
	}

	// Update palette from current theme
	m.rainEffect.UpdatePalette(animations.GetRainPalette(m.currentTheme))

	// Update rain simulation
	m.rainEffect.Update(m.animationFrame)

	// Render rain background
	rainBackground := m.rainEffect.Render()

	// For now, just return rain (content will be overlaid by main rendering)
	return rainBackground
}

// Matrix effect background rendering
func (m model) addMatrixEffect(content string, width, height int) string {
	if m.matrixEffect == nil {
		return content
	}

	// Only resize if dimensions actually changed - Problem: Calling Resize() every frame causes ghosting/flickering
	// Store last dimensions to avoid unnecessary reinits
	if m.lastMatrixWidth != width || m.lastMatrixHeight != height {
		m.matrixEffect.Resize(width, height)
		m.lastMatrixWidth = width
		m.lastMatrixHeight = height
	}

	// Update palette from current theme
	m.matrixEffect.UpdatePalette(animations.GetMatrixPalette(m.currentTheme))

	// Update matrix simulation
	m.matrixEffect.Update(m.animationFrame)

	// Render matrix background
	matrixBackground := m.matrixEffect.Render()

	// For now, just return matrix (content will be overlaid by main rendering)
	return matrixBackground
}
func (m model) getBackgroundColor() color.Color {
	// CHANGED 2025-10-01 21:30 - Always return BgBase to prevent bleeding - Problem: Different colors cause visible bleed
	return BgBase
}

func (m model) renderMainForm(width int) string {
	var parts []string

	// Session selection (always visible at top)
	sessionContent := m.renderSessionSelector(width)
	parts = append(parts, sessionContent)

	// Add spacing
	parts = append(parts, "")

	// Current input based on mode
	switch m.mode {
	case ModeLogin:
		usernameLabel := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.getFocusColor(FocusUsername)).
			Width(10).
			Render("Username:")

		// CHANGED 2025-10-01 22:35 - Remove Foreground, use BgBase only - Problem: Foreground on wrapper causes bleeding
		inputStyle := lipgloss.NewStyle().
			Background(BgBase).
			Padding(0, 1)

		usernameInput := inputStyle.Render(m.usernameInput.View())

		usernameRow := lipgloss.JoinHorizontal(
			lipgloss.Left,
			usernameLabel,
			" ",
			usernameInput,
		)
		parts = append(parts, usernameRow)

	case ModePassword:
		passwordLabel := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.getFocusColor(FocusPassword)).
			Width(10).
			Render("Password:")

		// CHANGED 2025-10-01 22:35 - Remove Foreground, use BgBase only - Problem: Foreground on wrapper causes bleeding
		inputStyle := lipgloss.NewStyle().
			Background(BgBase).
			Padding(0, 1)

		passwordInput := inputStyle.Render(m.passwordInput.View())

		// Add spinner if user is typing password
		var passwordRow string
		if m.passwordInput.Value() != "" {
			// Show spinner when password is being entered
			spinnerView := lipgloss.NewStyle().
				Foreground(Primary).
				Render(m.spinner.View())

			passwordRow = lipgloss.JoinHorizontal(
				lipgloss.Left,
				passwordLabel,
				" ",
				passwordInput,
				" ",
				spinnerView,
			)
		} else {
			passwordRow = lipgloss.JoinHorizontal(
				lipgloss.Left,
				passwordLabel,
				" ",
				passwordInput,
			)
		}
		parts = append(parts, passwordRow)

		// CHANGED 2025-10-05 - Display error message below password in main form - Problem: BUG #4 - Error display
		if m.errorMessage != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555")).
				Bold(true)
			parts = append(parts, "")
			parts = append(parts, errorStyle.Render("✗ "+m.errorMessage))
		}

	case ModeLoading:
		loadingStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(Accent).
			Align(lipgloss.Center).
			Width(width)

		// Show animated spinner
		loadingText := loadingStyle.Render("Authenticating... " + m.spinner.View())
		parts = append(parts, loadingText)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m model) renderSessionSelector(width int) string {
	// Session label with focus indication
	sessionLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.getFocusColor(FocusSession)).
		Width(10).
		Render("Session:")

	// Current session display
	var sessionDisplay string
	if m.selectedSession != nil {
		sessionText := fmt.Sprintf("%s (%s)", m.selectedSession.Name, m.selectedSession.Type)

		borderColor := BorderDefault
		if m.focusState == FocusSession {
			borderColor = BorderFocus
		}

		// CHANGED 2025-10-02 04:42 - Use Inline(true) to prevent border from adding height - Problem: Border causes vertical misalignment
		sessionStyle := lipgloss.NewStyle().
			Foreground(FgPrimary).
			Background(BgBase).
			Padding(0, 1).
			Bold(true).
			Inline(true) // Force single-line rendering, ignore border height

		if m.sessionDropdownOpen || m.focusState == FocusSession {
			sessionStyle = sessionStyle.
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(borderColor)
		}

		sessionDisplay = sessionStyle.Render(sessionText)
	} else {
		sessionDisplay = lipgloss.NewStyle().
			Foreground(FgMuted).
			Width(width - 14).
			Render("No session selected")
	}

	// Dropdown indicator
	dropdownIndicator := "▼"
	if m.sessionDropdownOpen {
		dropdownIndicator = "▲"
	}
	indicatorStyle := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Render(dropdownIndicator)

	sessionRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		sessionLabel,
		" ",
		sessionDisplay,
		" ",
		indicatorStyle,
	)

	// CHANGED 2025-10-02 12:21 - Minimal spacing, show dropdown when open - Problem: Too much vertical gap
	if m.sessionDropdownOpen && len(m.sessions) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, sessionRow, m.renderSessionDropdown(width))
	}

	// When closed, just return session row (no gap)
	return sessionRow
}

func (m model) renderSessionDropdown(width int) string {
	maxDropdownHeight := 8
	dropdownContent := make([]string, 0, min(len(m.sessions), maxDropdownHeight))

	start := 0
	end := len(m.sessions)

	// Scroll logic if too many sessions
	if len(m.sessions) > maxDropdownHeight {
		if m.sessionIndex >= maxDropdownHeight/2 {
			start = m.sessionIndex - maxDropdownHeight/2
			end = start + maxDropdownHeight
			if end > len(m.sessions) {
				end = len(m.sessions)
				start = end - maxDropdownHeight
			}
		} else {
			end = maxDropdownHeight
		}
	}

	for i := start; i < end; i++ {
		session := m.sessions[i]
		sessionText := fmt.Sprintf("%s (%s)", session.Name, session.Type)

		var sessionStyle lipgloss.Style
		if i == m.sessionIndex {
			// CHANGED 2025-10-01 22:35 - Use BgBase only - Problem: Different backgrounds cause bleeding
			sessionStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Background(BgBase).
				Bold(true).
				Padding(0, 1)
		} else {
			// CHANGED 2025-10-01 22:35 - Use BgBase only - Problem: Different backgrounds cause bleeding
			sessionStyle = lipgloss.NewStyle().
				Foreground(FgSecondary).
				Background(BgBase).
				Padding(0, 1)
		}

		dropdownContent = append(dropdownContent, sessionStyle.Render(sessionText))
	}

	// Add scroll indicators if needed
	if start > 0 {
		scrollUp := lipgloss.NewStyle().Foreground(FgMuted).Render("  ↑ more above")
		dropdownContent = append([]string{scrollUp}, dropdownContent...)
	}
	if end < len(m.sessions) {
		scrollDown := lipgloss.NewStyle().Foreground(FgMuted).Render("  ↓ more below")
		dropdownContent = append(dropdownContent, scrollDown)
	}

	dropdown := lipgloss.JoinVertical(lipgloss.Left, dropdownContent...)

	// Add border to dropdown
	// CHANGED 2025-10-01 18:00 - Remove Width() to prevent background bleeding - Problem: Background color extending beyond dropdown content
	// CHANGED 2025-10-01 21:30 - Use BgBase explicitly - Problem: All backgrounds must be identical
	// CHANGED 2025-10-02 04:00 - Add left margin to align with session text - Problem: Dropdown not aligned with Session: label
	dropdownStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderFocus).
		Background(BgBase).
		Padding(0, 1).
		MarginLeft(11) // "Session:" label is 10 chars wide + 1 space

	return dropdownStyle.Render(dropdown)
}

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
			// CHANGED 2025-10-01 22:35 - Use BgBase only - Problem: Different backgrounds cause bleeding
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
	// CHANGED 2025-10-01 21:30 - Use BgBase explicitly - Problem: All backgrounds must be identical
	powerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Danger).
		Background(BgBase).
		Padding(2, 4)

	powermenu := powerStyle.Render(innerContent)

	// CHANGED 2025-10-06 - Return power menu without Place(), let View() handle centering - Problem: Place() creates uncolored padding causing ghosting
	return powermenu
}

// CHANGED 2025-09-30 - Added menu view rendering with CRUSH-style ASCII framing
func (m model) renderMenuView(termWidth, termHeight int) string {
	var content []string

	// CHANGED 2025-09-30 14:58 - Dynamic menu titles based on mode - Problem: Need different titles for main menu vs submenus
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
		title = "/// Backgrounds ///" // CHANGED 2025-10-03 17:15 - Add backgrounds title
	case ModeWallpaperSubmenu:
		title = "/// Wallpapers ////" // CHANGED 2025-10-03 17:15 - Add wallpapers title
	default:
		title = "///// Menu //////"
	}

	// Will calculate title width after rendering menu items
	content = append(content, "") // Placeholder for title
	content = append(content, "")

	// CHANGED 2025-10-04 - Add pagination for long menus - Problem: Wallpaper list too long, show max 9 items
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
		// CHANGED 2025-10-01 15:30 - Widened menu to 32 - Problem: TransIsHardJob wraps at width 24
		var style lipgloss.Style
		if i == m.menuIndex {
			// CHANGED 2025-10-01 22:35 - Use BgBase only - Problem: Different backgrounds cause bleeding
			// CHANGED 2025-10-06 - Removed Align() - Problem: Align() causes ghosting in fullscreen kitty
			style = lipgloss.NewStyle().
				Bold(true).
				Foreground(Accent).
				Background(BgBase).
				Padding(0, 2)
		} else {
			// CHANGED 2025-10-06 - Removed Align() - Problem: Align() causes ghosting in fullscreen kitty
			style = lipgloss.NewStyle().
				Foreground(FgSecondary).
				Background(BgBase).
				Padding(0, 2)
		}
		content = append(content, style.Render(option))
	}

	// Show scroll indicator at bottom if not at end
	if totalItems > maxVisibleItems && endIdx < totalItems {
		// CHANGED 2025-10-06 - Removed Align(Center) - Problem: Align causes ghosting in fullscreen kitty
		indicatorStyle := lipgloss.NewStyle().Foreground(FgMuted)
		content = append(content, indicatorStyle.Render("▼ More below ▼"))
	}

	// Help
	content = append(content, "")
	// CHANGED 2025-10-06 - Removed Align(Center) - Problem: Align causes ghosting
	helpStyle := lipgloss.NewStyle().Foreground(FgMuted)
	content = append(content, helpStyle.Render("↑↓ Navigate • Enter Select • Esc Close"))

	// CHANGED 2025-10-06 - Calculate title width from widest rendered content line - Problem: Title needs to center within menu width
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

	// CHANGED 2025-10-06 - Use Left instead of Center - Problem: Center alignment causes ghosting
	innerContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Create bordered menu with ASCII-style framing
	// CHANGED 2025-10-01 21:30 - Use BgBase explicitly - Problem: All backgrounds must be identical
	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Accent).
		Background(BgBase).
		Padding(2, 4)

	menu := menuStyle.Render(innerContent)

	// CHANGED 2025-10-06 - Return menu without Place(), let View() handle centering - Problem: Place() creates uncolored padding causing ghosting in fullscreen
	return menu
}

// CHANGED 2025-10-01 14:17 - Added F5 release notes view rendering function - Problem: User requested F5 release notes popup functionality
// CHANGED 2025-10-01 15:15 - Updated with NOTES_POPUP.txt format - Problem: User wants specific format with ASCII art header
func (m model) renderReleaseNotesView(termWidth, termHeight int) string {
	// CHANGED 2025-10-03 16:35 - Rewrite to match NOTES popup format - Problem: Needed popup-style rendering like Menu/Power

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

	// CHANGED 2025-10-03 16:50 - Define width first, then build content - Problem: Need width for centering individual elements
	popupWidth := min(100, termWidth-8)

	// Build content
	var contentLines []string

	// CHANGED 2025-10-03 17:10 - Hardcode ASCII center position, left-align all text - Problem: Lipgloss centering not working, just manually position ASCII
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

	// CHANGED 2025-10-03 16:50 - Remove global Align, center individual elements instead - Problem: Global center broke ASCII art
	// Create bordered box (matching Menu/Power popup style)

	notesStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Background(BgBase).
		Padding(2, 3).
		Width(popupWidth)

	notesBox := notesStyle.Render(innerContent)

	// CHANGED 2025-10-06 - Return notes without Place(), let View() handle centering - Problem: Place() creates uncolored padding causing ghosting
	return notesBox
}

func (m model) renderMainHelp() string {
	switch m.mode {
	case ModeLogin, ModePassword:
		// CHANGED 2025-10-03 17:15 - Reorder function keys to F1-F4 logical sequence - Problem: User wanted Menu=F1, Sessions=F2, Notes=F3, Power=F4
		// CHANGED 2025-10-07 19:20 - Add Page Up/Down navigation for ASCII variants - Problem: User wants to cycle ASCII art
		if m.sessionDropdownOpen {
			return "↑↓ Navigate • ⇞⇟ ASCII • Enter Select • Esc Close • Tab Focus • F1 Menu • F2 Sessions • F3 Notes • F4 Power"
		}
		return "Tab Focus • ↑↓ Sessions • ⇞⇟ ASCII • F1 Menu • F2 Sessions • F3 Notes • F4 Power • Enter Continue"
	case ModeLoading:
		return "Please wait..."
	default:
		return "Ctrl+C Quit"
	}
}

// Animation helper functions
func (m model) getAnimatedColor() color.Color {
	colors := []color.Color{Primary, Secondary, Accent}
	index := (m.animationFrame / 20) % len(colors)
	return colors[index]
}

func (m model) getAnimatedBorderColor() color.Color {
	colors := []color.Color{BorderDefault, Primary, Secondary}
	index := (m.borderFrame / 5) % len(colors)
	return colors[index]
}

func (m model) getFocusColor(target FocusState) color.Color {
	if m.focusState == target {
		return Primary
	}
	return FgSecondary
}

func (m model) getSessionArt(sessionName string) string {
	// CHANGED 2025-09-30 15:20 - Use pre-made ASCII from config files - Problem: User wants clean ASCII from .conf files instead of generated figlet
	return m.getSessionASCII()
}

func (m model) authenticate(username, password string) tea.Cmd {
	return func() tea.Msg {
		// CHANGED 2025-10-05 - Add nil check for IPC client - Problem: Nil pointer defense-in-depth
		if m.ipcClient == nil {
			return fmt.Errorf("IPC client not initialized - greeter must be run by greetd")
		}

		// Create session
		if err := m.ipcClient.CreateSession(username); err != nil {
			return err
		}
		// Receive auth message
		resp, err := m.ipcClient.ReceiveResponse()
		if err != nil {
			return err
		}

		// CHANGED 2025-10-05 - Handle Error response from CreateSession - Problem: User might not exist or other early errors
		if errResp, ok := resp.(ipc.Error); ok {
			return fmt.Errorf("authentication failed: %s - %s", errResp.ErrorType, errResp.Description)
		}

		if _, ok := resp.(ipc.AuthMessage); ok {
			if m.config.Debug {
				logDebug(" Received auth message")
			}
			// Send password as response
			if err := m.ipcClient.PostAuthMessageResponse(&password); err != nil {
				return err
			}
			// Receive success or error
			resp, err := m.ipcClient.ReceiveResponse()
			if err != nil {
				return err
			}

			// CHANGED 2025-10-05 - Handle Error response (wrong password) - Problem: Wrong password returns Error, not Success
			if errResp, ok := resp.(ipc.Error); ok {
				return fmt.Errorf("authentication failed: %s - %s", errResp.ErrorType, errResp.Description)
			}

			if _, ok := resp.(ipc.Success); ok {
				// Start session
				if m.selectedSession == nil {
					return fmt.Errorf("no session selected")
				}
				cmd := []string{m.selectedSession.Exec}
				env := []string{} // Can be populated if needed
				if err := m.ipcClient.StartSession(cmd, env); err != nil {
					return err
				}
				return "success"
			} else {
				return fmt.Errorf("expected success or error, got %T", resp)
			}
		} else {
			return fmt.Errorf("expected auth message or error, got %T", resp)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CHANGED 2025-10-02 11:10 - Add ANSI stripping for compositing - Problem: Need to check visible chars in UI lines
// REFACTORED 2025-10-02 - Moved to internal/ui/utils.go
func stripAnsi(s string) string {
	return ui.StripAnsi(s)
}

// CHANGED 2025-10-02 11:13 - Character-by-character line merging - Problem: Need to overlay UI on fire per-character
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

func main() {
	// CHANGED 2025-10-01 - Removed SetColorProfile - not available in lipgloss v2
	// Color profile is now automatically detected via colorprofile package

	// Load configuration from file first
	// CHANGED 2025-09-29 - Added config file loading with command-line override support
	fileConfig, err := loadConfig("bubble-greet.conf")
	if err != nil {
		fmt.Printf("Warning: Could not load config file: %v\n", err)
	}

	// Define command-line flags with config file values as defaults
	config := fileConfig

	flag.BoolVar(&config.TestMode, "test", false, "Enable test mode (no actual authentication)")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug output")
	flag.StringVar(&config.Greeting, "greeting", "", "Custom greeting message")
	flag.BoolVar(&config.ShowTime, "time", false, "Display current time")
	flag.BoolVar(&config.ShowIssue, "issue", false, "Display system issue file")
	flag.IntVar(&config.Width, "width", 80, "Width of the main prompt")
	flag.StringVar(&config.ThemeName, "theme", "", "Theme name to use")
	flag.StringVar(&config.FontPath, "font", config.FontPath, "Path to figlet font file")

	// Add help text
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "bubble-greet - A beautiful terminal greeter for greetd\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  Config file: bubble-greet.conf\n")
		fmt.Fprintf(os.Stderr, "  Font: %s\n", config.FontPath)
		fmt.Fprintf(os.Stderr, "  Palettes: %d custom palettes loaded\n", len(config.Palettes))
		fmt.Fprintf(os.Stderr, "\nKey Bindings:\n")
		fmt.Fprintf(os.Stderr, "  Tab       Cycle focus between elements\n")
		fmt.Fprintf(os.Stderr, "  ↑↓       Navigate sessions when focused\n")
		fmt.Fprintf(os.Stderr, "  F3        Toggle session dropdown\n")
		fmt.Fprintf(os.Stderr, "  F4        Power menu\n")
		fmt.Fprintf(os.Stderr, "  Enter     Continue to next step\n")
		fmt.Fprintf(os.Stderr, "  Esc       Cancel/go back\n")
		fmt.Fprintf(os.Stderr, "  Ctrl+C    Quit\n")
	}

	flag.Parse()

	// CHANGED 2025-10-06 - Initialize debug logging - Problem: Need persistent logs
	initDebugLog()
	logDebug("=== sysc-greet started ===")
	logDebug("Version: sysc-greet greeter")
	logDebug("Test mode: %v", config.TestMode)
	logDebug("Debug mode: %v", config.Debug)
	logDebug("Theme: %s", config.ThemeName)
	logDebug("GREETD_SOCK: %s", os.Getenv("GREETD_SOCK"))
	logDebug("WAYLAND_DISPLAY: %s", os.Getenv("WAYLAND_DISPLAY"))
	logDebug("XDG_RUNTIME_DIR: %s", os.Getenv("XDG_RUNTIME_DIR"))

	if config.Debug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Debug log: /tmp/sysc-greet-debug.log\n")
	}

	// Initialize Bubble Tea program with proper screen management
	// CHANGED 2025-09-29 - Handle TTY access gracefully for different environments
	opts := []tea.ProgramOption{}

	// Check if we can access TTY before using alt screen
	if _, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
		// No TTY access - use basic program options
		if config.Debug {
			logDebug(" No TTY access, running without alt screen")
		}
	} else {
		// TTY available - use full screen features
		opts = append(opts, tea.WithAltScreen())
		if !config.TestMode {
			opts = append(opts, tea.WithMouseCellMotion())
		}
	}

	p := tea.NewProgram(initialModel(config), opts...)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// CHANGED 2025-10-02 03:50 - ASCII-2: Fancy template-based border
func (m model) renderASCII2BorderLayout(termWidth, termHeight int) string {
	// CHANGED 2025-10-03 16:00 - Complete rewrite to match ASCII_TEMPLATE.png reference - Problem: Wrong spacing, gradients directly adjacent to content
	// Fancy gradient border matching the reference template with proper wide spacing

	// CHANGED 2025-10-07 19:30 - Calculate border dynamically based on ASCII art width - Problem: Fixed width causes line shifting
	// Build content section FIRST to determine required width
	var contentLines []string

	// CHANGED 2025-10-03 15:45 - Split ASCII art into lines to prevent border corruption - Problem: Multi-line ASCII art treated as single line causing warping
	// CHANGED 2025-10-09 20:25 - Enforce mandatory 2-line gap between ASCII and input fields - Problem: Smaller ASCII positioned too high
	// WM/Session ASCII art
	if m.selectedSession != nil {
		art := m.getSessionASCII()
		if art != "" {
			// Split multi-line ASCII art into separate lines
			artLines := strings.Split(art, "\n")
			for _, line := range artLines {
				contentLines = append(contentLines, line)
			}
			// MANDATORY 2-line gap after ASCII art
			contentLines = append(contentLines, "")
			contentLines = append(contentLines, "")
		}
	}

	// Session display
	if len(m.sessions) > 0 && m.selectedSession != nil {
		sessionText := fmt.Sprintf("[ %s (%s) ]", m.selectedSession.Name, m.selectedSession.Type)
		sessionLine := lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Render(sessionText)
		contentLines = append(contentLines, sessionLine)
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

	// CHANGED 2025-10-07 19:30 - Calculate border width based on actual content - Problem: Dynamic sizing to fit ASCII art
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
	if borderWidth > min(120, termWidth-20) {
		borderWidth = min(120, termWidth-20) // Maximum width
	}

	// Now render borders and content
	var lines []string

	// CHANGED 2025-10-03 16:00 - Recreate top border matching template corners
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

	// CHANGED 2025-10-03 16:15 - Broken border design - gradient only at top, clean middle, gradient at bottom - Problem: Template shows NO side borders in content area
	// Top gradient fade: ▓ → ▒ → ░ (only 2 lines for shorter height)
	gradientChars := []string{"▓", "▒"}
	gradientColors := []color.Color{Secondary, Accent}

	for i, char := range gradientChars {
		gradLine := char + strings.Repeat(" ", borderWidth-2) + char
		lines = append(lines, lipgloss.NewStyle().Foreground(gradientColors[i]).Render(gradLine))
	}

	// CHANGED 2025-10-03 16:15 - Clean content area with NO side borders - Problem: Side borders interfere with ASCII art
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

	// CHANGED 2025-10-03 16:00 - Bottom decorations matching template
	// Bottom corner fade (mirroring top)
	bottomLine3 := "█▄" + strings.Repeat(" ", borderWidth-4) + "▄█"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(bottomLine3))

	bottomLine2 := "▀▄▄" + strings.Repeat(" ", borderWidth-6) + "▄▄▀"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(bottomLine2))

	bottomLine1 := "▀▀▀▀" + strings.Repeat("█", borderWidth-8) + "▀▀▀▀"
	lines = append(lines, lipgloss.NewStyle().Foreground(Primary).Render(bottomLine1))

	// CHANGED 2025-10-03 15:45 - Add help text below border - Problem: Missing F2/F3/F4/F5 key bindings help
	// Build bordered content
	borderedContent := strings.Join(lines, "\n")

	// Add help text below border
	// CHANGED 2025-10-06 - Removed Width(termWidth) - Problem: Full-width help text stretches border to screen edges
	helpText := m.renderMainHelp()
	helpStyle := lipgloss.NewStyle().
		Foreground(FgMuted)

	// Join border and help text vertically
	contentWithHelp := lipgloss.JoinVertical(lipgloss.Center, borderedContent, "", helpStyle.Render(helpText))

	// CHANGED 2025-10-06 - Return content without Place(), let View() handle centering - Problem: Place() creates uncolored padding
	return contentWithHelp
}
