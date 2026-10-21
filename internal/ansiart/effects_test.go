package ansiart

import (
	"testing"
)

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
	lap := float64(p)/speed + 4

	brightens := make([]int, w*h)
	prev := make([]bool, w*h)
	var maxChangedFrac float64
	const fps = 20
	for f := 0; f < 10*fps; f++ {
		g := NewGrid(w, h)
		copy(g.Cells, src.Cells)
		g.Shimmer(f, speed, band, 4)
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

func TestShimmerOffIdentity(t *testing.T) {
	g := NewGrid(10, 5)
	for i := range g.Cells {
		g.Cells[i] = Cell{Ch: 'x', Fg: 3, Bg: 0}
	}
	before := append([]Cell(nil), g.Cells...)
	g.Shimmer(100, 0, 6, 4)
	for i := range g.Cells {
		if g.Cells[i] != before[i] {
			t.Fatalf("cell %d changed with effect off: %+v vs %+v", i, g.Cells[i], before[i])
		}
	}
}

func TestShimmerLapRest(t *testing.T) {
	const w, h = 10, 6
	d, p := PerimeterDist(0, 0, w, h)
	base := int8(4)
	lapFrames := int(float64(p)/24*20) + 1
	if got := ShimmerFG(base, d, p, lapFrames+20, 24, 6, 4); got != base {
		t.Fatalf("cell brightened during rest: fg %d, want %d", got, base)
	}
	if got := ShimmerFG(base, d, p, 0, 24, 6, 4); got != base|8 {
		t.Fatalf("head cell fg = %d, want %d", got, base|8)
	}
}
