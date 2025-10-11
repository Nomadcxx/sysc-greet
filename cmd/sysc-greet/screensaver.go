package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// ScreensaverConfig holds screensaver configuration
type ScreensaverConfig struct {
	IdleTimeout   int      // Idle timeout in minutes
	TimeFormat    string   // Time format string
	DateFormat    string   // Date format string
	ASCIIVariants []string // Multiple ASCII art variants
	ClockSize     string   // Clock size: "small", "medium", "large"
}

// loadScreensaverConfig loads screensaver configuration
func loadScreensaverConfig() ScreensaverConfig {
	// Default config with one ASCII variant
	defaultASCII := `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀
//  SEE YOU SPACE COWBOY //`

	config := ScreensaverConfig{
		IdleTimeout:   5,
		TimeFormat:    "15:04:05",
		DateFormat:    "Monday, January 2, 2006",
		ASCIIVariants: []string{defaultASCII},
		ClockSize:     "medium",
	}

	// Try to load from config file
	paths := []string{
		"/usr/share/sysc-greet/ascii_configs/screensaver.conf",
		"ascii_configs/screensaver.conf",
		"screensaver.conf",
	}

	var file *os.File
	var err error
	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		return config
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentASCII []string
	inASCII := false

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments (but not inside ASCII sections)
		if !inASCII && (strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "") {
			continue
		}

		// Check if we're starting a new ASCII section (ascii_1=, ascii_2=, etc.)
		if strings.HasPrefix(line, "ascii_") && strings.Contains(line, "=") {
			// Save previous ASCII if we have one
			if len(currentASCII) > 0 {
				config.ASCIIVariants = append(config.ASCIIVariants, strings.Join(currentASCII, "\n"))
				currentASCII = []string{}
			}
			inASCII = true
			// Check if there's content on same line after =
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && parts[1] != "" {
				currentASCII = append(currentASCII, parts[1])
			}
			continue
		}

		// If in ASCII section, collect lines until we hit a config key
		if inASCII {
			// Check if this line is a config key (idle_timeout=, etc.)
			if strings.Contains(line, "=") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				// This is a config line, end ASCII section
				inASCII = false
				if len(currentASCII) > 0 {
					config.ASCIIVariants = append(config.ASCIIVariants, strings.Join(currentASCII, "\n"))
					currentASCII = []string{}
				}
				// Continue to parse this line as config
			} else {
				// Still in ASCII section
				currentASCII = append(currentASCII, line)
				continue
			}
		}

		// Parse config lines
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "idle_timeout":
			if timeout, err := strconv.Atoi(value); err == nil {
				config.IdleTimeout = timeout
			}
		case "time_format":
			config.TimeFormat = value
		case "date_format":
			config.DateFormat = value
		case "clock_size":
			config.ClockSize = value
		}
	}

	// Save final ASCII variant if we have one
	if len(currentASCII) > 0 {
		config.ASCIIVariants = append(config.ASCIIVariants, strings.Join(currentASCII, "\n"))
	}

	// If we loaded variants from file, replace default
	if len(config.ASCIIVariants) > 1 {
		config.ASCIIVariants = config.ASCIIVariants[1:] // Remove default, keep loaded variants
	}

	return config
}

// CHANGED 2025-10-10 - Large ASCII digit patterns for clock display - Problem: Need large clock like clock-tui
var largeDigits = map[rune][][]string{
	'0': {
		{"███", "███", "███"},
		{"█ █", "█ █", "█ █"},
		{"█ █", "█ █", "█ █"},
		{"█ █", "█ █", "█ █"},
		{"███", "███", "███"},
	},
	'1': {
		{"  █", "  █", "  █"},
		{" ██", " ██", " ██"},
		{"  █", "  █", "  █"},
		{"  █", "  █", "  █"},
		{"███", "███", "███"},
	},
	'2': {
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"███", "███", "███"},
		{"█  ", "█  ", "█  "},
		{"███", "███", "███"},
	},
	'3': {
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"███", "███", "███"},
	},
	'4': {
		{"█ █", "█ █", "█ █"},
		{"█ █", "█ █", "█ █"},
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"  █", "  █", "  █"},
	},
	'5': {
		{"███", "███", "███"},
		{"█  ", "█  ", "█  "},
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"███", "███", "███"},
	},
	'6': {
		{"███", "███", "███"},
		{"█  ", "█  ", "█  "},
		{"███", "███", "███"},
		{"█ █", "█ █", "█ █"},
		{"███", "███", "███"},
	},
	'7': {
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"  █", "  █", "  █"},
		{"  █", "  █", "  █"},
		{"  █", "  █", "  █"},
	},
	'8': {
		{"███", "███", "███"},
		{"█ █", "█ █", "█ █"},
		{"███", "███", "███"},
		{"█ █", "█ █", "█ █"},
		{"███", "███", "███"},
	},
	'9': {
		{"███", "███", "███"},
		{"█ █", "█ █", "█ █"},
		{"███", "███", "███"},
		{"  █", "  █", "  █"},
		{"███", "███", "███"},
	},
	':': {
		{"   ", "   ", "   "},
		{" █ ", " █ ", " █ "},
		{"   ", "   ", "   "},
		{" █ ", " █ ", " █ "},
		{"   ", "   ", "   "},
	},
	' ': {
		{"   ", "   ", "   "},
		{"   ", "   ", "   "},
		{"   ", "   ", "   "},
		{"   ", "   ", "   "},
		{"   ", "   ", "   "},
	},
}

