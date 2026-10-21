package main

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	lipgloss "github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

func TestValidFormDim(t *testing.T) {
	for _, dim := range []float64{0, 0.55, 1} {
		if !validFormDim(dim) {
			t.Errorf("validFormDim(%v) = false, want true", dim)
		}
	}
	for _, dim := range []float64{-0.01, 1.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if validFormDim(dim) {
			t.Errorf("validFormDim(%v) = true, want false", dim)
		}
	}
}

func seethroughTestModel(form string, opts seeThroughOpts) *seeThroughLayer {
	return newSeeThroughLayer(form, opts)
}

func TestSeeThroughHoleRules(t *testing.T) {
	applyTheme("dracula", true)
	key := BgBase
	keyStyle := lipgloss.NewStyle().Background(key)
	content := keyStyle.Render("   ") + lipgloss.NewStyle().Foreground(color.White).Render("X") + keyStyle.Render("   ")
	l := seethroughTestModel(content, seeThroughOpts{Mode: formSeeThrough, Key: key, Halo: 0, Dim: 0.55, GlassAlpha: 0.5})

	if !l.isHole(l.buf.CellAt(0, 0)) {
		t.Error("key-bg blank is not a hole")
	}
	if l.isHole(l.buf.CellAt(3, 0)) {
		t.Error("ink cell treated as hole")
	}

	cursor := seethroughTestModel(lipgloss.NewStyle().Reverse(true).Render(" "), seeThroughOpts{Mode: formSeeThrough, Key: key})
	if cursor.isHole(cursor.buf.CellAt(0, 0)) {
		t.Error("reverse-video cursor cell treated as hole")
	}

	wide := seethroughTestModel("界Y", seeThroughOpts{Mode: formSeeThrough, Key: key})
	if wide.isHole(wide.buf.CellAt(1, 0)) {
		t.Error("wide-glyph continuation treated as hole")
	}
	if wide.isHole(wide.buf.CellAt(0, 0)) {
		t.Error("wide-glyph head treated as hole")
	}
}

func TestSeeThroughClearsWideGlyphPairs(t *testing.T) {
	applyTheme("dracula", true)
	scr := uv.NewScreenBuffer(4, 1)
	scr.SetCell(0, 0, &uv.Cell{Content: "界", Width: 2})

	l := newSeeThroughLayer("Z", seeThroughOpts{Mode: formSeeThrough, Key: BgBase})
	l.Draw(scr, scr.Bounds())

	if got := scr.CellAt(0, 0).Content; got != "Z" {
		t.Errorf("head cell = %q, want Z", got)
	}
	if got := scr.CellAt(1, 0); got.Width != 1 || got.Content != " " {
		t.Errorf("orphan continuation not cleared: %#v", *got)
	}
}

func TestFrostedKeepsWideGlyphs(t *testing.T) {
	applyTheme("dracula", true)
	scr := uv.NewScreenBuffer(3, 1)
	scr.SetCell(0, 0, &uv.Cell{Content: "界", Width: 2, Style: uv.Style{Fg: color.RGBA{255, 255, 255, 255}}})

	l := newSeeThroughLayer("   ", seeThroughOpts{Mode: formFrosted, Key: BgBase, Dim: 0.55, GlassAlpha: 0.5})
	l.Draw(scr, scr.Bounds())

	if got := scr.CellAt(0, 0); got.Content != "界" {
		t.Errorf("frosted wide glyph lost: %#v", *got)
	}
	if got := scr.CellAt(1, 0); got.Width != 0 {
		t.Errorf("wide placeholder destroyed: %#v", *got)
	}
}

func TestSeeThroughHalo(t *testing.T) {
	applyTheme("dracula", true)
	l := newSeeThroughLayer("  X  ", seeThroughOpts{Mode: formSeeThrough, Key: BgBase, Halo: 1})

	scr := uv.NewScreenBuffer(5, 1)
	for x := 0; x < 5; x++ {
		scr.SetCell(x, 0, &uv.Cell{Content: "~", Width: 1})
	}
	l.Draw(scr, scr.Bounds())

	for _, x := range []int{1, 3} {
		c := scr.CellAt(x, 0)
		if c.Content != " " || !sameColor(c.Style.Bg, BgBase) {
			t.Errorf("halo cell %d not knocked out: %+v", x, c)
		}
	}
	if got := scr.CellAt(4, 0).Content; got != "~" {
		t.Errorf("cell outside halo was changed: %q", got)
	}
	if got := scr.CellAt(2, 0).Content; got != "X" {
		t.Errorf("ink cell = %q, want X", got)
	}
}

