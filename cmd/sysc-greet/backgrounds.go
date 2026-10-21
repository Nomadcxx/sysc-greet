package main

import (
	"image/color"
	"strings"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
)

// Background Effects - Extracted during Phase 5 refactoring
// This file contains background animation rendering (fire, matrix, rain, fireworks)
// IMPORTANT: These render functions are called from View() (value receiver).
// All state mutation (Update, Resize) must happen in Update() handlers, never here.

// applyBackgroundAnimation routes to the appropriate background effect based on selection
func (m model) applyBackgroundAnimation(content string) string {
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
	return syscgoanimations.NewAK47Effect(width, height, theme)
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
	return syscgoanimations.NewL2A1SAREffect(width, height, theme)
}

func (m model) addL2A1SAREffect(content string) string {
	if m.l2a1SAREffect == nil {
		return content
	}
	return m.l2a1SAREffect.Render()
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