// CHANGED 2025-10-10 - Medium ASCII digit patterns - Problem: Need configurable clock sizes
var mediumDigits = map[rune][][]string{
	'0': {
		{"██", "██"},
		{"█ █", "█ █"},
		{"█ █", "█ █"},
		{"██", "██"},
	},
	'1': {
		{" █", " █"},
		{"██", "██"},
		{" █", " █"},
		{"██", "██"},
	},
	'2': {
		{"██", "██"},
		{" █", " █"},
		{"█ ", "█ "},
		{"██", "██"},
	},
	'3': {
		{"██", "██"},
		{" █", " █"},
		{" █", " █"},
		{"██", "██"},
	},
	'4': {
		{"█ █", "█ █"},
		{"██", "██"},
		{" █", " █"},
		{" █", " █"},
	},
	'5': {
		{"██", "██"},
		{"█ ", "█ "},
		{" █", " █"},
		{"██", "██"},
	},
	'6': {
		{"██", "██"},
		{"█ ", "█ "},
		{"█ █", "█ █"},
		{"██", "██"},
	},
	'7': {
		{"██", "██"},
		{" █", " █"},
		{" █", " █"},
		{" █", " █"},
	},
	'8': {
		{"██", "██"},
		{"█ █", "█ █"},
		{"█ █", "█ █"},
		{"██", "██"},
	},
	'9': {
		{"██", "██"},
		{"█ █", "█ █"},
		{" █", " █"},
		{"██", "██"},
	},
	':': {
		{"  ", "  "},
		{"█ ", "█ "},
		{"█ ", "█ "},
		{"  ", "  "},
	},
	' ': {
		{"  ", "  "},
		{"  ", "  "},
		{"  ", "  "},
		{"  ", "  "},
	},
}

// CHANGED 2025-10-10 - Render large ASCII clock - Problem: Need large clock display
func renderLargeClock(timeStr string, size string) []string {
	var digits map[rune][][]string
	var sizeIndex int

	switch size {
	case "large":
		digits = largeDigits
		sizeIndex = 2 // Use third variant (widest)
	case "medium":
		digits = mediumDigits
		sizeIndex = 1 // Use second variant
	default:
		// Small - just return the plain string
		return []string{timeStr}
	}

	// Get the height of digits
	if len(digits['0']) == 0 {
		return []string{timeStr}
	}
	height := len(digits['0'])

	// Build each line of the clock
	var lines []string
	for row := 0; row < height; row++ {
		var line strings.Builder
		for _, ch := range timeStr {
			digitLines, ok := digits[ch]
			if !ok {
				// Unknown character, use space
				digitLines = digits[' ']
			}
			if row < len(digitLines) && sizeIndex < len(digitLines[row]) {
				line.WriteString(digitLines[row][sizeIndex])
				line.WriteString(" ") // Space between digits
			}
		}
		lines = append(lines, line.String())
	}

	return lines
}

// CHANGED 2025-10-10 - Implement ASCII cycling, large clock, and theme colors - Problem: Need cycling + theme-aware screensaver
func renderScreensaverView(m model, termWidth, termHeight int) string {
	config := loadScreensaverConfig()

	// CHANGED 2025-10-10 - Get theme-specific colors - Problem: Screensaver should respect current theme like animations
	palette := animations.GetScreensaverPalette(m.currentTheme)
	// palette: [background, ascii_primary, ascii_secondary, clock_primary, clock_secondary, date_color]

	// Create lipgloss styles using theme colors
	asciiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[1]))
	clockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[3])).Bold(true)
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[5]))

	// CHANGED 2025-10-10 - Cycle through ASCII variants every 5 minutes - Problem: User requested ASCII cycling
	// Calculate which ASCII variant to show based on elapsed time since screensaver started
	minutesElapsed := int(time.Since(m.idleTimer).Minutes())
	variantIndex := (minutesElapsed / 5) % len(config.ASCIIVariants)
	selectedASCII := config.ASCIIVariants[variantIndex]

	// Get current time and date
	currentTime := m.screensaverTime
	timeStr := currentTime.Format(config.TimeFormat)
	dateStr := currentTime.Format(config.DateFormat)

	// CHANGED 2025-10-10 - Render large clock with theme colors - Problem: User wants larger clock like clock-tui
	clockLines := renderLargeClock(timeStr, config.ClockSize)

	// Build content lines: ASCII art, blank line, clock, date
	var contentLines []string

	// Add ASCII art lines (split by newline) with theme color
	asciiLines := strings.Split(selectedASCII, "\n")
	for _, line := range asciiLines {
		contentLines = append(contentLines, asciiStyle.Render(line))
	}
	contentLines = append(contentLines, "") // Blank line

	// Add clock lines with theme color
	for _, line := range clockLines {
		contentLines = append(contentLines, clockStyle.Render(line))
	}
	contentLines = append(contentLines, "") // Blank line

	// Add date with theme color
	contentLines = append(contentLines, dateStyle.Render(dateStr))

	// Calculate vertical centering
	totalLines := len(contentLines)
	verticalPadding := (termHeight - totalLines) / 2
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	// Build the full screen content
	var lines []string

	// Add top padding
	for i := 0; i < verticalPadding; i++ {
		lines = append(lines, "")
	}

	// CHANGED 2025-10-10 - Improved centering for multi-line content - Problem: Verify ASCII and time centered
	for _, line := range contentLines {
		// Center each line horizontally using lipgloss width for proper styled text
		lineWidth := lipgloss.Width(line)
		horizontalPadding := (termWidth - lineWidth) / 2
		if horizontalPadding > 0 {
			centeredLine := strings.Repeat(" ", horizontalPadding) + line
			lines = append(lines, centeredLine)
		} else {
			lines = append(lines, line)
		}
	}

	// Add bottom padding
	for i := len(lines); i < termHeight; i++ {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// handleScreensaverInput handles input in screensaver mode
func handleScreensaverInput(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	// Exit screensaver on any key press
	m.mode = ModeLogin
	m.idleTimer = time.Now()
	return m, nil
}
