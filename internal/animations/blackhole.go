package animations

import (
	"math"
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// BlackholeEffect implements a black hole starfield effect
type BlackholeEffect struct {
	width   int      // Terminal width
	height  int      // Terminal height
	palette []string // Theme color palette
	frame   int      // Animation frame counter

	// Blackhole configuration
	blackholeRadius int
	blackholeChars  []BlackholeChar
	starChars       []StarChar
	phase           string // "forming", "consuming", "collapsing", "exploding", "complete"

	// Animation timing
	formationDelay    int
	consumptionDelay  int
	nextCharDelay     int
	maxConsumePerTick int

	// Character symbols
	starSymbols []rune
}

// BlackholeChar represents a character that forms the black hole border
type BlackholeChar struct {
	X       float64 // X position
	Y       float64 // Y position
	TargetX float64 // Target X position
	TargetY float64 // Target Y position
	Speed   float64 // Movement speed
	Symbol  rune    // Character symbol
	Color   string  // Character color
	Phase   string  // "forming", "rotating", "collapsing", "exploding"
}

// StarChar represents a star in the starfield
type StarChar struct {
	X       float64 // X position
	Y       float64 // Y position
	TargetX float64 // Target X position (center)
	TargetY float64 // Target Y position (center)
	Speed   float64 // Movement speed
	Symbol  rune    // Character symbol
	Color   string  // Character color
	Phase   string  // "starfield", "consumed", "exploding"
}

// NewBlackholeEffect creates a new black hole effect with given dimensions and theme palette
func NewBlackholeEffect(width, height int, palette []string) *BlackholeEffect {
	b := &BlackholeEffect{
		width:             width,
		height:            height,
		palette:           palette,
		frame:             0,
		blackholeRadius:   max(3, min(width/3, height/3)),
		blackholeChars:    []BlackholeChar{},
		starChars:         []StarChar{},
		phase:             "forming",
		formationDelay:    5,
		consumptionDelay:  0,
		nextCharDelay:     0,
		maxConsumePerTick: max(2, min(15, width*height/100)),
		starSymbols:       []rune{'*', '\'', '`', '¤', '•', '°', '·'},
	}
	b.init()
	return b
}

// Initialize the black hole effect
func (b *BlackholeEffect) init() {
	centerX := float64(b.width) / 2.0
	centerY := float64(b.height) / 2.0

	// Create black hole border characters
	numBlackholeChars := b.blackholeRadius * 3
	b.blackholeChars = make([]BlackholeChar, numBlackholeChars)

	// Position black hole characters in a circle
	for i := 0; i < numBlackholeChars; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numBlackholeChars)
		radius := float64(b.blackholeRadius)

		// Start from random positions outside the screen
		startX := rand.Float64() * float64(b.width)
		startY := rand.Float64() * float64(b.height)

		// Target position on the black hole circle
		targetX := centerX + math.Cos(angle)*radius
		targetY := centerY + math.Sin(angle)*radius

		b.blackholeChars[i] = BlackholeChar{
			X:       startX,
			Y:       startY,
			TargetX: targetX,
			TargetY: targetY,
			Speed:   0.5 + rand.Float64()*0.5,
			Symbol:  '*',
			Color:   b.getBlackholeColor(),
			Phase:   "forming",
		}
	}

	// Create starfield characters
	numStars := (b.width * b.height) / 4
	b.starChars = make([]StarChar, numStars)

	for i := 0; i < numStars; i++ {
		// Random position in the starfield
		x := rand.Float64() * float64(b.width)
		y := rand.Float64() * float64(b.height)

		b.starChars[i] = StarChar{
			X:       x,
			Y:       y,
			TargetX: centerX,
			TargetY: centerY,
			Speed:   0.1 + rand.Float64()*0.3,
			Symbol:  b.starSymbols[rand.Intn(len(b.starSymbols))],
			Color:   b.getRandomStarColor(),
			Phase:   "starfield",
		}
	}
}

// getBlackholeColor returns the color for black hole characters
func (b *BlackholeEffect) getBlackholeColor() string {
	// Use white or the first color in the palette for the black hole
	if len(b.palette) > 0 {
		return b.palette[0]
	}
	return "#ffffff"
}

