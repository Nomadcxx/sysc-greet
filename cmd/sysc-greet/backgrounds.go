package main

import (
	"image/color"
	"os"
	"strings"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/termios"
)

func applyTerminalCellAspect(effect interface{ SetCellAspect(int, int) }) {
	size, err := termios.GetWinsize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}
	setCellAspect(effect, int(size.Col), int(size.Row), int(size.Xpixel), int(size.Ypixel))
}

func setCellAspect(effect interface{ SetCellAspect(int, int) }, cols, rows, pixelWidth, pixelHeight int) {
	if cols <= 0 || rows <= 0 || pixelWidth <= 0 || pixelHeight <= 0 {
		return
	}
	effect.SetCellAspect((pixelWidth+cols/2)/cols, (pixelHeight+rows/2)/rows)
}

func (m model) usesKittyMeshBackground() bool {
	return m.kittyGraphics && (m.mode == ModeLogin || m.mode == ModePassword) &&
		(m.selectedBackground == "ak-47" || m.selectedBackground == "l2a1-sar" || isRotatingModelBackground(m.selectedBackground))
}

func (m model) viewBackgroundColor() color.Color {
	if m.usesKittyMeshBackground() {
		return nil
	}
	return BgBase
}

func (m model) kittyCreditBackground() string {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	background := blankTerminalContent(width, height)
	if m.kittyCredit == "" {
		return background
	}
	lines := strings.Split(background, "\n")
	creditLines := strings.Split(lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(m.kittyCredit), "\n")
	if len(creditLines) > height {
		creditLines = creditLines[len(creditLines)-height:]
	}
	copy(lines[height-len(creditLines):], creditLines)
	return strings.Join(lines, "\n")
}

func (m model) renderKittyMeshFrame() (syscgoanimations.MeshFrame, bool) {
	switch m.selectedBackground {
	case "ak-47":
		if m.ak47Effect != nil {
			return m.ak47Effect.RenderFrame(), true
		}
	case "l2a1-sar":
		if m.l2a1SAREffect != nil {
			return m.l2a1SAREffect.RenderFrame(), true
		}
	default:
		if isRotatingModelBackground(m.selectedBackground) && m.rotatingModelEffect != nil {
			return m.rotatingModelEffect.RenderFrame(), true
		}
	}
	return syscgoanimations.MeshFrame{}, false
}

func syncKittyMeshFrame(previous model, next *model, msg tea.Msg) tea.Cmd {
	if !next.usesKittyMeshBackground() {
		next.kittyCredit = ""
		if sequence := next.kittyImages.clear(); sequence != "" {
			return tea.Raw(sequence)
		}
		return nil
	}

	refresh := !previous.usesKittyMeshBackground() ||
		previous.width != next.width || previous.height != next.height ||
		previous.selectedBackground != next.selectedBackground || previous.currentTheme != next.currentTheme
	switch msg.(type) {
	case bgTickMsg, tea.WindowSizeMsg:
		refresh = true
	}
	if !refresh {
		return nil
	}
	frame, ok := next.renderKittyMeshFrame()
	if !ok {
		return nil
	}
	sequence, err := next.kittyImages.render(frame, next.width, next.height)
	if err != nil {
		logDebug("Could not render Kitty mesh background: %v", err)
		next.kittyGraphics = false
		next.kittyCredit = ""
		if clear := next.kittyImages.clear(); clear != "" {
			return tea.Raw(clear)
		}
		return nil
	}
	next.kittyCredit = frame.Credit
	return tea.Raw(sequence)
}

var rotatingModelBackgrounds = [...]struct{ id, label string }{
	{"jesus-christ-statue", "Jesus Christ Statue"},
	{"jason", "Jason"},
	{"nymphe-der-flora", "Nymphe der Flora"},
	{"lowe", "Löwe"},
	{"roman-gladiator", "Roman Gladiator"},
	{"grabfigur", "Grabfigur"},
	{"laocoon-and-his-sons", "Laocoon and His Sons"},
	{"napoleon-bust", "Napoleon Bust"},
	{"brandenburg-gate", "Brandenburg Gate"},
	{"louis-xiv", "Louis XIV"},
	{"engel", "Engel"},
	{"st-george", "St George"},
	{"santa-muerte", "Santa Muerte"},
	{"aspasia", "Aspasia"},
	{"st-nepomuk", "St Nepomuk"},
	{"lion-statue", "Lion Statue"},
	{"f-14-tomcat", "F-14 Tomcat"},
	{"f-15e-strike-eagle", "F-15E Strike Eagle"},
	{"f-22-raptor", "F-22 Raptor"},
	{"fa-18e-super-hornet", "F/A-18E Super Hornet"},
}

func isRotatingModelBackground(id string) bool {
	for _, background := range rotatingModelBackgrounds {
		if background.id == id {
			return true
		}
	}
	return false
}

func rotatingModelBackgroundID(label string) (string, bool) {
	for _, background := range rotatingModelBackgrounds {
		if background.label == label {
			return background.id, true
		}
	}
	return "", false
}

