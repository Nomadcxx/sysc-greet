package animations

import (
	"math"
	"math/rand"
	"strings"
)

type PlasmaPhase int

const (
	PlasmaFadeIn PlasmaPhase = iota
	PlasmaRun
)

type plasmaBlob struct {
	freqX  float64
	freqY  float64
	phaseX float64
	phaseY float64
	radius float64
}

type PlasmaAnimation struct {
	width, height int
	palette       []string
	theme         string

	blobs        []plasmaBlob
	phase        PlasmaPhase
	frameCount   int
	fadeInFrames int
	paletteShift float64

	builder strings.Builder
}

func NewPlasmaEffect(width, height int, palette []string, theme string) *PlasmaAnimation {
	p := &PlasmaAnimation{
		width:        width,
		height:       height,
		palette:      palette,
		theme:        theme,
		phase:        PlasmaFadeIn,
		fadeInFrames: 40,
		paletteShift: 0,
	}
	p.initBlobs()
	return p
}

func (p *PlasmaAnimation) initBlobs() {
	p.blobs = make([]plasmaBlob, 8)
	for i := range p.blobs {
		p.blobs[i] = plasmaBlob{
			freqX:  0.3 + rand.Float64()*0.7,
			freqY:  0.2 + rand.Float64()*0.6,
			phaseX: rand.Float64() * math.Pi * 2,
			phaseY: rand.Float64() * math.Pi * 2,
			radius: 6.0 + rand.Float64()*6.0,
		}
		// Ensure no two blobs have identical frequency ratios
		if i%2 == 0 {
			p.blobs[i].freqX *= 1.0 + float64(i)*0.15
		} else {
			p.blobs[i].freqY *= 1.0 + float64(i)*0.12
		}
	}
}

func (p *PlasmaAnimation) UpdatePalette(palette []string) {
	p.palette = palette
}

func (p *PlasmaAnimation) Update() {
	p.frameCount++
	if p.phase == PlasmaFadeIn && p.frameCount >= p.fadeInFrames {
		p.phase = PlasmaRun
	}
	p.paletteShift += 0.02
}

func (p *PlasmaAnimation) Resize(width, height int) {
	p.width = width
	p.height = height
}

func (p *PlasmaAnimation) Reset() {
	p.frameCount = 0
	p.phase = PlasmaFadeIn
	p.paletteShift = 0
	p.initBlobs()
}

