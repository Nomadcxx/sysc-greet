package animations

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// PourEffect implements a character pouring animation from different directions
type PourEffect struct {
	width                  int
	height                 int
	text                   string
	pourDirection          string
	pourSpeed              int
	movementSpeed          float64
	gap                    int
	startingColor          string
	finalGradientStops     []string
	finalGradientSteps     int
	finalGradientFrames    int
	finalGradientDirection string
	phase                  string
	frameCount             int
	holdCounter            int // Frames to hold after completion

	chars          []PourCharacter
	groups         [][]int // Indices of characters grouped by row/column
	currentGroup   int
	currentInGroup int
	gapCounter     int
	alternateDir   bool // Alternate pouring direction
}

// PourCharacter represents a single character in the pour animation
type PourCharacter struct {
	original        rune
	finalX          int
	finalY          int
	startX          int
	startY          int
	currentX        float64
	currentY        float64
	visible         bool
	color           string
	finalColor      string
	progress        float64
	gradientStep    int
	gradientCounter int
}

// PourConfig holds configuration for the pour effect
type PourConfig struct {
	Width                  int
	Height                 int
	Text                   string
	PourDirection          string
	PourSpeed              int
	MovementSpeed          float64
	Gap                    int
	StartingColor          string
	FinalGradientStops     []string
	FinalGradientSteps     int
	FinalGradientFrames    int
	FinalGradientDirection string
}

// NewPourEffect creates a new pour effect with given configuration
func NewPourEffect(config PourConfig) *PourEffect {
	effect := &PourEffect{
		width:                  config.Width,
		height:                 config.Height,
		text:                   config.Text,
		pourDirection:          config.PourDirection,
		pourSpeed:              config.PourSpeed,
		movementSpeed:          config.MovementSpeed,
		gap:                    config.Gap,
		startingColor:          config.StartingColor,
		finalGradientStops:     config.FinalGradientStops,
		finalGradientSteps:     config.FinalGradientSteps,
		finalGradientFrames:    config.FinalGradientFrames,
		finalGradientDirection: config.FinalGradientDirection,
		phase:                  "pouring",
		frameCount:             0,
		currentGroup:           0,
		currentInGroup:         0,
		gapCounter:             0,
		alternateDir:           false,
	}

	effect.init()
	return effect
}

// Initialize the pour effect with characters and their animations
func (p *PourEffect) init() {
	lines := strings.Split(p.text, "\n")

	// Calculate centered position for text block (like beams does it)
	startY := (p.height - len(lines)) / 2
	if startY < 0 {
		startY = 0
	}

	// Find max width for block centering
	maxWidth := 0
	for _, line := range lines {
		if len([]rune(line)) > maxWidth {
			maxWidth = len([]rune(line))
		}
	}

	blockStartX := (p.width - maxWidth) / 2
	if blockStartX < 0 {
		blockStartX = 0
	}

	// Map text to terminal coordinates
	for lineIdx, line := range lines {
		runes := []rune(line)
		for charIdx, char := range runes {
			if char == ' ' || char == '\t' {
				continue // Skip whitespace
			}

			finalX := blockStartX + charIdx
			finalY := startY + lineIdx

			// Skip characters that would be off-screen
			if finalX >= p.width || finalY >= p.height {
				continue
			}

			// Calculate gradient color based on terminal coordinates
			color := p.getGradientColorForCoord(finalX, finalY)

			// Get starting position based on pour direction
			startX, startY := p.getStartPosition(finalX, finalY)

			p.chars = append(p.chars, PourCharacter{
				original:        char,
				finalX:          finalX,
				finalY:          finalY,
				startX:          startX,
				startY:          startY,
				currentX:        float64(startX),
				currentY:        float64(startY),
				visible:         false,
				color:           p.startingColor,
				finalColor:      color,
				progress:        0.0,
				gradientStep:    0,
				gradientCounter: 0,
			})
		}
	}

	// Group characters by row or column based on direction
	p.createGroups()
}

// Get starting position based on pour direction
func (p *PourEffect) getStartPosition(finalX, finalY int) (int, int) {
	switch p.pourDirection {
	case "down":
		return finalX, 0
	case "up":
		return finalX, p.height - 1
	case "left":
		return p.width - 1, finalY
	case "right":
		return 0, finalY
	default:
		return finalX, 0
	}
}

// Create groups of characters by row or column
func (p *PourEffect) createGroups() {
	if p.pourDirection == "up" || p.pourDirection == "down" {
		p.groupByRows()
	} else {
		p.groupByColumns()
	}
}

