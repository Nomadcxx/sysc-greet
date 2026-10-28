package main

import (
	"image"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

type clearTerminalLayer struct {
	width, height int
}

func (l clearTerminalLayer) Bounds() image.Rectangle {
	return image.Rect(0, 0, max(1, l.width), max(1, l.height))
}

func (l clearTerminalLayer) Draw(scr uv.Screen, area image.Rectangle) {
	area = area.Intersect(l.Bounds())
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			scr.SetCell(x, y, nil)
		}
	}
}

func (m model) borderedLoginCanvas(termWidth, termHeight int) (*lipgloss.Canvas, bool) {
	if m.borderRegistry == nil || termWidth < 80 || termHeight < 24 {
		return nil, false
	}
	if m.mode != ModeLogin && m.mode != ModePassword && m.mode != ModeLoading {
		return nil, false
	}
	formKey, cacheForm := m.loginViewKey(termWidth, termHeight)
	renderer := m.borderRegistry.get(m.selectedBorderStyle)
	if renderer == nil {
		return nil, false
	}

	formWidth := min(96, max(15, termWidth-24))
	if m.config.BorderLayout == "fill" {
		if probe, ok := renderer.Frame(&m, termWidth, termHeight, 1, 1); ok && probe.Grid != nil {
			interiorWidth := probe.Grid.W - probe.Left - probe.Right - 4
			formWidth = min(formWidth, max(15, interiorWidth))
		}
	}
	formParts := make([]string, 0, 4)
	if m.config.ShowTime && !suppressOuterBorder(m.selectedBackground) {
		timeLine := lipgloss.NewStyle().Foreground(FgSecondary).Width(formWidth).Align(lipgloss.Center).
			Render(time.Now().Format("15:04:05 Mon Jan 02, 2006"))
		formParts = append(formParts, timeLine, "")
	}
	if ticker := m.renderTypewriterLine(min(100, termWidth-8) - 6); ticker != "" {
		formParts = append(formParts, ticker, "")
	}
	var sessionArt string
	if m.selectedSession != nil && termHeight >= 26 {
		sessionArt = m.getSessionArt(m.selectedSession.Name)
	}
	buildForm := func(includeSessionArt bool) string {
		parts := append([]string(nil), formParts...)
		if includeSessionArt && sessionArt != "" {
			parts = append(parts, sessionArt)
			if termHeight >= 30 {
				parts = append(parts, "", "")
			}
		}
		parts = append(parts, m.renderMainForm(formWidth))
		return lipgloss.JoinVertical(lipgloss.Center, parts...)
	}
	form := ""
	if cacheForm && m.loginViewCache.formValid && m.loginViewCache.formKey == formKey {
		form = m.loginViewCache.form
	} else {
		form = buildForm(true)
	}
	frame, ok := renderer.Frame(&m, termWidth, termHeight, lipgloss.Width(form), lipgloss.Height(form))
	if !ok && sessionArt != "" {
		form = buildForm(false)
		frame, ok = renderer.Frame(&m, termWidth, termHeight, lipgloss.Width(form), lipgloss.Height(form))
	}
	if !ok {
		return nil, false
	}
	if cacheForm {
		m.loginViewCache.formKey = formKey
		m.loginViewCache.form = form
		m.loginViewCache.formValid = true
	}

	layers := m.borderBackgroundLayers(termWidth, termHeight)
	idleFrame, shimmerActive := borderShimmerFrame(time.Since(m.idleTimer))
	shimmerActive = shimmerActive && m.borderEffect == "shimmer" && renderer.SupportsEffect("shimmer")
	paletteMode := m.borderPaletteMode
	if !validBorderPalette(paletteMode) {
		paletteMode = "theme"
	}
	failureAge := time.Duration(0)
	hasFailure := !m.borderFailureAt.IsZero()
	if hasFailure {
		failureAge = time.Since(m.borderFailureAt)
	}
	layers = append(layers, lipgloss.NewLayer(nfoGridLayer{
		Grid: frame.Grid, Palette: borderPalette(paletteMode), PaletteMode: paletteMode,
		Profile: colorProfile, Left: frame.Left, Top: frame.Top, Right: frame.Right, Bottom: frame.Bottom,
		Shimmer: shimmerActive, ShimmerPrimary: frame.ShimmerPrimary, Frame: idleFrame, Speed: frame.Speed, Rest: 4,
		FailureAge: failureAge, HasFailure: hasFailure,
	}).X(frame.X).Y(frame.Y))

	layers = append(layers, lipgloss.NewLayer(m.loginContentLayer(form)).X(frame.FormX).Y(frame.FormY))
	help := lipgloss.NewStyle().Foreground(FgMuted).Render(m.renderMainHelpFit(termWidth - 2))
	layers = append(layers, lipgloss.NewLayer(m.loginContentLayer(help)).
		X(max(0, (termWidth-lipgloss.Width(help))/2)).Y(frame.HelpY))

	return lipgloss.NewCanvas(layers...), true
}

func (m model) borderBackgroundLayers(termWidth, termHeight int) []*lipgloss.Layer {
	layers := []*lipgloss.Layer{lipgloss.NewLayer(clearTerminalLayer{width: termWidth, height: termHeight}).X(0).Y(0)}
	if m.mode != ModeLogin && m.mode != ModePassword {
		return layers
	}
	if m.enableFire || m.selectedBackground == "fire" || m.selectedBackground == "fire+rain" {
		fireHeight := (termHeight * 2) / 5
		layers = append(layers, lipgloss.NewLayer(m.addFireEffect("")).X(0).Y(termHeight-fireHeight))
		return layers
	}
	var background string
	switch m.selectedBackground {
	case "ascii-rain":
		background = m.addAsciiRain("")
	case "matrix":
		background = m.addMatrixEffect("")
	case "fireworks":
		background = m.addFireworksEffect("")
	case "sonar":
		background = m.addSonarEffect("")
	case "cracktro":
		background = m.addCracktroEffect("")
	case "plasma":
		background = m.addPlasmaEffect("")
	case "aquarium":
		background = m.addAquariumEffect("")
	case "ak-47":
		background = m.addAK47Effect("")
	case "l2a1-sar":
		background = m.addL2A1SAREffect("")
	default:
		if isMeshBackground(m.selectedBackground) {
			background = m.addMeshBackgroundEffect("")
		}
	}
	if background != "" {
		layers = append(layers, lipgloss.NewLayer(background).X(0).Y(0))
	}
	return layers
}