func TestFrostedColorMath(t *testing.T) {
	applyTheme("dracula", true)
	key := BgBase
	l := newSeeThroughLayer("   ", seeThroughOpts{Mode: formFrosted, Key: key, Dim: 0.55, GlassAlpha: 0.5})

	scr := uv.NewScreenBuffer(1, 1)
	under := &uv.Cell{Content: "#", Width: 1, Style: uv.Style{
		Fg: color.RGBA{255, 255, 255, 255},
		Bg: color.RGBA{0, 0, 0, 255},
	}}
	scr.SetCell(0, 0, under)
	l.Draw(scr, scr.Bounds())

	got := scr.CellAt(0, 0)
	if got.Content != "#" {
		t.Fatalf("frosted glyph coarsened to %q", got.Content)
	}
	kr, kg, kb, _ := key.RGBA()
	keyCh := [3]float64{float64(kr >> 8), float64(kg >> 8), float64(kb >> 8)}
	compare := func(name string, c color.Color, base [3]float64, mix float64) {
		r, g, b, _ := c.RGBA()
		gotCh := [3]float64{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
		for i := 0; i < 3; i++ {
			want := base[i]*(1-mix) + keyCh[i]*mix + .5
			if diff := gotCh[i] - want; diff > 1 || diff < -1 {
				t.Errorf("%s channel %d = %v, want %v (±1)", name, i, gotCh[i], want)
			}
		}
	}
	compare("fg", got.Style.Fg, [3]float64{255, 255, 255}, 0.55)
	compare("bg", got.Style.Bg, [3]float64{0, 0, 0}, 0.5)
}

func TestLayerCache(t *testing.T) {
	applyTheme("dracula", true)
	m := layoutModel(80, 24)
	m.formLayerCache = &layerCache{}
	m.formMode = "seethrough"
	content := "hello"

	a := m.loginContentLayer(content)
	b := m.loginContentLayer(content)
	if pa, pb := a.(*seeThroughLayer), b.(*seeThroughLayer); pa != pb {
		t.Fatal("same content and opts did not reuse the cached layer")
	}
	c := m.loginContentLayer(content + "!")
	if pa, pc := a.(*seeThroughLayer), c.(*seeThroughLayer); pa == pc {
		t.Fatal("changed content reused the cached layer")
	}
}

func TestOpaqueKeepsStringLayer(t *testing.T) {
	m := layoutModel(80, 24)
	m.formMode = "opaque"
	m.selectedBackground = "ak-47"
	content := "plain"
	if got := m.loginContentLayer(content); got != any(content) {
		t.Fatalf("opaque path did not return the plain string layer: %T", got)
	}
	m.formMode = "auto"
	m.selectedBackground = "none"
	if got := m.loginContentLayer(content); got != any(content) {
		t.Fatalf("auto/none did not return the plain string layer: %T", got)
	}
}

func TestEffectiveFormModeDefaults(t *testing.T) {
	m := layoutModel(80, 24)
	m.formMode = "auto"
	cases := map[string]formMode{
		"none":       formOpaque,
		"pattern":    formOpaque,
		"ticker":     formOpaque,
		"ak-47":      formSeeThrough,
		"l2a1-sar":   formSeeThrough,
		"sonar":      formSeeThrough,
		"cracktro":   formSeeThrough,
		"plasma":     formSeeThrough,
		"matrix":     formFrosted,
		"fire":       formFrosted,
		"ascii-rain": formFrosted,
		"fireworks":  formFrosted,
		"aquarium":   formFrosted,
	}
	for bg, want := range cases {
		m.selectedBackground = bg
		if got := m.effectiveFormMode(); got != want {
			t.Errorf("background %q: mode = %v, want %v", bg, got, want)
		}
	}
}

func BenchmarkSeeThrough(b *testing.B) {
	applyTheme("dracula", true)
	for _, tc := range []struct {
		name string
		mode formMode
	}{{"seethrough", formSeeThrough}, {"frosted", formFrosted}} {
		for _, size := range [][2]int{{120, 36}, {200, 55}} {
			b.Run(fmt.Sprintf("%s/%dx%d", tc.name, size[0], size[1]), func(b *testing.B) {
				m := layoutModel(size[0], size[1])
				m.selectedBackground = "ak-47"
				content := m.renderMainView(size[0], size[1])
				bg := "bg placeholder\n"
				opts := seeThroughOpts{Mode: tc.mode, Key: BgBase, Halo: 1, Dim: 0.55, GlassAlpha: 0.5}
				layer := newSeeThroughLayer(content, opts)
				scr := uv.NewScreenBuffer(size[0], size[1])
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					uv.NewStyledString(bg).Draw(scr, scr.Bounds())
					layer.Draw(scr, scr.Bounds())
				}
			})
		}
	}
}