// Group characters by rows (for vertical pouring)
func (p *PourEffect) groupByRows() {
	// Create map of Y coordinate to character indices
	rowMap := make(map[int][]int)
	for i, char := range p.chars {
		rowMap[char.finalY] = append(rowMap[char.finalY], i)
	}

	// Get sorted row coordinates
	rows := make([]int, 0, len(rowMap))
	for y := range rowMap {
		rows = append(rows, y)
	}
	sort.Ints(rows)

	// Create groups in order (top to bottom for down, bottom to top for up)
	p.groups = make([][]int, 0)
	
	if p.pourDirection == "down" {
		// Pour top to bottom in order
		for _, y := range rows {
			p.groups = append(p.groups, rowMap[y])
		}
	} else {
		// Pour bottom to top in order
		for i := len(rows) - 1; i >= 0; i-- {
			p.groups = append(p.groups, rowMap[rows[i]])
		}
	}
}

// Group characters by columns (for horizontal pouring)
func (p *PourEffect) groupByColumns() {
	// Create map of X coordinate to character indices
	colMap := make(map[int][]int)
	for i, char := range p.chars {
		colMap[char.finalX] = append(colMap[char.finalX], i)
	}

	// Get sorted column coordinates
	cols := make([]int, 0, len(colMap))
	for x := range colMap {
		cols = append(cols, x)
	}
	sort.Ints(cols)

	// Create groups in order (left to right for right, right to left for left)
	p.groups = make([][]int, 0)
	
	if p.pourDirection == "right" {
		// Pour left to right in order
		for _, x := range cols {
			p.groups = append(p.groups, colMap[x])
		}
	} else {
		// Pour right to left in order
		for i := len(cols) - 1; i >= 0; i-- {
			p.groups = append(p.groups, colMap[cols[i]])
		}
	}
}

// Calculate gradient color for a specific coordinate
func (p *PourEffect) getGradientColorForCoord(x, y int) string {
	if len(p.finalGradientStops) == 0 {
		return "#ffffff"
	}
	if len(p.finalGradientStops) == 1 {
		return p.finalGradientStops[0]
	}

	var ratio float64

	if p.finalGradientDirection == "vertical" {
		// Vertical gradient based on Y position
		if p.height > 1 {
			ratio = float64(y) / float64(p.height-1)
		}
	} else {
		// Horizontal gradient based on X position
		if p.width > 1 {
			ratio = float64(x) / float64(p.width-1)
		}
	}

	// Map ratio to gradient stops
	step := int(ratio * float64(len(p.finalGradientStops)-1))
	if step >= len(p.finalGradientStops) {
		step = len(p.finalGradientStops) - 1
	}
	if step < 0 {
		step = 0
	}

	return p.finalGradientStops[step]
}

// Easing function for smooth movement
func (p *PourEffect) easeInQuad(t float64) float64 {
	return t * t
}

// Update advances the pour animation by one frame
func (p *PourEffect) Update() {
	p.frameCount++

	switch p.phase {
	case "pouring":
		p.updatePouringPhase()
	case "complete":
		p.updateCompletePhase()
	case "hold":
		p.updateHoldPhase()
	}
}

// updateCompletePhase handles the completion phase
func (p *PourEffect) updateCompletePhase() {
	// Continue updating movement and gradients
	p.updateCharacterMovement()
	p.updateCharacterGradients()

	// Check if all characters have finished moving AND reached their final colors
	allComplete := true
	for i := range p.chars {
		// Check if character hasn't finished moving
		if p.chars[i].progress < 1.0 {
			allComplete = false
			break
		}
		// Check if character hasn't reached final color
		if p.chars[i].gradientStep < p.finalGradientSteps {
			allComplete = false
			break
		}
	}

	if allComplete {
		p.phase = "hold"
		p.holdCounter = 0
	}
}

// updateHoldPhase handles holding the final state before looping
func (p *PourEffect) updateHoldPhase() {
	p.holdCounter++

	// Hold for 6 seconds at ~20fps = 120 frames
	if p.holdCounter >= 120 {
		p.Reset()
	}
}

// Update the pouring phase of the animation
func (p *PourEffect) updatePouringPhase() {
	// Handle gap between group pours
	if p.gapCounter > 0 {
		p.gapCounter--
		p.updateCharacterMovement()
		p.updateCharacterGradients()
		return
	}

	// Check if all groups are complete
	if p.currentGroup >= len(p.groups) {
		p.phase = "complete"
		p.updateCharacterMovement()
		p.updateCharacterGradients()
		return
	}

	// Pour characters from current group
	group := p.groups[p.currentGroup]
	poured := 0

	for poured < p.pourSpeed && p.currentInGroup < len(group) {
		charIdx := group[p.currentInGroup]
		if charIdx >= 0 && charIdx < len(p.chars) {
			p.chars[charIdx].visible = true
		}
		p.currentInGroup++
		poured++
	}

	// Check if current group is complete
	if p.currentInGroup >= len(group) {
		p.currentGroup++
		p.currentInGroup = 0
		p.gapCounter = p.gap
	}

	// Update all characters
	p.updateCharacterMovement()
	p.updateCharacterGradients()
}

