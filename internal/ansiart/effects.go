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

// Shimmer brightens a band travelling clockwise around the frame. t is
// active time in seconds; speed is cells per second; band is the half
// width in cells. Cells brighten by having their fg raised to the bright
// twin (index | 8) when the wrapped distance to the head is within band,
// scaled by the smooth falloff (1-delta/B)^2.
// Flash safety: each cell brightens at most once per lap; callers must
// keep speed <= P (at most one lap per second per cell) and band <= P/8.
func (g *Grid) Shimmer(t, speed, band float64) {
	if speed <= 0 || band <= 0 {
		return
	}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.At(x, y)
			if c.Bg == -1 && c.Ch == ' ' {
				continue
			}
			d, p := PerimeterDist(x, y, g.W, g.H)
			v := math.Min(speed, float64(p))
			b := math.Min(band, float64(p)/8)
			head := math.Mod(v*t, float64(p))
			dd := math.Abs(float64(d) - head)
			dd = math.Min(dd, float64(p)-dd)
			if dd >= b {
				continue
			}
			gain := 1 - dd/b
			if gain*gain <= 0.5 {
				continue
			}
			if c.Fg >= 0 && c.Fg < 8 {
				c.Fg += 8
			}
		}
	}
}

// Cycle rotates palette slots along a ramp: fg slot ramp[i] becomes
// ramp[(i+step) mod len(ramp)]. Pure colour; glyphs never change.
func (g *Grid) Cycle(ramp []int8, step int) {
	if len(ramp) == 0 {
		return
	}
	idx := map[int8]int{}
	for i, v := range ramp {
		idx[v] = i
	}
	step %= len(ramp)
	if step < 0 {
		step += len(ramp)
	}
	for i := range g.Cells {
		if k, ok := idx[g.Cells[i].Fg]; ok {
			g.Cells[i].Fg = ramp[(k+step)%len(ramp)]
		}
	}
}
