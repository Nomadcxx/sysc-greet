package animations

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// AnimationPhase represents the current act of the sonar animation.
type AnimationPhase int

const (
	// PhaseDrip — Act 1: scanline sweeps once down the screen.
	PhaseDrip AnimationPhase = iota

	// PhaseGrid — Act 2: tron-style horizon grid ignites left-to-right.
	PhaseGrid

	// PhaseHold — Act 3: grid holds steady, sonar pulses radiate from centre.
	PhaseHold
)

// AshParticle represents falling ash with horizontal drift.
type AshParticle struct {
	x          float64
	y          float64
	char       rune
	velocityY  float64
	driftPhase float64
	driftAmp   float64
	driftFreq  float64
	layer      int // 0=light, 1=dense
	opacity    float64
}

// gridCell marks a background grid glyph.
type gridCell struct {
	x, y int
	ch   rune
}

// SonarRing is a single expanding pulse. Birth frame stores when it was
// emitted so radius + intensity can be derived at render time.
type SonarRing struct {
	birthFrame int
}

// SonarAnimation is a three-act cyberpunk sequence (scan sweep → tron grid
// ignition → multi-ring sonar pulses) radiating from the viewport centre
// out to the edges.
type SonarAnimation struct {
	width, height int
	palette       []string
	theme         string

	centerX int
	centerY int

	gridCells []gridCell
	horizonY  int

	ashParticles []AshParticle

	rings         []SonarRing
	spawnInterval int
	lifespan      int
	maxRadius     float64

	phase          AnimationPhase
	frameCount     int
	scanFrames     int
	gridStartFrame int
	gridDuration   int
	holdStartFrame int

	builder strings.Builder
}

// NewSonarEffect builds the sonar scene.
// Palette slots: 0 = dim grid, 1 = bright pulse, 2 = scanline, 5/6 = ash.
func NewSonarEffect(width, height int, palette []string, theme string) *SonarAnimation {
	s := &SonarAnimation{
		width:          width,
		height:         height,
		palette:        palette,
		theme:          theme,
		centerX:        width / 2,
		centerY:        height / 2,
		phase:          PhaseDrip,
		scanFrames:     60,
		gridDuration:   50,
		gridStartFrame: -1,
		holdStartFrame: -1,
		spawnInterval:  22,
		lifespan:       110,
	}

	dx := float64(width) / 2.0
	dy := float64(height) / 2.0 * 2.0
	s.maxRadius = math.Sqrt(dx*dx+dy*dy) + 1.0

	s.buildGrid()

	for i := 0; i < 50; i++ {
		s.spawnAshParticle()
	}

	return s
}

// UpdatePalette swaps colors (used on theme change).
func (s *SonarAnimation) UpdatePalette(palette []string) {
	s.palette = palette
}

func (s *SonarAnimation) buildGrid() {
	s.horizonY = s.centerY
	s.gridCells = s.gridCells[:0]

	rowOffsets := []int{2, 4, 7, 11, 16}
	for _, off := range rowOffsets {
		y := s.horizonY + off
		if y < 0 || y >= s.height {
			continue
		}
		for x := 0; x < s.width; x += 2 {
			s.gridCells = append(s.gridCells, gridCell{x: x, y: y, ch: '─'})
		}
	}

	lanes := []int{-6, -3, 3, 6}
	for _, lane := range lanes {
		for y := s.horizonY + 1; y < s.height; y++ {
			dy := y - s.horizonY
			x := s.centerX + lane*dy/3
			if x < 0 || x >= s.width {
				continue
			}
			s.gridCells = append(s.gridCells, gridCell{x: x, y: y, ch: '│'})
		}
	}
}

func (s *SonarAnimation) spawnAshParticle() {
	densityMod := 0.9 + 0.1*math.Sin(float64(s.frameCount)/100.0)
	if rand.Float64() > densityMod {
		return
	}

	layer := 0
	if rand.Float64() < 0.3 {
		layer = 1
	}

	var char rune
	var velocityY, opacity float64
	if layer == 0 {
		chars := []rune{'.', ',', '\'', '·'}
		char = chars[rand.Intn(len(chars))]
		velocityY = 0.5 + rand.Float64()
		opacity = 0.4 + rand.Float64()*0.3
	} else {
		chars := []rune{'▒', '░', '·', '*'}
		char = chars[rand.Intn(len(chars))]
		velocityY = 0.3 + rand.Float64()*0.5
		opacity = 0.6 + rand.Float64()*0.3
	}

	s.ashParticles = append(s.ashParticles, AshParticle{
		x:          float64(rand.Intn(s.width)),
		y:          0,
		char:       char,
		velocityY:  velocityY,
		driftPhase: rand.Float64() * 2 * math.Pi,
		driftAmp:   0.1 + rand.Float64()*0.4,
		driftFreq:  0.05 + rand.Float64()*0.05,
		layer:      layer,
		opacity:    opacity,
	})
}

