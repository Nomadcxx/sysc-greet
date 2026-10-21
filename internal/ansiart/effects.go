package ansiart

import "math"

// PerimeterDist returns the clockwise distance of cell (x, y) along the
// frame's centre-line ring, starting at the top-left corner, and the ring
// length P. Each cell projects onto the nearest outer edge (ties resolve
// top, right, bottom, left).
func PerimeterDist(x, y, w, h int) (d, p int) {
	if w < 2 || h < 2 {
		return 0, max(1, 2*(w-1)+2*(h-1))
	}
	p = 2*(w-1) + 2*(h-1)
	dt := y
	dr := w - 1 - x
	db := h - 1 - y
	dl := x
	switch {
	case dt <= dr && dt <= db && dt <= dl:
		d = x
	case dr <= db && dr <= dl:
		d = (w - 1) + y
	case db <= dl:
		d = (w - 1) + (h - 1) + (w - 1 - x)
	default:
		d = 2*(w-1) + (h - 1) + (h - 1 - y)
	}
	return d, p
}

// ShimmerWave holds the phase and falloff for one frame across a perimeter.
type ShimmerWave struct {
	p      int
	head   float64
	band   float64
	active bool
}

// NewShimmerWave computes the frame's wave position once. frame is quantized
// to 20 fps; rest is the pause after each lap. Speed and band are capped to
// one lap/s and P/8 for flash safety.
func NewShimmerWave(frame, p int, speed, band, rest float64) ShimmerWave {
	if p < 1 || frame < 0 || speed <= 0 || band <= 0 || math.IsNaN(speed) || math.IsNaN(band) {
		return ShimmerWave{}
	}
	v := math.Min(speed, float64(p))
	b := math.Min(band, float64(p)/8)
	if b <= 0 {
		return ShimmerWave{}
	}
	lap := float64(p) / v
	if rest < 0 || math.IsNaN(rest) || math.IsInf(rest, 0) {
		rest = 0
	}
	elapsed := math.Mod(float64(frame)/20, lap+rest)
	if elapsed >= lap {
		return ShimmerWave{}
	}
	return ShimmerWave{
		p: p, head: math.Mod(v*elapsed, float64(p)), band: b, active: true,
	}
}

// Gain returns the smooth falloff for one perimeter cell.
func (w ShimmerWave) Gain(d int) float64 {
	if !w.active {
		return 0
	}
	dd := math.Abs(float64(d) - w.head)
	dd = math.Min(dd, float64(w.p)-dd)
	if dd >= w.band {
		return 0
	}
	gain := 1 - dd/w.band
	return gain * gain
}

// ShimmerGain returns the smooth falloff for one perimeter cell.
func ShimmerGain(d, p, frame int, speed, band, rest float64) float64 {
	return NewShimmerWave(frame, p, speed, band, rest).Gain(d)
}

// ShimmerFG returns the 16-colour bright twin for a cell lit by the shimmer.
func ShimmerFG(fg int8, d, p, frame int, speed, band, rest float64) int8 {
	if fg >= 0 && fg < 8 && ShimmerGain(d, p, frame, speed, band, rest) > 0.5 {
		return fg | 8
	}
	return fg
}

func (g *Grid) Shimmer(frame int, speed, band, rest float64) {
	_, p := PerimeterDist(0, 0, g.W, g.H)
	wave := NewShimmerWave(frame, p, speed, band, rest)
	for i := range g.Cells {
		c := &g.Cells[i]
		if c.Bg == -1 && c.Ch == ' ' {
			continue
		}
		x, y := i%g.W, i/g.W
		d, _ := PerimeterDist(x, y, g.W, g.H)
		if c.Fg >= 0 && c.Fg < 8 && wave.Gain(d) > 0.5 {
			c.Fg |= 8
		}
	}
}
