package animations

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
)

// BeamsTextEffect implements beams that travel across text, illuminating characters
type BeamsTextEffect struct {
	width  int
	height int
	text   string

	// Configuration
	beamRowSymbols      []rune
	beamColumnSymbols   []rune
	beamDelay           int
	beamRowSpeedRange   [2]int
	beamColumnSpeedRange [2]int
	beamGradientStops   []string
	beamGradientSteps   int
	beamGradientFrames  int
	finalGradientStops  []string
	finalGradientSteps  int
	finalGradientFrames int
	finalWipeSpeed      int

	// Character data
	chars []BeamsCharacter

	// Beam groups
	rowGroups    []BeamsGroup
	columnGroups []BeamsGroup

	// Final wipe diagonal groups
	diagonalGroups [][]int

	// Animation state
	phase          string
	frameCount     int
	beamDelayCount int
	currentDiag    int
	holdCounter    int // Frames to hold after completion before reset

	rng *rand.Rand
}

// BeamsCharacter represents a single character in the beams animation
type BeamsCharacter struct {
	original rune
	x        int
	y        int

	// Animation state
	visible         bool
	currentSymbol   rune
	currentColor    string
	sceneActive     string
	sceneFrame      int
	beamGradient    []string
	fadeGradient    []string
	brightenGradient []string
}

// BeamsGroup represents a group of characters for beam animation
type BeamsGroup struct {
	charIndices        []int
	direction          string
	speed              float64
	nextCharCounter    float64
	currentCharIndex   int
	symbols            []rune
	beamGradientStops  []string
	beamGradientSteps  int
	beamGradientFrames int
	beamLength         int
}

// BeamsTextConfig holds configuration for the beams text effect
type BeamsTextConfig struct {
	Width                int
	Height               int
	Text                 string
	BeamRowSymbols       []rune
	BeamColumnSymbols    []rune
	BeamDelay            int
	BeamRowSpeedRange    [2]int
	BeamColumnSpeedRange [2]int
	BeamGradientStops    []string
	BeamGradientSteps    int
	BeamGradientFrames   int
	FinalGradientStops   []string
	FinalGradientSteps   int
	FinalGradientFrames  int
	FinalWipeSpeed       int
}

// NewBeamsTextEffect creates a new beams text effect
func NewBeamsTextEffect(config BeamsTextConfig) *BeamsTextEffect {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set defaults
	if len(config.BeamRowSymbols) == 0 {
		config.BeamRowSymbols = []rune{'▂', '▁', '_'}
	}
	if len(config.BeamColumnSymbols) == 0 {
		config.BeamColumnSymbols = []rune{'▌', '▍', '▎', '▏'}
	}
	if config.BeamDelay == 0 {
		config.BeamDelay = 2
	}
	if config.BeamRowSpeedRange[0] == 0 {
		config.BeamRowSpeedRange = [2]int{8, 32} // Further reduced by 20%
	}
	if config.BeamColumnSpeedRange[0] == 0 {
		config.BeamColumnSpeedRange = [2]int{6, 12} // Further reduced by 20%
	}
	if len(config.BeamGradientStops) == 0 {
		// Default gradient if none provided
		config.BeamGradientStops = []string{"#8b5cf6", "#06b6d4", "#10b981"}
	}
	if len(config.FinalGradientStops) == 0 {
		// Default gradient if none provided
		config.FinalGradientStops = []string{"#8b5cf6", "#06b6d4", "#10b981"}
	}
	if config.BeamGradientSteps == 0 {
		config.BeamGradientSteps = 5
	}
	if config.BeamGradientFrames == 0 {
		config.BeamGradientFrames = 1
	}
	if config.FinalGradientSteps == 0 {
		config.FinalGradientSteps = 8
	}
	if config.FinalGradientFrames == 0 {
		config.FinalGradientFrames = 1
	}
	if config.FinalWipeSpeed == 0 {
		config.FinalWipeSpeed = 3
	}

	b := &BeamsTextEffect{
		width:                config.Width,
		height:               config.Height,
		text:                 config.Text,
		beamRowSymbols:       config.BeamRowSymbols,
		beamColumnSymbols:    config.BeamColumnSymbols,
		beamDelay:            config.BeamDelay,
		beamRowSpeedRange:    config.BeamRowSpeedRange,
		beamColumnSpeedRange: config.BeamColumnSpeedRange,
		beamGradientStops:    config.BeamGradientStops,
		beamGradientSteps:    config.BeamGradientSteps,
		beamGradientFrames:   config.BeamGradientFrames,
		finalGradientStops:   config.FinalGradientStops,
		finalGradientSteps:   config.FinalGradientSteps,
		finalGradientFrames:  config.FinalGradientFrames,
		finalWipeSpeed:       config.FinalWipeSpeed,
		phase:                "beams",
		frameCount:           0,
		beamDelayCount:       0,
		currentDiag:          0,
		holdCounter:          0,
		rng:                  rng,
	}

	b.init()
	return b
}

