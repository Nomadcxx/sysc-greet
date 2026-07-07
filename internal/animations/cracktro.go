package animations

import (
	"math"
	"math/rand"
	"strings"
)

type CracktroPhase int

const (
	CracktroBoot CracktroPhase = iota
	CracktroRun
)

type cracktroStar struct {
	x     float64
	y     int
	speed float64
	layer int
}

type cracktroBar struct {
	phase    float64
	speed    float64
	colorIdx int
}

type CracktroAnimation struct {
	width, height int
	palette       []string
	theme         string

	stars      []cracktroStar
	bars       []cracktroBar
	phase      CracktroPhase
	frameCount int
	bootFrames int

	builder strings.Builder
}

func NewCracktroEffect(width, height int, palette []string, theme string) *CracktroAnimation {
	c := &CracktroAnimation{
		width:      width,
		height:     height,
		palette:    palette,
		theme:      theme,
		phase:      CracktroBoot,
		bootFrames: 45,
	}

	c.initCracktroStars()
	c.initCracktroBars()
	return c
}

func (c *CracktroAnimation) initCracktroStars() {
	c.stars = nil
	totalStars := (c.width * c.height) / 25
	if totalStars < 40 {
		totalStars = 40
	}
	if totalStars > 200 {
		totalStars = 200
	}
	for i := 0; i < totalStars; i++ {
		layer := rand.Intn(3)
		var speed float64
		switch layer {
		case 0:
			speed = 0.3 + rand.Float64()*0.2
		case 1:
			speed = 0.7 + rand.Float64()*0.3
		case 2:
			speed = 1.2 + rand.Float64()*0.5
		}
		c.stars = append(c.stars, cracktroStar{
			x:     rand.Float64() * float64(c.width),
			y:     rand.Intn(c.height),
			speed: speed,
			layer: layer,
		})
	}
}

func (c *CracktroAnimation) initCracktroBars() {
	c.bars = []cracktroBar{
		{phase: 0, speed: 0.06, colorIdx: 3},
		{phase: math.Pi * 0.5, speed: 0.05, colorIdx: 4},
		{phase: math.Pi, speed: 0.07, colorIdx: 3},
		{phase: math.Pi * 1.5, speed: 0.055, colorIdx: 4},
	}
}

func (c *CracktroAnimation) UpdatePalette(palette []string) {
	c.palette = palette
}

func (c *CracktroAnimation) Update() {
	c.frameCount++

	if c.phase == CracktroBoot && c.frameCount >= c.bootFrames {
		c.phase = CracktroRun
	}

	for i := range c.stars {
		c.stars[i].x -= c.stars[i].speed
		if c.stars[i].x < 0 {
			c.stars[i].x = float64(c.width) + rand.Float64()*10
			c.stars[i].y = rand.Intn(c.height)
		}
	}

	for i := range c.bars {
		c.bars[i].phase += c.bars[i].speed
	}
}

func (c *CracktroAnimation) Resize(width, height int) {
	c.width = width
	c.height = height
	c.initCracktroStars()
}

func (c *CracktroAnimation) Reset() {
	c.frameCount = 0
	c.phase = CracktroBoot
	c.initCracktroStars()
	c.initCracktroBars()
}

func (c *CracktroAnimation) Render() string {
	c.builder.Reset()

	bootFade := 1.0
	if c.phase == CracktroBoot {
		bootFade = float64(c.frameCount) / float64(c.bootFrames)
		if bootFade > 1.0 {
			bootFade = 1.0
		}
	}

	starMap := make(map[[2]int]int, len(c.stars))
	for _, s := range c.stars {
		x := int(s.x)
		if x >= 0 && x < c.width && s.y >= 0 && s.y < c.height {
			if c.phase == CracktroBoot && s.speed/1.7 > bootFade*0.8 {
				continue
			}
			pos := [2]int{x, s.y}
			if existing, ok := starMap[pos]; !ok || s.layer > existing {
				starMap[pos] = s.layer
			}
		}
	}

	type barHit struct {
		colorIdx  int
		intensity float64
	}
	barMap := make(map[int]barHit)
	for _, bar := range c.bars {
		centerY := float64(c.height)/2.0 + math.Sin(bar.phase)*float64(c.height)*0.38
		for dy := -2; dy <= 2; dy++ {
			y := int(centerY) + dy
			if y < 0 || y >= c.height {
				continue
			}
			dist := math.Abs(float64(dy)) / 2.5
			intensity := (1.0 - dist*0.4) * bootFade
			if intensity < 0.2 {
				intensity = 0.2
			}
			if existing, ok := barMap[y]; !ok || intensity > existing.intensity {
				barMap[y] = barHit{colorIdx: bar.colorIdx, intensity: intensity}
			}
		}
	}

	for y := 0; y < c.height; y++ {
		// Run-length SGR state: emit a color code only when the color changes
		// along the row, reset before default-colored cells and at row end
		lastR, lastG, lastB := -1, -1, -1

		for x := 0; x < c.width; x++ {
			pos := [2]int{x, y}

			var cr, cg, cb int
			var ch rune
			colored := true

			if bh, ok := barMap[y]; ok {
				col := c.cracktroSlot(bh.colorIdx)
				r, g, b := hexToRGBCracktro(col)
				cr = cracktroClamp(int(float64(r)*bh.intensity), 0, 255)
				cg = cracktroClamp(int(float64(g)*bh.intensity), 0, 255)
				cb = cracktroClamp(int(float64(b)*bh.intensity), 0, 255)
				if bh.intensity > 0.8 {
					ch = '\u2588'
				} else if bh.intensity > 0.6 {
					ch = '\u2593'
				} else if bh.intensity > 0.4 {
					ch = '\u2592'
				} else {
					ch = '\u2591'
				}
			} else if layer, ok := starMap[pos]; ok {
				col := c.cracktroSlot(layer)
				cr, cg, cb = hexToRGBCracktro(col)
				switch layer {
				case 0:
					ch = '\u00b7'
				case 1:
					ch = '\u2219'
				case 2:
					ch = '*'
				}
			} else {
				colored = false
			}

			if !colored {
				if lastR != -1 {
					c.builder.WriteString("\x1b[0m")
					lastR, lastG, lastB = -1, -1, -1
				}
				c.builder.WriteByte(' ')
				continue
			}

			if cr != lastR || cg != lastG || cb != lastB {
				writeRGBSGR(&c.builder, cr, cg, cb)
				lastR, lastG, lastB = cr, cg, cb
			}
			c.builder.WriteRune(ch)
		}

		if lastR != -1 {
			c.builder.WriteString("\x1b[0m")
		}
		if y < c.height-1 {
			c.builder.WriteByte('\n')
		}
	}

	return c.builder.String()
}

func (c *CracktroAnimation) cracktroSlot(i int) string {
	if len(c.palette) == 0 {
		return "#ffffff"
	}
	if i >= len(c.palette) {
		return c.palette[len(c.palette)-1]
	}
	return c.palette[i]
}

func cracktroClamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func hexToRGBCracktro(hex string) (int, int, int) {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 255, 255, 255
	}
	r := hexByteCracktro(hex[0])*16 + hexByteCracktro(hex[1])
	g := hexByteCracktro(hex[2])*16 + hexByteCracktro(hex[3])
	b := hexByteCracktro(hex[4])*16 + hexByteCracktro(hex[5])
	return r, g, b
}

func hexByteCracktro(c byte) int {
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