func (p *PlasmaAnimation) Render() string {
	p.builder.Reset()

	fade := 1.0
	if p.phase == PlasmaFadeIn {
		fade = float64(p.frameCount) / float64(p.fadeInFrames)
		if fade > 1.0 {
			fade = 1.0
		}
	}

	t := float64(p.frameCount) * 0.03

	// Precompute blob positions on Lissajous curves
	type blobPos struct {
		cx, cy float64
	}
	positions := make([]blobPos, len(p.blobs))
	for i, blob := range p.blobs {
		positions[i] = blobPos{
			cx: float64(p.width) * 0.5 * (1.0 + 0.8*math.Sin(blob.freqX*t+blob.phaseX)),
			cy: float64(p.height) * 0.5 * (1.0 + 0.8*math.Cos(blob.freqY*t+blob.phaseY)),
		}
	}

	numColors := len(p.palette)
	if numColors == 0 {
		numColors = 1
	}

	// Chunky rendering: 2 chars wide per visual pixel
	chunkyW := p.width / 2

	for y := 0; y < p.height; y++ {
		// CRT scanline dimming: alternate rows at 85%
		scanlineDim := 1.0
		if y%2 == 1 {
			scanlineDim = 0.85
		}

		// Run-length SGR state: only emit a color code when the pixel color
		// changes along the row (replaces one Fprintf + reset per pixel)
		lastR, lastG, lastB := -1, -1, -1

		for cx := 0; cx < chunkyW; cx++ {
			// Center of this chunky pixel
			px := float64(cx*2) + 1.0
			py := float64(y)

			// Compute metaball energy field
			energy := 0.0
			var weightedR, weightedG, weightedB float64

			for i, blob := range p.blobs {
				dx := px - positions[i].cx
				dy := (py - positions[i].cy) * 2.0 // aspect ratio correction
				distSq := dx*dx + dy*dy
				if distSq < 1.0 {
					distSq = 1.0
				}
				contribution := (blob.radius * blob.radius) / distSq
				energy += contribution

				// Color contribution: palette cycling per blob
				colorPhase := p.paletteShift + float64(i)*0.7
				colorIdx := int(math.Floor(colorPhase)) % numColors
				if colorIdx < 0 {
					colorIdx += numColors
				}
				col := p.plasmaSlot(colorIdx)
				r, g, b := hexToRGBPlasma(col)

				weightedR += float64(r) * contribution
				weightedG += float64(g) * contribution
				weightedB += float64(b) * contribution
			}

			// Determine render character and color based on energy field
			var ch rune
			var intensity float64

			if energy > 0.8 {
				// Inside blob core: full block
				ch = '\u2588'
				intensity = 1.0
			} else if energy > 0.5 {
				// Dense blob surface
				ch = '\u2593'
				intensity = 0.85
			} else if energy > 0.3 {
				// Glow zone: medium shade
				ch = '\u2592'
				intensity = 0.65
			} else if energy > 0.15 {
				// Outer halo: light shade
				ch = '\u2591'
				intensity = 0.4
			} else if energy > 0.05 {
				// Far glow: subtle dot pattern
				ch = '\u00B7'
				intensity = 0.25
			} else {
				// Deep background: very dim noise
				ch = '\u2219'
				intensity = 0.1
			}

			// Normalize weighted color
			if energy > 0 {
				weightedR /= energy
				weightedG /= energy
				weightedB /= energy
			} else {
				// For zero-energy cells, use a cycling palette color
				colorPhase := p.paletteShift + px*0.02 + py*0.03
				colorIdx := int(math.Floor(colorPhase)) % numColors
				if colorIdx < 0 {
					colorIdx += numColors
				}
				col := p.plasmaSlot(colorIdx)
				r, g, b := hexToRGBPlasma(col)
				weightedR = float64(r)
				weightedG = float64(g)
				weightedB = float64(b)
			}

			// Apply intensity, scanline dimming, and fade-in
			finalR := plasmaClamp(int(weightedR*intensity*scanlineDim*fade), 0, 255)
			finalG := plasmaClamp(int(weightedG*intensity*scanlineDim*fade), 0, 255)
			finalB := plasmaClamp(int(weightedB*intensity*scanlineDim*fade), 0, 255)

			// Write chunky pixel (2 chars wide)
			if finalR != lastR || finalG != lastG || finalB != lastB {
				p.builder.WriteString("\x1b[38;2;")
				writeByteDecimal(&p.builder, finalR)
				p.builder.WriteByte(';')
				writeByteDecimal(&p.builder, finalG)
				p.builder.WriteByte(';')
				writeByteDecimal(&p.builder, finalB)
				p.builder.WriteByte('m')
				lastR, lastG, lastB = finalR, finalG, finalB
			}
			p.builder.WriteRune(ch)
			p.builder.WriteRune(ch)
		}

		// Reset at row end so colors never bleed into layer padding
		p.builder.WriteString("\x1b[0m")

		// Fill remaining columns if width is odd
		if p.width%2 != 0 {
			p.builder.WriteByte(' ')
		}

		if y < p.height-1 {
			p.builder.WriteByte('\n')
		}
	}

	return p.builder.String()
}

func (p *PlasmaAnimation) plasmaSlot(i int) string {
	if len(p.palette) == 0 {
		return "#ffffff"
	}
	if i >= len(p.palette) {
		return p.palette[len(p.palette)-1]
	}
	if i < 0 {
		return p.palette[0]
	}
	return p.palette[i]
}

func plasmaClamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func hexToRGBPlasma(hex string) (int, int, int) {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 255, 255, 255
	}
	r := hexBytePlasma(hex[0])*16 + hexBytePlasma(hex[1])
	g := hexBytePlasma(hex[2])*16 + hexBytePlasma(hex[3])
	b := hexBytePlasma(hex[4])*16 + hexBytePlasma(hex[5])
	return r, g, b
}

func hexBytePlasma(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}
