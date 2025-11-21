package animations

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// BeamsEffect implements beams that travel across rows and columns
type BeamsEffect struct {
	width   int      // Terminal width
	height  int      // Terminal height
	palette []string // Theme color palette
	frame   int      // Animation frame counter

	// Beam configuration
	rowSymbols    []rune // Symbols for row beams
	columnSymbols []rune // Symbols for column beams
	delay         int    // Delay between beam groups
	speedRange    [2]int // Speed range for beams

	// Active beams
	activeBeams   []Beam
	pendingBeams  []Beam
	nextBeamDelay int
}

// Beam represents a single beam traveling across the screen
type Beam struct {
	X         int     // X position (for column beams) or starting X (for row beams)
	Y         int     // Y position (for row beams) or starting Y (for column beams)
	Direction string  // "row" or "column"
	Length    int     // Length of the beam
	Speed     float64 // Speed of the beam
	Position  float64 // Current position along the path
	Symbol    rune    // Current symbol
	Color     string  // Current color
}

// NewBeamsEffect creates a new beams effect with given dimensions and theme palette
func NewBeamsEffect(width, height int, palette []string) *BeamsEffect {
	b := &BeamsEffect{
		width:         width,
		height:        height,
		palette:       palette,
		frame:         0,
		rowSymbols:    []rune{'▂', '▁', '_'},
		columnSymbols: []rune{'▌', '▍', '▎', '▏'},
		delay:         10,
		speedRange:    [2]int{10, 40},
		activeBeams:   []Beam{},
		pendingBeams:  []Beam{},
		nextBeamDelay: 0,
	}
	b.init()
	return b
}

// Initialize beams
func (b *BeamsEffect) init() {
	// Create initial beams
	b.pendingBeams = make([]Beam, 0, (b.height+b.width)*2)

	// Add row beams (one for each row)
	for y := 0; y < b.height; y++ {
		beam := Beam{
			X:         0,
			Y:         y,
			Direction: "row",
			Length:    3 + rand.Intn(5), // Random length 3-7
			Speed:     float64(rand.Intn(b.speedRange[1]-b.speedRange[0])+b.speedRange[0]) * 0.1,
			Position:  -float64(b.width), // Start off screen
			Symbol:    b.rowSymbols[rand.Intn(len(b.rowSymbols))],
			Color:     b.getRandomColor(),
		}
		b.pendingBeams = append(b.pendingBeams, beam)
	}

	// Add column beams (one for each column)
	for x := 0; x < b.width; x++ {
		beam := Beam{
			X:         x,
			Y:         0,
			Direction: "column",
			Length:    3 + rand.Intn(5), // Random length 3-7
			Speed:     float64(rand.Intn(b.speedRange[1]-b.speedRange[0])+b.speedRange[0]) * 0.1,
			Position:  -float64(b.height), // Start off screen
			Symbol:    b.columnSymbols[rand.Intn(len(b.columnSymbols))],
			Color:     b.getRandomColor(),
		}
		b.pendingBeams = append(b.pendingBeams, beam)
	}

	// Shuffle the pending beams for random activation
	rand.Shuffle(len(b.pendingBeams), func(i, j int) {
		b.pendingBeams[i], b.pendingBeams[j] = b.pendingBeams[j], b.pendingBeams[i]
	})
}

// UpdatePalette changes the beam color palette (for theme switching)
func (b *BeamsEffect) UpdatePalette(palette []string) {
	b.palette = palette
}

// Resize reinitializes the beams effect with new dimensions
func (b *BeamsEffect) Resize(width, height int) {
	b.width = width
	b.height = height
	b.init()
}

// getRandomColor returns a random color from the palette
func (b *BeamsEffect) getRandomColor() string {
	if len(b.palette) == 0 {
		return "#ffffff" // Default white
	}
	return b.palette[rand.Intn(len(b.palette))]
}

// Update advances the beams simulation by one frame
func (b *BeamsEffect) Update(frame int) {
	b.frame = frame

	// Update delay counter
	if b.nextBeamDelay > 0 {
		b.nextBeamDelay--
	}

	// Add new beams from pending list
	if b.nextBeamDelay == 0 && len(b.pendingBeams) > 0 {
		// Add a group of 1-5 beams
		groupSize := rand.Intn(5) + 1
		for i := 0; i < groupSize && len(b.pendingBeams) > 0; i++ {
			beam := b.pendingBeams[0]
			b.activeBeams = append(b.activeBeams, beam)
			b.pendingBeams = b.pendingBeams[1:]
		}
		b.nextBeamDelay = b.delay
	}

	// Update active beams
	for i := range b.activeBeams {
		// Move beam
		b.activeBeams[i].Position += b.activeBeams[i].Speed

		// Randomly change symbol and color occasionally
		if rand.Float64() < 0.1 {
			if b.activeBeams[i].Direction == "row" {
				b.activeBeams[i].Symbol = b.rowSymbols[rand.Intn(len(b.rowSymbols))]
			} else {
				b.activeBeams[i].Symbol = b.columnSymbols[rand.Intn(len(b.columnSymbols))]
			}
		}
		if rand.Float64() < 0.05 {
			b.activeBeams[i].Color = b.getRandomColor()
		}
	}

	// Remove beams that have moved off screen
	filtered := b.activeBeams[:0]
	for _, beam := range b.activeBeams {
		// Keep beams that are still visible or partially visible
		if beam.Direction == "row" {
			if beam.Position > -float64(beam.Length) && beam.Position < float64(b.width) {
				filtered = append(filtered, beam)
			}
		} else { // column
			if beam.Position > -float64(beam.Length) && beam.Position < float64(b.height) {
				filtered = append(filtered, beam)
			}
		}
	}
	b.activeBeams = filtered

	// If no more beams, restart
	if len(b.activeBeams) == 0 && len(b.pendingBeams) == 0 {
		b.init()
	}
}

// Render converts the beams to colored text output
func (b *BeamsEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, b.height)
	colors := make([][]string, b.height)
	for i := range canvas {
		canvas[i] = make([]rune, b.width)
		colors[i] = make([]string, b.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Draw active beams on canvas
	for _, beam := range b.activeBeams {
		if beam.Direction == "row" {
			// Draw horizontal beam
			startX := int(beam.Position)
			for i := 0; i < beam.Length; i++ {
				x := startX + i
				if x >= 0 && x < b.width && beam.Y >= 0 && beam.Y < b.height {
					canvas[beam.Y][x] = beam.Symbol
					colors[beam.Y][x] = beam.Color
				}
			}
		} else { // column
			// Draw vertical beam
			startY := int(beam.Position)
			for i := 0; i < beam.Length; i++ {
				y := startY + i
				if y >= 0 && y < b.height && beam.X >= 0 && beam.X < b.width {
					canvas[y][beam.X] = beam.Symbol
					colors[y][beam.X] = beam.Color
				}
			}
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < b.height; y++ {
		var line strings.Builder
		for x := 0; x < b.width; x++ {
			char := canvas[y][x]
			if char != ' ' && colors[y][x] != "" {
				// Render colored character
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(colors[y][x])).
					Render(string(char))
				line.WriteString(styled)
			} else {
				line.WriteRune(char)
			}
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}