// init initializes characters and beam groups
func (b *BeamsTextEffect) init() {
	lines := strings.Split(b.text, "\n")

	// Calculate centered position
	startY := (b.height - len(lines)) / 2
	if startY < 0 {
		startY = 0
	}

	maxWidth := 0
	for _, line := range lines {
		if len([]rune(line)) > maxWidth {
			maxWidth = len([]rune(line))
		}
	}

	blockStartX := (b.width - maxWidth) / 2
	if blockStartX < 0 {
		blockStartX = 0
	}

	// Create characters from text
	for lineIdx, line := range lines {
		runes := []rune(line)
		for charIdx, char := range runes {
			if char == ' ' || char == '\t' {
				continue
			}

			x := blockStartX + charIdx
			y := startY + lineIdx

			if x >= b.width || y >= b.height {
				continue
			}

			beamGradient := b.createGradient(b.beamGradientStops, b.beamGradientSteps)
			fadeGradient := b.createFadeGradient(beamGradient[len(beamGradient)-1], 5)
			brightenGradient := b.createGradient(b.finalGradientStops, b.finalGradientSteps)

			b.chars = append(b.chars, BeamsCharacter{
				original:         char,
				x:                x,
				y:                y,
				visible:          false,
				currentSymbol:    char,
				currentColor:     "",
				sceneActive:      "",
				sceneFrame:       0,
				beamGradient:     beamGradient,
				fadeGradient:     fadeGradient,
				brightenGradient: brightenGradient,
			})
		}
	}

	b.createRowGroups()
	b.createColumnGroups()
	b.shuffleGroups()
	b.createDiagonalGroups()
}

// createRowGroups creates beam groups for each row
func (b *BeamsTextEffect) createRowGroups() {
	rowMap := make(map[int][]int)
	for i, char := range b.chars {
		rowMap[char.y] = append(rowMap[char.y], i)
	}

	for _, indices := range rowMap {
		sort.Slice(indices, func(i, j int) bool {
			return b.chars[indices[i]].x < b.chars[indices[j]].x
		})

		if b.rng.Float64() < 0.5 {
			for i := 0; i < len(indices)/2; i++ {
				j := len(indices) - 1 - i
				indices[i], indices[j] = indices[j], indices[i]
			}
		}

		speed := float64(b.rng.Intn(b.beamRowSpeedRange[1]-b.beamRowSpeedRange[0])+b.beamRowSpeedRange[0]) * 0.1

		b.rowGroups = append(b.rowGroups, BeamsGroup{
			charIndices:        indices,
			direction:          "row",
			speed:              speed,
			nextCharCounter:    0,
			currentCharIndex:   0,
			symbols:            b.beamRowSymbols,
			beamGradientStops:  b.beamGradientStops,
			beamGradientSteps:  b.beamGradientSteps,
			beamGradientFrames: b.beamGradientFrames,
			beamLength:         len(b.beamRowSymbols),
		})
	}
}

