package main

import (
	"image"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

type countWriter struct {
	bytes int
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

type countingLayer struct {
	draws int
}

func (l *countingLayer) Draw(tea.Screen, tea.Rectangle) {
	l.draws++
}

func TestStableLayerSkipsUnchangedDraws(t *testing.T) {
	inner := &countingLayer{}
	layer := &stableLayer{inner: inner}
	firstScreen := uv.NewScreenBuffer(8, 2)
	firstArea := firstScreen.Bounds()

	layer.Draw(firstScreen, firstArea)
	layer.Draw(firstScreen, firstArea)
	if inner.draws != 1 {
		t.Fatalf("unchanged draw count = %d, want 1", inner.draws)
	}

	layer.Draw(firstScreen, image.Rect(0, 0, 7, 2))
	if inner.draws != 2 {
		t.Fatalf("draw count after area change = %d, want 2", inner.draws)
	}

	secondScreen := uv.NewScreenBuffer(8, 2)
	layer.Draw(secondScreen, secondScreen.Bounds())
	if inner.draws != 3 {
		t.Fatalf("draw count after screen change = %d, want 3", inner.draws)
	}
}

func TestStableLayerMarksUnchangedScreenUntouched(t *testing.T) {
	layer := &stableLayer{inner: &countingLayer{}}
	screen := uv.NewScreenBuffer(8, 2)
	area := screen.Bounds()
	layer.Draw(screen, area)
	for i := range screen.Touched {
		screen.Touched[i] = &uv.LineData{FirstCell: -1, LastCell: -1}
	}

	layer.Draw(screen, area)
	for i, line := range screen.Touched {
		if line != nil {
			t.Fatalf("unchanged screen line %d remains marked for renderer diff: %+v", i, line)
		}
	}
}

func TestStableLayerForwardsCanvasInterfaces(t *testing.T) {
	canvas := lipgloss.NewCanvas(lipgloss.NewLayer("hello").ID("greeting"))
	layer := &stableLayer{inner: canvas}

	if got, want := layer.Bounds(), canvas.Bounds(); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
	if got, want := layer.Height(), canvas.Bounds().Dy(); got != want {
		t.Fatalf("height = %d, want %d", got, want)
	}
	if got, want := layer.Hit(0, 0), "greeting"; got != want {
		t.Fatalf("hit = %q, want %q", got, want)
	}
}

func TestComposeLayerReusesOnlyIdenticalFrames(t *testing.T) {
	m := model{frameCache: &frameCache{}}
	base := m.composeLayer("background", 0, 0, "content", 1, 2, 12, 4)
	if got := m.composeLayer("background", 0, 0, "content", 1, 2, 12, 4); got != base {
		t.Fatal("identical frame did not reuse its layer")
	}

	cases := []struct {
		name string
		bg   string
		bgX  int
		bgY  int
		body string
		x    int
		y    int
		w    int
		h    int
	}{
		{name: "background", bg: "changed", bgX: 0, bgY: 0, body: "content", x: 1, y: 2, w: 12, h: 4},
		{name: "background position", bg: "background", bgX: 1, bgY: 0, body: "content", x: 1, y: 2, w: 12, h: 4},
		{name: "background vertical position", bg: "background", bgX: 0, bgY: 1, body: "content", x: 1, y: 2, w: 12, h: 4},
		{name: "content", bg: "background", bgX: 0, bgY: 0, body: "changed", x: 1, y: 2, w: 12, h: 4},
		{name: "content position", bg: "background", bgX: 0, bgY: 0, body: "content", x: 2, y: 2, w: 12, h: 4},
		{name: "content vertical position", bg: "background", bgX: 0, bgY: 0, body: "content", x: 1, y: 3, w: 12, h: 4},
		{name: "terminal width", bg: "background", bgX: 0, bgY: 0, body: "content", x: 1, y: 2, w: 13, h: 4},
		{name: "terminal height", bg: "background", bgX: 0, bgY: 0, body: "content", x: 1, y: 2, w: 12, h: 5},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := model{frameCache: &frameCache{}}
			base := m.composeLayer("background", 0, 0, "content", 1, 2, 12, 4)
			if got := m.composeLayer(tc.bg, tc.bgX, tc.bgY, tc.body, tc.x, tc.y, tc.w, tc.h); got == base {
				t.Fatalf("changed %s reused the prior layer", tc.name)
			}
		})
	}
}

func TestViewReusesUnchangedLayerAndReplacesChangedView(t *testing.T) {
	applyTheme("dracula", true)
	m := model{
		frameCache:   &frameCache{},
		width:        80,
		height:       24,
		mode:         ModePower,
		powerOptions: []string{"Reboot", "Shutdown", "Cancel"},
	}

	first := m.View().Layer
	if got := m.View().Layer; got != first {
		t.Fatal("View did not reuse the unchanged composed layer")
	}

	m.mode = ModeReleaseNotes
	changed := m.View().Layer
	if changed == first {
		t.Fatal("mode change reused the prior composed layer")
	}

	m.width++
	if resized := m.View().Layer; resized == changed {
		t.Fatal("terminal resize reused the prior composed layer")
	}
}