// getRandomStarColor returns a random color for star characters
func (b *BlackholeEffect) getRandomStarColor() string {
	starColors := []string{
		"#ffcc0d", "#ff7326", "#ff194d",
		"#bf2669", "#702a8c", "#049dbf",
	}

	// Mix with theme palette if available
	if len(b.palette) > 0 {
		// Take some colors from the theme palette
		for i := 0; i < min(3, len(b.palette)); i++ {
			starColors = append(starColors, b.palette[i])
		}
	}

	return starColors[rand.Intn(len(starColors))]
}

// UpdatePalette changes the effect color palette (for theme switching)
func (b *BlackholeEffect) UpdatePalette(palette []string) {
	b.palette = palette
}

// Resize reinitializes the black hole effect with new dimensions
func (b *BlackholeEffect) Resize(width, height int) {
	b.width = width
	b.height = height
	b.blackholeRadius = max(3, min(width/3, height/3))
	b.init()
}

// Update advances the black hole simulation by one frame
func (b *BlackholeEffect) Update(frame int) {
	b.frame = frame

	switch b.phase {
	case "forming":
		// Move black hole characters toward their target positions
		allFormed := true
		for i := range b.blackholeChars {
			if b.blackholeChars[i].Phase == "forming" {
				// Move toward target position
				dx := b.blackholeChars[i].TargetX - b.blackholeChars[i].X
				dy := b.blackholeChars[i].TargetY - b.blackholeChars[i].Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0.1 {
					// Move toward target
					b.blackholeChars[i].X += dx * b.blackholeChars[i].Speed / distance
					b.blackholeChars[i].Y += dy * b.blackholeChars[i].Speed / distance
					allFormed = false
				} else {
					// Reached target, start rotating
					b.blackholeChars[i].Phase = "rotating"
				}
			}
		}

		if allFormed {
			b.phase = "consuming"
		}

	case "consuming":
		// Consume star characters
		if b.nextCharDelay > 0 {
			b.nextCharDelay--
		} else {
			// Consume a batch of stars
			consumed := 0
			for i := range b.starChars {
				if b.starChars[i].Phase == "starfield" && consumed < b.maxConsumePerTick {
					b.starChars[i].Phase = "consumed"
					b.starChars[i].Speed = 0.3 + rand.Float64()*0.2
					consumed++
				}
			}

			// Increase consumption rate over time
			b.maxConsumePerTick = min(b.maxConsumePerTick+1, 20)
			b.nextCharDelay = rand.Intn(10)
		}

		// Move consumed stars toward black hole center
		allStarsConsumed := true
		centerX := float64(b.width) / 2.0
		centerY := float64(b.height) / 2.0

		for i := range b.starChars {
			if b.starChars[i].Phase == "consumed" {
				// Move toward center
				dx := centerX - b.starChars[i].X
				dy := centerY - b.starChars[i].Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0.5 {
					b.starChars[i].X += dx * b.starChars[i].Speed / distance
					b.starChars[i].Y += dy * b.starChars[i].Speed / distance
					allStarsConsumed = false
				} else {
					// Reached center, fade out
					b.starChars[i].Phase = "collapsed"
					b.starChars[i].Symbol = ' '
				}
			} else if b.starChars[i].Phase == "starfield" {
				allStarsConsumed = false
			}
		}

		if allStarsConsumed {
			b.phase = "collapsing"
		}

	case "collapsing":
		// Collapse the black hole
		for i := range b.blackholeChars {
			if b.blackholeChars[i].Phase == "rotating" {
				// Expand outward then collapse inward
				centerX := float64(b.width) / 2.0
				centerY := float64(b.height) / 2.0

				// Expand the black hole radius
				expandedRadius := float64(b.blackholeRadius + 3)
				angle := math.Atan2(b.blackholeChars[i].Y-centerY, b.blackholeChars[i].X-centerX)
				b.blackholeChars[i].TargetX = centerX + math.Cos(angle)*expandedRadius
				b.blackholeChars[i].TargetY = centerY + math.Sin(angle)*expandedRadius

				// Move toward expanded position
				dx := b.blackholeChars[i].TargetX - b.blackholeChars[i].X
				dy := b.blackholeChars[i].TargetY - b.blackholeChars[i].Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0.1 {
					b.blackholeChars[i].X += dx * b.blackholeChars[i].Speed / distance
					b.blackholeChars[i].Y += dy * b.blackholeChars[i].Speed / distance
				} else {
					// Reached expanded position, now collapse to center
					b.blackholeChars[i].TargetX = centerX
					b.blackholeChars[i].TargetY = centerY
					b.blackholeChars[i].Phase = "collapsing"
				}
			} else if b.blackholeChars[i].Phase == "collapsing" {
				// Move toward center
				centerX := float64(b.width) / 2.0
				centerY := float64(b.height) / 2.0
				dx := centerX - b.blackholeChars[i].X
				dy := centerY - b.blackholeChars[i].Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0.1 {
					b.blackholeChars[i].X += dx * b.blackholeChars[i].Speed / distance
					b.blackholeChars[i].Y += dy * b.blackholeChars[i].Speed / distance
				} else {
					// Reached center, start exploding
					b.blackholeChars[i].Phase = "exploding"
					b.blackholeChars[i].Symbol = []rune("◦◎◉●◉◎◦")[b.frame%7]
				}
			}
		}

		// Check if all black hole chars have collapsed
		allCollapsed := true
		for _, bhChar := range b.blackholeChars {
			if bhChar.Phase != "exploding" {
				allCollapsed = false
				break
			}
		}

		if allCollapsed {
			b.phase = "exploding"
		}

	case "exploding":
		// Explode stars outward from center
		centerX := float64(b.width) / 2.0
		centerY := float64(b.height) / 2.0

		for i := range b.starChars {
			if b.starChars[i].Phase == "collapsed" {
				// Set explosion direction
				angle := rand.Float64() * 2.0 * math.Pi
				distance := 3.0 + rand.Float64()*5.0
				b.starChars[i].TargetX = centerX + math.Cos(angle)*distance
				b.starChars[i].TargetY = centerY + math.Sin(angle)*distance
				b.starChars[i].Speed = 0.2 + rand.Float64()*0.3
				b.starChars[i].Phase = "exploding"
				b.starChars[i].Symbol = '*'
				b.starChars[i].Color = b.getRandomStarColor()
			} else if b.starChars[i].Phase == "exploding" {
				// Move toward target position
				dx := b.starChars[i].TargetX - b.starChars[i].X
				dy := b.starChars[i].TargetY - b.starChars[i].Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0.1 {
					b.starChars[i].X += dx * b.starChars[i].Speed / distance
					b.starChars[i].Y += dy * b.starChars[i].Speed / distance
				} else {
					// Reached target, stabilize
					b.starChars[i].Phase = "stable"
				}
			}
		}

		b.phase = "complete"
	}

	// Update black hole character symbols/rotation
	if b.phase == "rotating" || b.phase == "consuming" {
		for i := range b.blackholeChars {
			if b.blackholeChars[i].Phase == "rotating" {
				// Rotate symbols for animation effect
				symbols := []rune{'*', '◦', '◎', '◉', '●'}
				b.blackholeChars[i].Symbol = symbols[(b.frame+i)%len(symbols)]
			}
		}
	}

	// Update exploding black hole characters
	if b.phase == "exploding" || b.phase == "complete" {
		for i := range b.blackholeChars {
			if b.blackholeChars[i].Phase == "exploding" {
				// Cycle through explosion symbols
				symbols := []rune("◦◎◉●◉◎◦")
				b.blackholeChars[i].Symbol = symbols[(b.frame/3+i)%len(symbols)]
				b.blackholeChars[i].Color = b.getRandomStarColor()
			}
		}
	}
}

// Render converts the black hole effect to colored text output
func (b *BlackholeEffect) Render() string {
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

	// Draw star characters first (background)
	for _, star := range b.starChars {
		x := int(star.X)
		y := int(star.Y)

		if x >= 0 && x < b.width && y >= 0 && y < b.height {
			canvas[y][x] = star.Symbol
			colors[y][x] = star.Color
		}
	}

	// Draw black hole characters on top
	for _, bhChar := range b.blackholeChars {
		x := int(bhChar.X)
		y := int(bhChar.Y)

		if x >= 0 && x < b.width && y >= 0 && y < b.height {
			canvas[y][x] = bhChar.Symbol
			colors[y][x] = bhChar.Color
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

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
