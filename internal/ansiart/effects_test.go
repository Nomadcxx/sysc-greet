package ansiart

import "testing"

func TestPerimeterDist(t *testing.T) {
	d, p := PerimeterDist(0, 0, 10, 6)
	if d != 0 {
		t.Errorf("top-left d = %d, want 0", d)
	}
	if p != 2*9+2*5 {
		t.Errorf("P = %d, want %d", p, 2*9+2*5)
	}
	if d, _ := PerimeterDist(9, 0, 10, 6); d != 9 {
		t.Errorf("top-right d = %d, want 9", d)
	}
	if d, _ := PerimeterDist(9, 5, 10, 6); d != 9+5 {
		t.Errorf("bottom-right d = %d, want 14", d)
	}
	if d, _ := PerimeterDist(0, 5, 10, 6); d != 2*9+5 {
		t.Errorf("bottom-left d = %d, want 23", d)
	}
}

func TestShimmerFlashSafety(t *testing.T) {
	const w, h = 200, 55
	src := NewGrid(w, h)
	for i := range src.Cells {
		src.Cells[i] = Cell{Ch: '█', Fg: 4, Bg: 0}
	}
	_, p := PerimeterDist(0, 0, w, h)
	speed, band := 24.0, 6.0
	lap := float64(p) / speed

	brightens := make([]int, w*h)
	prev := make([]bool, w*h)
	var maxChangedFrac float64
	const fps = 20
	for f := 0; f < 10*fps; f++ {
		g := NewGrid(w, h)
		copy(g.Cells, src.Cells)
		g.Shimmer(float64(f)/fps, speed, band)
		changed := 0
		for i := range g.Cells {
			lit := g.Cells[i].Fg >= 8
			if lit && !prev[i] {
				brightens[i]++
			}
			if lit != prev[i] {
				changed++
			}
			prev[i] = lit
		}
		if frac := float64(changed) / float64(w*h); frac > maxChangedFrac {
			maxChangedFrac = frac
		}
	}
	for i, n := range brightens {
		laps := int(10/lap) + 1
		if n > laps {
			t.Fatalf("cell %d brightened %d times in 10s (lap %.2fs)", i, n, lap)
		}
	}
	if maxChangedFrac >= 0.25 {
		t.Fatalf("changed area %.2f of screen reaches 25%%", maxChangedFrac)
	}
}

func TestShimmerOffAndCycleIdentity(t *testing.T) {
	g := NewGrid(10, 5)
	for i := range g.Cells {
		g.Cells[i] = Cell{Ch: 'x', Fg: 3, Bg: 0}
	}
	before := append([]Cell(nil), g.Cells...)
	g.Shimmer(5, 0, 6)
	g.Cycle(nil, 3)
	g.Cycle([]int8{3, 11}, 0)
	for i := range g.Cells {
		if g.Cells[i] != before[i] {
			t.Fatalf("cell %d changed with effect off: %+v vs %+v", i, g.Cells[i], before[i])
		}
	}
}

func TestCyclePureColor(t *testing.T) {
	g := NewGrid(4, 1)
	for x := 0; x < 4; x++ {
		*g.At(x, 0) = Cell{Ch: 'x', Fg: 4, Bg: 0}
	}
	g.Cycle([]int8{4, 6, 14, 6}, 1)
	for x := 0; x < 4; x++ {
		if c := g.At(x, 0); c.Ch != 'x' {
			t.Errorf("glyph changed at %d: %+v", x, *c)
		}
	}
	if c := g.At(0, 0); c.Fg != 6 {
		t.Errorf("ramp step fg = %d, want 6", c.Fg)
	}
}