// Background Effects - Extracted during Phase 5 refactoring
// This file contains background selection and render helpers.
// IMPORTANT: These render functions are called from View() (value receiver).
// All state mutation (Update, Resize) must happen in Update() handlers, never here.

// applyBackgroundAnimation routes to the appropriate background effect based on selection
func (m model) applyBackgroundAnimation(content string) string {
	if isRotatingModelBackground(m.selectedBackground) {
		return m.addRotatingModelEffect(content)
	}
	switch m.selectedBackground {
	case "fire": // Add fire effect rendering
		return m.addFireEffect(content)
	case "matrix":
		return m.addMatrixEffect(content)
	case "ascii-rain": // CHANGED 2025-10-08 - Add ascii rain effect
		return m.addAsciiRain(content)
	case "fireworks": // Add fireworks effect rendering
		return m.addFireworksEffect(content)
	case "aquarium":
		return m.addAquariumEffect(content)
	case "sonar":
		return m.addSonarEffect(content)
	case "cracktro":
		return m.addCracktroEffect(content)
	case "plasma":
		return m.addPlasmaEffect(content)
	case "ak-47":
		return m.addAK47Effect(content)
	case "l2a1-sar":
		return m.addL2A1SAREffect(content)
	case "none":
		fallthrough
	default:
		return content
	}
}

func (m model) newAK47Effect() *syscgoanimations.AK47Animation {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	// ponytail: sysc-Go accepts named palettes only; custom greeter palettes need an AK-47 API that accepts palette values.
	theme := strings.ReplaceAll(m.currentTheme, " ", "-")
	effect := syscgoanimations.NewAK47Effect(width, height, theme)
	applyTerminalCellAspect(effect)
	return effect
}

func (m model) addAK47Effect(content string) string {
	if m.ak47Effect == nil {
		return content
	}
	if m.usesKittyMeshBackground() {
		return m.kittyCreditBackground()
	}
	return m.ak47Effect.Render()
}

func (m model) newL2A1SAREffect() *syscgoanimations.L2A1SARAnimation {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	theme := strings.ReplaceAll(m.currentTheme, " ", "-")
	effect := syscgoanimations.NewL2A1SAREffect(width, height, theme)
	applyTerminalCellAspect(effect)
	return effect
}

func (m model) addL2A1SAREffect(content string) string {
	if m.l2a1SAREffect == nil {
		return content
	}
	if m.usesKittyMeshBackground() {
		return m.kittyCreditBackground()
	}
	return m.l2a1SAREffect.Render()
}

func (m model) newRotatingModelEffect(id string) *syscgoanimations.RotatingModelAnimation {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	theme := strings.ReplaceAll(m.currentTheme, " ", "-")
	effect, err := syscgoanimations.NewRotatingModelEffect(id, width, height, theme)
	if err != nil {
		logDebug("Could not create rotating model background %q: %v", id, err)
		return nil
	}
	applyTerminalCellAspect(effect)
	return effect
}

func (m model) addRotatingModelEffect(content string) string {
	if m.rotatingModelEffect == nil {
		return content
	}
	if m.usesKittyMeshBackground() {
		return m.kittyCreditBackground()
	}
	return m.rotatingModelEffect.Render()
}

// addFireEffect renders the fire background (render only, no state mutation)
func (m model) addFireEffect(content string) string {
	if m.fireEffect == nil {
		return content
	}
	return m.fireEffect.Render()
}

// addAsciiRain renders the ASCII rain background (render only, no state mutation)
func (m model) addAsciiRain(content string) string {
	if m.rainEffect == nil {
		return content
	}
	return m.rainEffect.Render()
}

// addMatrixEffect renders the Matrix-style background (render only, no state mutation)
func (m model) addMatrixEffect(content string) string {
	if m.matrixEffect == nil {
		return content
	}
	return m.matrixEffect.Render()
}

// addFireworksEffect renders the fireworks background (render only, no state mutation)
func (m model) addFireworksEffect(content string) string {
	if m.fireworksEffect == nil {
		return content
	}
	return m.fireworksEffect.Render()
}

// addAquariumEffect renders the aquarium background (render only, no state mutation)
func (m model) addAquariumEffect(content string) string {
	if m.aquariumEffect == nil {
		return content
	}
	return m.aquariumEffect.Render()
}

// addSonarEffect renders the sonar background (render only, no state mutation)
func (m model) addSonarEffect(content string) string {
	if m.sonarEffect == nil {
		return content
	}
	return m.sonarEffect.Render()
}

// addCracktroEffect renders the cracktro background (render only, no state mutation)
func (m model) addCracktroEffect(content string) string {
	if m.cracktroEffect == nil {
		return content
	}
	return m.cracktroEffect.Render()
}

// addPlasmaEffect renders the plasma background (render only, no state mutation)
func (m model) addPlasmaEffect(content string) string {
	if m.plasmaEffect == nil {
		return content
	}
	return m.plasmaEffect.Render()
}

// getBackgroundColor returns the background color (always BgBase to prevent bleeding)
func (m model) getBackgroundColor() color.Color {
	// Always return BgBase to prevent bleeding
	return BgBase
}
