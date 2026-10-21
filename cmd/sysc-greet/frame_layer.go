package main

import (
	"image"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

type stableLayer struct {
	canvas           *lipgloss.Canvas
	lastScreenBuffer uv.ScreenBuffer
	hasLastBuffer    bool
	lastArea         tea.Rectangle
}

func (l *stableLayer) Draw(screen tea.Screen, area tea.Rectangle) {
	buffer, isBuffer := screenBuffer(screen)
	if l.hasLastBuffer && isBuffer &&
		l.lastScreenBuffer.Buffer == buffer.Buffer && l.lastScreenBuffer.Method == buffer.Method &&
		l.lastArea == area {
		// Ultraviolet counts its clean-line sentinels as touched. An empty
		// non-nil slice makes the renderer take its no-change fast path.
		if buffer.Touched == nil {
			buffer.Touched = make([]*uv.LineData, 0)
		} else {
			buffer.Touched = buffer.Touched[:0]
		}
		return
	}
	l.canvas.Draw(screen, area)
	l.lastScreenBuffer, l.hasLastBuffer = buffer, isBuffer
	l.lastArea = area
}

func screenBuffer(screen tea.Screen) (uv.ScreenBuffer, bool) {
	switch screen := screen.(type) {
	case uv.ScreenBuffer:
		return screen, true
	case *uv.ScreenBuffer:
		if screen != nil {
			return *screen, true
		}
	}
	return uv.ScreenBuffer{}, false
}

func (l *stableLayer) Bounds() image.Rectangle {
	return l.canvas.Bounds()
}

func (l *stableLayer) Hit(x, y int) string {
	return l.canvas.Hit(x, y)
}

type frameCache struct {
	background  string
	backgroundX int
	backgroundY int
	content     string
	contentX    int
	contentY    int
	width       int
	height      int
	layer       *stableLayer
}

func (m model) composeLayer(background string, backgroundX, backgroundY int, content string, contentX, contentY, width, height int) tea.Layer {
	if cache := m.frameCache; cache != nil && cache.layer != nil &&
		cache.background == background && cache.backgroundX == backgroundX && cache.backgroundY == backgroundY &&
		cache.content == content && cache.contentX == contentX && cache.contentY == contentY &&
		cache.width == width && cache.height == height {
		return cache.layer
	}

	layer := &stableLayer{canvas: lipgloss.NewCanvas(
		lipgloss.NewLayer(background).X(backgroundX).Y(backgroundY),
		lipgloss.NewLayer(content).X(contentX).Y(contentY),
	)}
	if cache := m.frameCache; cache != nil {
		*cache = frameCache{
			background:  background,
			backgroundX: backgroundX,
			backgroundY: backgroundY,
			content:     content,
			contentX:    contentX,
			contentY:    contentY,
			width:       width,
			height:      height,
			layer:       layer,
		}
	}
	return layer
}
