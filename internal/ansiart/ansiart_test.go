package ansiart

import (
	"strings"
	"testing"
)

func TestCP437RoundTrip(t *testing.T) {
	for i := 0; i < 256; i++ {
		if CP437[i] == 0 {
			t.Fatalf("CP437[%d] is zero", i)
		}
	}
	for i := 32; i < 256; i++ {
		b, ok := RuneToCP437(CP437[i])
		if !ok || b != byte(i) {
			t.Errorf("RuneToCP437(CP437[%d]) = %d, %v", i, b, ok)
		}
	}
	if b, ok := RuneToCP437(' '); !ok || b != 32 {
		t.Errorf("space maps to %d, %v; printable slots must win", b, ok)
	}
}

func TestDecodeANSIDeferredWrap(t *testing.T) {
	body := []byte(strings.Repeat("A", 80) + "\r\nX")
	g := DecodeANSI(body, 80, false)
	if g.H != 2 {
		t.Fatalf("80 columns + CRLF advanced %d rows, want 1", g.H-1)
	}
	if c := g.At(0, 1); c == nil || c.Ch != 'X' {
		t.Fatalf("X after wrap = %+v", c)
	}
}

func TestDecodeANSIWrapWithoutCRLF(t *testing.T) {
	body := []byte(strings.Repeat("A", 81))
	g := DecodeANSI(body, 80, false)
	if g.H != 2 {
		t.Fatalf("81 chars at width 80 gave %d rows, want 2", g.H)
	}
	if c := g.At(0, 1); c == nil || c.Ch != 'A' {
		t.Fatalf("wrapped glyph = %+v", c)
	}
}

func TestDecodeANSICUFAndSGR(t *testing.T) {
	body := []byte("\x1b[1;1H\x1b[5CX\x1b[31;42mY\x1b[0mZ")
	g := DecodeANSI(body, 40, false)
	if c := g.At(5, 0); c == nil || c.Ch != 'X' {
		t.Fatalf("CUF compression put X at %+v", c)
	}
	if c := g.At(6, 0); c == nil || c.Ch != 'Y' || c.Fg != 1 || c.Bg != 2 {
		t.Fatalf("SGR cell = %+v", c)
	}
	if c := g.At(7, 0); c == nil || c.Ch != 'Z' || c.Fg != 7 || c.Bg != 0 {
		t.Fatalf("reset cell = %+v", c)
	}
}

func TestDecodeANSIBoldAndICE(t *testing.T) {
	g := DecodeANSI([]byte("\x1b[1mA\x1b[31;1mB\x1b[5mC\x1b[0m"), 40, false)
	if c := g.At(0, 0); c.Fg != 15 {
		t.Errorf("bold on fg 7 = %d, want 15 (bright)", c.Fg)
	}
	if c := g.At(1, 0); c.Fg != 9 {
		t.Errorf("bold on fg 1 = %d, want 9 (bright red)", c.Fg)
	}
	if c := g.At(2, 0); c.Bg != 0 {
		t.Errorf("blink without iCE brightened bg: %d", c.Bg)
	}
	g = DecodeANSI([]byte("\x1b[5mB"), 40, true)
	if c := g.At(0, 0); c.Bg != 8 {
		t.Errorf("blink with iCE bg = %d, want 8", c.Bg)
	}
}

func TestSauceRoundTrip(t *testing.T) {
	in := Sauce{
		Title: "sysc nfo frame 01", Author: "tester", Group: "sysc", Date: "20261020",
		DataType: 1, FileType: 1, TInfo1: 24, TInfo2: 9,
		Flags: 0, TInfoS: "IBM VGA",
		Comments: []string{"nineslice l=8 t=3 r=8 b=3 anchor=center"},
	}
	raw := EncodeSAUCE(in)
	body, out := ParseSAUCE(raw)
	if len(body) != 0 {
		t.Fatalf("body not empty: %q", body)
	}
	if out == nil {
		t.Fatal("no SAUCE parsed")
	}
	if out.Title != in.Title || out.Author != in.Author || out.Group != in.Group || out.Date != in.Date {
		t.Errorf("text fields: %+v", out)
	}
	if out.TInfo1 != 24 || out.TInfo2 != 9 || out.TInfoS != "IBM VGA" {
		t.Errorf("info fields: %+v", out)
	}
	if len(out.Comments) != 1 || out.Comments[0] != in.Comments[0] {
		t.Errorf("comments: %q", out.Comments)
	}
	if out.ICE() {
		t.Error("ICE set on non-iCE record")
	}
}

