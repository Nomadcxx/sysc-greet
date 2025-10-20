package animations

import (
	"math"
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"gonum.org/v1/gonum/spatial/r2"
)

// Particle represents a single firework particle
type Particle struct {
	pos              r2.Vec  // Current position
	p0, p1, p2, p3   r2.Vec  // Bezier control points
	t                float64 // Progress (0-1)
	char             rune    // Character to display
	style            lipgloss.Style
	phase            int    // 0=launch, 1=explosion, 2=fall
	color            string // Current color
	targetX, targetY int    // Final position
}

// FireworksEffect implements fireworks animation
type FireworksEffect struct {
	width, height int
	particles     []Particle
	palette       []string
	frame         int
	shells        [][]int // Indices of particles in each shell
	launchDelay   int
	activeShells  int
}

// NewFireworksEffect creates a new fireworks effect
func NewFireworksEffect(width, height int, palette []string) *FireworksEffect {
	fw := &FireworksEffect{
		width:        width,
		height:       height,
		palette:      palette,
		frame:        0,
		launchDelay:  0,
		activeShells: 0,
	}
	fw.init()
	return fw
}

// Initialize fireworks with particles
func (fw *FireworksEffect) init() {
	// Create particles - for sysc-greet we'll create a fixed number of particles
	// that continuously animate rather than animating text characters
	particleCount := fw.width * 2
	fw.particles = make([]Particle, particleCount)

	// Various characters for firework particles
	chars := []rune{'*', '•', '○', '●', '◦', '◉', '◌', '◍', '◎', '◐', '+', 'x', '✦', '✧', '✨', '✪', '✫', '✬', '✭', '✮'}

	for i := 0; i < particleCount; i++ {
		fw.particles[i] = Particle{
			char:  chars[rand.Intn(len(chars))],
			t:     1, // Set to 1 so particles don't render until launched
			phase: 0,
			pos:   r2.Vec{X: -100, Y: -100}, // Off-screen initially
		}
	}

	// Create shells (groups of particles that explode together)
	shellSize := 25
	fw.shells = nil // Clear existing shells
	for i := 0; i < len(fw.particles); i += shellSize {
		end := i + shellSize
		if end > len(fw.particles) {
			end = len(fw.particles)
		}
		indices := make([]int, end-i)
		for j := range indices {
			indices[j] = i + j
		}
		fw.shells = append(fw.shells, indices)
	}
}

// UpdatePalette changes the fireworks color palette
func (fw *FireworksEffect) UpdatePalette(palette []string) {
	fw.palette = palette
}

// Resize reinitializes the fireworks effect with new dimensions
func (fw *FireworksEffect) Resize(width, height int) {
	// Don't reinit if dimensions haven't actually changed
	if fw.width == width && fw.height == height {
		return
	}
	
	fw.width = width
	fw.height = height
	
	// Reset animation state but don't call init() to avoid resetting all particles
	fw.activeShells = 0
	fw.launchDelay = 0
	
	// Only reinit if this is a significant size change or first init
	// (Check if we have particles - if not, this is first init)
	if len(fw.particles) == 0 {
		fw.init()
	}
}

// evaluateBezier evaluates a cubic bezier curve at parameter t
func evaluateBezier(p0, p1, p2, p3 r2.Vec, t float64) r2.Vec {
	it := 1 - t
	return r2.Vec{
		X: it*it*it*p0.X + 3*it*it*t*p1.X + 3*it*t*t*p2.X + t*t*t*p3.X,
		Y: it*it*it*p0.Y + 3*it*it*t*p1.Y + 3*it*t*t*p2.Y + t*t*t*p3.Y,
	}
}

