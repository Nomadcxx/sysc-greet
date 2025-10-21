package animations

import (
"strings"
"time"
)

// PrintEffect creates a typewriter/printer effect for ASCII art
// Inspired by TTE's print effect - prints line by line with a print head
// Used by both screensaver and ASCII effects
type PrintEffect struct {
lines       []string      // ASCII art lines
currentLine int           // Current line being printed
currentCol  int           // Current column position in line
revealed    []string      // Fully revealed lines
lastUpdate  time.Time     // Last update time
charDelay   time.Duration // Delay between printing characters
complete    bool          // Animation complete
}

// NewPrintEffect creates a new print effect for ASCII art
func NewPrintEffect(asciiArt string, charDelay time.Duration) *PrintEffect {
lines := strings.Split(asciiArt, "\n")

// Remove empty trailing lines
for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
lines = lines[:len(lines)-1]
}

return &PrintEffect{
lines:       lines,
currentLine: 0,
currentCol:  0,
revealed:    []string{},
lastUpdate:  time.Now(),
charDelay:   charDelay,
complete:    false,
}
}

// Tick advances the print effect animation based on current time
// Animation loops automatically after completion
func (p *PrintEffect) Tick(currentTime time.Time) {
// Check if animation is complete
if p.currentLine >= len(p.lines) {
if !p.complete {
p.complete = true
p.lastUpdate = currentTime
}
// Wait 2 seconds after completion before restarting
if currentTime.Sub(p.lastUpdate) >= time.Second*2 {
// Restart the animation
p.currentLine = 0
p.currentCol = 0
p.revealed = []string{}
p.complete = false
p.lastUpdate = currentTime
}
return
}

// Check if enough time has passed to print next character
if currentTime.Sub(p.lastUpdate) >= p.charDelay {
currentLineText := p.lines[p.currentLine]

if p.currentCol < len([]rune(currentLineText)) {
// Print next character (use rune count for proper multi-byte handling)
p.currentCol++
p.lastUpdate = currentTime
} else {
// Line complete - move to next line
p.revealed = append(p.revealed, currentLineText)
p.currentLine++
p.currentCol = 0
p.lastUpdate = currentTime
}
}
}

// GetVisibleLines returns the currently visible lines (for rendering)
func (p *PrintEffect) GetVisibleLines() []string {
if p.complete {
return p.lines
}

var result []string
result = append(result, p.revealed...)

// Add currently printing line with trailing effect
if p.currentLine < len(p.lines) {
currentLineText := p.lines[p.currentLine]
runes := []rune(currentLineText)

var currentLine string

if p.currentCol > 0 {
// Show revealed portion
if p.currentCol <= len(runes) {
currentLine = string(runes[:p.currentCol])
} else {
currentLine = currentLineText
}

// Add trail effect: show shading blocks trailing the print head
// Trail: ░▒▓█ (3 shade blocks before print head)
currentLine += "░▒▓"
} else {
// Just starting - show partial trail
currentLine = "▒▓"
}

// Add print head character
currentLine += "█"

result = append(result, currentLine)
}

return result
}

// IsComplete returns whether the animation is complete
func (p *PrintEffect) IsComplete() bool {
return p.complete
}

// Reset restarts the print effect animation
func (p *PrintEffect) Reset(asciiArt string) {
lines := strings.Split(asciiArt, "\n")
for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
lines = lines[:len(lines)-1]
}

p.lines = lines
p.currentLine = 0
p.currentCol = 0
p.revealed = []string{}
p.lastUpdate = time.Now()
p.complete = false
}
