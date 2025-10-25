package main

import (
	"image/color"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
)

// Background Effects - Extracted during Phase 5 refactoring
// This file contains background animation effects (fire, matrix, rain, fireworks)

// applyBackgroundAnimation routes to the appropriate background effect based on selection
func (m model) applyBackgroundAnimation(content string, width, height int) string {
	switch m.selectedBackground {
	case "fire": // Add fire effect rendering
		return m.addFireEffect(content, width, height)
	case "matrix":
		return m.addMatrixEffect(content, width, height)
	case "ascii-rain": // CHANGED 2025-10-08 - Add ascii rain effect
		return m.addAsciiRain(content, width, height)
	case "fireworks": // Add fireworks effect rendering
		return m.addFireworksEffect(content, width, height)
	case "aquarium":
		return m.addAquariumEffect(content, width, height)
	case "none":
		fallthrough
	default:
		return content
	}
}

// addFireEffect renders the fire background using the internal fire effect engine
// Fire effect background rendering
func (m model) addFireEffect(content string, width, height int) string {
	if m.fireEffect == nil {
		return content
	}

	// CHANGED 2025-10-06 - Only resize if dimensions actually changed
	// Store last dimensions to avoid unnecessary reinits
	if m.lastFireWidth != width || m.lastFireHeight != height {
		m.fireEffect.Resize(width, height)
		m.lastFireWidth = width
		m.lastFireHeight = height
	}

	// Update palette from current theme
	m.fireEffect.UpdatePalette(animations.GetFirePalette(m.currentTheme))

	// Update fire simulation
	m.fireEffect.Update(m.animationFrame)

	// Render fire background
	fireBackground := m.fireEffect.Render()

	// For now, just return fire (content will be overlaid by main rendering)
	return fireBackground
}

// addAsciiRain renders the ASCII rain background effect
// CHANGED 2025-10-08 - Rain effect background rendering
func (m model) addAsciiRain(content string, width, height int) string {
	if m.rainEffect == nil {
		return content
	}

	// CHANGED 2025-10-08 - Only resize if dimensions actually changed
	// Store last dimensions to avoid unnecessary reinits
	if m.lastRainWidth != width || m.lastRainHeight != height {
		m.rainEffect.Resize(width, height)
		m.lastRainWidth = width
		m.lastRainHeight = height
	}

	// Update palette from current theme
	m.rainEffect.UpdatePalette(animations.GetRainPalette(m.currentTheme))

	// Update rain simulation
	m.rainEffect.Update(m.animationFrame)

	// Render rain background
	rainBackground := m.rainEffect.Render()

	// For now, just return rain (content will be overlaid by main rendering)
	return rainBackground
}

// addMatrixEffect renders the Matrix-style background effect
// Matrix effect background rendering
func (m model) addMatrixEffect(content string, width, height int) string {
	if m.matrixEffect == nil {
		return content
	}

	// Only resize if dimensions actually changed
	// Store last dimensions to avoid unnecessary reinits
	if m.lastMatrixWidth != width || m.lastMatrixHeight != height {
		m.matrixEffect.Resize(width, height)
		m.lastMatrixWidth = width
		m.lastMatrixHeight = height
	}

	// Update palette from current theme
	m.matrixEffect.UpdatePalette(animations.GetMatrixPalette(m.currentTheme))

	// Update matrix simulation
	m.matrixEffect.Update(m.animationFrame)

	// Render matrix background
	matrixBackground := m.matrixEffect.Render()

	// For now, just return matrix (content will be overlaid by main rendering)
	return matrixBackground
}

// addFireworksEffect renders the fireworks background effect
func (m model) addFireworksEffect(content string, width, height int) string {
	if m.fireworksEffect == nil {
		return content
	}

	// Only resize if dimensions actually changed
	if m.lastFireworksWidth != width || m.lastFireworksHeight != height {
		m.fireworksEffect.Resize(width, height)
		m.lastFireworksWidth = width
		m.lastFireworksHeight = height
	}

	// Update palette from current theme
	m.fireworksEffect.UpdatePalette(animations.GetFireworksPalette(m.currentTheme))

	// Update fireworks simulation
	m.fireworksEffect.Update(m.animationFrame)

	// Render fireworks background
	fireworksBackground := m.fireworksEffect.Render()

	// For now, just return fireworks (content will be overlaid by main rendering)
	return fireworksBackground
}

// addAquariumEffect renders the aquarium background effect
func (m model) addAquariumEffect(content string, width, height int) string {
	if m.aquariumEffect == nil {
		return content
	}

	// Resize if dimensions changed
	if m.lastAquariumWidth != width || m.lastAquariumHeight != height {
		m.aquariumEffect.Resize(width, height)
		m.lastAquariumWidth = width
		m.lastAquariumHeight = height
	}

	// Update palette from current theme
	fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor, _ := getThemeColorsForAquarium(m.currentTheme)
	m.aquariumEffect.UpdatePalette(fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor)

	// Render aquarium (it handles its own updates in the main loop)
	aquariumBackground := m.aquariumEffect.Render()

	// Return aquarium as full background
	return aquariumBackground
}

// getBackgroundColor returns the background color (always BgBase to prevent bleeding)
func (m model) getBackgroundColor() color.Color {
	// Always return BgBase to prevent bleeding
	return BgBase
}