// launchShell launches a group of particles
func (fw *FireworksEffect) launchShell(shellIndex int) {
	if shellIndex >= len(fw.shells) {
		return
	}

	indices := fw.shells[shellIndex]
	centerX := float64(rand.Intn(fw.width-20) + 10)              // Keep away from edges
	centerY := float64(fw.height - 1)                            // Start from bottom
	explodeY := float64(rand.Intn(fw.height/3) + fw.height/5) // Explosion in upper third

	for _, idx := range indices {
		p := &fw.particles[idx]
		p.t = 0
		p.phase = 0
		p.pos = r2.Vec{X: centerX, Y: centerY}

		// Launch path - straight up with slight curve
		p.p0 = r2.Vec{X: centerX, Y: centerY}
		p.p1 = r2.Vec{X: centerX + (rand.Float64()-0.5)*2, Y: centerY - (centerY-explodeY)*0.3}
		p.p2 = r2.Vec{X: centerX + (rand.Float64()-0.5)*2, Y: explodeY + 5}
		p.p3 = r2.Vec{X: centerX, Y: explodeY}

		// Set initial color
		if len(fw.palette) > 0 {
			p.color = fw.palette[len(fw.palette)-1] // Use brightest color for launch
		} else {
			p.color = "#FFFFFF"
		}
		p.style = lipgloss.NewStyle().Foreground(lipgloss.Color(p.color))
	}
}

// explodeShell makes particles explode from their current position
func (fw *FireworksEffect) explodeShell(shellIndex int) {
	if shellIndex >= len(fw.shells) {
		return
	}

	indices := fw.shells[shellIndex]
	if len(indices) == 0 {
		return
	}

	// Use position of first particle as explosion center
	centerX := fw.particles[indices[0]].pos.X
	centerY := fw.particles[indices[0]].pos.Y
	explodeRadius := float64(20 + rand.Intn(25)) // Larger explosion radius

	for _, idx := range indices {
		p := &fw.particles[idx]
		p.t = 0
		p.phase = 1

		// Random angle for explosion direction
		angle := rand.Float64() * 2 * math.Pi
		targetX := centerX + explodeRadius*math.Cos(angle)
		targetY := centerY + explodeRadius*math.Sin(angle)*0.6 // Slightly elliptical

		// Bezier path for explosion - arc upward then fall
		p.p0 = r2.Vec{X: centerX, Y: centerY}
		p.p1 = r2.Vec{X: centerX + (targetX-centerX)*0.3, Y: centerY - 8} // Stronger upward curve
		p.p2 = r2.Vec{X: centerX + (targetX-centerX)*0.7, Y: targetY - 5} // Mid curve
		p.p3 = r2.Vec{X: targetX, Y: targetY}
		
		// Assign a color for this explosion
		if len(fw.palette) > 0 {
			p.color = fw.palette[rand.Intn(len(fw.palette))]
			p.style = lipgloss.NewStyle().Foreground(lipgloss.Color(p.color))
		}
	}
}

// fallParticles makes particles fall to bottom of screen
func (fw *FireworksEffect) fallParticles(shellIndex int) {
	if shellIndex >= len(fw.shells) {
		return
	}

	indices := fw.shells[shellIndex]
	for _, idx := range indices {
		p := &fw.particles[idx]
		if p.phase != 1 {
			continue
		}

		p.t = 0
		p.phase = 2

		startX := p.pos.X
		startY := p.pos.Y
		endX := startX + (rand.Float64()-0.5)*10 // Slight horizontal drift
		endY := float64(fw.height - 1)

		// Bezier path for falling - slight curve
		p.p0 = r2.Vec{X: startX, Y: startY}
		p.p1 = r2.Vec{X: startX + (endX-startX)*0.3, Y: startY + (endY-startY)*0.3}
		p.p2 = r2.Vec{X: startX + (endX-startX)*0.7, Y: startY + (endY-startY)*0.7}
		p.p3 = r2.Vec{X: endX, Y: endY}
	}
}