func (s *SonarAnimation) updateAsh() {
	toKeep := s.ashParticles[:0]
	for _, p := range s.ashParticles {
		p.y += p.velocityY
		p.driftPhase += p.driftFreq
		p.x += math.Sin(p.driftPhase) * p.driftAmp

		if p.layer == 0 && p.y > float64(s.height)*0.7 {
			fadeStart := float64(s.height) * 0.7
			fadeRange := float64(s.height) * 0.3
			fadeProgress := (p.y - fadeStart) / fadeRange
			p.opacity *= 1.0 - fadeProgress*0.5
		}

		accumZone := 3
		if p.layer == 1 {
			if p.y < float64(s.height+accumZone) {
				toKeep = append(toKeep, p)
			}
		} else if p.y < float64(s.height) {
			toKeep = append(toKeep, p)
		}
	}
	s.ashParticles = toKeep

	spawnCount := 2 + rand.Intn(3)
	for i := 0; i < spawnCount; i++ {
		s.spawnAshParticle()
	}
}

// Update advances the animation by one frame.
func (s *SonarAnimation) Update() {
	s.frameCount++
	s.updateAsh()

	switch s.phase {
	case PhaseDrip:
		if s.frameCount >= s.scanFrames {
			s.phase = PhaseGrid
			s.gridStartFrame = s.frameCount
		}
	case PhaseGrid:
		if s.frameCount-s.gridStartFrame >= s.gridDuration {
			s.phase = PhaseHold
			s.holdStartFrame = s.frameCount
		}
	case PhaseHold:
		if (s.frameCount-s.holdStartFrame)%s.spawnInterval == 0 {
			s.rings = append(s.rings, SonarRing{birthFrame: s.frameCount})
		}
		kept := s.rings[:0]
		for _, r := range s.rings {
			if s.frameCount-r.birthFrame < s.lifespan {
				kept = append(kept, r)
			}
		}
		s.rings = kept
	}
}

func (s *SonarAnimation) scanRow() int {
	if s.phase != PhaseDrip {
		return -1
	}
	if s.frameCount >= s.scanFrames {
		return -1
	}
	return s.frameCount * s.height / s.scanFrames
}

func (s *SonarAnimation) gridLitFraction() float64 {
	switch s.phase {
	case PhaseDrip:
		return 0
	case PhaseGrid:
		t := float64(s.frameCount-s.gridStartFrame) / float64(s.gridDuration)
		if t > 1 {
			t = 1
		}
		return t
	default:
		return 1
	}
}

func (s *SonarAnimation) ringRadius(r SonarRing) (float64, float64) {
	age := s.frameCount - r.birthFrame
	if age < 0 || age >= s.lifespan {
		return 0, 0
	}
	t := float64(age) / float64(s.lifespan)
	return t * s.maxRadius, 1.0 - t
}

func (s *SonarAnimation) onRing(x, y int, radius float64) bool {
	dx := float64(x - s.centerX)
	dy := float64(y-s.centerY) * 2.0
	d := math.Sqrt(dx*dx + dy*dy)
	return math.Abs(d-radius) < 0.9
}

// hexLerp linearly interpolates between two hex colour strings.
func hexLerp(from, to string, t float64) string {
	if t <= 0 {
		return from
	}
	if t >= 1 {
		return to
	}
	r1, g1, b1 := hexToRGBSonar(from)
	r2, g2, b2 := hexToRGBSonar(to)
	r := r1 + int(math.Round(t*float64(r2-r1)))
	g := g1 + int(math.Round(t*float64(g2-g1)))
	b := b1 + int(math.Round(t*float64(b2-b1)))
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// hexToRGBSonar parses a hex color string. Named to avoid clashing with
// other effects that may define hexToRGB.
func hexToRGBSonar(hex string) (int, int, int) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	var r, g, b int
	if len(hex) == 6 {
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	}
	return r, g, b
}