// createColumnGroups creates beam groups for each column
func (b *BeamsTextEffect) createColumnGroups() {
	colMap := make(map[int][]int)
	for i, char := range b.chars {
		colMap[char.x] = append(colMap[char.x], i)
	}

	for _, indices := range colMap {
		sort.Slice(indices, func(i, j int) bool {
			return b.chars[indices[i]].y < b.chars[indices[j]].y
		})

		if b.rng.Float64() < 0.5 {
			for i := 0; i < len(indices)/2; i++ {
				j := len(indices) - 1 - i
				indices[i], indices[j] = indices[j], indices[i]
			}
		}

		speed := float64(b.rng.Intn(b.beamColumnSpeedRange[1]-b.beamColumnSpeedRange[0])+b.beamColumnSpeedRange[0]) * 0.1

		b.columnGroups = append(b.columnGroups, BeamsGroup{
			charIndices:        indices,
			direction:          "column",
			speed:              speed,
			nextCharCounter:    0,
			currentCharIndex:   0,
			symbols:            b.beamColumnSymbols,
			beamGradientStops:  b.beamGradientStops,
			beamGradientSteps:  b.beamGradientSteps,
			beamGradientFrames: b.beamGradientFrames,
			beamLength:         len(b.beamColumnSymbols),
		})
	}
}

// shuffleGroups shuffles row and column groups
func (b *BeamsTextEffect) shuffleGroups() {
	allGroups := append(b.rowGroups, b.columnGroups...)

	for i := len(allGroups) - 1; i > 0; i-- {
		j := b.rng.Intn(i + 1)
		allGroups[i], allGroups[j] = allGroups[j], allGroups[i]
	}

	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]

	for _, group := range allGroups {
		if group.direction == "row" {
			b.rowGroups = append(b.rowGroups, group)
		} else {
			b.columnGroups = append(b.columnGroups, group)
		}
	}
}

// createDiagonalGroups creates diagonal groups for final wipe
func (b *BeamsTextEffect) createDiagonalGroups() {
	diagMap := make(map[int][]int)
	for i, char := range b.chars {
		diag := char.x + char.y
		diagMap[diag] = append(diagMap[diag], i)
	}

	keys := make([]int, 0, len(diagMap))
	for k := range diagMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		b.diagonalGroups = append(b.diagonalGroups, diagMap[k])
	}
}

// createGradient creates a color gradient
func (b *BeamsTextEffect) createGradient(stops []string, steps int) []string {
	if len(stops) == 0 {
		return []string{"#ffffff"}
	}
	if len(stops) == 1 {
		return []string{stops[0]}
	}

	var gradient []string
	stepsPerSegment := steps / (len(stops) - 1)

	for i := 0; i < len(stops)-1; i++ {
		c1 := parseBeamsHexColor(stops[i])
		c2 := parseBeamsHexColor(stops[i+1])

		for j := 0; j < stepsPerSegment; j++ {
			t := float64(j) / float64(stepsPerSegment)
			r := uint8(float64(c1[0])*(1-t) + float64(c2[0])*t)
			g := uint8(float64(c1[1])*(1-t) + float64(c2[1])*t)
			b := uint8(float64(c1[2])*(1-t) + float64(c2[2])*t)
			gradient = append(gradient, formatBeamsHexColor([3]uint8{r, g, b}))
		}
	}

	gradient = append(gradient, stops[len(stops)-1])
	return gradient
}

// createFadeGradient creates a fade to dark gradient
func (b *BeamsTextEffect) createFadeGradient(startColor string, steps int) []string {
	rgb := parseBeamsHexColor(startColor)
	targetRGB := [3]uint8{
		uint8(float64(rgb[0]) * 0.3),
		uint8(float64(rgb[1]) * 0.3),
		uint8(float64(rgb[2]) * 0.3),
	}

	var gradient []string
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		r := uint8(float64(rgb[0])*(1-t) + float64(targetRGB[0])*t)
		g := uint8(float64(rgb[1])*(1-t) + float64(targetRGB[1])*t)
		b := uint8(float64(rgb[2])*(1-t) + float64(targetRGB[2])*t)
		gradient = append(gradient, formatBeamsHexColor([3]uint8{r, g, b}))
	}

	return gradient
}

// Update advances the animation by one frame
func (b *BeamsTextEffect) Update() {
	b.frameCount++

	if b.phase == "beams" {
		b.updateBeamsPhase()
	} else if b.phase == "final_wipe" {
		b.updateFinalWipePhase()
	} else if b.phase == "hold" {
		b.updateHoldPhase()
	}

	b.updateCharacterAnimations()
}