// Update character movement animation
func (p *PourEffect) updateCharacterMovement() {
	for i := range p.chars {
		char := &p.chars[i]
		if !char.visible {
			continue
		}

		// Update progress
		char.progress += p.movementSpeed
		if char.progress > 1.0 {
			char.progress = 1.0
		}

		// Apply easing
		easedProgress := p.easeInQuad(char.progress)

		// Calculate new position
		char.currentX = float64(char.startX) + (float64(char.finalX)-float64(char.startX))*easedProgress
		char.currentY = float64(char.startY) + (float64(char.finalY)-float64(char.startY))*easedProgress

		// Snap to final position when complete
		if char.progress >= 1.0 {
			char.currentX = float64(char.finalX)
			char.currentY = float64(char.finalY)
		}
	}
}

// Update character gradient animation
func (p *PourEffect) updateCharacterGradients() {
	for i := range p.chars {
		char := &p.chars[i]
		if !char.visible || char.progress < 1.0 {
			continue
		}

		// Update gradient counter
		char.gradientCounter++

		// Change gradient step
		if char.gradientCounter >= p.finalGradientFrames {
			char.gradientCounter = 0
			char.gradientStep++

			// Interpolate from starting color to final color
			if char.gradientStep <= p.finalGradientSteps {
				ratio := float64(char.gradientStep) / float64(p.finalGradientSteps)
				if ratio > 1.0 {
					ratio = 1.0
				}
				char.color = p.interpolateColor(p.startingColor, char.finalColor, ratio)
			} else {
				char.color = char.finalColor
			}
		}
	}
}

// Interpolate between two colors
func (p *PourEffect) interpolateColor(startColor, endColor string, ratio float64) string {
	startR, startG, startB := p.parseHexColor(startColor)
	endR, endG, endB := p.parseHexColor(endColor)

	r := int(float64(startR) + float64(endR-startR)*ratio)
	g := int(float64(startG) + float64(endG-startG)*ratio)
	b := int(float64(startB) + float64(endB-startB)*ratio)

	r = int(math.Max(0, math.Min(255, float64(r))))
	g = int(math.Max(0, math.Min(255, float64(g))))
	b = int(math.Max(0, math.Min(255, float64(b))))

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Parse hex color string to RGB values
func (p *PourEffect) parseHexColor(hex string) (int, int, int) {
	if len(hex) < 7 || hex[0] != '#' {
		return 255, 255, 255
	}

	r, _ := strconv.ParseInt(hex[1:3], 16, 64)
	g, _ := strconv.ParseInt(hex[3:5], 16, 64)
	b, _ := strconv.ParseInt(hex[5:7], 16, 64)

	return int(r), int(g), int(b)
}

// Render converts the pour effect to colored text output
func (p *PourEffect) Render() string {
	buffer := make([][]string, p.height)
	for i := range buffer {
		buffer[i] = make([]string, p.width)
		for j := range buffer[i] {
			buffer[i][j] = " "
		}
	}

	// Render visible characters
	for _, char := range p.chars {
		if char.visible {
			x := int(math.Round(char.currentX))
			y := int(math.Round(char.currentY))

			if y >= 0 && y < p.height && x >= 0 && x < p.width {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(char.color))
				buffer[y][x] = style.Render(string(char.original))
			}
		}
	}

	// Convert buffer to string
	var lines []string
	for _, line := range buffer {
		lines = append(lines, strings.Join(line, ""))
	}

	return strings.Join(lines, "\n")
}

// Reset restarts the animation from the beginning
func (p *PourEffect) Reset() {
	p.phase = "pouring"
	p.frameCount = 0
	p.currentGroup = 0
	p.currentInGroup = 0
	p.gapCounter = 0

	for i := range p.chars {
		startX, startY := p.getStartPosition(p.chars[i].finalX, p.chars[i].finalY)
		p.chars[i].visible = false
		p.chars[i].startX = startX
		p.chars[i].startY = startY
		p.chars[i].currentX = float64(startX)
		p.chars[i].currentY = float64(startY)
		p.chars[i].progress = 0.0
		p.chars[i].color = p.startingColor
		p.chars[i].gradientStep = 0
		p.chars[i].gradientCounter = 0
	}
}