// Update advances the fireworks simulation
func (fw *FireworksEffect) Update(frame int) {
	fw.frame = frame

	// Launch new shell if delay is over
	if fw.launchDelay <= 0 && fw.activeShells < len(fw.shells) {
		fw.launchShell(fw.activeShells)
		fw.launchDelay = 15 + rand.Intn(20) // 15-35 frames between shells (faster)
		fw.activeShells++
	}
	fw.launchDelay--

	// Track which shells need phase transitions
	shellsToExplode := make(map[int]bool)
	shellsToFall := make(map[int]bool)

	// Update all particles
	for i := range fw.particles {
		p := &fw.particles[i]
		
		// Skip inactive particles (t=1 and phase=0 means waiting for next launch)
		if p.t >= 1 && p.phase == 0 {
			continue
		}

		// Different speeds for different phases
		speed := 0.03 // Default speed
		switch p.phase {
		case 0: // Launch - faster
			speed = 0.05
		case 1: // Explosion - medium
			speed = 0.03
		case 2: // Fall - faster
			speed = 0.04
		}
		
		p.t += speed

		// Update position along bezier path
		if p.t <= 1 {
			p.pos = evaluateBezier(p.p0, p.p1, p.p2, p.p3, p.t)
		}

		// Handle phase transitions
		if p.t >= 1 {
			p.t = 1
			// Find which shell this particle belongs to
			for shellIdx, indices := range fw.shells {
				for _, idx := range indices {
					if idx == i {
						switch p.phase {
						case 0: // Launch complete, mark for explosion
							shellsToExplode[shellIdx] = true
						case 1: // Explosion complete, mark for fall
							shellsToFall[shellIdx] = true
						case 2: // Fall complete, hide particle
							p.t = 1 // Keep at end so it doesn't render
							p.phase = 0
							p.pos = r2.Vec{X: -100, Y: -100} // Move off-screen
						}
						break
					}
				}
			}
		}

		// Update color based on phase
		if len(fw.palette) > 0 {
			switch p.phase {
			case 0: // Launch - bright color
				p.color = fw.palette[len(fw.palette)-1] // Brightest
			case 1: // Explosion - random color
				if p.t < 0.1 || rand.Float64() < 0.05 { // Change color occasionally
					p.color = fw.palette[rand.Intn(len(fw.palette))]
				}
			case 2: // Fall - fade to darker colors
				fadeIdx := int(p.t * float64(len(fw.palette)-1))
				if fadeIdx >= len(fw.palette) {
					fadeIdx = len(fw.palette) - 1
				}
				p.color = fw.palette[fadeIdx]
			}
			p.style = lipgloss.NewStyle().Foreground(lipgloss.Color(p.color))
		}
	}

	// Execute phase transitions for shells
	for shellIdx := range shellsToExplode {
		fw.explodeShell(shellIdx)
	}
	for shellIdx := range shellsToFall {
		fw.fallParticles(shellIdx)
	}

	// Reset when all shells have been launched and completed
	if fw.activeShells >= len(fw.shells) {
		allComplete := true
		for i := range fw.particles {
			if fw.particles[i].phase != 0 || fw.particles[i].t < 1 {
				allComplete = false
				break
			}
		}
		if allComplete {
			fw.activeShells = 0
			fw.launchDelay = 0
		}
	}
}

// Render converts the fireworks to colored text output
func (fw *FireworksEffect) Render() string {
	// Create empty canvas
	canvas := make([][]rune, fw.height)
	styles := make([][]lipgloss.Style, fw.height)
	for i := range canvas {
		canvas[i] = make([]rune, fw.width)
		styles[i] = make([]lipgloss.Style, fw.width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Place particles on canvas
	for _, p := range fw.particles {
		// Only render particles that are actively animating
		if p.t >= 1 {
			continue
		}
		x, y := int(p.pos.X), int(p.pos.Y)
		if x >= 0 && x < fw.width && y >= 0 && y < fw.height {
			canvas[y][x] = p.char
			styles[y][x] = p.style
		}
	}

	// Convert to colored string
	var lines []string
	for y := 0; y < fw.height; y++ {
		var line strings.Builder
		for x := 0; x < fw.width; x++ {
			char := canvas[y][x]
			if char != ' ' {
				line.WriteString(styles[y][x].Render(string(char)))
			} else {
				line.WriteRune(char)
			}
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}
