package animations

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// MatrixEffect implements Matrix digital rain animation using particle-based streaks
type MatrixEffect struct {
	width   int      // Terminal width
	height  int      // Terminal height
	palette []string // Theme color palette
	chars   []rune   // Matrix characters

	// Particle-based implementation - individual streaks that move down screen
	streaks []MatrixStreak // Active streaks
	frame   int            // Animation frame counter
}

// MatrixStreak represents a single vertical streak falling down the screen
type MatrixStreak struct {
	X       int  // X position (column)
	Y       int  // Y position of head
	Length  int  // Length of streak
	Speed   int  // Movement speed (frames per pixel)
	Counter int  // Frame counter for movement
	Active  bool // Whether streak is active
}

// MatrixChar represents a single character in a streak
type MatrixChar struct {
	Char  rune
	Color string
}

// NewMatrixEffect creates a new Matrix effect with given dimensions and theme palette
func NewMatrixEffect(width, height int, palette []string) *MatrixEffect {
	m := &MatrixEffect{
		width:   width,
		height:  height,
		palette: palette,
		// Use a mix of Latin, Greek, and Japanese characters like the original Matrix effect
		chars: []rune{
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'α', 'β', 'γ', 'δ', 'ε', 'ζ', 'η', 'θ', 'ι', 'κ', 'λ', 'μ',
			'ν', 'ξ', 'ο', 'π', 'ρ', 'σ', 'τ', 'υ', 'φ', 'χ', 'ψ', 'ω',
			'А', 'Б', 'В', 'Г', 'Д', 'Е', 'Ж', 'З', 'И', 'Й', 'К', 'Л', 'М',
			'Н', 'О', 'П', 'Р', 'С', 'Т', 'У', 'Ф', 'Х', 'Ц', 'Ч', 'Ш', 'Щ',
			'░', '▒', '▓', '█', '▀', '▄', '▌', '▐', '■', '□', '▪', '▫',
		},
		streaks: make([]MatrixStreak, 0, 100), // Pre-allocate capacity
		frame:   0,
	}
	m.init()
	return m
}

// Initialize Matrix effect with some initial streaks
func (m *MatrixEffect) init() {
	// Create initial streaks across width
	for i := 0; i < m.width; i++ {
		if rand.Float64() < 0.1 { // 10% chance of initial streak
			streak := MatrixStreak{
				X:       i,
				Y:       -rand.Intn(m.height), // Start above screen
				Length:  rand.Intn(15) + 5,    // Length 5-20
				Speed:   rand.Intn(3) + 1,     // Speed 1-3
				Counter: 0,
				Active:  true,
			}
			m.streaks = append(m.streaks, streak)
		}
	}
}

// UpdatePalette changes the Matrix color palette (for theme switching)
func (m *MatrixEffect) UpdatePalette(palette []string) {
	m.palette = palette
}

// Resize reinitializes the Matrix effect with new dimensions
func (m *MatrixEffect) Resize(width, height int) {
	m.width = width
	m.height = height
	m.init()
}

// getRandomColor returns a random color from the theme palette
func (m *MatrixEffect) getRandomColor() string {
	if len(m.palette) == 0 {
		return "#00ff00" // Default green if no palette
	}
	return m.palette[rand.Intn(len(m.palette))]
}

// getHeadColor returns the bright color for the head of the streak
func (m *MatrixEffect) getHeadColor() string {
	if len(m.palette) == 0 {
		return "#ffffff" // Default white if no palette
	}
	// Use the brightest color from the palette for heads
	if len(m.palette) > 0 {
		return m.palette[len(m.palette)-1]
	}
	return m.palette[0]
}

// getTrailColor returns a dimmer color for trail positions
func (m *MatrixEffect) getTrailColor(position, length int) string {
	if len(m.palette) == 0 {
		return "#00aa00" // Default dimmer green
	}

	// Calculate fade factor (0.0 = head, 1.0 = tail)
	fadeFactor := float64(position) / float64(length)

	// Use different colors based on position in trail
	if fadeFactor < 0.2 {
		// Bright trail near head
		if len(m.palette) > 0 {
			return m.palette[len(m.palette)-1]
		}
		return m.palette[0]
	} else if fadeFactor < 0.5 {
		// Medium trail
		if len(m.palette) > 2 {
			return m.palette[len(m.palette)-2]
		}
		return m.palette[0]
	} else {
		// Dim trail
		return m.palette[0]
	}
}

// Update advances the Matrix simulation by one frame
func (m *MatrixEffect) Update(frame int) {
	m.frame = frame

	// Update existing streaks
	activeStreaks := m.streaks[:0] // Reuse slice for efficiency
	for _, streak := range m.streaks {
		if !streak.Active {
			continue
		}

		// Increment counter
		streak.Counter++

		// Move streak if it's time
		if streak.Counter >= streak.Speed {
			streak.Counter = 0
			streak.Y++ // Move down

			// Deactivate if streak has moved well past screen
			if streak.Y-streak.Length > m.height {
				streak.Active = false
			}
		}

		if streak.Active {
			activeStreaks = append(activeStreaks, streak)
		}
	}
	m.streaks = activeStreaks

	// Add new streaks randomly
	for i := 0; i < m.width; i++ {
		// Low probability to create new streaks
		if rand.Float64() < 0.02 && len(m.streaks) < 150 { // Limit total streaks
			streak := MatrixStreak{
				X:       i,
				Y:       -rand.Intn(5),     // Start just above screen
				Length:  rand.Intn(15) + 5, // Length 5-20
				Speed:   rand.Intn(3) + 1,  // Speed 1-3
				Counter: 0,
				Active:  true,
			}
			m.streaks = append(m.streaks, streak)
		}
	}
}

// Render converts the Matrix streaks to colored text output
func (m *MatrixEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, m.height)
	colors := make([][]string, m.height)
	for i := range canvas {
		canvas[i] = make([]rune, m.width)
		colors[i] = make([]string, m.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
			colors[i][j] = ""
		}
	}

	// Render each active streak
	for _, streak := range m.streaks {
		if !streak.Active {
			continue
		}

		// Render the streak - from head downward
		for i := 0; i < streak.Length; i++ {
			yPos := streak.Y + i // Head at streak.Y, trail going down
			if yPos >= 0 && yPos < m.height && streak.X >= 0 && streak.X < m.width {
				// Get character
				char := m.chars[rand.Intn(len(m.chars))]

				// Get color based on position in streak
				var color string
				if i == 0 {
					// Head is brightest
					color = m.getHeadColor()
				} else {
					// Trail fades
					color = m.getTrailColor(i, streak.Length)
				}

				// Place character on canvas
				canvas[yPos][streak.X] = char
				colors[yPos][streak.X] = color
			}
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < m.height; y++ {
		var line strings.Builder
		for x := 0; x < m.width; x++ {
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
