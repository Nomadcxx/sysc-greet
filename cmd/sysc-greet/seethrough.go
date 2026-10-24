package main

import (
	"encoding/binary"
	"hash/fnv"
	"image"
	"image/color"
	"math"

	lipgloss "github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

type formMode uint8

const (
	formOpaque formMode = iota
	formSeeThrough
	formFrosted
)

type seeThroughOpts struct {
	Mode       formMode
	Key        color.Color
	Halo       int
	Dim        float64
	GlassAlpha float64
}

type seeThroughLayer struct {
	buf       uv.ScreenBuffer
	w, h      int
	ink, halo []bool
	opts      seeThroughOpts
}

func newSeeThroughLayer(content string, o seeThroughOpts) *seeThroughLayer {
	w, h := lipgloss.Width(content), lipgloss.Height(content)
	buf := uv.NewScreenBuffer(w, h)
	uv.NewStyledString(content).Draw(buf, buf.Bounds())
	l := &seeThroughLayer{buf: buf, w: w, h: h, opts: o}
	l.buildMasks()
	return l
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	return ar == br && ag == bg && ab == bb
}

func mixColor(a, b color.Color, t float64) color.Color {
	if a == nil {
		return b
	}
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	f := func(x, y uint32) uint8 {
		return uint8(float64(x>>8)*(1-t) + float64(y>>8)*t + .5)
	}
	return color.RGBA{f(ar, br), f(ag, bg), f(ab, bb), 255}
}

func (l *seeThroughLayer) isHole(c *uv.Cell) bool {
	if c == nil {
		return true
	}
	if c.Width == 0 {
		return false
	}
	blank := c.Content == " " || c.Content == ""
	return blank && c.Style.Attrs == 0 && (c.Style.Bg == nil || sameColor(c.Style.Bg, l.opts.Key))
}

func (l *seeThroughLayer) buildMasks() {
	l.ink = make([]bool, l.w*l.h)
	l.halo = make([]bool, l.w*l.h)
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.ink[y*l.w+x] = !l.isHole(l.buf.CellAt(x, y))
		}
	}
	if l.opts.Halo <= 0 {
		return
	}
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			if l.ink[y*l.w+x] {
				continue
			}
			for dx := -l.opts.Halo; dx <= l.opts.Halo; dx++ {
				xx := x + dx
				if xx >= 0 && xx < l.w && l.ink[y*l.w+xx] {
					l.halo[y*l.w+x] = true
					break
				}
			}
		}
	}
}

func (l *seeThroughLayer) Bounds() image.Rectangle {
	return image.Rect(0, 0, l.w, l.h)
}

func (l *seeThroughLayer) Draw(scr uv.Screen, area image.Rectangle) {
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			px, py := area.Min.X+x, area.Min.Y+y
			if !(image.Point{X: px, Y: py}).In(area) {
				continue
			}
			src := l.buf.CellAt(x, y)
			if src.Width == 0 {
				// Wide-glyph continuation: the head's write already covers
				// the pair; writing it again would destroy the head.
				continue
			}
			if l.ink[y*l.w+x] {
				c := src.Clone()
				if l.opts.Mode == formSeeThrough && sameColor(c.Style.Bg, l.opts.Key) {
					c.Style.Bg = nil
				}
				scr.SetCell(px, py, c)
				continue
			}
			under := scr.CellAt(px, py)
			if under != nil && under.Width == 0 {
				// Continuation of an underlying wide glyph; the head's
				// iteration owns the pair.
				continue
			}
			if l.halo[y*l.w+x] {
				scr.SetCell(px, py, &uv.Cell{Content: " ", Width: 1, Style: uv.Style{Bg: l.opts.Key}})
				continue
			}
			if l.opts.Mode == formFrosted && under != nil {
				c := under.Clone()
				c.Style.Fg = mixColor(c.Style.Fg, l.opts.Key, l.opts.Dim)
				if c.Style.Bg != nil {
					c.Style.Bg = mixColor(c.Style.Bg, l.opts.Key, l.opts.GlassAlpha)
				}
				scr.SetCell(px, py, c)
			}
		}
	}
}

type layerCache struct {
	key   uint64
	layer *seeThroughLayer
}

func (c *layerCache) get(content string, o seeThroughOpts) *seeThroughLayer {
	h := fnv.New64a()
	h.Write([]byte(content))
	var meta [18]byte
	meta[0] = byte(o.Mode)
	meta[1] = byte(o.Halo)
	binary.LittleEndian.PutUint64(meta[2:], math.Float64bits(o.Dim))
	binary.LittleEndian.PutUint64(meta[10:], math.Float64bits(o.GlassAlpha))
	h.Write(meta[:])
	if o.Key != nil {
		r, g, b, _ := o.Key.RGBA()
		h.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)})
	}
	key := h.Sum64()
	if c.layer != nil && c.key == key {
		return c.layer
	}
	l := newSeeThroughLayer(content, o)
	c.key, c.layer = key, l
	return l
}

func (m model) effectiveFormMode() formMode {
	switch m.formMode {
	case "opaque":
		return formOpaque
	case "seethrough":
		return formSeeThrough
	case "frosted":
		return formFrosted
	}
	switch m.selectedBackground {
	case "none", "pattern", "":
		return formOpaque
	case "ak-47", "l2a1-sar", "sonar", "cracktro", "plasma":
		return formSeeThrough
	case "matrix", "fire", "ascii-rain", "fireworks", "aquarium":
		return formFrosted
	default:
		if isMeshBackground(m.selectedBackground) {
			return formSeeThrough
		}
		return formOpaque
	}
}

func validFormMode(mode string) bool {
	switch mode {
	case "", "auto", "opaque", "seethrough", "frosted":
		return true
	default:
		return false
	}
}

func validFormDim(dim float64) bool {
	return dim >= 0 && dim <= 1
}

func (m model) loginContentLayer(content string) any {
	mode := m.effectiveFormMode()
	if mode == formOpaque {
		return content
	}
	o := seeThroughOpts{
		Mode:       mode,
		Key:        BgBase,
		Halo:       m.config.FormHalo,
		Dim:        m.config.FormDim,
		GlassAlpha: 0.5,
	}
	if o.Halo < 0 {
		o.Halo = 0
	}
	if o.Halo > 2 {
		o.Halo = 2
	}
	if m.formLayerCache == nil {
		return newSeeThroughLayer(content, o)
	}
	return m.formLayerCache.get(content, o)
}