// updateBeamsPhase handles beam movement
func (b *BeamsTextEffect) updateBeamsPhase() {
	if b.beamDelayCount > 0 {
		b.beamDelayCount--
		return
	}

	groupsToActivate := b.rng.Intn(5) + 1
	activated := false

	for i := 0; i < groupsToActivate; i++ {
		for j := range b.rowGroups {
			if b.rowGroups[j].currentCharIndex == 0 && b.rowGroups[j].nextCharCounter == 0 {
				b.rowGroups[j].nextCharCounter = 0.01
				activated = true
				break
			}
		}

		for j := range b.columnGroups {
			if b.columnGroups[j].currentCharIndex == 0 && b.columnGroups[j].nextCharCounter == 0 {
				b.columnGroups[j].nextCharCounter = 0.01
				activated = true
				break
			}
		}
	}

	if activated {
		b.beamDelayCount = b.beamDelay
	}

	allGroupsComplete := true

	for i := range b.rowGroups {
		if b.updateGroup(&b.rowGroups[i]) {
			allGroupsComplete = false
		}
	}

	for i := range b.columnGroups {
		if b.updateGroup(&b.columnGroups[i]) {
			allGroupsComplete = false
		}
	}

	if allGroupsComplete {
		b.phase = "final_wipe"
	}
}

// updateGroup updates a beam group
func (b *BeamsTextEffect) updateGroup(group *BeamsGroup) bool {
	if group.nextCharCounter == 0 {
		return false
	}

	if group.currentCharIndex >= len(group.charIndices) {
		return false
	}

	group.nextCharCounter += group.speed

	charsToActivate := int(group.nextCharCounter)
	group.nextCharCounter -= float64(charsToActivate)

	for i := 0; i < charsToActivate && group.currentCharIndex < len(group.charIndices); i++ {
		charIdx := group.charIndices[group.currentCharIndex]
		char := &b.chars[charIdx]

		if group.direction == "row" {
			char.sceneActive = "beam_row"
		} else {
			char.sceneActive = "beam_column"
		}
		char.sceneFrame = 0
		char.visible = true

		symbolIndex := 0
		char.currentSymbol = group.symbols[symbolIndex]

		for j := 1; j < group.beamLength && group.currentCharIndex-j >= 0; j++ {
			trailCharIdx := group.charIndices[group.currentCharIndex-j]
			trailChar := &b.chars[trailCharIdx]

			if trailChar.sceneActive == "beam_row" || trailChar.sceneActive == "beam_column" {
				symbolIdx := j
				if symbolIdx >= len(group.symbols) {
					symbolIdx = len(group.symbols) - 1
				}
				trailChar.currentSymbol = group.symbols[symbolIdx]
			}
		}

		group.currentCharIndex++
	}

	return true
}

// updateFinalWipePhase handles final diagonal wipe
func (b *BeamsTextEffect) updateFinalWipePhase() {
	for i := 0; i < b.finalWipeSpeed && b.currentDiag < len(b.diagonalGroups); i++ {
		for _, charIdx := range b.diagonalGroups[b.currentDiag] {
			char := &b.chars[charIdx]
			char.sceneActive = "brighten"
			char.sceneFrame = 0
			char.visible = true
			char.currentSymbol = char.original
		}
		b.currentDiag++
	}

	if b.currentDiag >= len(b.diagonalGroups) {
		allComplete := true
		for i := range b.chars {
			char := &b.chars[i]
			if char.sceneActive == "brighten" {
				gradientLen := len(char.brightenGradient)
				framesPerStep := b.finalGradientFrames
				totalFrames := gradientLen * framesPerStep
				if char.sceneFrame < totalFrames {
					allComplete = false
					break
				}
			}
		}

		if allComplete {
			b.phase = "hold"
			b.holdCounter = 0
		}
	}
}

// updateHoldPhase handles the hold period before reset
func (b *BeamsTextEffect) updateHoldPhase() {
	b.holdCounter++
	
	// Hold for 6 seconds at ~20fps = 120 frames
	if b.holdCounter >= 120 {
		b.Reset()
	}
}

