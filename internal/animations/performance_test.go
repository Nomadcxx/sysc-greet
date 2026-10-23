package animations

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestAquariumRenderCoalescesAdjacentCellsWithTheSameColor(t *testing.T) {
	a := &AquariumEffect{
		width:  2,
		height: 3,
		fish: []Fish{{
			x:       0,
			y:       0,
			pattern: []string{"XX"},
			color:   "#112233",
		}},
	}

	firstRow := strings.Split(a.Render(), "\n")[0]
	if got := ansi.Strip(firstRow); got != "XX" {
		t.Fatalf("stripped first row = %q, want %q", got, "XX")
	}
	if got := ansi.StringWidth(firstRow); got != 2 {
		t.Fatalf("first row width = %d, want 2", got)
	}
	if got := strings.Count(firstRow, "\x1b[38;2;17;34;51m"); got != 1 {
		t.Fatalf("same-color run emitted %d foreground sequences, want 1", got)
	}
	if !strings.HasSuffix(firstRow, "\x1b[0m") {
		t.Fatal("first row did not reset its foreground before the next row")
	}
}

func TestSonarRenderKeepsAllocationCountUnderBudget(t *testing.T) {
	s := NewSonarEffect(213, 54, GetSonarPalette("dracula"), "dracula")
	for i := 0; i < 100; i++ {
		s.Update()
	}

	allocs := testing.AllocsPerRun(5, func() {
		_ = s.Render()
	})
	if allocs > 1000 {
		t.Fatalf("Render allocated %.0f objects, want at most 1000", allocs)
	}
}

func TestSonarRenderPreservesInterpolatedRingColorAndCellWidth(t *testing.T) {
	s := &SonarAnimation{
		width:      3,
		height:     1,
		palette:    []string{"#000000", "#808080"},
		centerX:    1,
		centerY:    0,
		maxRadius:  3,
		lifespan:   110,
		phase:      PhaseHold,
		frameCount: 55,
		rings:      []SonarRing{{birthFrame: 0}},
	}

	frame := s.Render()
	if got := ansi.Strip(frame); got != "· ·" {
		t.Fatalf("stripped frame = %q, want %q", got, "· ·")
	}
	if got := ansi.StringWidth(frame); got != 3 {
		t.Fatalf("frame width = %d, want 3", got)
	}
	if got := strings.Count(frame, "\x1b[38;2;64;64;64m"); got != 2 {
		t.Fatalf("interpolated ring color emitted %d times, want 2", got)
	}
}

func BenchmarkAquariumRender213x54(b *testing.B) {
	a := NewAquariumEffect(AquariumConfig{
		Width:         213,
		Height:        54,
		FishColors:    []string{"#ff6666", "#66ff66", "#6666ff"},
		WaterColors:   []string{"#4a9eff", "#c2b280"},
		SeaweedColors: []string{"#336633", "#558855"},
		BubbleColor:   "#ffffff",
		DiverColor:    "#ffcc00",
		BoatColor:     "#cc8844",
		MermaidColor:  "#ff66cc",
	})
	for i := 0; i < 300; i++ {
		a.Update()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Update()
		_ = a.Render()
	}
}

func BenchmarkSonarRender213x54(b *testing.B) {
	s := NewSonarEffect(213, 54, GetSonarPalette("dracula"), "dracula")
	for i := 0; i < 100; i++ {
		s.Update()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update()
		_ = s.Render()
	}
}