func TestSliceIndexSymmetry(t *testing.T) {
	src := 6
	lo, hi := 2, 2
	for out := 21; out <= 81; out++ {
		for o := 0; o < out; o++ {
			s := sliceIndex(o, out, src, lo, hi, true)
			if s < 0 || s >= src {
				t.Fatalf("out=%d o=%d: source %d out of range", out, o, s)
			}
		}
	}
	if got := sliceIndex(0, 5, 6, 2, 2, true); got != 0 {
		t.Errorf("leading corner = %d, want 0", got)
	}
	if got := sliceIndex(4, 5, 6, 2, 2, true); got != 5 {
		t.Errorf("trailing corner = %d, want 5", got)
	}
	mid := sliceIndex(2, 5, 6, 2, 2, true)
	if mid != 2 && mid != 3 {
		t.Errorf("centre tile = %d, want 2 or 3", mid)
	}
}

func symmetricBorder(t *testing.T) *Border {
	t.Helper()
	g := NewGrid(6, 3)
	pattern := [6]rune{'A', 'B', 'C', 'C', 'B', 'A'}
	for y := 0; y < 3; y++ {
		for x := 0; x < 6; x++ {
			*g.At(x, y) = Cell{Ch: pattern[x], Fg: 7, Bg: 0}
		}
	}
	return &Border{Art: g, Left: 2, Top: 1, Right: 2, Bottom: 1, CenterAnchor: true, BlackIsClear: true, Scale: 1}
}

func TestBorderRenderSymmetry(t *testing.T) {
	b := symmetricBorder(t)
	for w := 21; w <= 81; w++ {
		g := b.Render(w, 9)
		for y := 0; y < g.H; y++ {
			for x := 0; x < g.W/2; x++ {
				if g.At(x, y).Ch != g.At(g.W-1-x, y).Ch {
					t.Fatalf("w=%d y=%d: not mirror-symmetric", w, y)
				}
			}
		}
	}
}

func TestBorderRenderClampAndInterior(t *testing.T) {
	b := symmetricBorder(t)
	minW, minH := b.MinSize()
	if minW != 5 || minH != 3 {
		t.Fatalf("MinSize = %dx%d, want 5x3", minW, minH)
	}
	g := b.Render(1, 1)
	if g.W < minW || g.H < minH {
		t.Fatalf("clamped render = %dx%d, want >= %dx%d", g.W, g.H, minW, minH)
	}
	g = b.Render(21, 9)
	for y := 1; y < 8; y++ {
		for x := 2; x < 19; x++ {
			if c := g.At(x, y); c.Bg != -1 {
				t.Fatalf("interior cell %d,%d not transparent: %+v", x, y, *c)
			}
		}
	}
}

func TestLoadBorderValidation(t *testing.T) {
	good := symmetricBorder(t)
	ans := good.Art.ToCP437ANS(Sauce{Title: "t", Comments: []string{"nineslice l=2 t=1 r=2 b=1 anchor=center name=test"}})

	b, err := LoadBorder(ans)
	if err != nil {
		t.Fatalf("valid asset rejected: %v", err)
	}
	if b.Left != 2 || b.Top != 1 || !b.CenterAnchor || b.Name != "test" || b.Scale != 1 {
		t.Fatalf("metadata: %+v", b)
	}

	if _, err := LoadBorder([]byte("not an ansi file")); err == nil {
		t.Error("missing SAUCE accepted")
	}

	noMeta := good.Art.ToCP437ANS(Sauce{Title: "t"})
	if _, err := LoadBorder(noMeta); err == nil {
		t.Error("missing nineslice accepted")
	}

	badInsets := good.Art.ToCP437ANS(Sauce{Title: "t", Comments: []string{"nineslice l=100 t=1 r=100 b=1 anchor=center"}})
	if _, err := LoadBorder(badInsets); err == nil {
		t.Error("oversized insets accepted")
	}
	for _, meta := range []string{
		"nineslice l=-1 t=1 r=2 b=1 anchor=center",
		"nineslice l=x t=1 r=2 b=1 anchor=center",
		"nineslice l=2 t=1 r=2 b=1 anchor=sideways",
		"nineslice l=2 t=1 r=2 b=1 anchor=center scale=4",
		"nineslice l=2 t=1 r=2 b=1 anchor=center title=1",
	} {
		raw := good.Art.ToCP437ANS(Sauce{Title: "t", Comments: []string{meta}})
		if _, err := LoadBorder(raw); err == nil {
			t.Errorf("invalid metadata accepted: %q", meta)
		}
	}

	badCursor := []byte("\x1b[999;1HX")
	meta := Sauce{Title: "t", DataType: 1, FileType: 1, TInfo1: 6, TInfo2: 3, Comments: []string{"nineslice l=2 t=1 r=2 b=1 anchor=center"}}
	badCursor = append(badCursor, EncodeSAUCE(meta)...)
	if _, err := LoadBorder(badCursor); err == nil {
		t.Error("out-of-bounds cursor accepted")
	}

	if _, err := LoadBorder(make([]byte, maxANSIFile+1)); err == nil {
		t.Error("oversized file accepted")
	}
}