// updateCharacterAnimations updates character scenes
func (b *BeamsTextEffect) updateCharacterAnimations() {
	for i := range b.chars {
		char := &b.chars[i]

		if !char.visible {
			continue
		}

		switch char.sceneActive {
		case "beam_row", "beam_column":
			gradientLen := len(char.beamGradient)
			if gradientLen == 0 {
				break
			}

			framesPerStep := b.beamGradientFrames
			totalFrames := gradientLen * framesPerStep

			if char.sceneFrame < totalFrames {
				step := char.sceneFrame / framesPerStep
				if step >= gradientLen {
					step = gradientLen - 1
				}
				char.currentColor = char.beamGradient[step]
				char.sceneFrame++
			} else {
				char.sceneActive = "fade"
				char.sceneFrame = 0
			}

		case "fade":
			fadeLen := len(char.fadeGradient)
			if fadeLen == 0 {
				char.sceneActive = ""
				char.currentSymbol = char.original
				break
			}

			if char.sceneFrame < fadeLen {
				char.currentColor = char.fadeGradient[char.sceneFrame]
				char.sceneFrame++
			} else {
				char.sceneActive = ""
				char.currentSymbol = char.original
			}

		case "brighten":
			gradientLen := len(char.brightenGradient)
			if gradientLen == 0 {
				break
			}

			framesPerStep := b.finalGradientFrames
			totalFrames := gradientLen * framesPerStep

			if char.sceneFrame < totalFrames {
				step := char.sceneFrame / framesPerStep
				if step >= gradientLen {
					step = gradientLen - 1
				}
				char.currentColor = char.brightenGradient[step]
				char.sceneFrame++
			}
		}
	}
}

// Render returns the current frame as colored text
func (b *BeamsTextEffect) Render() string {
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

	for _, char := range b.chars {
		if !char.visible {
			continue
		}

		if char.y >= 0 && char.y < b.height && char.x >= 0 && char.x < b.width {
			canvas[char.y][char.x] = char.currentSymbol
			colors[char.y][char.x] = char.currentColor
		}
	}

	var lines []string
	for y := 0; y < b.height; y++ {
		var line strings.Builder
		for x := 0; x < b.width; x++ {
			char := canvas[y][x]
			if char != ' ' && colors[y][x] != "" {
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

// Reset restarts the animation
func (b *BeamsTextEffect) Reset() {
	b.phase = "beams"
	b.frameCount = 0
	b.beamDelayCount = 0
	b.currentDiag = 0
	b.holdCounter = 0

	for i := range b.chars {
		b.chars[i].visible = false
		b.chars[i].sceneActive = ""
		b.chars[i].sceneFrame = 0
		b.chars[i].currentSymbol = b.chars[i].original
		b.chars[i].currentColor = ""
	}

	for i := range b.rowGroups {
		b.rowGroups[i].nextCharCounter = 0
		b.rowGroups[i].currentCharIndex = 0
	}
	for i := range b.columnGroups {
		b.columnGroups[i].nextCharCounter = 0
		b.columnGroups[i].currentCharIndex = 0
	}
}

// UpdateText changes the text and reinitializes
func (b *BeamsTextEffect) UpdateText(text string) {
	b.text = text
	b.chars = b.chars[:0]
	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]
	b.diagonalGroups = b.diagonalGroups[:0]
	b.init()
}

// parseBeamsHexColor converts hex to RGB
func parseBeamsHexColor(hex string) [3]uint8 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return [3]uint8{255, 255, 255}
	}

	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return [3]uint8{r, g, b}
}

// formatBeamsHexColor converts RGB to hex
func formatBeamsHexColor(rgb [3]uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
}

// Resize reinitializes with new dimensions
func (b *BeamsTextEffect) Resize(width, height int) {
	b.width = width
	b.height = height
	b.chars = b.chars[:0]
	b.rowGroups = b.rowGroups[:0]
	b.columnGroups = b.columnGroups[:0]
	b.diagonalGroups = b.diagonalGroups[:0]
	b.init()
}

// adjustBeamsColorBrightness adjusts color brightness
func adjustBeamsColorBrightness(color string, factor float64) string {
	rgb := parseBeamsHexColor(color)
	r := uint8(math.Min(255, float64(rgb[0])*factor))
	g := uint8(math.Min(255, float64(rgb[1])*factor))
	b := uint8(math.Min(255, float64(rgb[2])*factor))
	return formatBeamsHexColor([3]uint8{r, g, b})
}
