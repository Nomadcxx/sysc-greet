package main

import (
	"image/color"
	"os"
	"strings"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
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

type meshBackgroundEffect interface {
	syscgoanimations.Animation
	Resize(int, int)
	SetCellAspect(int, int)
}

var meshBackgrounds = [...]struct{ id, label string }{
	{"higgs-field", "Higgs Field"},
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

func isMeshBackground(id string) bool {
	for _, background := range meshBackgrounds {
		if background.id == id {
			return true
		}
	}
	return false
}

func meshBackgroundID(label string) (string, bool) {
	for _, background := range meshBackgrounds {
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
	if isMeshBackground(m.selectedBackground) {
		return m.addMeshBackgroundEffect(content)
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
	return m.l2a1SAREffect.Render()
}

func (m model) newMeshBackgroundEffect(id string) meshBackgroundEffect {
	width, height := m.width, m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	theme := strings.ReplaceAll(m.currentTheme, " ", "-")
	if id == "higgs-field" {
		effect := syscgoanimations.NewHiggsFieldEffect(width, height, theme)
		applyTerminalCellAspect(effect)
		return effect
	}
	effect, err := syscgoanimations.NewRotatingModelEffect(id, width, height, theme)
	if err != nil {
		logDebug("Could not create rotating model background %q: %v", id, err)
		return nil
	}
	applyTerminalCellAspect(effect)
	return effect
}

func (m model) addMeshBackgroundEffect(content string) string {
	if m.meshEffect == nil {
		return content
	}
	return m.meshEffect.Render()
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
