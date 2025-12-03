package animations

import (
	"math/rand"
	"strings"
	"time"
)

// Created ticker.go for animated ticker effects

// TickerAnimation provides animated loading/thinking effect
type TickerAnimation struct {
	frame      int
	lastUpdate time.Time
	frameDur   time.Duration
}

// NewTickerAnimation creates a new ticker animation
func NewTickerAnimation() *TickerAnimation {
	return &TickerAnimation{
		frame:      0,
		lastUpdate: time.Now(),
		frameDur:   time.Millisecond * 150, // 150ms per frame
	}
}

// GetFrame returns the current animation frame
// Returns a string like "⠋", "⠙", "⠹", etc. (braille spinner)
func (t *TickerAnimation) GetFrame() string {
	now := time.Now()
	if now.Sub(t.lastUpdate) >= t.frameDur {
		t.frame = (t.frame + 1) % len(spinnerFrames)
		t.lastUpdate = now
	}
	return spinnerFrames[t.frame]
}

// GetTitle returns the animated title replacing "SESSIONS"
func (t *TickerAnimation) GetTitle(width int) string {
	spinner := t.GetFrame()

	// Create animated title like: "⠋ LOADING ⠋"
	text := " LOADING "
	decorated := spinner + text + spinner

	// Center it with dashes
	remaining := width - len(decorated)
	if remaining < 0 {
		remaining = 0
	}
	leftDashes := remaining / 2
	rightDashes := remaining - leftDashes

	return strings.Repeat("─", leftDashes) + decorated + strings.Repeat("─", rightDashes)
}

// Braille spinner frames (smooth rotation effect)
var spinnerFrames = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// RoastingTicker provides scrolling text with WM-specific roasts
type RoastingTicker struct {
	offset      int
	lastUpdate  time.Time
	frameDur    time.Duration
	roasts      []string      // Array of individual roast phrases
	currentWM   string
	roastIndex  int           // Which roast we're currently showing
	paused      bool
	pauseUntil  time.Time
}

// NewRoastingTicker creates a scrolling roast ticker
// If customRoasts is empty, falls back to hardcoded defaults
func NewRoastingTicker(wmName string, customRoasts string) *RoastingTicker {
	roastText := customRoasts
	if roastText == "" {
		roastText = getRoastForWM(wmName)
	}

	return &RoastingTicker{
		offset:     0,
		lastUpdate: time.Now(),
		frameDur:   time.Millisecond * 33, // CHANGED 2025-10-04 - Reduced speed by 30% (25ms -> 33ms)
		roasts:     splitRoasts(roastText),
		currentWM:  wmName,
		roastIndex: 0,
		paused:     false,
		pauseUntil: time.Now(),
	}
}

// UpdateWM changes the roast text when WM selection changes
func (r *RoastingTicker) UpdateWM(wmName string, customRoasts string) {
	if wmName != r.currentWM {
		roastText := customRoasts
		if roastText == "" {
			roastText = getRoastForWM(wmName)
		}

		r.roasts = splitRoasts(roastText)
		r.currentWM = wmName
		r.offset = 0
		r.roastIndex = 0
		r.lastUpdate = time.Now()
		r.paused = false
	}
}

