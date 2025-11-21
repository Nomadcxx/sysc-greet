package animations

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// FireEffect implements PSX DOOM-style fire algorithm
type FireEffect struct {
	width   int      // Terminal width
	height  int      // Terminal height
	buffer  []int    // Heat values (0-36), size = width * height
	palette []string // Hex color codes from theme
	chars   []rune   // Fire characters for density
}

// NewFireEffect creates a new fire effect with given dimensions and theme palette
func NewFireEffect(width, height int, palette []string) *FireEffect {
	f := &FireEffect{
		width:   width,
		height:  height,
		palette: palette,
		chars:   []rune{' ', '░', '▒', '▓', '█'},
	}
	f.init()
	return f
}

// Initialize fire buffer with bottom row as heat source
func (f *FireEffect) init() {
	f.buffer = make([]int, f.width*f.height)

	// Set bottom row to maximum heat (fire source)
	for i := 0; i < f.width; i++ {
		f.buffer[(f.height-1)*f.width+i] = 36
	}
}

// UpdatePalette changes the fire color palette (for theme switching)
func (f *FireEffect) UpdatePalette(palette []string) {
	f.palette = palette
}

// Resize reinitializes the fire effect with new dimensions
func (f *FireEffect) Resize(width, height int) {
	f.width = width
	f.height = height
	f.init()
}

// spreadFire propagates heat upward with random decay
func (f *FireEffect) spreadFire(from int) {
	// Random horizontal offset (0-3) for chaos
	offset := rand.Intn(4)
	to := from - f.width - offset + 1

	// Bounds check
	if to < 0 || to >= len(f.buffer) {
		return
	}

	// Calculate target row
	toY := to / f.width
	hardLimit := f.height / 10          // Top 10% - absolute no-go zone
	fadeZoneStart := (f.height * 4) / 5 // Top 80% - start heavy decay

	// Hard limit - no propagation into top 10%
	if toY < hardLimit {
		return
	}

	// Random decay (0 or 1)
	decay := rand.Intn(2)

	// Aggressive decay in fade zone (between 10% and 80% from top)
	if toY < fadeZoneStart {
		// Add 2-6 extra decay for smooth gradient fade
		decay += rand.Intn(5) + 2
	}

	newHeat := f.buffer[from] - decay
	if newHeat < 0 {
		newHeat = 0
	}

	f.buffer[to] = newHeat
}

// Update advances the fire simulation by one frame
func (f *FireEffect) Update(frame int) {
	// Process all pixels from bottom to top
	// (Fire spreads upward, must process bottom row first)
	for y := f.height - 1; y > 0; y-- {
		for x := 0; x < f.width; x++ {
			index := y*f.width + x
			f.spreadFire(index)
		}
	}
}

// Render converts the fire buffer to colored text output
func (f *FireEffect) Render() string {
	var lines []string

	// Render across full height - low heat at top will naturally fade to black/background
	for y := 0; y < f.height; y++ {
		var line strings.Builder
		for x := 0; x < f.width; x++ {
			heat := f.buffer[y*f.width+x]

			// Skip very low heat (natural fade to background)
			if heat < 3 {
				line.WriteString(" ")
				continue
			}

			// Map heat to character (0-36 heat → 5 chars)
			charIndex := heat / 7
			if charIndex >= len(f.chars) {
				charIndex = len(f.chars) - 1
			}
			char := f.chars[charIndex]

			// Map heat to color from palette
			colorIndex := heat * (len(f.palette) - 1) / 36
			if colorIndex >= len(f.palette) {
				colorIndex = len(f.palette) - 1
			}
			colorHex := f.palette[colorIndex]

			// Render colored character
			styled := lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorHex)).
				Render(string(char))
			line.WriteString(styled)
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}
