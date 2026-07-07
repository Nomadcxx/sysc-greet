package animations

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Layer compositing requires every effect frame to be a full width x height
// grid, and colors must never remain open past a row's end or they bleed into
// layer padding (ghosting, see e265c41). These tests pin that contract for
// the run-length SGR renderers.

func checkFrame(t *testing.T, name, frame string, width, height int) {
	t.Helper()
	lines := strings.Split(frame, "\n")
	if len(lines) != height {
		t.Fatalf("%s: got %d rows, want %d", name, len(lines), height)
	}
	for i, line := range lines {
		if w := ansi.StringWidth(line); w != width {
			t.Errorf("%s: row %d visible width %d, want %d", name, i, w, width)
		}
		if idx := strings.LastIndex(line, "\x1b[38;"); idx != -1 {
			if !strings.Contains(line[idx:], "\x1b[0m") {
				t.Errorf("%s: row %d leaves color state open past row end", name, i)
			}
		}
	}
}

func TestFireRenderInvariants(t *testing.T) {
	f := NewFireEffect(120, 40, GetDefaultFirePalette())
	for i := 0; i < 30; i++ {
		f.Update(i)
	}
	checkFrame(t, "fire", f.Render(), 120, 40)
}

func TestMatrixRenderInvariants(t *testing.T) {
	m := NewMatrixEffect(120, 40, GetMatrixPalette("default"))
	for i := 0; i < 30; i++ {
		m.Update(i)
	}
	checkFrame(t, "matrix", m.Render(), 120, 40)
}

func TestRainRenderInvariants(t *testing.T) {
	r := NewRainEffect(120, 40, GetRainPalette("default"))
	for i := 0; i < 30; i++ {
		r.Update(i)
	}
	checkFrame(t, "rain", r.Render(), 120, 40)
}

func TestPlasmaRenderInvariants(t *testing.T) {
	for _, width := range []int{120, 121} { // even and odd: odd pads the last column
		p := NewPlasmaEffect(width, 40, GetPlasmaPalette("default"), "default")
		for i := 0; i < 30; i++ {
			p.Update()
		}
		checkFrame(t, "plasma", p.Render(), width, 40)
	}
}