func TestBorderTitleCapUsesMetadataAndCellWidths(t *testing.T) {
	b := symmetricBorder(t)
	b.Art = NewGrid(24, 9)
	for i := range b.Art.Cells {
		b.Art.Cells[i] = Cell{Ch: '-', Fg: 7, Bg: -1}
	}
	data := b.Art.ToCP437ANS(Sauce{
		DataType: 1, FileType: 1, Comments: []string{
			"nineslice l=3 t=3 r=3 b=3 anchor=center title=0",
		},
	})

	border, err := LoadBorder(data)
	if err != nil {
		t.Fatalf("load titled border: %v", err)
	}
	if border.TitleRow != 0 {
		t.Fatalf("TitleRow = %d, want 0", border.TitleRow)
	}
	grid := border.RenderWithTitle(24, 9, "NIRI")
	if got := grid.At(5, 0).Ch; got != '▐' {
		t.Fatalf("title cap starts with %q, want ▐", got)
	}
	if got := grid.At(10, 0); got.Ch != 'N' || got.Fg != 15 {
		t.Fatalf("title text cell = %+v, want N in fg 15", *got)
	}
	if got := grid.At(18, 0).Ch; got != '▌' {
		t.Fatalf("title cap ends with %q, want ▌", got)
	}
	if got := border.RenderWithTitle(24, 9, "界").At(5, 0).Ch; got != '-' {
		t.Fatalf("wide title rune corrupted the frame: got %q, want original edge", got)
	}

	truncated := border.RenderWithTitle(24, 9, "12345678901234567890")
	var title strings.Builder
	for x := 8; x < 16; x++ {
		title.WriteRune(truncated.At(x, 0).Ch)
	}
	if got, want := title.String(), "1234567…"; got != want {
		t.Fatalf("truncated title = %q, want %q", got, want)
	}
}

func TestLoadBorderRejectsAmbiguousAndUnsafeMetadata(t *testing.T) {
	good := symmetricBorder(t)
	for _, meta := range []string{
		"nineslice l=2 l=3 t=1 r=2 b=1 anchor=center",
		"nineslice l=2 t=1 r=2 b=1 anchor=center label=\x1b[31m",
	} {
		data := good.Art.ToCP437ANS(Sauce{Comments: []string{meta}})
		if _, err := LoadBorder(data); err == nil {
			t.Errorf("accepted unsafe metadata: %q", meta)
		}
	}
}

func TestLoadBorderRejectsDuplicateMetadata(t *testing.T) {
	good := symmetricBorder(t)
	meta := "nineslice l=2 t=1 r=2 b=1 anchor=center"
	data := good.Art.ToCP437ANS(Sauce{Comments: []string{meta, meta}})
	if _, err := LoadBorder(data); err == nil || !strings.Contains(err.Error(), "duplicate nineslice metadata") {
		t.Fatalf("duplicate nineslice metadata error = %v", err)
	}
}

func TestToCP437ANSRoundTrip(t *testing.T) {
	g := NewGrid(24, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 24; x++ {
			*g.At(x, y) = Cell{Ch: 'A' + rune((x+y)%26), Fg: int8(x % 16), Bg: int8(y % 8)}
		}
	}
	raw := g.ToCP437ANS(Sauce{Title: "round trip"})
	body, s := ParseSAUCE(raw)
	if s == nil || s.TInfo1 != 24 || s.TInfo2 != 3 {
		t.Fatalf("SAUCE: %+v", s)
	}
	back := DecodeANSI(body, int(s.TInfo1), s.ICE())
	for y := 0; y < 3; y++ {
		for x := 0; x < 24; x++ {
			if *back.At(x, y) != *g.At(x, y) {
				t.Fatalf("cell %d,%d = %+v, want %+v", x, y, *back.At(x, y), *g.At(x, y))
			}
		}
	}
}
