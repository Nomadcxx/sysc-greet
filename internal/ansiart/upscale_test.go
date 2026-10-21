package ansiart

import "testing"

func TestUpscaleBlockGlyphs(t *testing.T) {
	for _, tc := range []struct {
		glyph rune
		scale int
		rows  []string
	}{
		{'█', 2, []string{"██", "██"}},
		{'▀', 2, []string{"██", "  "}},
		{'▄', 2, []string{"  ", "██"}},
		{'▌', 2, []string{"█ ", "█ "}},
		{'▐', 2, []string{" █", " █"}},
		{'░', 2, []string{"░░", "░░"}},
		{'▀', 3, []string{"███", "▀▀▀", "   "}},
		{'▄', 3, []string{"   ", "▄▄▄", "███"}},
		{'▌', 3, []string{"█▌ ", "█▌ ", "█▌ "}},
		{'▐', 3, []string{" ▐█", " ▐█", " ▐█"}},
		{'░', 3, []string{"░░░", "░░░", "░░░"}},
		{'┌', 3, []string{"   ", " ┌─", " │ "}},
	} {
		g := NewGrid(1, 1)
		g.Cells[0] = Cell{Ch: tc.glyph, Fg: 14, Bg: -1}
		b := (&Border{Art: g, Left: 0, Top: 0, Right: 0, Bottom: 0, Scale: 1}).Upscale(tc.scale)
		if b.Scale != tc.scale || b.Art.W != tc.scale || b.Art.H != tc.scale {
			t.Fatalf("upscaled %q has scale %d and size %dx%d", tc.glyph, b.Scale, b.Art.W, b.Art.H)
		}
		for y, want := range tc.rows {
			for x, wantRune := range []rune(want) {
				if got := b.Art.At(x, y).Ch; got != wantRune {
					t.Errorf("upscale %q at %d,%d = %q, want %q", tc.glyph, x, y, got, wantRune)
				}
			}
		}
	}
}
