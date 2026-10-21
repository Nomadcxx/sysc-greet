package main

import (
	"image"
	"path/filepath"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
)

type renderCountWriter struct {
	bytes int
}

func (w *renderCountWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

func idleRenderModel(tb testing.TB) model {
	tb.Helper()
	oldDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	tb.Cleanup(func() { dataDir = oldDataDir })

	applyTheme("dracula", true)
	m := initialModel(Config{TestMode: true}, false)
	session := sessions.Session{Name: "niri", Exec: "niri", Type: "Wayland"}
	m.selectedSession = &session
	m.width, m.height = 213, 54
	m.mode = ModeLogin
	m.selectedBackground = "none"
	return m
}

func newIdleRenderHarness(width, height int, profile colorprofile.Profile) (*uv.ScreenBuffer, *uv.TerminalRenderer, *renderCountWriter) {
	screen := uv.NewScreenBuffer(width, height)
	writer := &renderCountWriter{}
	renderer := uv.NewTerminalRenderer(writer, []string{"TERM=xterm-256color"})
	renderer.SetColorProfile(profile)
	renderer.EnterAltScreen()
	return &screen, renderer, writer
}

func drawIdleFrame(layer tea.Layer, screen *uv.ScreenBuffer, renderer *uv.TerminalRenderer) error {
	layer.Draw(*screen, screen.Bounds())
	renderer.Render(screen.Buffer)
	return renderer.Flush()
}

func TestIdleFlushEmitsNothing(t *testing.T) {
	profiles := []struct {
		name    string
		profile colorprofile.Profile
	}{
		{name: "TrueColor", profile: colorprofile.TrueColor},
		{name: "ANSI256", profile: colorprofile.ANSI256},
		{name: "ANSI", profile: colorprofile.ANSI},
		{name: "Ascii", profile: colorprofile.Ascii},
	}

	for _, tt := range profiles {
		t.Run(tt.name, func(t *testing.T) {
			m := idleRenderModel(t)
			screen, renderer, writer := newIdleRenderHarness(m.width, m.height, tt.profile)
			if err := drawIdleFrame(m.View().Layer, screen, renderer); err != nil {
				t.Fatal(err)
			}
			writer.bytes = 0

			for i := 0; i < 4; i++ {
				if err := drawIdleFrame(m.View().Layer, screen, renderer); err != nil {
					t.Fatal(err)
				}
			}
			if screen.Touched == nil || len(screen.Touched) != 0 {
				t.Fatalf("unchanged layer left %d touched lines; want an empty non-nil slice", len(screen.Touched))
			}
			if writer.bytes != 0 {
				t.Fatalf("unchanged view wrote %d bytes over four flushes", writer.bytes)
			}
		})
	}
}

func TestCachedViewLayerInvalidatesOnContentAndResize(t *testing.T) {
	m := idleRenderModel(t)
	screen, renderer, writer := newIdleRenderHarness(m.width, m.height, colorprofile.TrueColor)
	if err := drawIdleFrame(m.View().Layer, screen, renderer); err != nil {
		t.Fatal(err)
	}

	writer.bytes = 0
	m.usernameInput.SetValue("changed username")
	contentChanged := m.View()
	if err := drawIdleFrame(contentChanged.Layer, screen, renderer); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("changed content produced no terminal output")
	}

	writer.bytes = 0
	m.width, m.height = 215, 56
	screen.Resize(m.width, m.height)
	resized := m.View()
	if err := drawIdleFrame(resized.Layer, screen, renderer); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("resized view produced no terminal output")
	}

	fresh := uv.NewScreenBuffer(m.width, m.height)
	resized.Layer.Draw(fresh, fresh.Bounds())
	if got, want := screen.String(), fresh.String(); got != want {
		t.Fatal("resized screen contents differ from a fresh render")
	}
}

func TestCachedViewLayerDrawsIntoNewScreenAndForwardsCanvasMethods(t *testing.T) {
	m := idleRenderModel(t)
	layer := m.View().Layer
	bounded, ok := layer.(interface{ Bounds() image.Rectangle })
	if !ok {
		t.Fatal("view layer does not expose bounds")
	}
	if got, want := bounded.Bounds(), image.Rect(0, 0, m.width, m.height); got != want {
		t.Fatalf("view layer bounds = %v, want %v", got, want)
	}
	if _, ok := layer.(tea.Hittable); !ok {
		t.Fatal("view layer does not forward hit testing")
	}

	first := uv.NewScreenBuffer(m.width, m.height)
	layer.Draw(first, first.Bounds())
	second := uv.NewScreenBuffer(m.width, m.height)
	layer.Draw(second, second.Bounds())
	if got, want := second.String(), first.String(); got != want {
		t.Fatal("cached layer did not draw into a new screen")
	}
}

func TestCachedViewLayerRedrawsAfterClearScreen(t *testing.T) {
	m := idleRenderModel(t)
	screen, renderer, writer := newIdleRenderHarness(m.width, m.height, colorprofile.ANSI256)
	layer := m.View().Layer
	if err := drawIdleFrame(layer, screen, renderer); err != nil {
		t.Fatal(err)
	}

	writer.bytes = 0
	renderer.Redraw(screen.Buffer) // Bubble Tea uses this to handle tea.ClearScreen.
	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("forced screen redraw emitted no output")
	}

	writer.bytes = 0
	if err := drawIdleFrame(layer, screen, renderer); err != nil {
		t.Fatal(err)
	}
	if writer.bytes != 0 {
		t.Fatalf("unchanged view wrote %d bytes after the forced redraw", writer.bytes)
	}
}

func TestAnimatedBackgroundInvalidatesCachedView(t *testing.T) {
	m := idleRenderModel(t)
	m.selectedBackground = "ak-47"
	updated, _ := m.Update(bgTickMsg(time.Now()))
	m = updated.(model)

	screen, renderer, writer := newIdleRenderHarness(m.width, m.height, colorprofile.TrueColor)
	if err := drawIdleFrame(m.View().Layer, screen, renderer); err != nil {
		t.Fatal(err)
	}

	writer.bytes = 0
	updated, _ = m.Update(bgTickMsg(time.Now().Add(time.Second)))
	m = updated.(model)
	if err := drawIdleFrame(m.View().Layer, screen, renderer); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("animated background frame change produced no terminal output")
	}
}

func BenchmarkIdleFlushSameView(b *testing.B) {
	m := idleRenderModel(b)
	view := m.View()
	screen, renderer, writer := newIdleRenderHarness(m.width, m.height, colorprofile.TrueColor)
	if err := drawIdleFrame(view.Layer, screen, renderer); err != nil {
		b.Fatal(err)
	}
	writer.bytes = 0
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := drawIdleFrame(view.Layer, screen, renderer); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(writer.bytes)/float64(b.N), "termbytes/op")
}
