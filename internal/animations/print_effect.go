package animations

import (
	"strings"
	"time"
	"unicode/utf8"
)

// PrintEffect manages the typewriter animation state
type PrintEffect struct {
	lines          []string
	currentLine    int
	currentChar    int
	charDelay      time.Duration
	printHeadChar  rune
	showPrintHead  bool
	complete       bool
	lastTickTime   time.Time
	cycleVariants  bool
	variantIndex   int
	allVariants    []string
	cycleDuration  time.Duration
	cycleStartTime time.Time
}

// NewPrintEffect creates a new typewriter animation effect
func NewPrintEffect(text string, charDelay time.Duration) *PrintEffect {
	lines := strings.Split(text, "\n")
	return &PrintEffect{
		lines:         lines,
		currentLine:   0,
		currentChar:   0,
		charDelay:     charDelay,
		printHeadChar: '█',
		showPrintHead: true,
		complete:      false,
		lastTickTime:  time.Now(),
	}
}

// NewPrintEffectWithVariants creates a print effect that cycles through multiple ASCII variants
func NewPrintEffectWithVariants(variants []string, charDelay, cycleDuration time.Duration) *PrintEffect {
	if len(variants) == 0 {
		return NewPrintEffect("", charDelay)
	}

	lines := strings.Split(variants[0], "\n")
	return &PrintEffect{
		lines:          lines,
		currentLine:    0,
		currentChar:    0,
		charDelay:      charDelay,
		printHeadChar:  '█',
		showPrintHead:  true,
		complete:       false,
		lastTickTime:   time.Now(),
		cycleVariants:  true,
		variantIndex:   0,
		allVariants:    variants,
		cycleDuration:  cycleDuration,
		cycleStartTime: time.Now(),
	}
}

// Tick advances the animation if enough time has passed
func (p *PrintEffect) Tick(currentTime time.Time) bool {
	// Check if we should cycle to next variant
	if p.cycleVariants && p.complete {
		elapsed := currentTime.Sub(p.cycleStartTime)
		if elapsed >= p.cycleDuration {
			// Move to next variant
			p.variantIndex = (p.variantIndex + 1) % len(p.allVariants)
			p.Reset(p.allVariants[p.variantIndex])
			return true
		}
		return false // Animation complete, waiting for cycle time
	}

	if p.complete {
		return false // No more changes
	}

	// Check if enough time has passed for next character
	if currentTime.Sub(p.lastTickTime) < p.charDelay {
		return false
	}

	p.lastTickTime = currentTime

	if p.currentLine >= len(p.lines) {
		p.complete = true
		p.showPrintHead = false
		return false
	}

	currentLineText := p.lines[p.currentLine]
	lineLength := utf8.RuneCountInString(currentLineText)

	if p.currentChar < lineLength {
		// Advance character in current line
		p.currentChar++
		return true
	} else {
		// Move to next line
		p.currentLine++
		p.currentChar = 0
		return true
	}
}

// Reset resets the animation with new text
func (p *PrintEffect) Reset(text string) {
	p.lines = strings.Split(text, "\n")
	p.currentLine = 0
	p.currentChar = 0
	p.complete = false
	p.showPrintHead = true
	p.lastTickTime = time.Now()
	p.cycleStartTime = time.Now()
}

// GetVisibleLines returns the currently visible lines (completed + current line being typed)
func (p *PrintEffect) GetVisibleLines() []string {
	var visible []string

	for lineIdx, line := range p.lines {
		if lineIdx < p.currentLine {
			// Completed line
			visible = append(visible, line)
		} else if lineIdx == p.currentLine {
			// Current line being typed
			if p.currentChar > 0 {
				runes := []rune(line)
				visibleRunes := runes[:min(p.currentChar, len(runes))]
				visibleText := string(visibleRunes)

				// Add print head if not at end of line
				if p.showPrintHead && p.currentChar < len(runes) {
					visibleText += string(p.printHeadChar)
				}

				visible = append(visible, visibleText)
			} else if p.showPrintHead {
				visible = append(visible, string(p.printHeadChar))
			}
		}
		// Lines below current line are not shown yet
	}

	return visible
}

// IsComplete returns whether the animation has finished
func (p *PrintEffect) IsComplete() bool {
	return p.complete
}

// SetSpeed updates the character delay
func (p *PrintEffect) SetSpeed(charDelay time.Duration) {
	p.charDelay = charDelay
}
