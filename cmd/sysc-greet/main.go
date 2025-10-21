package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/ipc"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	themesOld "github.com/Nomadcxx/sysc-greet/internal/themes"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mbndr/figlet4go"
)

// Version info - set via ldflags during build
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// CHANGED 2025-10-06 - Add debug logging to file
var debugLog *log.Logger

func initDebugLog() {
	logFile, err := os.OpenFile("/tmp/sysc-greet-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to stderr if can't open log file
		debugLog = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
		return
	}
	debugLog = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func logDebug(format string, args ...interface{}) {
	if debugLog != nil {
		debugLog.Printf(format, args...)
	}
}

// TTY-safe colors with profile detection
var (
	// Detect color profile once at startup
	colorProfile = colorprofile.Detect(os.Stdout, os.Environ())
	complete     = lipgloss.Complete(colorProfile)

	// Backgrounds - using Complete() for TTY compatibility
	BgBase     color.Color
	BgElevated color.Color
	BgSubtle   color.Color
	BgActive   color.Color

	// Primary brand colors
	Primary   color.Color
	Secondary color.Color
	Accent    color.Color
	Warning   color.Color
	Danger    color.Color

	// Text colors
	FgPrimary   color.Color
	FgSecondary color.Color
	FgMuted     color.Color
	FgSubtle    color.Color

	// Border colors
	BorderDefault color.Color
	BorderFocus   color.Color
)

func init() {
	// Initialize colors with TTY fallbacks
	// Dark background - fallback to black on TTY
	BgBase = complete(
		lipgloss.Color("0"),       // ANSI black
		lipgloss.Color("235"),     // ANSI256 dark gray
		lipgloss.Color("#1a1a1a"), // TrueColor charcoal
	)
	BgElevated = BgBase
	BgSubtle = BgBase
	BgActive = BgBase

	// Primary violet - fallback to magenta on TTY
	Primary = complete(
		lipgloss.Color("5"),       // ANSI magenta
		lipgloss.Color("141"),     // ANSI256 purple
		lipgloss.Color("#8b5cf6"), // TrueColor violet
	)

	// Secondary cyan
	Secondary = complete(
		lipgloss.Color("6"),       // ANSI cyan
		lipgloss.Color("45"),      // ANSI256 cyan
		lipgloss.Color("#06b6d4"), // TrueColor cyan
	)

	// Accent green
	Accent = complete(
		lipgloss.Color("2"),       // ANSI green
		lipgloss.Color("42"),      // ANSI256 green
		lipgloss.Color("#10b981"), // TrueColor emerald
	)

	// Warning amber
	Warning = complete(
		lipgloss.Color("3"),       // ANSI yellow
		lipgloss.Color("214"),     // ANSI256 orange
		lipgloss.Color("#f59e0b"), // TrueColor amber
	)

	// Danger red
	Danger = complete(
		lipgloss.Color("1"),       // ANSI red
		lipgloss.Color("196"),     // ANSI256 red
		lipgloss.Color("#ef4444"), // TrueColor red
	)

	// Primary text - white
	FgPrimary = complete(
		lipgloss.Color("7"),       // ANSI white
		lipgloss.Color("255"),     // ANSI256 white
		lipgloss.Color("#f8fafc"), // TrueColor white
	)

	// Secondary text - light gray
	FgSecondary = complete(
		lipgloss.Color("7"),       // ANSI white
		lipgloss.Color("252"),     // ANSI256 light gray
		lipgloss.Color("#cbd5e1"), // TrueColor light gray
	)

	// Muted text - gray
	FgMuted = complete(
		lipgloss.Color("8"),       // ANSI bright black
		lipgloss.Color("244"),     // ANSI256 gray
		lipgloss.Color("#94a3b8"), // TrueColor gray
	)

	// Subtle text - dark gray
	FgSubtle = complete(
		lipgloss.Color("8"),       // ANSI bright black
		lipgloss.Color("240"),     // ANSI256 dark gray
		lipgloss.Color("#64748b"), // TrueColor dark gray
	)

	// Border default - dark gray
	BorderDefault = complete(
		lipgloss.Color("8"),       // ANSI bright black
		lipgloss.Color("238"),     // ANSI256 dark gray
		lipgloss.Color("#374151"), // TrueColor gray
	)

	BorderFocus = Primary
}

// Theme management functions moved to theme.go
// Includes: applyTheme, setThemeWallpaper, getAnimatedColor, getAnimatedBorderColor, getFocusColor

// Utility functions moved to utils.go
// Includes: centerText, stripANSI, stripAnsi, min, extractCharsWithAnsi

// Color palette definitions for different WMs/sessions
// CHANGED 2025-09-29 - Added custom color palettes for different session types
type ColorPalette struct {
	Name   string
	Colors []string // Hex colors for the rainbow effect
}

// Fire effect implementation (PSX DOOM algorithm)

var sessionPalettes = map[string]ColorPalette{
	"GNOME": {
		Name:   "GNOME Blue",
		Colors: []string{"#4285f4", "#34a853", "#fbbc05", "#ea4335", "#9c27b0", "#ff9800"},
	},
	"KDE": {
		Name:   "KDE Plasma",
		Colors: []string{"#3daee9", "#1cdc9a", "#f67400", "#da4453", "#8e44ad", "#f39c12"},
	},
	"Hyprland": {
		Name:   "Hyprland Neon",
		Colors: []string{"#89b4fa", "#a6e3a1", "#f9e2af", "#fab387", "#f38ba8", "#cba6f7"},
	},
	"Sway": {
		Name:   "Sway Minimal",
		Colors: []string{"#458588", "#98971a", "#d79921", "#cc241d", "#b16286", "#689d6a"},
	},
	"i3": {
		Name:   "i3 Classic",
		Colors: []string{"#458588", "#98971a", "#d79921", "#cc241d", "#b16286", "#689d6a"},
	},
	"Xfce": {
		Name:   "Xfce Fresh",
		Colors: []string{"#4e9a06", "#f57900", "#cc0000", "#75507b", "#3465a4", "#c4a000"},
	},
	"default": {
		Name:   "Glamorous",
		Colors: []string{"#8b5cf6", "#06b6d4", "#10b981", "#f59e0b", "#ef4444", "#ec4899"},
	},
}

// ASCII art generator with proper Unicode block character support
// CHANGED 2025-09-29 - Fixed Unicode block character handling issue in figlet4go
// Removed old session art generation

// CHANGED 2025-09-30 - Use real figlet binary instead of broken custom parser
// Fallback to figlet4go
func renderWithFiglet4goFallback(text, fontPath string, debug bool) (string, error) {
	ascii := figlet4go.NewAsciiRender()
	ascii.LoadFont(fontPath) // Ignore errors, use default if needed
	return ascii.Render(text)
}

// Parse figlet font file directly with proper Unicode support
// CHANGED 2025-09-29 - Core fix for Unicode block character rendering + encoding
// Parse figlet font file directly with proper Unicode support
// Added ASCII config loading system

// CHANGED 2025-10-01 - Enhanced ASCIIConfig with animation controls
// Added multi-ASCII variant support
type ASCIIConfig struct {
	Name               string
	ASCII              string   // DEPRECATED: Use ASCIIVariants instead
	ASCIIVariants      []string // Support multiple ASCII art variants (ascii_1, ascii_2, etc.)
	MaxASCIIHeight     int      // Track max height across all variants for normalization
	Colors             []string
	AnimationStyle     string  // "gradient", "wave", "pulse", "rainbow", "matrix", "typewriter", "glow", "static"
	AnimationSpeed     float64 // 0.1 (slow) to 2.0 (fast), default 1.0
	AnimationDirection string  // "left", "right", "up", "down", "center-out", "random"
}

// Parse multiple ASCII variants (ascii_1, ascii_2, etc.)
// Load ASCII configuration from file

// ASCII art and animation functions moved to ascii.go
// Includes: loadASCIIConfig, getSessionASCII, getSessionASCIIMonochrome, getSessionArt
// Animation: applyASCIIAnimation, applySmoothGradient, applyWaveAnimation, applyPulseAnimation,
// applyRainbowAnimation, applyMatrixAnimation, applyTypewriterAnimation, applyGlowAnimation, applyStaticColors
// Helpers: interpolateColors, parseHexColor

// CHANGED 2025-10-14 - Removed loadConfig() function and Config.Palettes field
// The sysc-greet.conf system was unused and confusing - hardcoded sessionPalettes provide all needed palettes

type Config struct {
	TestMode  bool
	Debug     bool
	ShowTime  bool
	ThemeName string
}

type ViewMode string

const (
	ModeLogin    ViewMode = "login"
	ModePassword ViewMode = "password"
	ModeLoading  ViewMode = "loading"
	ModePower    ViewMode = "power"
	ModeMenu     ViewMode = "menu"
	// Added new menu modes for structured menu system
	ModeThemesSubmenu      ViewMode = "themes_submenu"
	ModeBordersSubmenu     ViewMode = "borders_submenu"
	ModeBackgroundsSubmenu ViewMode = "backgrounds_submenu"
	ModeWallpaperSubmenu   ViewMode = "wallpaper_submenu" // CHANGED 2025-10-03 - Add wallpaper submenu for gslapper videos
	// CHANGED 2025-10-01 - Added release notes mode
	ModeReleaseNotes ViewMode = "release_notes"
	// CHANGED 2025-10-10 - Added screensaver mode
	ModeScreensaver ViewMode = "screensaver"
)

type FocusState int

const (
	FocusSession FocusState = iota
	FocusUsername
	FocusPassword
)

type model struct {
	usernameInput   textinput.Model
	passwordInput   textinput.Model
	spinner         spinner.Model
	sessions        []sessions.Session
	selectedSession *sessions.Session
	sessionIndex    int
	ipcClient       *ipc.Client
	theme           themesOld.Theme
	mode            ViewMode
	config          Config
	startTime       time.Time

	// Terminal dimensions
	width  int
	height int

	// Power menu
	powerOptions []string
	powerIndex   int

	// Session dropdown
	sessionDropdownOpen bool

	// Menu system
	menuOptions []string
	menuIndex   int
	// Added fields for functional menu system
	customASCIIText        string
	selectedBorderStyle    string
	selectedBackground     string
	currentTheme           string
	borderAnimationEnabled bool
	selectedFont           string
	// CHANGED 2025-10-01 - Added animation control fields
	selectedAnimationStyle     string
	selectedAnimationSpeed     float64
	selectedAnimationDirection string
	animationStyleOptions      []string
	animationDirectionOptions  []string

	// Focus management
	focusState FocusState

	// Animation state
	animationFrame int
	pulseColor     int
	borderFrame    int

	// Fire effect instance
	fireEffect     *animations.FireEffect
	lastFireWidth  int
	lastFireHeight int

	// CHANGED 2025-10-08 - Rain effect instance
	rainEffect     *animations.RainEffect
	lastRainWidth  int
	lastRainHeight int

	// Matrix effect instance
	matrixEffect     *animations.MatrixEffect
	lastMatrixWidth  int
	lastMatrixHeight int

	// Fireworks effect instance
	fireworksEffect     *animations.FireworksEffect
	lastFireworksWidth  int
	lastFireworksHeight int

	// CHANGED 2025-10-04 - Separate flags for multiple backgrounds
	enableFire bool

	// CHANGED 2025-10-05 - Add error message for authentication failures
	errorMessage string

	// CHANGED 2025-10-10 - Screensaver fields
	idleTimer         time.Time               // Time when idle started
	screensaverTime   time.Time               // Current time for screensaver display
	screensaverPrint  *animations.PrintEffect // CHANGED 2025-10-11 - Print effect animation for screensaver
	screensaverActive bool                    // CHANGED 2025-10-11 - Track if screensaver just activated

	// ASCII navigation fields for multi-variant support
	asciiArtIndex      int         // Current variant index (0-indexed)
	asciiArtCount      int         // Total variants available
	asciiMaxHeight     int         // Max height for normalization
	currentASCIIConfig ASCIIConfig // Cached config for current session

	capsLockOn bool // CAPS LOCK state detected via kitty keyboard protocol
}

type sessionSelectedMsg sessions.Session
type powerSelectedMsg string
type tickMsg time.Time

func doTick() tea.Cmd {
	// CHANGED 2025-10-04 - Reduced tick interval to 30ms for smoother ticker animation
	return tea.Tick(time.Millisecond*30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel(config Config, screensaverMode bool) model {
	// Setup username input with proper styling
	ti := textinput.New()
	ti.Prompt = ""      // Remove prompt, will be added by layout
	ti.Placeholder = "" // Remove placeholder
	// Updated for textinput v2 API
	ti.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	ti.Styles.Focused.Text = lipgloss.NewStyle().Foreground(FgPrimary)
	ti.Styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(FgMuted).Italic(true)

	// Setup password input
	pi := textinput.New()
	pi.Prompt = ""      // Remove prompt, will be added by layout
	pi.Placeholder = "" // Remove placeholder
	pi.EchoMode = textinput.EchoPassword
	// Updated for textinput v2 API
	pi.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	pi.Styles.Focused.Text = lipgloss.NewStyle().Foreground(FgPrimary)
	pi.Styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(FgMuted).Italic(true)

	// Load sessions
	sess, _ := sessions.LoadSessions()
	if config.TestMode && len(sess) == 0 {
		sess = []sessions.Session{
			{Name: "GNOME", Exec: "gnome-session", Type: "X11"},
			{Name: "KDE Plasma", Exec: "startplasma-x11", Type: "X11"},
			{Name: "Sway", Exec: "sway", Type: "Wayland"},
			{Name: "Hyprland", Exec: "Hyprland", Type: "Wayland"},
			{Name: "i3", Exec: "i3", Type: "X11"},
			{Name: "Xfce Session", Exec: "startxfce4", Type: "X11"},
		}
	}
	if config.Debug {
		logDebug(" Loaded %d sessions", len(sess))
		for _, s := range sess {
			fmt.Printf("  - %s (%s)\n", s.Name, s.Type)
		}
	}

	// Setup animated spinner
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(Primary)

	var ipcClient *ipc.Client
	var selectedSession *sessions.Session
	var sessionIndex int

	if !config.TestMode {
		// CHANGED 2025-10-05 - Proper IPC client error handling
		logDebug("Attempting to create IPC client...")
		client, err := ipc.NewClient()
		if err != nil {
			// CRITICAL: If IPC fails, we cannot authenticate with greetd
			// Log the error and exit rather than continue with nil client
			logDebug("FATAL: IPC client creation failed: %v", err)
			fmt.Fprintf(os.Stderr, "FATAL: Failed to create IPC client: %v\n", err)
			fmt.Fprintf(os.Stderr, "GREETD_SOCK environment variable: %s\n", os.Getenv("GREETD_SOCK"))
			fmt.Fprintf(os.Stderr, "This greeter must be run by greetd with GREETD_SOCK set.\n")
			os.Exit(1)
		}
		ipcClient = client
		logDebug("IPC client created successfully")

		// Load cached session and find its index
		cached, err := cache.LoadSelectedSession()
		if err != nil && config.Debug {
			logDebug(" Failed to load cached session: %v", err)
		} else if cached != nil {
			selectedSession = cached
			// Find the index of the cached session
			for i, s := range sess {
				if s.Name == cached.Name && s.Type == cached.Type {
					sessionIndex = i
					break
				}
			}
		}
	}

	// Default to first session if none selected
	if selectedSession == nil && len(sess) > 0 {
		selectedSession = &sess[0]
		sessionIndex = 0
	}

	// Load themes from directory
	themesDir := "themes"
	loadedThemes, err := themesOld.LoadThemesFromDir(themesDir)
	if err != nil && config.Debug {
		logDebug(" Failed to load themes: %v", err)
	}

	// Use specified theme if available, otherwise default
	currentTheme := themesOld.DefaultTheme
	if config.ThemeName != "" {
		if theme, ok := loadedThemes[config.ThemeName]; ok {
			currentTheme = theme
		}
	} else if theme, ok := loadedThemes["gnome"]; ok {
		currentTheme = theme
	}

	// Set initial focus
	ti.Focus()

	// REMOVED 2025-10-17 - Don't apply Dracula at initialization
	// The cached theme will be loaded immediately after model creation (line 558)
	// Applying Dracula here causes a race condition with the cached theme wallpaper

	// CHANGED 2025-10-11 - Determine initial mode
	initialMode := ModeLogin
	if screensaverMode {
		initialMode = ModeScreensaver
	}

	m := model{
		usernameInput:       ti,
		passwordInput:       pi,
		spinner:             sp,
		sessions:            sess,
		selectedSession:     selectedSession,
		sessionIndex:        sessionIndex,
		ipcClient:           ipcClient,
		theme:               currentTheme,
		mode:                initialMode,
		config:              config,
		startTime:           time.Now(),
		width:               80,
		height:              24,
		powerOptions:        []string{"Reboot", "Shutdown", "Cancel"},
		powerIndex:          0,
		sessionDropdownOpen: false,
		focusState:          FocusUsername,
		animationFrame:      0,
		pulseColor:          0,
		borderFrame:         0,
		// Initialize default border and background settings
		// Set Dracula as default theme and disable border animation
		selectedBorderStyle:    "classic",
		selectedBackground:     "none",
		currentTheme:           "dracula",
		borderAnimationEnabled: false,
		selectedFont:           "/usr/share/bubble-greet/fonts/dos_rebel.flf", // Absolute path
		customASCIIText:        "",
		// CHANGED 2025-10-01 - Initialize animation control defaults
		selectedAnimationStyle:     "gradient",
		selectedAnimationSpeed:     1.0,
		selectedAnimationDirection: "right",
		animationStyleOptions:      []string{"gradient", "wave", "pulse", "rainbow", "matrix", "typewriter", "glow", "static"},
		animationDirectionOptions:  []string{"right", "left", "up", "down", "center-out"},
		// CHANGED 2025-10-10 - Initialize screensaver timers
		idleTimer:       time.Now(),
		screensaverTime: time.Now(),
		// Initialize fire effect with default size
		fireEffect: animations.NewFireEffect(80, 30, animations.GetDefaultFirePalette()),
		// CHANGED 2025-10-08 - Initialize rain effect with default size
		rainEffect: animations.NewRainEffect(80, 30, animations.GetRainPalette("default")),
		// Initialize matrix effect with default size
		matrixEffect: animations.NewMatrixEffect(80, 30, animations.GetMatrixPalette("default")),
		// Initialize fireworks effect with default size
		fireworksEffect: animations.NewFireworksEffect(80, 30, animations.GetFireworksPalette("default")),
	}

	// CHANGED 2025-10-03 - Load cached preferences including session
	// CHANGED 2025-10-03 - Skip cache in test mode
	// FIXED 2025-10-17 - Apply Dracula as fallback if no cached theme exists
	themeApplied := false
	if !m.config.TestMode {
		if prefs, err := cache.LoadPreferences(); err == nil && prefs != nil {
			if prefs.Theme != "" {
				m.currentTheme = prefs.Theme
				applyTheme(prefs.Theme, m.config.TestMode)
				themeApplied = true
			}
			if prefs.Background != "" {
				m.selectedBackground = prefs.Background
			}
			if prefs.BorderStyle != "" {
				m.selectedBorderStyle = prefs.BorderStyle
			}
			if prefs.Session != "" {
				// Find matching session in m.sessions
				for i, s := range m.sessions {
					if s.Name == prefs.Session {
						m.selectedSession = &m.sessions[i]
						m.sessionIndex = i
						break
					}
				}
			}
			// FIXED 2025-10-17 - Load username and auto-advance to password if matches current session
			if prefs.Username != "" && m.selectedSession != nil && prefs.Session == m.selectedSession.Name {
				m.usernameInput.SetValue(prefs.Username)
				// FIXED 2025-10-17 - Automatically switch to password mode when username is cached
				m.mode = ModePassword
				m.focusState = FocusPassword
				m.usernameInput.Blur()
				m.passwordInput.Focus()
				logDebug("Loaded cached username '%s' for session: %s - auto-advancing to password", prefs.Username, m.selectedSession.Name)
			}
		}
	}

	// FIXED 2025-10-17 - Apply Dracula as fallback if no cached theme was loaded
	if !themeApplied {
		applyTheme("dracula", m.config.TestMode)
		logDebug("No cached theme found - applied Dracula as default")
	}

	// CHANGED 2025-10-11 - Initialize print effect if starting in screensaver mode
	if screensaverMode {
		ssConfig := loadScreensaverConfig()
		if ssConfig.AnimateOnStart && ssConfig.AnimationType == "print" && len(ssConfig.ASCIIVariants) > 0 {
			selectedASCII := ssConfig.ASCIIVariants[0]
			charDelay := time.Duration(ssConfig.AnimationSpeed) * time.Millisecond
			m.screensaverPrint = animations.NewPrintEffect(selectedASCII, charDelay)
			m.screensaverActive = true
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	// Request keyboard enhancements to get CAPS LOCK state reporting
	// RequestUniformKeyLayout enables kitty flags 4+8 which includes lock key state reporting
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		doTick(),
		tea.RequestUniformKeyLayout,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		logDebug("Terminal resized: %dx%d", msg.Width, msg.Height)
		return m, nil

	case tickMsg:
		m.animationFrame++
		m.pulseColor = (m.pulseColor + 1) % 100
		m.borderFrame = (m.borderFrame + 1) % 20

		// CHANGED 2025-10-10 - Update screensaver time and check for activation
		m.screensaverTime = time.Time(msg)

		// CHANGED 2025-10-11 - Tick print effect animation if in screensaver mode
		if m.mode == ModeScreensaver && m.screensaverPrint != nil {
			m.screensaverPrint.Tick(m.screensaverTime)
		}

		// Check for screensaver activation using configurable timeout
		if m.mode == ModeLogin || m.mode == ModePassword {
			ssConfig := loadScreensaverConfig()
			idleDuration := time.Since(m.idleTimer)
			if idleDuration >= time.Duration(ssConfig.IdleTimeout)*time.Minute && m.mode != ModeScreensaver {
				m.mode = ModeScreensaver
				m.screensaverActive = true // CHANGED 2025-10-11 - Mark screensaver as just activated

				// CHANGED 2025-10-11 - Initialize print effect animation if enabled
				if ssConfig.AnimateOnStart && ssConfig.AnimationType == "print" {
					// Get the ASCII variant to animate
					variantIndex := 0 // Start with first variant
					if len(ssConfig.ASCIIVariants) > 0 {
						selectedASCII := ssConfig.ASCIIVariants[variantIndex]
						charDelay := time.Duration(ssConfig.AnimationSpeed) * time.Millisecond
						m.screensaverPrint = animations.NewPrintEffect(selectedASCII, charDelay)
					}
				}
			}
		}

		// CHANGED 2025-10-04 - Update fire when enableFire is true
		if (m.enableFire || m.selectedBackground == "fire" || m.selectedBackground == "fire+rain") && m.fireEffect != nil {
			m.fireEffect.Update(m.animationFrame)
		}

		// CHANGED 2025-10-08 - Update rain when ascii-rain is selected
		if m.selectedBackground == "ascii-rain" && m.rainEffect != nil {
			m.rainEffect.Update(m.animationFrame)
		}

		// Update matrix when matrix background is selected
		if m.selectedBackground == "matrix" && m.matrixEffect != nil {
			m.matrixEffect.Update(m.animationFrame)
		}

		// Update fireworks when fireworks background is selected
		if m.selectedBackground == "fireworks" && m.fireworksEffect != nil {
			m.fireworksEffect.Update(m.animationFrame)
		}

		cmds = append(cmds, doTick())

	case sessionSelectedMsg:
		session := sessions.Session(msg)

		// FIXED 2025-10-17 - Track previous session to detect changes
		previousSession := ""
		if m.selectedSession != nil {
			previousSession = m.selectedSession.Name
		}

		m.selectedSession = &session
		// Update session index
		for i, s := range m.sessions {
			if s.Name == session.Name && s.Type == session.Type {
				m.sessionIndex = i
				break
			}
		}
		m.sessionDropdownOpen = false

		// FIXED 2025-10-17 - Clear and reload username when session changes
		if previousSession != "" && previousSession != session.Name {
			logDebug("Session changed from '%s' to '%s', clearing username", previousSession, session.Name)
			m.usernameInput.SetValue("") // Clear current username

			// Load cached username for NEW session if available
			if !m.config.TestMode {
				if prefs, err := cache.LoadPreferences(); err == nil && prefs != nil {
					if prefs.Username != "" && prefs.Session == session.Name {
						m.usernameInput.SetValue(prefs.Username)
						logDebug("Loaded cached username for new session: %s", session.Name)
					}
				}
			}
		}

		if m.config.Debug {
			logDebug(" Selected session: %s", session.Name)
		}
		if m.config.TestMode {
			fmt.Println("Test mode: Selected session:", session.Name)
			return m, tea.Quit
		} else {
			// Save to cache
			if err := cache.SaveSelectedSession(session); err != nil && m.config.Debug {
				logDebug(" Failed to save session: %v", err)
			}
			// CHANGED 2025-10-03 - Save session preference
			// CHANGED 2025-10-03 - Skip saving in test mode
			// FIXED 2025-10-17 - Save current username value (already loaded for this session above)
			if !m.config.TestMode {
				cache.SavePreferences(cache.UserPreferences{
					Theme:       m.currentTheme,
					Background:  m.selectedBackground,
					BorderStyle: m.selectedBorderStyle,
					Session:     session.Name,
					Username:    m.usernameInput.Value(), // Save current username value
				})
			}
			return m, tea.Batch(cmds...)
		}

	case powerSelectedMsg:
		action := string(msg)
		switch action {
		case "Reboot":
			if m.config.TestMode {
				fmt.Println("Test mode: Would reboot system")
				return m, tea.Quit
			}
			// FIXED 2025-10-17 - Use shutdown command for proper system reboot from greeter session
			// shutdown -r now coordinates with init system and works from any session
			fmt.Println("Rebooting...")
			exec.Command("shutdown", "-r", "now").Start()
			return m, tea.Quit
		case "Shutdown":
			if m.config.TestMode {
				fmt.Println("Test mode: Would shutdown system")
				return m, tea.Quit
			}
			// FIXED 2025-10-17 - Use shutdown command for proper system poweroff from greeter session
			// shutdown -h now coordinates with init system and works from any session
			fmt.Println("Shutting down...")
			exec.Command("shutdown", "-h", "now").Start()
			return m, tea.Quit
		case "Cancel":
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			cmds = append(cmds, textinput.Blink)
		}

	case string:
		if msg == "success" {
			// Removed delay workaround
			// Now we properly wait for greetd's success response in StartSession() before returning
			// This ensures greetd has finished session initialization regardless of hardware speed

			// FIXED 2025-10-17 - Save username to cache on successful login
			if !m.config.TestMode && m.selectedSession != nil {
				sessionName := m.selectedSession.Name
				username := m.usernameInput.Value()
				cache.SavePreferences(cache.UserPreferences{
					Theme:       m.currentTheme,
					Background:  m.selectedBackground,
					BorderStyle: m.selectedBorderStyle,
					Session:     sessionName,
					Username:    username,
				})
				logDebug("Saved username '%s' for session: %s", username, sessionName)
			}

			fmt.Println("Session started successfully")
			return m, tea.Quit
		} else {
			// FIXED 2025-10-17 - Return to login mode (not password mode) so user can fix username
			m.errorMessage = msg
			m.mode = ModeLogin
			m.usernameInput.SetValue("") // Clear username field
			m.passwordInput.SetValue("") // Clear password field
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			m.focusState = FocusUsername
			return m, textinput.Blink
		}
	case error:
		// FIXED 2025-10-17 - Return to login mode (not password mode) so user can fix username
		m.errorMessage = msg.Error()
		m.mode = ModeLogin
		m.usernameInput.SetValue("") // Clear username field
		m.passwordInput.SetValue("") // Clear password field
		m.usernameInput.Focus()
		m.passwordInput.Blur()
		m.focusState = FocusUsername
		return m, textinput.Blink

	case tea.KeyMsg:
		// CHANGED 2025-10-21 - Detect CAPS LOCK from kitty keyboard protocol
		// Kitty keyboard protocol sends CAPS LOCK and NUM LOCK as ModCapsLock and ModNumLock
		key := msg.Key()
		m.capsLockOn = (key.Mod & tea.ModCapsLock) != 0
		
		if m.config.Debug {
			// Log ALL key presses to debug what modifiers are being sent
			fmt.Fprintf(os.Stderr, "KEY: %q | Mod=%08b (%d) | CapsLock=%v\n", 
				key.Text, key.Mod, key.Mod, m.capsLockOn)
		}
		
		// CHANGED 2025-10-12 - Handle screensaver exit on any key press
		if m.mode == ModeScreensaver {
			return handleScreensaverInput(m, msg)
		}
		newModel, cmd := m.handleKeyInput(msg)
		m = newModel
		cmds = append(cmds, cmd)

	case tea.MouseMsg:
		// CHANGED 2025-10-12 - Exit screensaver and reset idle timer on mouse movement
		if m.mode == ModeScreensaver {
			m.mode = ModeLogin
			m.idleTimer = time.Now()
			return m, nil
		}
		// Reset idle timer on any mouse input in normal modes
		m.idleTimer = time.Now()
	}

	// Update components based on current mode and focus
	switch m.mode {
	case ModeLogin:
		if m.focusState == FocusUsername {
			var cmd tea.Cmd
			m.usernameInput, cmd = m.usernameInput.Update(msg)
			cmds = append(cmds, cmd)
			// FIXED 2025-10-17 - Clear error message when user starts typing in login mode
			if m.errorMessage != "" && len(m.usernameInput.Value()) > 0 {
				m.errorMessage = ""
			}
		}
	case ModePassword:
		if m.focusState == FocusPassword {
			var cmd tea.Cmd
			m.passwordInput, cmd = m.passwordInput.Update(msg)
			cmds = append(cmds, cmd)
			// CHANGED 2025-10-05 - Clear error message when user starts typing
			if m.errorMessage != "" && len(m.passwordInput.Value()) > 0 {
				m.errorMessage = ""
			}
		}
	case ModeLoading:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleKeyInput(msg tea.KeyMsg) (model, tea.Cmd) {
	// CHANGED 2025-10-12 - Reset idle timer on any key press to prevent screensaver activation
	m.idleTimer = time.Now()

	// Updated for tea.KeyMsg v2 API
	if m.config.Debug {
		keyStr := msg.String()
		fmt.Fprintf(os.Stderr, "KEY DEBUG: String='%s'\n", keyStr)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		// CHANGED 2025-10-10 - Disable Ctrl+C in production mode
		// Only allow Ctrl+C/Q to quit in test mode (when ipcClient is nil)
		if m.ipcClient == nil {
			// Test mode - allow quit
			return m, tea.Quit
		}
		// Production mode - ignore Ctrl+C/Q (security measure)
		return m, nil

	case "f1":
		// Remapped F1 to Menu
		// Main menu
		if m.mode == ModeLogin || m.mode == ModePassword {
			m.mode = ModeMenu
			m.menuIndex = 0

			// Build new structured menu
			m.menuOptions = []string{
				"Close Menu",
				"Themes",
				"Borders",
				"Backgrounds",
				"Wallpaper",
			}
			return m, nil
		}

	case "f2":
		// Remapped F2 to Sessions
		// Toggle session dropdown
		if m.mode == ModeLogin || m.mode == ModePassword {
			m.sessionDropdownOpen = !m.sessionDropdownOpen
			return m, nil
		}

	case "f3":
		// Remapped F3 to Notes
		// Release notes popup
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.config.Debug {
				fmt.Println("Debug: Opening release notes")
			}
			m.mode = ModeReleaseNotes
			m.usernameInput.Blur()
			m.passwordInput.Blur()
			return m, nil
		}

	case "f4":
		// F4 remains Power
		// Power menu
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.config.Debug {
				fmt.Println("Debug: Opening power menu")
			}
			m.mode = ModePower
			m.usernameInput.Blur()
			m.passwordInput.Blur()
			return m, nil
		}

	case "tab":
		// Cycle focus through form elements
		if m.mode == ModeLogin {
			switch m.focusState {
			case FocusSession:
				m.focusState = FocusUsername
				m.usernameInput.Focus()
			case FocusUsername:
				m.focusState = FocusSession
				m.usernameInput.Blur()
			}
			return m, textinput.Blink
		} else if m.mode == ModePassword {
			switch m.focusState {
			case FocusSession:
				m.focusState = FocusPassword
				m.passwordInput.Focus()
			case FocusPassword:
				m.focusState = FocusSession
				m.passwordInput.Blur()
			}
			return m, textinput.Blink
		}

	case "esc":
		switch m.mode {
		case ModePassword:
			// CHANGED 2025-10-18 22:05 - Allow ESC to return from password mode to login mode
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.passwordInput.SetValue("") // Clear password field
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			return m, textinput.Blink
		case ModePower:
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			return m, textinput.Blink
		case ModeMenu:
			// CHANGED 2025-09-30 - Add escape from menu
			m.mode = ModeLogin
			return m, nil
		// Add escape handling for submenus
		case ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeWallpaperSubmenu:
			// Go back to main menu
			m.mode = ModeMenu
			m.menuOptions = []string{
				"Close Menu",
				"Themes",
				"Borders",
				"Backgrounds",
				"Wallpaper",
			}
			m.menuIndex = 0
			return m, nil
		// Add escape handling for release notes
		case ModeReleaseNotes:
			// Return to login mode
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			return m, textinput.Blink
		default:
			if m.sessionDropdownOpen {
				m.sessionDropdownOpen = false
				return m, nil
			}
		}

	case "up", "k":
		if m.sessionDropdownOpen {
			if m.sessionIndex > 0 {
				m.sessionIndex--
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		} else if m.mode == ModePower {
			if m.powerIndex > 0 {
				m.powerIndex--
			}
			return m, nil
		} else if m.mode == ModeMenu || m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu {
			// Removed ModeVideoWallpapersSubmenu from navigation
			if m.menuIndex > 0 {
				m.menuIndex--
			}
			return m, nil
		} else if m.focusState == FocusSession {
			// Navigate sessions when session selector is focused
			if m.sessionIndex > 0 {
				m.sessionIndex--
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		}

	case "down", "j":
		if m.sessionDropdownOpen {
			if m.sessionIndex < len(m.sessions)-1 {
				m.sessionIndex++
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		} else if m.mode == ModePower {
			if m.powerIndex < len(m.powerOptions)-1 {
				m.powerIndex++
			}
		} else if m.mode == ModeMenu || m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu {
			// Removed ModeVideoWallpapersSubmenu from navigation
			if m.menuIndex < len(m.menuOptions)-1 {
				m.menuIndex++
			}
			return m, nil
		} else if m.focusState == FocusSession {
			// Navigate sessions when session selector is focused
			if m.sessionIndex < len(m.sessions)-1 {
				m.sessionIndex++
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
			}
			return m, nil
		}

	// Add Page Up/Down handlers for ASCII variant cycling
	case "pgup", "page up":
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.selectedSession != nil {
				// Load config to get variant count
				sessionName := strings.ToLower(strings.Fields(m.selectedSession.Name)[0])
				var configFileName string
				switch sessionName {
				case "gnome":
					configFileName = "gnome_desktop"
				case "i3":
					configFileName = "i3wm"
				case "bspwm":
					configFileName = "bspwm_manager"
				case "plasma":
					configFileName = "kde"
				case "xmonad":
					configFileName = "xmonad"
				default:
					configFileName = sessionName
				}

				configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
				if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
					m.asciiArtCount = len(asciiConfig.ASCIIVariants)
					m.asciiMaxHeight = asciiConfig.MaxASCIIHeight

					// Cycle backwards (decrement index with wraparound)
					m.asciiArtIndex--
					if m.asciiArtIndex < 0 {
						m.asciiArtIndex = m.asciiArtCount - 1
					}
				}
			}
			return m, nil
		}

	case "pgdn", "page down":
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.selectedSession != nil {
				// Load config to get variant count
				sessionName := strings.ToLower(strings.Fields(m.selectedSession.Name)[0])
				var configFileName string
				switch sessionName {
				case "gnome":
					configFileName = "gnome_desktop"
				case "i3":
					configFileName = "i3wm"
				case "bspwm":
					configFileName = "bspwm_manager"
				case "plasma":
					configFileName = "kde"
				case "xmonad":
					configFileName = "xmonad"
				default:
					configFileName = sessionName
				}

				configPath := fmt.Sprintf("/usr/share/sysc-greet/ascii_configs/%s.conf", configFileName)
				if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
					m.asciiArtCount = len(asciiConfig.ASCIIVariants)
					m.asciiMaxHeight = asciiConfig.MaxASCIIHeight

					// Cycle forwards (increment index with wraparound)
					m.asciiArtIndex++
					if m.asciiArtIndex >= m.asciiArtCount {
						m.asciiArtIndex = 0
					}
				}
			}
			return m, nil
		}

	case "enter":
		if m.sessionDropdownOpen {
			// Select current session from dropdown
			session := m.sessions[m.sessionIndex]
			m.sessionDropdownOpen = false
			return m, func() tea.Msg { return sessionSelectedMsg(session) }
		}

		// Add submenu selection handling
		if m.mode == ModeMenu {
			selectedOption := m.menuOptions[m.menuIndex]
			switch selectedOption {
			case "Close Menu":
				m.mode = ModeLogin
				return m, nil
			case "Themes":
				newModel, cmd := m.navigateToThemesSubmenu()
				return newModel.(model), cmd
			case "Borders":
				newModel, cmd := m.navigateToBordersSubmenu()
				return newModel.(model), cmd
			case "Backgrounds":
				newModel, cmd := m.navigateToBackgroundsSubmenu()
				return newModel.(model), cmd
			case "Wallpaper":
				newModel, cmd := m.navigateToWallpaperSubmenu()
				return newModel.(model), cmd
			}
			return m, nil
		}

		// Handle submenu selections
		if m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu {
			selectedOption := m.menuOptions[m.menuIndex]

			// Handle "← Back" option for all submenus
			if selectedOption == "← Back" {
				m.mode = ModeMenu
				m.menuOptions = []string{
					"Close Menu",
					"Themes",
					"Borders",
					"Backgrounds",
					"Wallpaper",
				}
				m.menuIndex = 0
				return m, nil
			}

			// Implement actual submenu functionality
			switch m.mode {
			case ModeThemesSubmenu:
				// Parse theme selection and apply it
				if strings.HasPrefix(selectedOption, "Theme: ") {
					themeName := strings.TrimPrefix(selectedOption, "Theme: ")
					m.currentTheme = themeName
					// Apply theme immediately
					applyTheme(themeName, m.config.TestMode)
					// CHANGED 2025-10-03 - Save theme preference
					// CHANGED 2025-10-03 - Skip saving in test mode
					if !m.config.TestMode {
						sessionName := ""
						if m.selectedSession != nil {
							sessionName = m.selectedSession.Name
						}
						cache.SavePreferences(cache.UserPreferences{
							Theme:       m.currentTheme,
							Background:  m.selectedBackground,
							BorderStyle: m.selectedBorderStyle,
							Session:     sessionName,
						})
					}
				}
				m.mode = ModeLogin
				return m, nil

			case ModeBordersSubmenu:
				// Restored ASCII border handling
				switch selectedOption {
				case "Style: Classic":
					m.selectedBorderStyle = "classic"
				case "Style: Modern":
					m.selectedBorderStyle = "modern"
				case "Style: Minimal":
					m.selectedBorderStyle = "minimal"
				case "Style: ASCII-1":
					m.selectedBorderStyle = "ascii1"
				case "Style: ASCII-2":
					m.selectedBorderStyle = "ascii2"
				case "Style: ASCII-3":
					m.selectedBorderStyle = "ascii3"
				case "Style: ASCII-4":
					m.selectedBorderStyle = "ascii4"
				case "Animation: Wave":
					m.borderAnimationEnabled = true
					m.selectedBorderStyle = "wave"
				case "Animation: Pulse":
					m.borderAnimationEnabled = true
					m.selectedBorderStyle = "pulse"
				case "Animation: Off":
					m.borderAnimationEnabled = false
				}
				// CHANGED 2025-10-03 - Save border preference
				// CHANGED 2025-10-03 - Skip saving in test mode
				if !m.config.TestMode {
					sessionName := ""
					if m.selectedSession != nil {
						sessionName = m.selectedSession.Name
					}
					cache.SavePreferences(cache.UserPreferences{
						Theme:       m.currentTheme,
						Background:  m.selectedBackground,
						BorderStyle: m.selectedBorderStyle,
						Session:     sessionName,
					})
				}
				m.mode = ModeLogin
				return m, nil

			case ModeBackgroundsSubmenu:
				// CHANGED 2025-10-04 - Toggle backgrounds instead of replacing
				// Strip checkbox prefix to get actual option name
				optionName := strings.TrimPrefix(selectedOption, "[✓] ")
				optionName = strings.TrimPrefix(optionName, "[ ] ")

				switch optionName {
				case "Fire":
					m.enableFire = !m.enableFire
				case "ASCII Rain": // CHANGED 2025-10-08 - Add ascii-rain option
					// Rain is exclusive - disable others
					m.enableFire = false
					if m.selectedBackground != "ascii-rain" {
						m.selectedBackground = "ascii-rain"
					} else {
						m.selectedBackground = "none"
					}
				case "Matrix": // Add matrix option
					// Matrix is exclusive - disable others
					m.enableFire = false
					if m.selectedBackground != "matrix" {
						m.selectedBackground = "matrix"
					} else {
						m.selectedBackground = "none"
					}
				case "Fireworks": // Add fireworks option
					// Fireworks is exclusive - disable others
					m.enableFire = false
					if m.selectedBackground != "fireworks" {
						m.selectedBackground = "fireworks"
					} else {
						m.selectedBackground = "none"
					}
				}

				// Update selectedBackground based on enabled flags
				// Priority: Fire > Matrix > ASCII Rain > Fireworks > none
				if m.enableFire {
					m.selectedBackground = "fire"
				} else if m.selectedBackground != "pattern" && m.selectedBackground != "ascii-rain" && m.selectedBackground != "matrix" && m.selectedBackground != "fireworks" {
					m.selectedBackground = "none"
				}
				// CHANGED 2025-10-03 - Save background preference
				// CHANGED 2025-10-03 - Skip saving in test mode
				if !m.config.TestMode {
					sessionName := ""
					if m.selectedSession != nil {
						sessionName = m.selectedSession.Name
					}
					cache.SavePreferences(cache.UserPreferences{
						Theme:       m.currentTheme,
						Background:  m.selectedBackground,
						BorderStyle: m.selectedBorderStyle,
						Session:     sessionName,
					})
				}
				// CHANGED 2025-10-06 - Refresh menu to update checkboxes
				newModel, cmd := m.navigateToBackgroundsSubmenu()
				return newModel.(model), cmd
			case ModeWallpaperSubmenu:
				// Use modular wallpaper handler
				newModel, cmd := m.handleWallpaperSelection(selectedOption)
				return newModel.(model), cmd
			}
			return m, nil
		}

		switch m.mode {
		case ModeLogin:
			if m.focusState == FocusSession {
				// Enter from session focus goes to username
				m.focusState = FocusUsername
				m.usernameInput.Focus()
				return m, textinput.Blink
			} else {
				// Enter from username goes to password
				if m.config.Debug {
					fmt.Println("Debug: Switching to password mode")
				}
				m.mode = ModePassword
				m.focusState = FocusPassword
				m.usernameInput.Blur()
				m.passwordInput.Focus()
				return m, textinput.Blink
			}

		case ModePassword:
			if m.focusState == FocusSession {
				// Enter from session focus goes to password input
				m.focusState = FocusPassword
				m.passwordInput.Focus()
				return m, textinput.Blink
			} else {
				// Enter from password submits
				username := m.usernameInput.Value()
				password := m.passwordInput.Value()
				if m.config.Debug {
					// SECURITY: Never log passwords - only log username for debugging
					logDebug(" Authentication attempt for user: %s", username)
				}
				if m.config.TestMode {
					fmt.Println("Test mode: Auth successful")
					return m, tea.Quit
				} else {
					if m.ipcClient == nil {
						fmt.Println("Error: No IPC client available")
						return m, tea.Quit
					}
					m.mode = ModeLoading
					return m, m.authenticate(username, password)
				}
			}

		case ModePower:
			if m.powerIndex < len(m.powerOptions) {
				option := m.powerOptions[m.powerIndex]
				return m, func() tea.Msg { return powerSelectedMsg(option) }
			}
		}
	}

	return m, nil
}

// Return tea.View with BackgroundColor set
func (m model) View() tea.View {
	// Use full terminal dimensions
	termWidth := m.width
	termHeight := m.height
	if termWidth == 0 {
		termWidth = 80
	}
	if termHeight == 0 {
		termHeight = 24
	}

	var content string
	switch m.mode {
	case ModePower:
		// Fixed missing power menu rendering
		content = m.renderPowerView(termWidth, termHeight)
	case ModeMenu, ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeWallpaperSubmenu:
		// Removed ModeVideoWallpapersSubmenu from rendering
		content = m.renderMenuView(termWidth, termHeight)
	case ModeReleaseNotes:
		// Added F5 release notes view rendering
		content = m.renderReleaseNotesView(termWidth, termHeight)
	case ModeScreensaver:
		// CHANGED 2025-10-10 - Added screensaver rendering
		content = renderScreensaverView(m, termWidth, termHeight)
	default:
		content = m.renderMainView(termWidth, termHeight)
	}

	var view tea.View

	// Check if fire background is enabled
	// CHANGED 2025-10-06 - Removed wallpaper check
	// CHANGED 2025-10-06 - Only show fire on main login screen, not in menus
	// CHANGED 2025-10-08 - Add ascii-rain background support
	// CHANGED 2025-10-18 22:00 - Enable background animations in password mode (username caching means most users see password mode)
	hasFireBackground := (m.enableFire || m.selectedBackground == "fire" || m.selectedBackground == "fire+rain") && (m.mode == ModeLogin || m.mode == ModePassword)
	hasRainBackground := (m.selectedBackground == "ascii-rain") && (m.mode == ModeLogin || m.mode == ModePassword)

	if hasFireBackground {
		// CHANGED 2025-10-06 - Use multi-layer approach: fire at bottom, centered UI on top
		fireHeight := (termHeight * 2) / 5 // Bottom 40% of terminal
		fireY := termHeight - fireHeight

		// Render fire
		backgroundContent := m.addFireEffect("", termWidth, fireHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: fire at bottom, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(fireY),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	} else if hasRainBackground {
		// CHANGED 2025-10-08 - Render ascii-rain as full background
		// Render rain as full background
		backgroundContent := m.addAsciiRain("", termWidth, termHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: rain as background, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(0),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	} else if m.selectedBackground == "matrix" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render matrix as full background
		backgroundContent := m.addMatrixEffect("", termWidth, termHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: matrix as background, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(0),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	} else if m.selectedBackground == "fireworks" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render fireworks as full background
		backgroundContent := m.addFireworksEffect("", termWidth, termHeight)

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: fireworks as background, UI centered on top
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(backgroundContent).X(0).Y(0),
			lipgloss.NewLayer(content).X(uiX).Y(uiY),
		)
		view.BackgroundColor = BgBase
		return view
	}

	// CHANGED 2025-10-06 - Use X/Y positioning instead of Place() to avoid ghosting
	// Calculate center position manually (CRUSH approach)
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)
	x := (termWidth - contentWidth) / 2
	y := (termHeight - contentHeight) / 2

	// Removed ticker fullscreen check
	// Use layer X/Y positioning instead of Place()
	view.Layer = lipgloss.NewCanvas(lipgloss.NewLayer(content).X(x).Y(y))
	view.BackgroundColor = BgBase
	return view
}

// CHANGED 2025-10-06 - Ensure content fills entire terminal to prevent ghosting
// Problem: In fullscreen kitty, ANSI codes mess with Bubble Tea's diff renderer cell counting
func ensureFullTerminalCoverage(content string, termWidth, termHeight int) string {
	lines := strings.Split(content, "\n")

	// Pad lines to exact terminal width with plain spaces (no ANSI styling)
	// CRITICAL: Use PLAIN spaces, not lipgloss.Render(), to avoid ANSI code length variations
	for i := range lines {
		// Strip ANSI to get actual visible width
		visibleWidth := len([]rune(stripAnsi(lines[i])))
		if visibleWidth < termWidth {
			// Use plain spaces - Bubble Tea renderer will fill with background color
			lines[i] += strings.Repeat(" ", termWidth-visibleWidth)
		}
	}

	// Create full-width empty line with plain spaces
	emptyLine := strings.Repeat(" ", termWidth)

	// Ensure we have exactly termHeight lines
	for len(lines) < termHeight {
		lines = append(lines, emptyLine)
	}

	// Trim to exactly termHeight lines
	if len(lines) > termHeight {
		lines = lines[:termHeight]
	}

	return strings.Join(lines, "\n")
}

// CHANGED 2025-10-01 - Replaced WM-named themes with common themes
// Moved navigation functions to menu.go and wallpaper.go

// Complete dual border redesign
func (m model) renderMainView(termWidth, termHeight int) string {
	return m.renderDualBorderLayout(termWidth, termHeight)
}

// Border rendering functions moved to borders.go
// Includes: renderDualBorderLayout, renderASCII1/2/3/4BorderLayout, renderASCIIBorderFallback,
// getInnerBorderStyle, getOuterBorderStyle, getInnerBorderColor, getOuterBorderColor

// Render form with monochrome colors

// Implement actual border style functionality

// Get inner border style based on user selection

// Background effect functions moved to backgrounds.go
// Includes: applyBackgroundAnimation, addMatrixRain, addFireEffect, addAsciiRain, addMatrixEffect, getBackgroundColor

// UI component functions moved to ui_components.go
// Includes: renderMonochromeForm, renderMainForm, renderSessionSelector, renderSessionDropdown, renderMainHelp

// Animation helper functions moved to theme.go

func (m model) authenticate(username, password string) tea.Cmd {
	return func() tea.Msg {
		// CHANGED 2025-10-05 - Add nil check for IPC client
		if m.ipcClient == nil {
			return fmt.Errorf("IPC client not initialized - greeter must be run by greetd")
		}

		// Create session
		if err := m.ipcClient.CreateSession(username); err != nil {
			return err
		}
		// Receive auth message
		resp, err := m.ipcClient.ReceiveResponse()
		if err != nil {
			return err
		}

		// CHANGED 2025-10-05 - Handle Error response from CreateSession
		if errResp, ok := resp.(ipc.Error); ok {
			return fmt.Errorf("authentication failed: %s - %s", errResp.ErrorType, errResp.Description)
		}

		if _, ok := resp.(ipc.AuthMessage); ok {
			if m.config.Debug {
				logDebug(" Received auth message")
			}
			// Send password as response
			if err := m.ipcClient.PostAuthMessageResponse(&password); err != nil {
				return err
			}
			// Receive success or error
			resp, err := m.ipcClient.ReceiveResponse()
			if err != nil {
				return err
			}

			// CHANGED 2025-10-05 - Handle Error response (wrong password)
			if errResp, ok := resp.(ipc.Error); ok {
				return fmt.Errorf("authentication failed: %s - %s", errResp.ErrorType, errResp.Description)
			}

			if _, ok := resp.(ipc.Success); ok {
				// Start session
				if m.selectedSession == nil {
					return fmt.Errorf("no session selected")
				}
				cmd := []string{m.selectedSession.Exec}
				env := []string{} // Can be populated if needed
				if err := m.ipcClient.StartSession(cmd, env); err != nil {
					return err
				}
				return "success"
			} else {
				return fmt.Errorf("expected success or error, got %T", resp)
			}
		} else {
			return fmt.Errorf("expected auth message or error, got %T", resp)
		}
	}
}

// Utility helper functions (min, stripAnsi, extractCharsWithAnsi, etc.) moved to utils.go

func main() {
	// CHANGED 2025-10-01 - Removed SetColorProfile - not available in lipgloss v2
	// Color profile is now automatically detected via colorprofile package
	// CHANGED 2025-10-14 - Removed sysc-greet.conf loading - hardcoded sessionPalettes provide all needed palettes

	// Initialize config with defaults
	config := Config{}

	var screensaverTestMode bool // CHANGED 2025-10-11 - Add screensaver test mode flag
	var showVersion bool

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.BoolVar(&config.TestMode, "test", false, "Enable test mode (no actual authentication)")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug output")
	flag.BoolVar(&screensaverTestMode, "screensaver", false, "Start directly in screensaver mode for testing")
	flag.StringVar(&config.ThemeName, "theme", "", "Theme name (dracula, gruvbox, material, nord, tokyo-night, catppuccin, solarized, monochrome, transishardjob)")
	flag.BoolVar(&config.ShowTime, "time", false, "") // Hidden flag - not shown in help


	// Add help text
	// CHANGED 2025-10-12 - Updated help text to reflect sysc-greet branding
	// CHANGED 2025-10-14 - Removed sysc-greet.conf references
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "sysc-greet - A terminal greeter for greetd\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		// Manually print flags (excluding hidden ones)
		fmt.Fprintf(os.Stderr, "  -debug\n")
		fmt.Fprintf(os.Stderr, "    	Enable debug output\n")
		fmt.Fprintf(os.Stderr, "  -screensaver\n")
		fmt.Fprintf(os.Stderr, "    	Start directly in screensaver mode for testing\n")
		fmt.Fprintf(os.Stderr, "  -test\n")
		fmt.Fprintf(os.Stderr, "    	Enable test mode (no actual authentication)\n")
		fmt.Fprintf(os.Stderr, "  -theme string\n")
		fmt.Fprintf(os.Stderr, "    	Theme name (dracula, gruvbox, material, nord, tokyo-night, catppuccin, solarized, monochrome, transishardjob)\n")
		fmt.Fprintf(os.Stderr, "  -v	Show version information (shorthand)\n")
		fmt.Fprintf(os.Stderr, "  -version\n")
		fmt.Fprintf(os.Stderr, "    	Show version information\n")
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  ASCII configs: /usr/share/sysc-greet/ascii_configs/\n")
		fmt.Fprintf(os.Stderr, "\nKey Bindings:\n")
		fmt.Fprintf(os.Stderr, "  Tab       Cycle focus between elements\n")
		fmt.Fprintf(os.Stderr, "  ↑↓       Navigate sessions when focused\n")
		fmt.Fprintf(os.Stderr, "  F3        Toggle session dropdown\n")
		fmt.Fprintf(os.Stderr, "  F4        Power menu\n")
		fmt.Fprintf(os.Stderr, "  Enter     Continue to next step\n")
		fmt.Fprintf(os.Stderr, "  Esc       Cancel/go back\n")
		fmt.Fprintf(os.Stderr, "  Ctrl+C    Quit\n")
	}

	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("sysc-greet %s\n", Version)
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
		os.Exit(0)
	}

	// SECURITY: Prevent test mode in production environment
	// Test mode bypasses authentication and should only be used for development
	if config.TestMode && os.Getenv("GREETD_SOCK") != "" {
		fmt.Fprintf(os.Stderr, "SECURITY ERROR: Test mode cannot be enabled in production (GREETD_SOCK is set)\n")
		fmt.Fprintf(os.Stderr, "Test mode bypasses authentication and should only be used for development.\n")
		os.Exit(1)
	}

	// CHANGED 2025-10-06 - Initialize debug logging
	initDebugLog()
	logDebug("=== sysc-greet started ===")
	logDebug("Version: sysc-greet greeter")
	logDebug("Test mode: %v", config.TestMode)
	logDebug("Debug mode: %v", config.Debug)
	logDebug("Theme: %s", config.ThemeName)
	logDebug("GREETD_SOCK: %s", os.Getenv("GREETD_SOCK"))
	logDebug("WAYLAND_DISPLAY: %s", os.Getenv("WAYLAND_DISPLAY"))
	logDebug("XDG_RUNTIME_DIR: %s", os.Getenv("XDG_RUNTIME_DIR"))

	if config.Debug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Debug log: /tmp/sysc-greet-debug.log\n")
	}

	// Initialize Bubble Tea program with proper screen management
	// CHANGED 2025-09-29 - Handle TTY access gracefully for different environments
	// CHANGED 2025-10-21 - Enable kitty keyboard protocol for CAPS LOCK detection
	opts := []tea.ProgramOption{}

	// Check if we can access TTY before using alt screen
	if _, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
		// No TTY access - use basic program options
		if config.Debug {
			logDebug(" No TTY access, running without alt screen")
		}
	} else {
		// TTY available - use full screen features
		opts = append(opts, tea.WithAltScreen())
		if !config.TestMode {
			opts = append(opts, tea.WithMouseCellMotion())
		}
	}
	
	p := tea.NewProgram(initialModel(config, screensaverTestMode), opts...)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// ASCII-2, ASCII-3, ASCII-4 border rendering functions moved to borders.go
