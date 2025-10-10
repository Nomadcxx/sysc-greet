package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// ScreensaverConfig holds screensaver configuration
type ScreensaverConfig struct {
	IdleTimeout int    // Idle timeout in minutes
	TimeFormat  string // Time format string
	DateFormat  string // Date format string
	ASCII       string // ASCII art
}

// loadScreensaverConfig loads screensaver configuration
func loadScreensaverConfig() ScreensaverConfig {
	// Default config
	config := ScreensaverConfig{
		IdleTimeout: 5,
		TimeFormat:  "15:04:05",
		DateFormat:  "Monday, January 2, 2006",
		ASCII: `▄▀▀▀▀ █   █ ▄▀▀▀▀ ▄▀▀▀▀    ▄▀    ▄▀
 ▀▀▀▄ ▀▀▀▀█  ▀▀▀▄ █      ▄▀    ▄▀
▀▀▀▀  ▀▀▀▀▀ ▀▀▀▀   ▀▀▀▀ ▀     ▀
//  SEE YOU SPACE COWBOY //`,
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
	var asciiLines []string
	inASCII := false

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Check if we're starting ASCII section
		if strings.HasPrefix(line, "ascii_1=") {
			inASCII = true
			// Check if there's content on same line
			content := strings.TrimPrefix(line, "ascii_1=")
			if content != "" {
				asciiLines = append(asciiLines, content)
			}
			continue
		}

		// If in ASCII section, collect lines
		if inASCII {
			asciiLines = append(asciiLines, line)
			continue
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
		}
	}

	if len(asciiLines) > 0 {
		config.ASCII = strings.Join(asciiLines, "\n")
	}

	return config
}

// renderScreensaverView displays the screensaver with ASCII art and clock
func renderScreensaverView(m model, termWidth, termHeight int) string {
	// Load screensaver config
	config := loadScreensaverConfig()

	// Get current time and date
	currentTime := m.screensaverTime
	timeStr := currentTime.Format(config.TimeFormat)
	dateStr := currentTime.Format(config.DateFormat)

	// Center the content vertically and horizontally
	contentLines := []string{
		config.ASCII,
		"",
		timeStr,
		dateStr,
	}

	// Calculate vertical centering
	totalLines := len(contentLines)
	verticalPadding := (termHeight - totalLines) / 2

	// Build the full screen content
	var lines []string

	// Add top padding
	for i := 0; i < verticalPadding; i++ {
		lines = append(lines, "")
	}

	// Add content lines (centered horizontally)
	for _, line := range contentLines {
		// Center each line horizontally
		lineWidth := len([]rune(line))
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
