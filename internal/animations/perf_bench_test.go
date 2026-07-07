package animations

import (
	"testing"
)

// Perf audit benchmarks — fullscreen kitty on 1080p ≈ 240x67 cells.
const (
	benchW = 240
	benchH = 67
)

func BenchmarkFireUpdate(b *testing.B) {
	f := NewFireEffect(benchW, benchH*2/5, GetDefaultFirePalette())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Update(i)
	}
}

func BenchmarkFireRender(b *testing.B) {
	f := NewFireEffect(benchW, benchH*2/5, GetDefaultFirePalette())
	for i := 0; i < 60; i++ {
		f.Update(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.Render()
	}
}

func BenchmarkMatrixUpdate(b *testing.B) {
	m := NewMatrixEffect(benchW, benchH, GetMatrixPalette("default"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Update(i)
	}
}

func BenchmarkMatrixRender(b *testing.B) {
	m := NewMatrixEffect(benchW, benchH, GetMatrixPalette("default"))
	for i := 0; i < 60; i++ {
		m.Update(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Render()
	}
}

func BenchmarkRainRender(b *testing.B) {
	r := NewRainEffect(benchW, benchH, GetRainPalette("default"))
	for i := 0; i < 60; i++ {
		r.Update(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Render()
	}
}

func BenchmarkPlasmaUpdateRender(b *testing.B) {
	p := NewPlasmaEffect(benchW, benchH, GetPlasmaPalette("default"), "default")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Update()
		_ = p.Render()
	}
}

func BenchmarkGetMatrixPalette(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetMatrixPalette("dracula")
	}
}