func TestIdleFlushEmitsNothingAcrossColorProfiles(t *testing.T) {
	profiles := []colorprofile.Profile{
		colorprofile.TrueColor,
		colorprofile.ANSI256,
		colorprofile.ANSI,
		colorprofile.Ascii,
	}
	for _, profile := range profiles {
		t.Run(profile.String(), func(t *testing.T) {
			const width, height = 24, 5
			writer := &countWriter{}
			renderer := uv.NewTerminalRenderer(writer, []string{"TERM=xterm-256color"})
			renderer.SetColorProfile(profile)
			screen := uv.NewScreenBuffer(width, height)
			m := model{frameCache: &frameCache{}}
			content := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("colored frame")
			layer := m.composeLayer(blankTerminalContent(width, height), 0, 0, content, 0, 0, width, height)

			for flush := 0; flush < 5; flush++ {
				layer.Draw(screen, screen.Bounds())
				renderer.Render(screen.Buffer)
				if err := renderer.Flush(); err != nil {
					t.Fatalf("flush %d: %v", flush, err)
				}
				if flush == 0 {
					writer.bytes = 0
					continue
				}
				if writer.bytes != 0 {
					t.Fatalf("flush %d wrote %d bytes for unchanged frame", flush, writer.bytes)
				}
			}
		})
	}
}

func TestChangedFrameAndResizeRedraw(t *testing.T) {
	const width, height = 24, 5
	writer := &countWriter{}
	renderer := uv.NewTerminalRenderer(writer, []string{"TERM=xterm-kitty", "COLORTERM=truecolor"})
	renderer.SetColorProfile(colorprofile.TrueColor)
	screen := uv.NewScreenBuffer(width, height)
	m := model{frameCache: &frameCache{}}

	layer := m.composeLayer(blankTerminalContent(width, height), 0, 0, "old", 0, 0, width, height)
	layer.Draw(screen, screen.Bounds())
	renderer.Render(screen.Buffer)
	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
	writer.bytes = 0

	layer = m.composeLayer(blankTerminalContent(width, height), 0, 0, "new", 0, 0, width, height)
	layer.Draw(screen, screen.Bounds())
	renderer.Render(screen.Buffer)
	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("changed content produced no terminal output")
	}
	writer.bytes = 0

	const resizedWidth = width + 1
	screen.Resize(resizedWidth, height)
	layer = m.composeLayer(blankTerminalContent(resizedWidth, height), 0, 0, "new", 0, 0, resizedWidth, height)
	layer.Draw(screen, screen.Bounds())
	renderer.Render(screen.Buffer)
	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("resized frame produced no terminal output")
	}
}

func TestCachedLayerRedrawsAfterErase(t *testing.T) {
	const width, height = 24, 5
	writer := &countWriter{}
	renderer := uv.NewTerminalRenderer(writer, []string{"TERM=xterm-kitty", "COLORTERM=truecolor"})
	renderer.SetColorProfile(colorprofile.TrueColor)
	screen := uv.NewScreenBuffer(width, height)
	m := model{frameCache: &frameCache{}}
	layer := m.composeLayer(blankTerminalContent(width, height), 0, 0, "content", 0, 0, width, height)
	layer.Draw(screen, screen.Bounds())
	renderer.Render(screen.Buffer)
	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
	writer.bytes = 0

	layer.Draw(screen, screen.Bounds())
	renderer.Erase()
	renderer.Render(screen.Buffer)
	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
	if writer.bytes == 0 {
		t.Fatal("cached layer did not repaint the screen after erase")
	}
}

func BenchmarkIdleFlushSameViewCachedLayer(b *testing.B) {
	const width, height = 213, 54
	writer := &countWriter{}
	renderer := uv.NewTerminalRenderer(writer, []string{"TERM=xterm-kitty", "COLORTERM=truecolor"})
	renderer.SetColorProfile(colorprofile.TrueColor)
	screen := uv.NewScreenBuffer(width, height)
	m := model{frameCache: &frameCache{}}
	content := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(strings.Repeat("idle frame", 10))
	layer := m.composeLayer(blankTerminalContent(width, height), 0, 0, content, 0, 0, width, height)
	layer.Draw(screen, screen.Bounds())
	renderer.Render(screen.Buffer)
	if err := renderer.Flush(); err != nil {
		b.Fatal(err)
	}
	writer.bytes = 0

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layer.Draw(screen, screen.Bounds())
		renderer.Render(screen.Buffer)
		if err := renderer.Flush(); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(writer.bytes)/float64(b.N), "termbytes/op")
}

func TestCachedMenuRedrawsAfterLoginFrame(t *testing.T) {
	const width, height = 132, 40
	m := initialModel(Config{TestMode: true}, false)
	m.width, m.height = width, height
	// The renderer draws every frame into one buffer for the program's life.
	screen := uv.NewScreenBuffer(width, height)
	draw := func(next model) string {
		next.View().Layer.Draw(screen, screen.Bounds())
		return screen.Buffer.String()
	}

	menu := m.navigateToMainMenu()
	if !strings.Contains(draw(menu), "Close Menu") {
		t.Fatal("first menu frame missing")
	}
	if strings.Contains(draw(m), "Close Menu") {
		t.Fatal("login frame still shows the menu")
	}
	if !strings.Contains(draw(menu), "Close Menu") {
		t.Fatal("reopened menu reused its cached layer and left the login frame on screen")
	}
}
