package animations

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// RainEffect implements ascii rain animation
type RainEffect struct {
	width   int      // Terminal width
	height  int      // Terminal height
	drops   []Drop   // Rain drops
	palette []string // Theme color palette
	chars   []rune   // Rain characters
	frame   int      // Animation frame counter
}

// Drop represents a single raindrop
type Drop struct {
	X     int    // X position
	Y     int    // Y position
	Speed int    // Fall speed
	Char  rune   // Character to display
	Color string // Color hex code
}

// NewRainEffect creates a new rain effect with given dimensions and theme palette
func NewRainEffect(width, height int, palette []string) *RainEffect {
	r := &RainEffect{
		width:   width,
		height:  height,
		palette: palette,
		chars:   []rune{'|', '│', '┤', '╡', '╢', '╖', '╕', '╣', '║', '╗', '╝', '╜', '╛', '┐', '└', '┴', '┬', '├', '─', '┼', '╞', '╟', '╚', '╔', '╩', '╦', '╠', '═', '╬', '╧', '╨', '╤', '╥', '╙', '╘'},
		frame:   0,
	}
	r.init()
	return r
}

// Initialize rain drops
func (r *RainEffect) init() {
	// Create initial raindrops
	r.drops = make([]Drop, 0, r.width*2)
	for i := 0; i < r.width; i++ {
		if rand.Float64() < 0.3 { // 30% chance of raindrop at each column
			drop := Drop{
				X:     i,
				Y:     rand.Intn(r.height),
				Speed: rand.Intn(3) + 1, // Speed 1-3
				Char:  r.chars[rand.Intn(len(r.chars))],
				Color: r.getRandomColor(),
			}
			r.drops = append(r.drops, drop)
		}
	}
}

// UpdatePalette changes the rain color palette (for theme switching)
func (r *RainEffect) UpdatePalette(palette []string) {
	r.palette = palette
	// Update colors for existing drops
	for i := range r.drops {
		r.drops[i].Color = r.getRandomColor()
	}
}

// Resize reinitializes the rain effect with new dimensions
func (r *RainEffect) Resize(width, height int) {
	r.width = width
	r.height = height
	r.init()
}

// getRandomColor returns a random color from the palette
func (r *RainEffect) getRandomColor() string {
	if len(r.palette) == 0 {
		return "#00ff00" // Default green
	}
	return r.palette[rand.Intn(len(r.palette))]
}

// Update advances the rain simulation by one frame
func (r *RainEffect) Update(frame int) {
	r.frame = frame

	// Move existing drops
	for i := range r.drops {
		// Move drop down by its speed
		r.drops[i].Y += r.drops[i].Speed

		// Randomly change character occasionally
		if rand.Float64() < 0.1 {
			r.drops[i].Char = r.chars[rand.Intn(len(r.chars))]
		}

		// Randomly change color occasionally
		if rand.Float64() < 0.05 {
			r.drops[i].Color = r.getRandomColor()
		}
	}

	// Remove drops that have fallen off screen
	filtered := r.drops[:0]
	for _, drop := range r.drops {
		if drop.Y < r.height {
			filtered = append(filtered, drop)
		}
	}
	r.drops = filtered

	// Add new drops at the top
	for i := 0; i < r.width; i++ {
		// Probability of new drop decreases as we approach max density
		dropProbability := 0.1 - (float64(len(r.drops))/float64(r.width*10))*0.08
		if rand.Float64() < dropProbability {
			drop := Drop{
				X:     i,
				Y:     0,
				Speed: rand.Intn(3) + 1, // Speed 1-3
				Char:  r.chars[rand.Intn(len(r.chars))],
				Color: r.getRandomColor(),
			}
			r.drops = append(r.drops, drop)
		}
	}
}

// Render converts the rain drops to colored text output
func (r *RainEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, r.height)
	for i := range canvas {
		canvas[i] = make([]rune, r.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Place drops on canvas
	for _, drop := range r.drops {
		if drop.X >= 0 && drop.X < r.width && drop.Y >= 0 && drop.Y < r.height {
			canvas[drop.Y][drop.X] = drop.Char
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < r.height; y++ {
		var line strings.Builder
		for x := 0; x < r.width; x++ {
			char := canvas[y][x]
			if char != ' ' {
				// Find the drop at this position to get its color
				color := "#00ff00" // Default color
				for _, drop := range r.drops {
					if drop.X == x && drop.Y == y {
						color = drop.Color
						break
					}
				}

				// Render colored character
				styled := lipgloss.NewStyle().
					Foreground(lipgloss.Color(color)).
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