// Render produces the current frame.
func (s *SonarAnimation) Render() string {
	s.builder.Reset()

	ashMap := make(map[[2]int]AshParticle, len(s.ashParticles))
	for _, p := range s.ashParticles {
		x := int(p.x)
		y := int(p.y)
		if x >= 0 && x < s.width && y >= 0 && y < s.height {
			ashMap[[2]int{x, y}] = p
		}
	}

	gridMap := make(map[[2]int]rune, len(s.gridCells))
	litFrac := s.gridLitFraction()
	litX := int(float64(s.width) * litFrac)
	for _, g := range s.gridCells {
		if g.x >= litX && s.phase != PhaseHold {
			continue
		}
		if _, exists := gridMap[[2]int{g.x, g.y}]; !exists {
			gridMap[[2]int{g.x, g.y}] = g.ch
		}
	}

	type ringHit struct {
		intensity float64
	}
	ringMap := make(map[[2]int]ringHit)
	for _, r := range s.rings {
		radius, intensity := s.ringRadius(r)
		if radius <= 0 {
			continue
		}
		for y := 0; y < s.height; y++ {
			dy := float64(y-s.centerY) * 2.0
			if math.Abs(dy) > radius+1 {
				continue
			}
			for x := 0; x < s.width; x++ {
				if s.onRing(x, y, radius) {
					if prev, ok := ringMap[[2]int{x, y}]; !ok || intensity > prev.intensity {
						ringMap[[2]int{x, y}] = ringHit{intensity: intensity}
					}
				}
			}
		}
	}

	scanY := s.scanRow()
	gridColor := s.paletteSlot(0)
	scanColor := s.paletteSlot(2)

	for y := 0; y < s.height; y++ {
		// Run-length SGR state: emit a color code only when the color changes
		// along the row, reset before default-colored cells and at row end
		lastR, lastG, lastB := -1, -1, -1

		for x := 0; x < s.width; x++ {
			pos := [2]int{x, y}

			var cr, cg, cb int
			var ch rune
			colored := true

			if ap, ok := ashMap[pos]; ok {
				var col string
				if ap.layer == 0 {
					col = s.paletteSlot(5)
				} else {
					col = s.paletteSlot(6)
				}
				cr, cg, cb = hexToRGBSonar(col)
				ch = ap.char
			} else if y == scanY {
				cr, cg, cb = hexToRGBSonar(scanColor)
				ch = '─'
			} else if h, ok := ringMap[pos]; ok {
				col := hexLerp(s.paletteSlot(0), s.paletteSlot(1), h.intensity)
				cr, cg, cb = hexToRGBSonar(col)
				ch = '·'
			} else if gch, ok := gridMap[pos]; ok {
				cr, cg, cb = hexToRGBSonar(gridColor)
				ch = gch
			} else {
				colored = false
			}

			if !colored {
				if lastR != -1 {
					s.builder.WriteString("\x1b[0m")
					lastR, lastG, lastB = -1, -1, -1
				}
				s.builder.WriteByte(' ')
				continue
			}

			if cr != lastR || cg != lastG || cb != lastB {
				writeRGBSGR(&s.builder, cr, cg, cb)
				lastR, lastG, lastB = cr, cg, cb
			}
			s.builder.WriteRune(ch)
		}

		if lastR != -1 {
			s.builder.WriteString("\x1b[0m")
		}
		if y < s.height-1 {
			s.builder.WriteByte('\n')
		}
	}

	return s.builder.String()
}

// paletteSlot returns the palette entry at index, falling back to the last
// valid entry if the palette is shorter than expected.
func (s *SonarAnimation) paletteSlot(i int) string {
	if len(s.palette) == 0 {
		return "#ffffff"
	}
	if i >= len(s.palette) {
		return s.palette[len(s.palette)-1]
	}
	return s.palette[i]
}

// Reset returns the animation to frame 0.
func (s *SonarAnimation) Reset() {
	s.frameCount = 0
	s.phase = PhaseDrip
	s.gridStartFrame = -1
	s.holdStartFrame = -1
	s.rings = s.rings[:0]
	s.ashParticles = s.ashParticles[:0]
	for i := 0; i < 50; i++ {
		s.spawnAshParticle()
	}
}

// Resize handles terminal size changes.
func (s *SonarAnimation) Resize(width, height int) {
	s.width = width
	s.height = height
	s.centerX = width / 2
	s.centerY = height / 2
	dx := float64(width) / 2.0
	dy := float64(height) / 2.0 * 2.0
	s.maxRadius = math.Sqrt(dx*dx+dy*dy) + 1.0
	s.buildGrid()
}