// splitRoasts splits a roast string on │ separator and cleans up
// Randomize roast order
func splitRoasts(roastText string) []string {
	// Split on │ separator
	parts := strings.Split(roastText, "│")

	var cleaned []string
	for i, part := range parts {
		// Trim whitespace
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		// For first part, strip "WM:" prefix if present
		if i == 0 {
			// Find first colon and strip everything before it
			colonIdx := strings.Index(trimmed, ":")
			if colonIdx > 0 && colonIdx < 20 { // Sanity check - WM names are short
				// Skip the "WM:" part and the space after colon
				trimmed = strings.TrimSpace(trimmed[colonIdx+1:])
			}
		}

		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	// Shuffle the roasts for random order
	rand.Shuffle(len(cleaned), func(i, j int) {
		cleaned[i], cleaned[j] = cleaned[j], cleaned[i]
	})

	return cleaned
}

// GetScrollingText returns the scrolling text for given width
// Cycle through individual roast phrases
func (r *RoastingTicker) GetScrollingText(width int) string {
	// Safety check
	if len(r.roasts) == 0 {
		return strings.Repeat(" ", width)
	}

	now := time.Now()

	// Check if we're in pause state
	if r.paused {
		if now.Before(r.pauseUntil) {
			// Still paused, show empty spaces
			return strings.Repeat(" ", width)
		}
		// Pause over, move to next roast and reset
		r.paused = false
		r.offset = 0
		r.roastIndex = (r.roastIndex + 1) % len(r.roasts)
	}

	// Get current roast phrase
	currentRoast := r.roasts[r.roastIndex]

	// Advance scroll position
	if now.Sub(r.lastUpdate) >= r.frameDur {
		r.offset++
		r.lastUpdate = now

		// Check if we've scrolled the entire message off screen
		// Total scroll distance = text length + width (to fully clear the view)
		if r.offset >= len(currentRoast)+width {
			// Start pause before next roast
			r.paused = true
			r.pauseUntil = now.Add(time.Second * 2) // 2 second pause between roasts
			r.offset = 0
			return strings.Repeat(" ", width)
		}
	}

	// Pad text with leading/trailing spaces
	paddedText := strings.Repeat(" ", width) + currentRoast + strings.Repeat(" ", width)

	start := r.offset
	end := start + width

	// Ensure we don't go out of bounds
	if end > len(paddedText) {
		end = len(paddedText)
	}
	if start >= len(paddedText) {
		return strings.Repeat(" ", width)
	}

	result := paddedText[start:end]

	// Force exact width, truncate if needed
	// Ensure exactly width characters (pad or truncate)
	runes := []rune(result)
	if len(runes) > width {
		// Truncate to exact width
		result = string(runes[:width])
	} else if len(runes) < width {
		// Pad to exact width
		result += strings.Repeat(" ", width-len(runes))
	}

	return result
}

// WM roast messages - funny quotes about each window manager
// Expanded roasts with community feedback
func getRoastForWM(wmName string) string {
	roasts := map[string]string{
		// GNOME - The Great Destroyer
		"GNOME": "GNOME: Removing features since 3.0 │ " +
			"If a soyboy was a desktop environment │ Where customization goes to die │ " +
			"If Lenin wore programming socks │ 'We know better than you' - The Desktop │ " +
			"Where Linux development meets ideology │ " +
			"System tray? Never heard of her │ Extensions break every update │ " +
			"Minimizing windows is a power user feature │ GNOME devs: Making users migrate to KDE since 2011 │ " +
			"Eulogizing legacy code as toxic masculinity since 2023 │ " +
			"We fixed your workflow by removing options you didn't know you hated │ " +
			"Your right-click was problematic; consider it canceled │ " +
			"We liberated your desktop from X11's toxic flexibility │ " +
			"Dynamic workspaces: Dynamic enough to disappear when you need them most │ " +
			"Dynamic theming: Themes that adapt to your apps, by breaking them│ " +
			"Open-source openness: Openly open to cancelling devs we dislike │ " +
			"Why is all my text aligning left? │ " +
			"False consciousness in flat design │ " +
			"Dunning Kruger neckbeards conflating political praxis with programming, many such cases│ ",

		"GNOME on Wayland": "GNOME on Wayland: Breaking what never worked │ " +
			"Screen sharing? That's a premium feature │ Variable refresh rate? Too advanced for you │ ",

		// KDE Plasma
		"KDE": "KDE: 5000 settings, 4950 you'll never use │ Bloatware masquerading as customization │ " +
			"RAM is cheap, right? RIGHT?! │ 47 daemons running to display a wallpaper │ " +
			"'Lightweight' said no one ever │ Akonadi has entered the chat (and consumed 2GB RAM) │ " +
			"KDE: When you want Windows-level resource usage on Linux │ " +
			"Breaking theming since Plasma 5 │ 15 different ways to crash kwin │ " +
			"Customization so deep, you'll need a map—and a PhD │ " +
			"Your desktop, now with 57 shades of widget—pick wisely │ ",

		"KDE Plasma": "KDE Plasma: Baloo indexing your soul at 100% CPU │ " +
			"Settings menu has settings for the settings │ Compositor crashed? Just restart it for the 5th time today │ " +
			"Customization so deep, you'll need a map—and a PhD │ " +
			"Your desktop, now with 57 shades of widget—pick wisely │ ",

		// The Chaotic Ones
		"Hyprland": ">install hyprland update >breaks again >downgrade again >repeat │ Breaking configs since yesterday │ " +
			"It really isn't that bad │ Animations over stability every time │ " +
			"'Let me just rewrite this core system real quick' - Vaxry │ " +
			"Your config worked yesterday? Not anymore! │ Git blame Vaxry │ " +
			"Voted number one tiling manager on Reddit 2023-2024 │ " +
			"The tiktok tiling manager - swipe right on stability │ ",

		// The Perfect
		"niri": "niri: Perfection in compositor form │ Scrollable tiling done right │ " +
			"No bugs, only features │ The chosen one │ Russian excellence │ " +
			"Because finite desktops were too confining │ ",

		"NIRI": "NIRI: Perfection in compositor form │ Scrollable tiling done right │ " +
			"No bugs, only features │ The chosen one │ Russian excellence │ " +
			"Because finite desktops were too confining │ ",

		// The Actually Good Ones
		"XFCE": "XFCE: The best desktop environment, period │ Lightweight, stable, customizable │ " +
			"Doesn't remove features you love │ Doesn't consume 4GB RAM │ " +
			"XFCE: Quietly being perfect while GNOME implodes │ GTK's last stand │ ",

		"Sway":      "Sway: i3 but we pretend X11 never existed │ Minimalism with Wayland pain │ ",
		"i3":        "i3: Tiling before it was cool (and bloated) │ The last WM that just works │ X11 gang represent │ " +
			"Status bars? Slap on i3bar.. │ ",
		"AwesomeWM": "AwesomeWM: Lua configs because sanity is overrated │ ",
		"awesome":   "Awesome: Lua configs because XML wasn't painful enough │ ",

		// The Memes
		"dwm":   "dwm: Recompile to change wallpaper │ Suckless: Because git patches are a lifestyle │ " +
			"Actually pretty based │ ",
		"bspwm": "bspwm: For when you want to write more shell scripts │ Binary space partitioning your sanity │ " +
			"Wayland? Over our dead keyboard shortcuts │ ",
		"qtile": "Qtile: Python configs for people who can't C │ ",
		"xmonad": "Xmonad: Haskell - because learning WM config should require a PhD │ " +
			"Recompile to apply config │ Monads in your window manager │ ",

		// The Forgotten Ones
		"Openbox":       "Openbox: The WM your grandma uses │ Right-click: The desktop experience │ ",
		"Fluxbox":       "Fluxbox: Like Openbox but with more Y2K vibes │ ",
		"Enlightenment": "Enlightenment: Still waiting for E17 │ Remember when this was the future? │ ",

		// GNOME Forks (GNOME 2 refugees)
		"Cinnamon": "Cinnamon: GNOME 2 cosplay │ Mint's apology for GNOME 3 │ " +
			"The number one voted desktop for Windows users │ ",
		"MATE":     "MATE: GNOME 2 but we actually mean it │ Keeping the dream alive │ ",
		"Budgie":   "Budgie: Solus says 'we can fix GNOME' │ Narrator: They couldn't │ ",

		// The Lightweights
		"LXQt": "LXQt: LXDE but now with more Q's │ Qt's lightweight cousin │ ",
		"LXDE": "LXDE: For when your potato is actually a potato │ ",

		// Wayland Pioneers
		"Wayfire": "Wayfire: Compiz nostalgia in Wayland │ Spinning cube! │ ",
		"River":   "River: Minimalism meets Zig │ For people who think Sway has too many features │ ",

		// The Tilers
		"leftwm":   "LeftWM: Rust btw │ Tiling for people who read r/unixporn │ " +
			"A rust tiling manager......................... │ ",
		"Herbstluftwm": "Herbstluftwm: German engineering applied to window management │ ",

		// Gaming WMs
		"Gamescope": "Gamescope: For when you game more than you configure │ ",
	}

	// Improved WM name matching with regex cleanup
	// Clean up the WM name - remove parentheses, "Session", "on", etc.
	cleanedName := wmName
	cleanedName = strings.ReplaceAll(cleanedName, "(Wayland)", "")
	cleanedName = strings.ReplaceAll(cleanedName, "(X11)", "")
	cleanedName = strings.ReplaceAll(cleanedName, "Session", "")
	cleanedName = strings.ReplaceAll(cleanedName, " on ", " ")
	cleanedName = strings.TrimSpace(cleanedName)
	wmLower := strings.ToLower(cleanedName)

	// First try exact match
	for key, roast := range roasts {
		if strings.ToLower(key) == wmLower {
			return roast
		}
	}

	// Then try partial match (contains)
	for key, roast := range roasts {
		keyLower := strings.ToLower(key)
		if strings.Contains(wmLower, keyLower) || strings.Contains(keyLower, wmLower) {
			return roast
		}
	}

	// Default roast for unknown WMs
	return "Your WM: So obscure even I don't have a roast for it │ "
}

// TypewriterTicker types out text one character at a time with a block cursor
// This provides a "typewriter" effect for the roast messages
type TypewriterTicker struct {
roasts       []string        // All roast messages
currentWM    string          // Current WM name
roastIndex   int             // Current message index
charIndex    int             // Current character being typed
lastUpdate   time.Time       // Last update time
charDelay    time.Duration   // Delay between characters (typing speed)
messageDelay time.Duration   // Delay after complete message
paused       bool            // Are we paused after message?
pauseUntil   time.Time       // When to unpause
}

// NewTypewriterTicker creates a new typewriter ticker
// If customRoasts is empty, falls back to hardcoded defaults
func NewTypewriterTicker(wmName string, customRoasts string) *TypewriterTicker {
	roastText := customRoasts
	if roastText == "" {
		roastText = getRoastForWM(wmName)
	}

	return &TypewriterTicker{
		roasts:       splitRoasts(roastText),
		currentWM:    wmName,
		roastIndex:   0,
		charIndex:    0,
		lastUpdate:   time.Now(),
		charDelay:    time.Millisecond * 50, // 50ms per character (adjustable typing speed)
		messageDelay: time.Second * 2,        // 2 second pause after complete message
		paused:       false,
		pauseUntil:   time.Now(),
	}
}

// UpdateWM changes the roast text when WM selection changes
func (t *TypewriterTicker) UpdateWM(wmName string, customRoasts string) {
	if wmName != t.currentWM {
		roastText := customRoasts
		if roastText == "" {
			roastText = getRoastForWM(wmName)
		}

		t.roasts = splitRoasts(roastText)
		t.currentWM = wmName
		t.roastIndex = 0
		t.charIndex = 0
		t.paused = false
		t.lastUpdate = time.Now()
	}
}

// GetTypewriterText returns the current typewriter text with block cursor
func (t *TypewriterTicker) GetTypewriterText(width int) string {
now := time.Now()

// Handle paused state (after complete message)
if t.paused {
if now.Before(t.pauseUntil) {
// Still paused - show complete message
if len(t.roasts) > 0 {
message := t.roasts[t.roastIndex]
if len(message) <= width {
// Center the message
padding := (width - len(message)) / 2
return strings.Repeat(" ", padding) + message + strings.Repeat(" ", width-len(message)-padding)
}
// Truncate if too long
return message[:width]
}
return strings.Repeat(" ", width)
}
// Pause over - move to next message
t.roastIndex = (t.roastIndex + 1) % len(t.roasts)
t.charIndex = 0
t.paused = false
t.lastUpdate = now
}

// Check if we need to type next character
if now.Sub(t.lastUpdate) >= t.charDelay {
if len(t.roasts) == 0 {
return strings.Repeat(" ", width)
}

currentMessage := t.roasts[t.roastIndex]

// Check if message is complete
if t.charIndex >= len(currentMessage) {
// Message complete - start pause
t.paused = true
t.pauseUntil = now.Add(t.messageDelay)
// Return complete message (will be displayed during pause)
if len(currentMessage) <= width {
padding := (width - len(currentMessage)) / 2
return strings.Repeat(" ", padding) + currentMessage + strings.Repeat(" ", width-len(currentMessage)-padding)
}
return currentMessage[:width]
}

// Type next character
t.charIndex++
t.lastUpdate = now
}

// Build current typed text with cursor
if len(t.roasts) == 0 {
return strings.Repeat(" ", width)
}

currentMessage := t.roasts[t.roastIndex]
typedText := currentMessage[:t.charIndex]

// Add block cursor (█)
result := typedText + "█"

// Center it if it fits
if len(result) <= width {
padding := (width - len(result)) / 2
return strings.Repeat(" ", padding) + result + strings.Repeat(" ", width-len(result)-padding)
}

// If too long, truncate (shouldn't happen with proper width)
if len(result) > width {
return result[:width]
}

return result
}
