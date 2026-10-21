package main

import (
	"image"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

type stableLayer struct {
	inner      tea.Layer
	lastScreen tea.Screen
	lastArea   tea.Rectangle
	hasDrawn   bool
}

func (l *stableLayer) Draw(screen tea.Screen, area tea.Rectangle) {
	if l.hasDrawn && l.lastScreen == screen && l.lastArea == area {
		markScreenUntouched(screen)
		return
	}

	l.inner.Draw(screen, area)
	l.lastScreen = screen
	l.lastArea = area
	l.hasDrawn = true
}

// Ultraviolet treats nil touched lines as unchanged, so its renderer can skip
// diffing the cached cells while preserving them for erase-triggered redraws.
func markScreenUntouched(screen tea.Screen) {
	var buffer *uv.Buffer
	switch screen := screen.(type) {
	case uv.ScreenBuffer:
		buffer = screen.Buffer
	}
	if buffer == nil {
		return
	}
	if len(buffer.Touched) != buffer.Height() {
		buffer.Touched = make([]*uv.LineData, buffer.Height())
		return
	}
	clear(buffer.Touched)
}

func (l *stableLayer) Hit(x, y int) string {
	if hit, ok := l.inner.(interface{ Hit(int, int) string }); ok {
		return hit.Hit(x, y)
	}
	return ""
}

func (l *stableLayer) Bounds() image.Rectangle {
	if bounded, ok := l.inner.(interface{ Bounds() image.Rectangle }); ok {
		return bounded.Bounds()
	}
	return image.Rectangle{}
}

func (l *stableLayer) Height() int {
	if height, ok := l.inner.(interface{ Height() int }); ok {
		return height.Height()
	}
	return l.Bounds().Dy()
}

type frameCache struct {
	background  string
	backgroundX int
	backgroundY int
	content     string
	contentX    int
	contentY    int
	terminalW   int
	terminalH   int
	layer       *stableLayer
}

func (m model) composeLayer(background string, backgroundX, backgroundY int, content string, contentX, contentY, terminalW, terminalH int) tea.Layer {
	cache := m.frameCache
	if cache != nil && cache.layer != nil &&
		cache.background == background &&
		cache.backgroundX == backgroundX &&
		cache.backgroundY == backgroundY &&
		cache.content == content &&
		cache.contentX == contentX &&
		cache.contentY == contentY &&
		cache.terminalW == terminalW &&
		cache.terminalH == terminalH {
		return cache.layer
	}

	layer := &stableLayer{inner: lipgloss.NewCanvas(
		lipgloss.NewLayer(background).X(backgroundX).Y(backgroundY),
		lipgloss.NewLayer(content).X(contentX).Y(contentY),
	)}
	if cache != nil {
		cache.background = background
		cache.backgroundX = backgroundX
		cache.backgroundY = backgroundY
		cache.content = content
		cache.contentX = contentX
		cache.contentY = contentY
		cache.terminalW = terminalW
		cache.terminalH = terminalH
		cache.layer = layer
	}
	return layer
}
