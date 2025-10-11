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
	IdleTimeout      int      // Idle timeout in minutes
	TimeFormat       string   // Time format string
	DateFormat       string   // Date format string
	ASCIIVariants    []string // Multiple ASCII art variants
	ClockSize        string   // Clock size: "small", "medium", "large"
	AnimateOnStart   bool     // Enable animation when screensaver starts
	AnimationType    string   // Animation type: "print", "none"
	AnimationSpeed   int      // Animation speed in milliseconds per character
}

// loadScreensaverConfig loads screensaver configuration
func loadScreensaverConfig() ScreensaverConfig {
	// Default config with one ASCII variant
	defaultASCII := `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀
//  SEE YOU SPACE COWBOY //`

	config := ScreensaverConfig{
		IdleTimeout:    5,
		TimeFormat:     "3:04:05 PM",
		DateFormat:     "Monday, January 2, 2006",
		ASCIIVariants:  []string{defaultASCII},
		ClockSize:      "medium",
		AnimateOnStart: true,
		AnimationType:  "print",
		AnimationSpeed: 20,
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
		case "animate_on_start":
			config.AnimateOnStart = (strings.ToLower(value) == "true")
		case "animation_type":
			config.AnimationType = value
		case "animation_speed":
			if speed, err := strconv.Atoi(value); err == nil {
				config.AnimationSpeed = speed
			}
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

// Large ASCII digits generated with figlet kompaktblk font
var largeDigits = map[rune][]string{
	'0': {
		"▄▀▀█▄ ",
		"█▄▀ █ ",
		" ▀▀▀  ",
	},
	'1': {
		" ▄█   ",
		"  █   ",
		"▀▀▀▀▀ ",
	},
	'2': {
		"▀▀▀▀█ ",
		"█▀▀▀▀ ",
		"▀▀▀▀▀ ",
	},
	'3': {
		"▀▀▀▀▄ ",
		"  ▀▀▄ ",
		"▀▀▀▀  ",
	},
	'4': {
		"█   █ ",
		"▀▀▀▀█ ",
		"    ▀ ",
	},
	'5': {
		"█▀▀▀▀ ",
		"▀▀▀▀█ ",
		"▀▀▀▀▀ ",
	},
	'6': {
		"█▀▀▀▀ ",
		"█▀▀▀█ ",
		"▀▀▀▀▀ ",
	},
	'7': {
		"▀▀▀▀█ ",
		"   █▀ ",
		"   ▀  ",
	},
	'8': {
		"█▀▀▀█ ",
		"█▀▀▀█ ",
		"▀▀▀▀▀ ",
	},
	'9': {
		"█▀▀▀█ ",
		"▀▀▀▀█ ",
		"▀▀▀▀▀ ",
	},
	':': {
		"  ▄   ",
		"      ",
		"  ▄   ",
	},
	' ': {
		"      ",
		"      ",
		"      ",
	},
	'A': {
		"▄▀▀▀▄ ",
		"█▀▀▀█ ",
		"▀   ▀ ",
	},
	'M': {
		"█▀▄▀█ ",
		"█   █ ",
		"▀   ▀ ",
	},
	'P': {
		"█▀▀▀▄ ",
		"█▀▀▀  ",
		"▀     ",
	},
}

// Medium ASCII digit patterns for clock display
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

// renderLargeClock renders time string using ASCII digit patterns
func renderLargeClock(timeStr string, size string) []string {
	switch size {
	case "large":
		// Get the height from first digit
		if len(largeDigits['0']) == 0 {
			return []string{timeStr}
		}
		height := len(largeDigits['0'])

		// Build each line of the clock
		var lines []string
		for row := 0; row < height; row++ {
			var line strings.Builder
			for _, ch := range timeStr {
				digitLines, ok := largeDigits[ch]
				if !ok {
					// Unknown character, use space
					digitLines = largeDigits[' ']
				}
				if row < len(digitLines) {
					line.WriteString(digitLines[row])
				}
			}
			lines = append(lines, line.String())
		}
		return lines

	case "medium":
		// Get the height from first digit
		if len(mediumDigits['0']) == 0 {
			return []string{timeStr}
		}
		height := len(mediumDigits['0'])

		// Build each line of the clock
		var lines []string
		for row := 0; row < height; row++ {
			var line strings.Builder
			for _, ch := range timeStr {
				digitLines, ok := mediumDigits[ch]
				if !ok {
					// Unknown character, use space
					digitLines = mediumDigits[' ']
				}
				if row < len(digitLines) && 1 < len(digitLines[row]) {
					line.WriteString(digitLines[row][1]) // Use second variant
					line.WriteString(" ")                // Space between digits
				}
			}
			lines = append(lines, line.String())
		}
		return lines

	default:
		// Small - just return the plain string
		return []string{timeStr}
	}
}

// renderScreensaverView renders the screensaver with ASCII art, clock, and date
func renderScreensaverView(m model, termWidth, termHeight int) string {
	config := loadScreensaverConfig()

	// Get theme-specific color palette
	palette := animations.GetScreensaverPalette(m.currentTheme)
	// palette: [background, ascii_primary, ascii_secondary, clock_primary, clock_secondary, date_color]

	// Create lipgloss styles using theme colors
	asciiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[1]))
	clockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[3])).Bold(true)
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[5]))
	printHeadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette[2])).Bold(true)

	// Cycle through ASCII variants every 5 minutes
	minutesElapsed := int(time.Since(m.idleTimer).Minutes())
	variantIndex := (minutesElapsed / 5) % len(config.ASCIIVariants)
	selectedASCII := config.ASCIIVariants[variantIndex]

	// Get current time and date
	currentTime := m.screensaverTime
	timeStr := currentTime.Format(config.TimeFormat)
	// Pad single-digit hours for consistent width in 12-hour format
	if strings.Contains(config.TimeFormat, "3:04") && len(timeStr) > 0 && timeStr[0] != '1' && timeStr[1] == ':' {
		timeStr = " " + timeStr
	}
	dateStr := strings.ToUpper(currentTime.Format(config.DateFormat))

	clockLines := renderLargeClock(timeStr, config.ClockSize)

	// Build content lines: ASCII art, blank line, clock, date
	var contentLines []string

	// Show print effect animation if enabled and in progress
	if config.AnimateOnStart && config.AnimationType == "print" && m.screensaverPrint != nil && !m.screensaverPrint.IsComplete() {
		// Animation in progress - show partially revealed ASCII
		visibleLines := m.screensaverPrint.GetVisibleLines()
		for _, line := range visibleLines {
			// Apply styling with print head highlighted
			styledLine := asciiStyle.Render(line)
			// Highlight the print head character if present
			if strings.Contains(line, "█") {
				// Replace print head with styled version
				parts := strings.Split(line, "█")
				if len(parts) == 2 {
					styledLine = asciiStyle.Render(parts[0]) + printHeadStyle.Render("█") + asciiStyle.Render(parts[1])
				}
			}
			contentLines = append(contentLines, styledLine)
		}
	} else {
		// No animation or animation complete - show full ASCII
		asciiLines := strings.Split(selectedASCII, "\n")
		for _, line := range asciiLines {
			contentLines = append(contentLines, asciiStyle.Render(line))
		}
	}
	contentLines = append(contentLines, "") // Blank line

	// Add clock lines with theme color
	for _, line := range clockLines {
		contentLines = append(contentLines, clockStyle.Render(line))
	}
	contentLines = append(contentLines, "") // Blank line

	// Add date with theme color
	contentLines = append(contentLines, dateStyle.Render(dateStr))

	// Join all content with center alignment
	content := lipgloss.JoinVertical(lipgloss.Center, contentLines...)

	// Use lipgloss Place to center both horizontally and vertically
	centeredContent := lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)

	return centeredContent
}

// handleScreensaverInput handles input in screensaver mode
func handleScreensaverInput(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	// Exit screensaver on any key press
	m.mode = ModeLogin
	m.idleTimer = time.Now()
	return m, nil
}
