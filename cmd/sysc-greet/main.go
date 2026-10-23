package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/mbndr/figlet4go"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/ipc"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	themesOld "github.com/Nomadcxx/sysc-greet/internal/themes"
)

// Version info - set via ldflags during build
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// Data directory for resources (ASCII configs, wallpapers, themes, fonts)
// Can be overridden at build time via: -ldflags="-X 'main.dataDir=/custom/path'"
// Defaults to /usr/share/sysc-greet for standard Linux builds
// NixOS flake injects the actual Nix store path at build time
var dataDir = "/usr/share/sysc-greet"

// CHANGED 2025-10-06 - Add debug logging to file
var debugLog *log.Logger

func initDebugLog() {
	// Try persistent location first ($HOME/.cache/sysc-greet/debug.log)
	// Falls back to /tmp/ if home dir unavailable
	logPaths := []string{"/tmp/sysc-greet-debug.log"}
	if home, err := os.UserHomeDir(); err == nil {
		cacheDir := filepath.Join(home, ".cache", "sysc-greet")
		if err := os.MkdirAll(cacheDir, 0o700); err == nil {
			if info, err := os.Lstat(cacheDir); err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
				if err := os.Chmod(cacheDir, 0o700); err == nil {
					logPaths = append([]string{filepath.Join(cacheDir, "debug.log")}, logPaths...)
				}
			}
		}
	}

	for _, logPath := range logPaths {
		logFile, err := openPrivateLog(logPath)
		if err == nil {
			debugLog = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
			return
		}
	}
}

func openPrivateLog(path string) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CLOEXEC|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, fmt.Errorf("debug log path %q is not a regular file", path)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		file.Close()
		return nil, fmt.Errorf("cannot verify owner of debug log path %q", path)
	}
	if err := validatePrivateLogOwner(path, stat.Uid); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func validatePrivateLogOwner(path string, ownerUID uint32) error {
	if ownerUID != uint32(os.Geteuid()) {
		return fmt.Errorf("debug log path %q is not owned by effective uid", path)
	}
	return nil
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
// Includes: applyTheme, setThemeWallpaper, getAnimatedColor, getFocusColor

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
	Color              string   // Optional hex color override for ASCII art (e.g., "#89b4fa")
	AnimationStyle     string   // "gradient", "wave", "pulse", "rainbow", "matrix", "typewriter", "glow", "static"
	AnimationSpeed     float64  // 0.1 (slow) to 2.0 (fast), default 1.0
	AnimationDirection string   // "left", "right", "up", "down", "center-out", "random"
	Roasts             string   // Custom roast messages separated by │
	ANSIStyle          bool     // Interpret embedded SGR colors for explicitly opted-in art
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
	FollowShell      *bool
	ShellThemeFile   string
	TestMode         bool
	TestSession      string
	SecondarySaver   bool
	Debug            bool
	ShowTime         bool
	ThemeName        string
	RememberUsername bool
	BorderLayout     string
	FormMode         string
	FormModeExplicit bool
	FormHalo         int
	FormDim          float64
	Ambient          ambientConfig
	Secondary        secondaryConfig
}

func shouldCreateGreetdIPC(config Config) bool {
	return !config.TestMode && !config.SecondarySaver
}

func shouldLoadUserPreferences(config Config) bool {
	return !config.TestMode && !config.SecondarySaver
}

type ViewMode string

const (
	ModeLogin    ViewMode = "login"
	ModePassword ViewMode = "password"
	ModeLoading  ViewMode = "loading"
	ModePower    ViewMode = "power"
	ModeMenu     ViewMode = "menu"
	// Added new menu modes for structured menu system
	ModeThemesSubmenu       ViewMode = "themes_submenu"
	ModeBordersSubmenu      ViewMode = "borders_submenu"
	ModeBackgroundsSubmenu  ViewMode = "backgrounds_submenu"
	ModeWallpaperSubmenu    ViewMode = "wallpaper_submenu"
	ModeASCIIEffectsSubmenu ViewMode = "ascii_effects_submenu"
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
	frameCache         *frameCache
	ambient            *ambientCollector
	followShell        bool
	hideAmbient        bool // Set only on render copies when optional status cannot fit.
	machine            metricsMsg
	weather            weatherMsg
	gpu                gpuMsg
	secondary          secondaryMsg
	secondaryRuntime   *secondaryStateBus
	secondaryReadyFile string
	secondaryReadySent bool
	usernameInput      textinput.Model
	passwordInput      textinput.Model
	spinner            spinner.Model
	sessions           []sessions.Session
	selectedSession    *sessions.Session
	sessionIndex       int
	ipcClient          *ipc.Client
	mode               ViewMode
	config             Config
	startTime          time.Time

	// Terminal dimensions
	width     int
	height    int
	altScreen bool

	// Power menu
	powerOptions []string
	powerIndex   int

	// Session dropdown
	sessionDropdownOpen bool

	// Menu system
	menuOptions []string
	menuIndex   int
	// Added fields for functional menu system
	customASCIIText     string
	selectedBorderStyle string
	borderEffect        string
	borderPaletteMode   string
	borderRegistry      *borderRegistry
	formMode            string
	formLayerCache      *layerCache
	selectedBackground  string
	currentTheme        string
	availableThemes     []string // Built-in + custom theme names
	selectedFont        string
	// CHANGED 2025-10-01 - Added animation control fields
	selectedAnimationStyle     string
	selectedAnimationSpeed     float64
	selectedAnimationDirection string
	animationStyleOptions      []string
	animationDirectionOptions  []string

	// Focus management
	focusState FocusState

	// Authentication tracking
	failedAttempts  int
	borderFailureAt time.Time

	// Animation state
	animationFrame   int
	animSpeed        string // "slow", "normal", "fast"
	bgAnimationFrame int    // frame counter for background effects

	// Background effect instances (resized in WindowSizeMsg, updated in tickMsg)
	fireEffect      *animations.FireEffect
	rainEffect      *animations.RainEffect
	matrixEffect    *animations.MatrixEffect
	fireworksEffect *animations.FireworksEffect

	// CHANGED 2025-10-04 - Separate flags for multiple backgrounds
	enableFire bool

	// CHANGED 2025-10-05 - Add error message for authentication failures
	errorMessage    string
	authNotice      string
	authPromptText  string
	authPromptType  ipc.AuthMessageType
	authPromptReply chan authPromptResult
	sendMessage     func(tea.Msg)

	// CHANGED 2025-10-10 - Screensaver fields
	idleTimer         time.Time               // Time when idle started
	screensaverTime   time.Time               // Current time for screensaver display
	screensaverStart  time.Time               // Shared activation time for mirrored saver variants
	screensaverPrint  *animations.PrintEffect // CHANGED 2025-10-11 - Print effect animation for screensaver
	screensaverActive bool                    // CHANGED 2025-10-11 - Track if screensaver just activated
	ssConfig          ScreensaverConfig       // Cached screensaver.conf; reading it per tick was 33 file opens/s
	ssConfigLoaded    time.Time               // Last cache refresh (refreshed at most every 60s from the tick handler)

	// ASCII navigation fields for multi-variant support
	asciiArtIndex      int         // Current variant index (0-indexed)
	asciiArtCount      int         // Total variants available
	asciiMaxHeight     int         // Max height for normalization
	currentASCIIConfig ASCIIConfig // Cached config for current session

	capsLockOn    bool // CAPS LOCK state detected via kitty keyboard protocol
	kittyUpgraded bool // Whether we've upgraded kitty keyboard flags for non-US layout support
	kittyGraphics bool
	kittyCredit   string
	kittyImages   kittyImagePresenter

	// ASCII Effects
	typewriterTicker    *animations.TypewriterTicker  // Typewriter ticker for session roasts
	printEffect         *animations.PrintEffect       // Print effect for ASCII art
	beamsEffect         *animations.BeamsTextEffect   // Beams text effect for ASCII art
	pourEffect          *animations.PourEffect        // Pour effect for ASCII art
	aquariumEffect      *animations.AquariumEffect    // Aquarium background effect
	sonarEffect         *animations.SonarAnimation    // Sonar background effect
	cracktroEffect      *animations.CracktroAnimation // Cracktro background effect
	plasmaEffect        *animations.PlasmaAnimation   // Plasma background effect
	ak47Effect          *syscgoanimations.AK47Animation
	l2a1SAREffect       *syscgoanimations.L2A1SARAnimation
	rotatingModelEffect *syscgoanimations.RotatingModelAnimation
	selectedWallpaper   string // gslapper video wallpaper (separate from background effect)
	gslapperLaunched    bool   // Track if gslapper was launched from cache
}

type sessionSelectedMsg sessions.Session
type powerSelectedMsg string
type authSuccessMsg struct {
	Username string
	Session  string
}
type tickMsg time.Time

func doTick() tea.Cmd {
	// CHANGED 2025-10-04 - Reduced tick interval to 30ms for smoother ticker animation
	return tea.Tick(time.Millisecond*30, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// doSlowTick keeps the UI tick chain alive at 1s while nothing on screen
// animates: enough for the screensaver idle check and clock. One chain only —
// see doBgTickIdle for why extra chains are forbidden.
func doSlowTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// uiAnimationsActive reports whether anything on screen needs the fast 30ms
// UI tick right now. When false, the tick drops to 1s and the animation
// counters freeze (a frozen banner gradient beats one stepping at 1fps).
func (m model) uiAnimationsActive() bool {
	// Someone is at the keyboard: keep everything smooth
	if time.Since(m.idleTimer) < time.Minute {
		return true
	}
	// Wall-clock driven animations visible on the login screens
	if m.mode == ModeLogin || m.mode == ModePassword {
		switch m.selectedBackground {
		case "ticker", "print", "beams", "pour":
			return true
		}
	}
	// Screensaver print animation in progress
	if m.mode == ModeScreensaver && m.screensaverPrint != nil && !m.screensaverPrint.IsComplete() {
		return true
	}
	// Lazy inits ride this tick (0d1299c); stay fast until they have fired
	if !m.gslapperLaunched && m.selectedWallpaper != "" {
		return true
	}
	switch m.selectedBackground {
	case "aquarium":
		return m.aquariumEffect == nil
	case "sonar":
		return m.sonarEffect == nil
	case "cracktro":
		return m.cracktroEffect == nil
	case "plasma":
		return m.plasmaEffect == nil
	}
	return false
}

type bgTickMsg time.Time

func doBgTick(speed string) tea.Cmd {
	interval := 30 * time.Millisecond
	switch speed {
	case "slow":
		interval = 60 * time.Millisecond
	case "fast":
		interval = 15 * time.Millisecond
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return bgTickMsg(t)
	})
}

// doBgTickIdle keeps the background tick chain alive on a slow heartbeat while
// no effect is animating. One chain only — spawning extra chains on events
// multiplies the frame rate (the "hyperspeed" bug class, issue #45).
func doBgTickIdle() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return bgTickMsg(t)
	})
}

// backgroundEffectActive reports whether the selected background needs
// per-frame bgTick updates right now. Effects only display on the login and
// password screens (see View), so other modes don't pay for animation.
func (m model) backgroundEffectActive() bool {
	if m.mode != ModeLogin && m.mode != ModePassword {
		return false
	}
	if m.enableFire {
		return true
	}
	switch m.selectedBackground {
	case "fire", "fire+rain", "ascii-rain", "matrix", "fireworks", "sonar", "cracktro", "plasma", "aquarium", "ak-47", "l2a1-sar":
		return true
	default:
		return isRotatingModelBackground(m.selectedBackground)
	}
}

// cycleAnimSpeed cycles to the next speed preset
func cycleAnimSpeed(current string) string {
	switch current {
	case "slow":
		return "normal"
	case "normal":
		return "fast"
	case "fast":
		return "slow"
	default:
		return "normal"
	}
}

// applyInputTheme styles the username and password inputs with the current
// theme colors. Call after applyTheme.
func (m *model) applyInputTheme() {
	for _, in := range []*textinput.Model{&m.usernameInput, &m.passwordInput} {
		in.Styles.Focused.Prompt = lipgloss.NewStyle().Foreground(Primary).Bold(true)
		in.Styles.Focused.Text = lipgloss.NewStyle().Foreground(FgPrimary)
		in.Styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(FgMuted).Italic(true)
	}
}

func normalizeTestSessionOption(testMode bool, name string) (string, error) {
	name = strings.TrimSpace(name)
	if !testMode {
		return "", fmt.Errorf("--test-session requires --test")
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("invalid --test-session profile: use a name without path separators")
	}
	return name, nil
}

func selectTestSession(available []sessions.Session, name string) ([]sessions.Session, *sessions.Session, int) {
	for i := range available {
		if strings.EqualFold(available[i].Name, name) {
			return available, &available[i], i
		}
	}

	available = append(available, sessions.Session{Name: name, Exec: name, Type: "Wayland"})
	index := len(available) - 1
	return available, &available[index], index
}

func sessionCommand(execLine string) []string {
	return strings.Fields(execLine)
}

func initialModel(config Config, screensaverMode bool) model {
	// Setup username input with proper styling
	ti := textinput.New()
	ti.Prompt = ""      // Remove prompt, will be added by layout
	ti.Placeholder = "" // Remove placeholder

	// Setup password input
	pi := textinput.New()
	pi.Prompt = ""      // Remove prompt, will be added by layout
	pi.Placeholder = "" // Remove placeholder
	pi.EchoMode = textinput.EchoPassword

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
			logDebug("  - %s (%s)", s.Name, s.Type)
		}
	}

	// Setup animated spinner
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(Primary)

	var ipcClient *ipc.Client
	var selectedSession *sessions.Session
	var sessionIndex int

	if shouldCreateGreetdIPC(config) {
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
	if config.TestMode && config.TestSession != "" {
		sess, selectedSession, sessionIndex = selectTestSession(sess, config.TestSession)
	}

	// Default to first session if none selected
	if selectedSession == nil && len(sess) > 0 {
		selectedSession = &sess[0]
		sessionIndex = 0
	}

	// Scan for custom themes
	themeDirs := []string{
		dataDir + "/themes",
		filepath.Join(os.Getenv("HOME"), ".config/sysc-greet/themes"),
	}
	customThemeNames := themesOld.ScanCustomThemes(themeDirs)

	// Combine built-in and custom themes
	availableThemes := themesOld.GetAvailableThemes()
	availableThemes = append(availableThemes, customThemeNames...)

	// Set initial focus
	ti.Focus()

	// CHANGED 2025-10-11 - Determine initial mode
	initialMode := ModeLogin
	if screensaverMode {
		initialMode = ModeScreensaver
	}

	m := model{
		frameCache:          &frameCache{},
		usernameInput:       ti,
		passwordInput:       pi,
		spinner:             sp,
		sessions:            sess,
		selectedSession:     selectedSession,
		sessionIndex:        sessionIndex,
		ipcClient:           ipcClient,
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
		animSpeed:           "normal",
		bgAnimationFrame:    0,
		// Initialize default border and background settings
		// Set Dracula as default theme and disable border animation
		selectedBorderStyle: defaultBorderStyle,
		borderEffect:        "off",
		borderPaletteMode:   "theme",
		borderRegistry:      newBorderRegistry(dataDir + "/borders"),
		formMode:            config.FormMode,
		formLayerCache:      &layerCache{},
		selectedBackground:  "none",
		currentTheme:        startupThemeName(config.ThemeName, "", availableThemes),
		followShell:         config.FollowShell == nil || *config.FollowShell,
		availableThemes:     availableThemes,
		selectedFont:        "/usr/share/bubble-greet/fonts/dos_rebel.flf", // Absolute path
		customASCIIText:     "",
		// CHANGED 2025-10-01 - Initialize animation control defaults
		selectedAnimationStyle:     "gradient",
		selectedAnimationSpeed:     1.0,
		selectedAnimationDirection: "right",
		animationStyleOptions:      []string{"gradient", "wave", "pulse", "rainbow", "matrix", "typewriter", "glow", "static"},
		animationDirectionOptions:  []string{"right", "left", "up", "down", "center-out"},
		// CHANGED 2025-10-10 - Initialize screensaver timers
		idleTimer:       time.Now(),
		screensaverTime: time.Now(),
		ssConfig:        loadScreensaverConfig(),
		ssConfigLoaded:  time.Now(),
		// Initialize fire effect with default size
		fireEffect: animations.NewFireEffect(80, 30, animations.GetDefaultFirePalette()),
		// CHANGED 2025-10-08 - Initialize rain effect with default size
		rainEffect: animations.NewRainEffect(80, 30, animations.GetRainPalette("default")),
		// Initialize matrix effect with default size
		matrixEffect: animations.NewMatrixEffect(80, 30, animations.GetMatrixPalette("default")),
		// Initialize fireworks effect with default size
		fireworksEffect: animations.NewFireworksEffect(80, 30, animations.GetFireworksPalette("default")),
		// Aquarium is nil by default, initialized when user enables it
		aquariumEffect: nil,
		// Sonar is nil by default, initialized when user enables it
		sonarEffect: nil,
		// Cracktro is nil by default, initialized when user enables it
		cracktroEffect: nil,
		// Plasma is nil by default, initialized when user enables it
		plasmaEffect: nil,
		// TypewriterTicker is nil by default, initialized when user enables it
		typewriterTicker: nil,
	}

	m.currentTheme = m.startupTheme("")

	// Test mode skips cached preferences, so SYSC_BG lets a background effect
	// be exercised directly: SYSC_BG=fire sysc-greet --test
	if config.TestMode {
		if bg := os.Getenv("SYSC_BG"); bg != "" {
			m.selectedBackground = bg
		}
	}

	// CHANGED 2025-10-03 - Load cached preferences including session
	// CHANGED 2025-10-03 - Skip cache in test mode
	// Apply the explicit CLI theme before restoring cached effects.
	themeApplied := false
	if shouldLoadUserPreferences(m.config) {
		if prefs, err := cache.LoadPreferences(); err != nil {
			logDebug("failed to load sysc-greet preferences: %v", err)
		} else if prefs != nil {
			if config.FollowShell == nil && prefs.FollowShell != nil {
				m.followShell = *prefs.FollowShell
			}
			m.currentTheme = m.startupTheme(prefs.Theme)
			home, _ := os.UserHomeDir()
			applyThemeWithWallpaper(m.currentTheme, m.config.TestMode, shouldSetCachedThemeWallpaper(prefs, home))
			themeApplied = true
			logDebug("Startup theme: %s", m.currentTheme)
			if prefs.Background != "" {
				m.selectedBackground = prefs.Background
				logDebug("Loaded cached background: %s", prefs.Background)
			}
			if prefs.Wallpaper != "" {
				m.selectedWallpaper = prefs.Wallpaper
				logDebug("Loaded cached wallpaper: %s", prefs.Wallpaper)
			}
			if prefs.BorderStyle != "" || prefs.BorderEffect != "" {
				style, effect := migrateBorderPrefs(prefs.BorderStyle, prefs.BorderEffect)
				m.selectedBorderStyle = preferredBorderStyle(style, m.borderRegistry)
				if effect == "shimmer" || effect == "off" {
					m.borderEffect = effect
				}
			}
			m.formMode = preferredFormMode(m.config.FormMode, prefs.FormMode, m.config.FormModeExplicit)
			if validBorderPalette(prefs.BorderPaletteMode) {
				m.borderPaletteMode = prefs.BorderPaletteMode
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
			// Load cached ASCII variant index (0 is valid - first variant)
			m.asciiArtIndex = prefs.ASCIIIndex
			if prefs.AnimSpeed != "" {
				m.animSpeed = prefs.AnimSpeed
				logDebug("Loaded cached anim speed: %s", prefs.AnimSpeed)
			}

			// Initialize ASCII effect objects based on cached selection
			if m.selectedSession != nil {
				configFileName := asciiConfigFileName(m.selectedSession.Name)
				configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)

				switch m.selectedBackground {
				case "ticker":
					customRoasts := ""
					if asciiConfig, err := loadASCIIConfig(configPath); err == nil {
						customRoasts = asciiConfig.Roasts
					}
					m.typewriterTicker = animations.NewTypewriterTicker(m.selectedSession.Name, customRoasts)
				case "print":
					if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])
						m.printEffect = animations.NewPrintEffect(ascii, time.Millisecond*3)
					}
				case "beams":
					if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])
						beamColors, finalColors := getThemeColorsForBeams(m.currentTheme)
						lines := strings.Split(ascii, "\n")
						asciiHeight := len(lines)
						asciiWidth := 0
						for _, line := range lines {
							if len([]rune(line)) > asciiWidth {
								asciiWidth = len([]rune(line))
							}
						}
						m.beamsEffect = animations.NewBeamsTextEffect(animations.BeamsTextConfig{
							Width:              asciiWidth,
							Height:             asciiHeight,
							Text:               ascii,
							BeamGradientStops:  beamColors,
							FinalGradientStops: finalColors,
						})
					}
				case "pour":
					if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])
						pourColors := getThemeColorsForPour(m.currentTheme)
						lines := strings.Split(ascii, "\n")
						asciiHeight := len(lines)
						asciiWidth := 0
						for _, line := range lines {
							if len([]rune(line)) > asciiWidth {
								asciiWidth = len([]rune(line))
							}
						}
						m.pourEffect = animations.NewPourEffect(animations.PourConfig{
							Width:                  asciiWidth,
							Height:                 asciiHeight,
							Text:                   ascii,
							PourDirection:          "down",
							PourSpeed:              1,
							MovementSpeed:          0.05,
							Gap:                    2,
							StartingColor:          "#ffffff",
							FinalGradientStops:     pourColors,
							FinalGradientSteps:     12,
							FinalGradientFrames:    5,
							FinalGradientDirection: "horizontal",
						})
					}
				case "aquarium":
					// selectedBackground already set on line 589
					// Leave m.aquariumEffect = nil, will initialize in WindowSizeMsg
				default:
					// For gslapper wallpapers, selectedBackground already set on line 589
					// Don't launch yet - wait for compositor in WindowSizeMsg
				}
			}

			// FIXED 2025-10-17 - Load username and auto-advance to password if matches current session
			if m.config.RememberUsername && prefs.Username != "" && m.selectedSession != nil && prefs.Session == m.selectedSession.Name {
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

	// Test mode and an absent cache still apply the resolved startup theme.
	if !themeApplied {
		applyThemeWithWallpaper(m.currentTheme, m.config.TestMode, !m.config.SecondarySaver)
		logDebug("Startup theme: %s", m.currentTheme)
	}

	// Inputs are built before the theme loads; restyle with theme colors
	m.applyInputTheme()
	m.spinner.Style = lipgloss.NewStyle().Foreground(Primary)

	// CHANGED 2025-10-11 - Initialize print effect if starting in screensaver mode
	if screensaverMode {
		m.screensaverStart = time.Now()
		ssConfig := m.ssConfig
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
	if m.config.SecondarySaver {
		return doTick()
	}
	return tea.Batch(
		m.ambient.metricsCmd(0),
		m.ambient.weatherCmd(0),
		m.ambient.gpuCmd(0),
		textinput.Blink,
		m.spinner.Tick,
		doTick(),
		doBgTick(m.animSpeed),
	)
}

func (m model) Update(msg tea.Msg) (result tea.Model, command tea.Cmd) {
	defer func() {
		if current, ok := result.(model); ok {
			current.publishSecondaryRuntime()
		}
	}()
	updated, cmd := m.update(msg)
	next, ok := updated.(model)
	if !ok {
		return updated, cmd
	}
	if imageCmd := syncKittyMeshFrame(m, &next, msg); imageCmd != nil {
		if cmd == nil {
			cmd = imageCmd
		} else {
			cmd = tea.Batch(cmd, imageCmd)
		}
	}
	return next, cmd
}

func (m model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.config.SecondarySaver {
		switch msg.(type) {
		case tea.KeyMsg, tea.MouseMsg:
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case gpuMsg:
		m.gpu = msg
		if m.config.Debug {
			logDebug("Ambient GPU: devices=%d issues=%v error=%v", len(msg.Snapshot.GPUs), msg.Snapshot.Issues, msg.Err)
		}
		return m, m.ambient.gpuCmd(2 * time.Second)
	case secondaryMsg:
		m.secondary = msg
		if m.config.Debug {
			logDebug("Secondary outputs: primary=%s active=%v error=%v", msg.Primary, msg.Outputs, msg.Err)
		}
		return m, nil
	case secondaryWakeMsg:
		if m.mode == ModeScreensaver {
			m = leaveScreensaver(m)
		}
		return m, nil
	case metricsMsg:
		m.machine = msg
		if m.config.Debug {
			logDebug("Ambient metrics: CPU valid=%v fraction=%.4f RAM valid=%v used=%d total=%d CPU error=%v RAM error=%v", msg.CPU.Valid, msg.CPU.Fraction, msg.MemoryValid, msg.Memory.UsedBytes, msg.Memory.TotalBytes, msg.CPUErr, msg.MemoryErr)
		}
		return m, m.ambient.metricsCmd(time.Second)
	case weatherMsg:
		delay := 15 * time.Minute
		if !msg.Valid {
			delay = 30 * time.Second
			if m.weather.Valid {
				m.weather.Stale = true
				m.weather.Err = msg.Err
			} else {
				m.weather = msg
			}
		} else {
			m.weather = msg
		}
		if m.config.Debug {
			logDebug("Ambient weather: valid=%v stale=%v temperature=%.1f code=%d error=%v", m.weather.Valid, m.weather.Stale, m.weather.Temperature, m.weather.Code, m.weather.Err)
		}
		return m, m.ambient.weatherCmd(delay)
	case tea.KeyboardEnhancementsMsg:
		// Upgrade kitty keyboard protocol to include alternate key and associated text reporting.
		// Without these flags, non-US keyboard layouts (e.g., German QWERTZ) get wrong characters
		// for shifted keys because kitty sends the base keycode without layout-aware text.
		// Mode 2 = add flags without removing existing ones (preserves CapsLock detection).
		if !m.kittyUpgraded {
			m.kittyUpgraded = true
			flags := ansi.KittyReportAlternateKeys | ansi.KittyReportAssociatedKeys
			logDebug("Upgrading kitty keyboard flags: adding %d (ReportAlternateKeys|ReportAssociatedKeys)", flags)
			return m, tea.Raw(ansi.KittyKeyboard(flags, 2))
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.config.SecondarySaver && !m.secondaryReadySent && m.secondaryReadyFile != "" {
			if err := writeSecondarySaverReady(m.secondaryReadyFile); err != nil {
				logDebug("secondary saver readiness: %v", err)
			} else {
				m.secondaryReadySent = true
			}
		}
		logDebug("Terminal resized: %dx%d", msg.Width, msg.Height)

		// Resize all active background effects to match new terminal dimensions
		fireHeight := (msg.Height * 2) / 5
		if m.fireEffect != nil {
			m.fireEffect.Resize(msg.Width, fireHeight)
		}
		if m.rainEffect != nil {
			m.rainEffect.Resize(msg.Width, msg.Height)
		}
		if m.matrixEffect != nil {
			m.matrixEffect.Resize(msg.Width, msg.Height)
		}
		if m.fireworksEffect != nil {
			m.fireworksEffect.Resize(msg.Width, msg.Height)
		}
		if m.sonarEffect != nil {
			m.sonarEffect.Resize(msg.Width, msg.Height)
		}
		if m.cracktroEffect != nil {
			m.cracktroEffect.Resize(msg.Width, msg.Height)
		}
		if m.plasmaEffect != nil {
			m.plasmaEffect.Resize(msg.Width, msg.Height)
		}
		if m.aquariumEffect != nil {
			m.aquariumEffect.Resize(msg.Width, msg.Height)
		}
		if m.ak47Effect != nil {
			m.ak47Effect.Resize(msg.Width, msg.Height)
			applyTerminalCellAspect(m.ak47Effect)
		}
		if m.l2a1SAREffect != nil {
			m.l2a1SAREffect.Resize(msg.Width, msg.Height)
			applyTerminalCellAspect(m.l2a1SAREffect)
		}
		if m.rotatingModelEffect != nil {
			m.rotatingModelEffect.Resize(msg.Width, msg.Height)
			applyTerminalCellAspect(m.rotatingModelEffect)
		}
		return m, nil

	case tickMsg:
		fastTick := m.uiAnimationsActive()
		if fastTick {
			// Counters only advance on fast ticks so idle frames are
			// byte-identical and the renderer diffs them to nothing
			m.animationFrame++
		}

		// Lazy init: create aquarium on first tick when we have real dimensions
		if m.selectedBackground == "aquarium" && m.aquariumEffect == nil && m.width > 0 && m.height > 0 {
			fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor, anchorColor := getThemeColorsForAquarium(m.currentTheme)
			m.aquariumEffect = animations.NewAquariumEffect(animations.AquariumConfig{
				Width:         m.width,
				Height:        m.height,
				FishColors:    fishColors,
				WaterColors:   waterColors,
				SeaweedColors: seaweedColors,
				BubbleColor:   bubbleColor,
				DiverColor:    diverColor,
				BoatColor:     boatColor,
				MermaidColor:  mermaidColor,
				AnchorColor:   anchorColor,
			})
			logDebug("Lazy init aquarium in tick: %dx%d", m.width, m.height)
		}

		// Lazy init: create sonar on first tick when we have real dimensions
		if m.selectedBackground == "sonar" && m.sonarEffect == nil && m.width > 0 && m.height > 0 {
			m.sonarEffect = animations.NewSonarEffect(m.width, m.height, animations.GetSonarPalette(m.currentTheme), m.currentTheme)
			logDebug("Lazy init sonar in tick: %dx%d", m.width, m.height)
		}

		// Lazy init: create cracktro on first tick when we have real dimensions
		if m.selectedBackground == "cracktro" && m.cracktroEffect == nil && m.width > 0 && m.height > 0 {
			m.cracktroEffect = animations.NewCracktroEffect(m.width, m.height, animations.GetCracktroPalette(m.currentTheme), m.currentTheme)
			logDebug("Lazy init cracktro in tick: %dx%d", m.width, m.height)
		}

		// Lazy init: create plasma on first tick when we have real dimensions
		if m.selectedBackground == "plasma" && m.plasmaEffect == nil && m.width > 0 && m.height > 0 {
			m.plasmaEffect = animations.NewPlasmaEffect(m.width, m.height, animations.GetPlasmaPalette(m.currentTheme), m.currentTheme)
			logDebug("Lazy init plasma in tick: %dx%d", m.width, m.height)
		}

		// Lazy init: launch gslapper on first tick when compositor is ready
		if !m.gslapperLaunched && m.width > 0 && m.selectedWallpaper != "" {
			launchGslapperWallpaper(m.selectedWallpaper)
			m.gslapperLaunched = true
			logDebug("Lazy init gslapper in tick: %s", m.selectedWallpaper)
		}

		// CHANGED 2025-10-10 - Update screensaver time and check for activation
		m.screensaverTime = time.Time(msg)

		// CHANGED 2025-10-11 - Tick print effect animation if in screensaver mode
		if m.mode == ModeScreensaver && m.screensaverPrint != nil {
			m.screensaverPrint.Tick(m.screensaverTime)
		}

		// Refresh the cached screensaver config at most once a minute so edits
		// to screensaver.conf still apply without restarting the greeter
		if time.Since(m.ssConfigLoaded) > time.Minute {
			m.ssConfig = loadScreensaverConfig()
			m.ssConfigLoaded = time.Now()
		}

		// Check for screensaver activation using configurable timeout
		if (m.mode == ModeLogin || m.mode == ModePassword) && m.authPromptReply == nil {
			ssConfig := m.ssConfig
			idleDuration := time.Since(m.idleTimer)
			if idleDuration >= time.Duration(ssConfig.IdleTimeout)*time.Minute && m.mode != ModeScreensaver {
				m.mode = ModeScreensaver
				m.screensaverStart = m.screensaverTime
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

		if m.selectedBackground == "print" && m.printEffect != nil {
			m.printEffect.Tick(m.screensaverTime)
		}

		if m.selectedBackground == "beams" && m.beamsEffect != nil {
			m.beamsEffect.Update()
		}

		if m.selectedBackground == "pour" && m.pourEffect != nil {
			m.pourEffect.Update()
		}

		if fastTick {
			cmds = append(cmds, doTick())
		} else {
			cmds = append(cmds, doSlowTick())
		}

	case bgTickMsg:
		if !m.backgroundEffectActive() {
			cmds = append(cmds, doBgTickIdle())
			return m, tea.Batch(cmds...)
		}

		m.bgAnimationFrame++
		if m.selectedBackground == "ak-47" && m.ak47Effect == nil {
			m.ak47Effect = m.newAK47Effect()
		}
		if m.selectedBackground == "l2a1-sar" && m.l2a1SAREffect == nil {
			m.l2a1SAREffect = m.newL2A1SAREffect()
		}
		if isRotatingModelBackground(m.selectedBackground) && m.rotatingModelEffect == nil {
			m.rotatingModelEffect = m.newRotatingModelEffect(m.selectedBackground)
		}

		if (m.enableFire || m.selectedBackground == "fire" || m.selectedBackground == "fire+rain") && m.fireEffect != nil {
			m.fireEffect.UpdatePalette(animations.GetFirePalette(m.currentTheme))
			m.fireEffect.Update(m.bgAnimationFrame)
		}

		if m.selectedBackground == "ascii-rain" && m.rainEffect != nil {
			m.rainEffect.UpdatePalette(animations.GetRainPalette(m.currentTheme))
			m.rainEffect.Update(m.bgAnimationFrame)
		}

		if m.selectedBackground == "matrix" && m.matrixEffect != nil {
			m.matrixEffect.UpdatePalette(animations.GetMatrixPalette(m.currentTheme))
			m.matrixEffect.Update(m.bgAnimationFrame)
		}

		if m.selectedBackground == "fireworks" && m.fireworksEffect != nil {
			m.fireworksEffect.UpdatePalette(animations.GetFireworksPalette(m.currentTheme))
			m.fireworksEffect.Update(m.bgAnimationFrame)
		}

		if m.selectedBackground == "sonar" && m.sonarEffect != nil {
			m.sonarEffect.UpdatePalette(animations.GetSonarPalette(m.currentTheme))
			m.sonarEffect.Update()
		}

		if m.selectedBackground == "cracktro" && m.cracktroEffect != nil {
			m.cracktroEffect.UpdatePalette(animations.GetCracktroPalette(m.currentTheme))
			m.cracktroEffect.Update()
		}

		if m.selectedBackground == "plasma" && m.plasmaEffect != nil {
			m.plasmaEffect.UpdatePalette(animations.GetPlasmaPalette(m.currentTheme))
			m.plasmaEffect.Update()
		}

		if m.selectedBackground == "aquarium" && m.aquariumEffect != nil {
			fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor, _ := getThemeColorsForAquarium(m.currentTheme)
			m.aquariumEffect.UpdatePalette(fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor)
			m.aquariumEffect.Update()
		}

		if m.selectedBackground == "ak-47" && m.ak47Effect != nil {
			m.ak47Effect.Update()
		}
		if m.selectedBackground == "l2a1-sar" && m.l2a1SAREffect != nil {
			m.l2a1SAREffect.Update()
		}
		if isRotatingModelBackground(m.selectedBackground) && m.rotatingModelEffect != nil {
			m.rotatingModelEffect.Update()
		}

		cmds = append(cmds, doBgTick(m.animSpeed))

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

		// Update typewriter ticker for new session
		if m.typewriterTicker != nil {
			if m.config.Debug {
				logDebug("Updating ticker for session: %s", session.Name)
			}
			m.typewriterTicker.UpdateWM(session.Name, getCustomRoastsForSession(session.Name))
		}

		// Update ASCII effects for new session
		m.resetPrintEffectForSession(session.Name)
		m.resetBeamsEffectForSession(session.Name)
		m.resetPourEffectForSession(session.Name)

		// FIXED 2025-10-17 - Clear and reload username when session changes
		if previousSession != "" && previousSession != session.Name {
			logDebug("Session changed from '%s' to '%s', clearing username", previousSession, session.Name)
			m.usernameInput.SetValue("") // Clear current username

			// Load cached username for NEW session if available
			if !m.config.TestMode {
				if prefs, err := cache.LoadPreferences(); err != nil {
					logDebug("failed to load sysc-greet preferences: %v", err)
				} else if prefs != nil {
					if m.config.RememberUsername && prefs.Username != "" && prefs.Session == session.Name {
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
			m.savePreferences()
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
			// FIXED 2026-03-23 - Don't quit greeter after issuing reboot
			// systemd 260+ kills all processes in session scope on greeter exit,
			// which races with the reboot command. Stay alive and let reboot kill us.
			exec.Command("systemctl", "reboot").Start()
			m.mode = ModeLoading
			return m, nil
		case "Shutdown":
			if m.config.TestMode {
				fmt.Println("Test mode: Would shutdown system")
				return m, tea.Quit
			}
			// FIXED 2026-03-23 - Don't quit greeter after issuing shutdown
			// systemd 260+ kills all processes in session scope on greeter exit,
			// which races with the poweroff command. Stay alive and let shutdown kill us.
			exec.Command("systemctl", "poweroff").Start()
			m.mode = ModeLoading
			return m, nil
		case "Cancel":
			m.mode = ModeLogin
			m.focusState = FocusUsername
			m.usernameInput.Focus()
			m.passwordInput.Blur()
			cmds = append(cmds, textinput.Blink)
		}

	case authSuccessMsg:
		m.clearAuthPrompt()
		m.failedAttempts = 0
		m.authNotice = ""
		m.saveSuccessfulLogin(msg.Username, msg.Session)
		fmt.Println("Session started successfully")
		return m, tea.Quit

	case authNoticeMsg:
		m.authNotice = string(msg)
		return m, nil

	case authPromptMsg:
		m = m.handleAuthPromptMessage(msg)
		return m, textinput.Blink

	case string:
		// FIXED 2025-10-17 - Return to login mode (not password mode) so user can fix username
		m.errorMessage = msg
		m.borderFailureAt = time.Now()
		m.mode = ModeLogin
		m.usernameInput.SetValue("") // Clear username field
		m.passwordInput.SetValue("") // Clear password field
		m.usernameInput.Focus()
		m.passwordInput.Blur()
		m.focusState = FocusUsername
		return m, textinput.Blink
	case error:
		// FIXED 2025-10-17 - Return to password mode so user can retry
		m.clearAuthPrompt()
		m.authNotice = ""
		m.errorMessage = msg.Error()
		m.borderFailureAt = time.Now()
		if !errors.Is(msg, errAuthPromptCancelled) && !errors.Is(msg, errAuthPromptTimeout) {
			m.failedAttempts++ // Track failed authentication attempts
		}
		m.mode = ModePassword
		// Keep username, only clear password
		m.passwordInput.SetValue("")
		m.passwordInput.Focus()
		m.usernameInput.Blur()
		m.focusState = FocusPassword
		return m, textinput.Blink

	case tea.KeyMsg:
		// CHANGED 2025-10-21 - Detect CAPS LOCK from kitty keyboard protocol
		// Kitty keyboard protocol sends CAPS LOCK and NUM LOCK as ModCapsLock and ModNumLock
		key := msg.Key()
		m.capsLockOn = (key.Mod & tea.ModCapsLock) != 0

		if m.config.Debug {
			// Log ALL key presses to debug what modifiers are being sent
			keyText := key.Text
			if m.passwordInputActive() {
				keyText = "<redacted>"
			}
			logDebug("KEY: %q | Mod=%08b (%d) | CapsLock=%v",
				keyText, key.Mod, key.Mod, m.capsLockOn)
		}
		if m.authPromptReply != nil {
			m, cmd := m.handleKeyInput(msg)
			return m, cmd
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
			m = leaveScreensaver(m)
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
		if m.passwordInputActive() {
			keyStr = "<redacted>"
		}
		logDebug("KEY DEBUG: String='%s'", keyStr)
	}
	if m.authPromptReply != nil {
		return m.handleAuthPromptKey(msg)
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
		// Main menu - works from any mode
		m.sessionDropdownOpen = false
		return m.navigateToMainMenu(), nil

	case "f2":
		// Remapped F2 to Sessions
		// Toggle session dropdown - works from any mode, returns to login
		m.mode = ModeLogin
		m.focusState = FocusUsername
		m.usernameInput.Focus()
		m.passwordInput.Blur()
		m.sessionDropdownOpen = !m.sessionDropdownOpen
		return m, textinput.Blink

	case "f3":
		// Remapped F3 to Notes
		// Release notes popup - works from any mode
		m.sessionDropdownOpen = false
		if m.config.Debug {
			logDebug("Opening release notes")
		}
		m.mode = ModeReleaseNotes
		m.usernameInput.Blur()
		m.passwordInput.Blur()
		return m, nil

	case "f4":
		// F4 remains Power
		// Power menu - works from any mode, resets to first option
		m.sessionDropdownOpen = false
		m.powerIndex = 0
		if m.config.Debug {
			logDebug("Opening power menu")
		}
		m.mode = ModePower
		m.usernameInput.Blur()
		m.passwordInput.Blur()
		return m, nil

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
		case ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeWallpaperSubmenu, ModeASCIIEffectsSubmenu:
			// Go back to main menu
			return m.navigateToMainMenu(), nil
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
				// Update typewriter ticker for new session
				if m.typewriterTicker != nil {
					m.typewriterTicker.UpdateWM(session.Name, getCustomRoastsForSession(session.Name))
				}
				// Update ASCII effects for new variant
				m.resetPrintEffectForSession(session.Name)
				m.resetBeamsEffectForSession(session.Name)
				m.resetPourEffectForSession(session.Name)

			}
			return m, nil
		} else if m.mode == ModePower {
			if m.powerIndex > 0 {
				m.powerIndex--
			}
			return m, nil
		} else if m.mode == ModeMenu || m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu || m.mode == ModeASCIIEffectsSubmenu {
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
				// Update typewriter ticker for new session
				if m.typewriterTicker != nil {
					m.typewriterTicker.UpdateWM(session.Name, getCustomRoastsForSession(session.Name))
				}
				// Update ASCII effects for new session
				m.resetPrintEffectForSession(session.Name)
				m.resetBeamsEffectForSession(session.Name)
				m.resetPourEffectForSession(session.Name)

			}
			return m, nil
		}

	case "down", "j":
		if m.sessionDropdownOpen {
			if m.sessionIndex < len(m.sessions)-1 {
				m.sessionIndex++
				session := m.sessions[m.sessionIndex]
				m.selectedSession = &session
				// Update typewriter ticker for new session
				if m.typewriterTicker != nil {
					m.typewriterTicker.UpdateWM(session.Name, getCustomRoastsForSession(session.Name))
				}
				// Update ASCII effects for new variant
				m.resetPrintEffectForSession(session.Name)
				m.resetBeamsEffectForSession(session.Name)
				m.resetPourEffectForSession(session.Name)

			}
			return m, nil
		} else if m.mode == ModePower {
			if m.powerIndex < len(m.powerOptions)-1 {
				m.powerIndex++
			}
		} else if m.mode == ModeMenu || m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu || m.mode == ModeASCIIEffectsSubmenu {
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
				// Update typewriter ticker for new session
				if m.typewriterTicker != nil {
					m.typewriterTicker.UpdateWM(session.Name, getCustomRoastsForSession(session.Name))
				}
				// Update ASCII effects for new session
				m.resetPrintEffectForSession(session.Name)
				m.resetBeamsEffectForSession(session.Name)
				m.resetPourEffectForSession(session.Name)

			}
			return m, nil
		}

	// Add Page Up/Down handlers for ASCII variant cycling
	case "pgup", "page up":
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.selectedSession != nil {
				// Load config to get variant count
				configFileName := asciiConfigFileName(m.selectedSession.Name)

				configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)
				if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
					m.asciiArtCount = len(asciiConfig.ASCIIVariants)
					m.asciiMaxHeight = asciiConfig.MaxASCIIHeight

					// Cycle backwards (decrement index with wraparound)
					m.asciiArtIndex--
					if m.asciiArtIndex < 0 {
						m.asciiArtIndex = m.asciiArtCount - 1
					}

					// Save ASCII index preference
					if m.selectedSession != nil {
						m.savePreferences()
					}

					// Reset print effect with new ASCII if enabled
					if m.selectedBackground == "print" && m.printEffect != nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])
						m.printEffect.Reset(ascii)
					}

					// Reset beams effect with new ASCII if enabled
					if m.selectedBackground == "beams" && m.beamsEffect != nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])

						lines := strings.Split(ascii, "\n")
						asciiHeight := len(lines)
						asciiWidth := 0
						for _, line := range lines {
							if len([]rune(line)) > asciiWidth {
								asciiWidth = len([]rune(line))
							}
						}

						m.beamsEffect.Resize(asciiWidth, asciiHeight)
						m.beamsEffect.UpdateText(ascii)
					}

					// Reset pour effect with new ASCII if enabled
					if m.selectedBackground == "pour" && m.pourEffect != nil && len(asciiConfig.ASCIIVariants) > 0 {
						session := m.selectedSession
						if session != nil {
							m.resetPourEffectForSession(session.Name)
						}
					}

				}
			}
			return m, nil
		}

	case "pgdn", "pgdown", "page down":
		if m.config.Debug {
			logDebug("Page Down pressed - mode: %v, session: %v", m.mode, m.selectedSession)
		}
		if m.mode == ModeLogin || m.mode == ModePassword {
			if m.selectedSession != nil {
				// Load config to get variant count
				configFileName := asciiConfigFileName(m.selectedSession.Name)

				configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)
				if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
					m.asciiArtCount = len(asciiConfig.ASCIIVariants)
					m.asciiMaxHeight = asciiConfig.MaxASCIIHeight

					// Cycle forwards (increment index with wraparound)
					m.asciiArtIndex++
					if m.asciiArtIndex >= m.asciiArtCount {
						m.asciiArtIndex = 0
					}

					// Save ASCII index preference
					if m.selectedSession != nil {
						m.savePreferences()
					}

					// Reset print effect with new ASCII if enabled
					if m.selectedBackground == "print" && m.printEffect != nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])
						m.printEffect.Reset(ascii)
					}

					// Reset beams effect with new ASCII if enabled
					if m.selectedBackground == "beams" && m.beamsEffect != nil && len(asciiConfig.ASCIIVariants) > 0 {
						variantIndex := m.asciiArtIndex
						if variantIndex >= len(asciiConfig.ASCIIVariants) {
							variantIndex = 0
						}
						ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])

						lines := strings.Split(ascii, "\n")
						asciiHeight := len(lines)
						asciiWidth := 0
						for _, line := range lines {
							if len([]rune(line)) > asciiWidth {
								asciiWidth = len([]rune(line))
							}
						}

						m.beamsEffect.Resize(asciiWidth, asciiHeight)
						m.beamsEffect.UpdateText(ascii)
					}

					// Reset pour effect with new ASCII if enabled
					if m.selectedBackground == "pour" && m.pourEffect != nil && len(asciiConfig.ASCIIVariants) > 0 {
						session := m.selectedSession
						if session != nil {
							m.resetPourEffectForSession(session.Name)
						}
					}

				}
			}
			return m, nil
		}

	case "left":
		if m.mode == ModeBackgroundsSubmenu && m.menuIndex == len(m.menuOptions)-1 {
			switch m.animSpeed {
			case "normal":
				m.animSpeed = "slow"
			case "fast":
				m.animSpeed = "normal"
			case "slow":
				m.animSpeed = "fast"
			default:
				m.animSpeed = "slow"
			}
			m.savePreferences()
			newModel, cmd := m.navigateToBackgroundsSubmenu()
			m2 := newModel.(model)
			m2.menuIndex = len(m2.menuOptions) - 1
			return m2, cmd
		}

	case "right":
		if m.mode == ModeBackgroundsSubmenu && m.menuIndex == len(m.menuOptions)-1 {
			m.animSpeed = cycleAnimSpeed(m.animSpeed)
			m.savePreferences()
			newModel, cmd := m.navigateToBackgroundsSubmenu()
			m2 := newModel.(model)
			m2.menuIndex = len(m2.menuOptions) - 1
			return m2, cmd
		}

	case "space":
		// Space toggles options in backgrounds submenu
		if m.mode == ModeBackgroundsSubmenu {
			selectedOption := m.menuOptions[m.menuIndex]

			// Handle speed selector row
			if strings.HasPrefix(selectedOption, "Speed: ") {
				m.animSpeed = cycleAnimSpeed(m.animSpeed)
				m.savePreferences()
				newModel, cmd := m.navigateToBackgroundsSubmenu()
				m2 := newModel.(model)
				m2.menuIndex = len(m2.menuOptions) - 1
				return m2, cmd
			}

			// Skip "← Back" row
			if selectedOption == "← Back" {
				return m, nil
			}

			// Toggle background effects
			optionName := strings.TrimPrefix(selectedOption, "[✓] ")
			optionName = strings.TrimPrefix(optionName, "[ ] ")

			switch optionName {
			case "Fire":
				m.enableFire = !m.enableFire
			case "ASCII Rain":
				m.enableFire = false
				if m.selectedBackground != "ascii-rain" {
					m.selectedBackground = "ascii-rain"
				} else {
					m.selectedBackground = "none"
				}
			case "Matrix":
				m.enableFire = false
				if m.selectedBackground != "matrix" {
					m.selectedBackground = "matrix"
				} else {
					m.selectedBackground = "none"
				}
			case "Fireworks":
				m.enableFire = false
				if m.selectedBackground != "fireworks" {
					m.selectedBackground = "fireworks"
				} else {
					m.selectedBackground = "none"
				}
			case "Sonar":
				m.enableFire = false
				if m.selectedBackground != "sonar" {
					m.selectedBackground = "sonar"
					width := m.width
					height := m.height
					if width == 0 {
						width = 80
					}
					if height == 0 {
						height = 30
					}
					m.sonarEffect = animations.NewSonarEffect(width, height, animations.GetSonarPalette(m.currentTheme), m.currentTheme)
				} else {
					m.selectedBackground = "none"
					m.sonarEffect = nil
				}
			case "Cracktro":
				m.enableFire = false
				if m.selectedBackground != "cracktro" {
					m.selectedBackground = "cracktro"
					width := m.width
					height := m.height
					if width == 0 {
						width = 80
					}
					if height == 0 {
						height = 30
					}
					m.cracktroEffect = animations.NewCracktroEffect(width, height, animations.GetCracktroPalette(m.currentTheme), m.currentTheme)
				} else {
					m.selectedBackground = "none"
					m.cracktroEffect = nil
				}
			case "Plasma":
				m.enableFire = false
				if m.selectedBackground != "plasma" {
					m.selectedBackground = "plasma"
					width := m.width
					height := m.height
					if width == 0 {
						width = 80
					}
					if height == 0 {
						height = 30
					}
					m.plasmaEffect = animations.NewPlasmaEffect(width, height, animations.GetPlasmaPalette(m.currentTheme), m.currentTheme)
				} else {
					m.selectedBackground = "none"
					m.plasmaEffect = nil
				}
			case "AK-47":
				m.enableFire = false
				if m.selectedBackground != "ak-47" {
					m.selectedBackground = "ak-47"
					m.ak47Effect = m.newAK47Effect()
				} else {
					m.selectedBackground = "none"
					m.ak47Effect = nil
				}
			case "Australian SLR":
				m.enableFire = false
				if m.selectedBackground != "l2a1-sar" {
					m.selectedBackground = "l2a1-sar"
					m.l2a1SAREffect = m.newL2A1SAREffect()
				} else {
					m.selectedBackground = "none"
					m.l2a1SAREffect = nil
				}
			case "Aquarium":
				m.enableFire = false
				if m.selectedBackground != "aquarium" {
					m.selectedBackground = "aquarium"
					width := m.width
					height := m.height
					if width == 0 {
						width = 80
					}
					if height == 0 {
						height = 30
					}
					fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor, anchorColor := getThemeColorsForAquarium(m.currentTheme)
					m.aquariumEffect = animations.NewAquariumEffect(animations.AquariumConfig{
						Width:         width,
						Height:        height,
						FishColors:    fishColors,
						WaterColors:   waterColors,
						SeaweedColors: seaweedColors,
						BubbleColor:   bubbleColor,
						DiverColor:    diverColor,
						BoatColor:     boatColor,
						MermaidColor:  mermaidColor,
						AnchorColor:   anchorColor,
					})
				} else {
					m.selectedBackground = "none"
					m.aquariumEffect = nil
				}
			default:
				if id, ok := rotatingModelBackgroundID(optionName); ok {
					m.enableFire = false
					if m.selectedBackground != id {
						m.selectedBackground = id
						m.rotatingModelEffect = m.newRotatingModelEffect(id)
					} else {
						m.selectedBackground = "none"
						m.rotatingModelEffect = nil
					}
				}
			}

			if m.enableFire {
				m.selectedBackground = "fire"
			} else if m.selectedBackground != "pattern" && m.selectedBackground != "ascii-rain" && m.selectedBackground != "matrix" && m.selectedBackground != "fireworks" && m.selectedBackground != "sonar" && m.selectedBackground != "cracktro" && m.selectedBackground != "plasma" && m.selectedBackground != "aquarium" && m.selectedBackground != "ak-47" && m.selectedBackground != "l2a1-sar" && m.selectedBackground != "ticker" && !isRotatingModelBackground(m.selectedBackground) {
				m.selectedBackground = "none"
			}

			m.savePreferences()

			savedIndex := m.menuIndex
			newModel, cmd := m.navigateToBackgroundsSubmenu()
			m2 := newModel.(model)
			m2.menuIndex = savedIndex
			return m2, cmd
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
			case "ASCII Effects":
				newModel, cmd := m.navigateToASCIIEffectsSubmenu()
				return newModel.(model), cmd
			}
			return m, nil
		}

		// Handle submenu selections
		if m.mode == ModeThemesSubmenu || m.mode == ModeBordersSubmenu || m.mode == ModeBackgroundsSubmenu || m.mode == ModeWallpaperSubmenu || m.mode == ModeASCIIEffectsSubmenu {
			selectedOption := m.menuOptions[m.menuIndex]

			// Handle "← Back" option for all submenus
			if selectedOption == "← Back" {
				return m.navigateToMainMenu(), nil
			}

			// Implement actual submenu functionality
			switch m.mode {
			case ModeThemesSubmenu:
				followingChanged := false
				if strings.HasSuffix(selectedOption, "Follow sysc-shell") {
					m.followShell = !m.followShell
					m.config.ThemeName = ""
					selectedOption = "Theme: " + m.startupTheme(m.currentTheme)
					followingChanged = true
				}
				// Parse theme selection and apply it.
				if strings.HasPrefix(selectedOption, "Theme: ") {
					themeName := strings.TrimPrefix(selectedOption, "Theme: ")
					if !followingChanged {
						m.followShell = false
					}
					m.currentTheme = themeName
					// Apply theme immediately
					applyTheme(themeName, m.config.TestMode)
					m.applyInputTheme()
					if m.selectedBackground == "ak-47" {
						m.ak47Effect = m.newAK47Effect()
					}
					if m.selectedBackground == "l2a1-sar" {
						m.l2a1SAREffect = m.newL2A1SAREffect()
					}
					if isRotatingModelBackground(m.selectedBackground) {
						m.rotatingModelEffect = m.newRotatingModelEffect(m.selectedBackground)
					}
					// CHANGED 2025-10-03 - Save theme preference
					// CHANGED 2025-10-03 - Skip saving in test mode
					m.savePreferences()

					// Reinitialize ASCII effects with new theme colors if active
					if m.selectedBackground == "beams" && m.beamsEffect != nil && m.selectedSession != nil {
						logDebug("Theme changed to %s - reinitializing beams", themeName)
						m.resetBeamsEffectForSession(m.selectedSession.Name)
					}
					if m.selectedBackground == "pour" && m.pourEffect != nil && m.selectedSession != nil {
						logDebug("Theme changed to %s - reinitializing pour", themeName)
						m.resetPourEffectForSession(m.selectedSession.Name)
					}
					// Aquarium updates palette automatically via UpdatePalette() in backgrounds.go
					m.mode = ModeLogin
				}
				return m, nil

			case ModeBordersSubmenu:
				optionName := strings.TrimPrefix(selectedOption, "[✓] ")
				optionName = strings.TrimPrefix(optionName, "[ ] ")
				if strings.HasPrefix(optionName, "Style: ") && m.borderRegistry != nil {
					label := strings.TrimPrefix(optionName, "Style: ")
					for _, renderer := range m.borderRegistry.styles() {
						if renderer.Label() == label {
							m.selectedBorderStyle = renderer.ID()
							break
						}
					}
				} else {
					switch optionName {
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
					case "Effect: Shimmer":
						if renderer := m.borderRegistry.get(m.selectedBorderStyle); renderer != nil && renderer.SupportsEffect("shimmer") {
							m.borderEffect = "shimmer"
						}
					case "Effect: Off":
						m.borderEffect = "off"
					case "Palette: Theme":
						m.borderPaletteMode = "theme"
					case "Palette: VGA":
						m.borderPaletteMode = "vga"
					case "Palette: Terminal":
						m.borderPaletteMode = "terminal"
					case "Form: Auto":
						m.formMode = "auto"
					case "Form: Opaque":
						m.formMode = "opaque"
					case "Form: See-through":
						m.formMode = "seethrough"
					case "Form: Frosted":
						m.formMode = "frosted"
					}
				}
				// CHANGED 2025-10-03 - Save border preference
				// CHANGED 2025-10-03 - Skip saving in test mode
				m.savePreferences()
				m.mode = ModeLogin
				return m, nil

			case ModeBackgroundsSubmenu:
				// Enter confirms and exits backgrounds submenu
				m.mode = ModeLogin
				return m, nil
			case ModeASCIIEffectsSubmenu:
				// Handle ASCII Effects submenu selections
				optionName := strings.TrimPrefix(selectedOption, "[✓] ")
				optionName = strings.TrimPrefix(optionName, "[ ] ")

				switch optionName {
				case "Typewriter":
					// Typewriter is exclusive - disable other backgrounds/effects
					m.enableFire = false
					if m.selectedBackground != "ticker" {
						m.selectedBackground = "ticker"
						// Initialize ticker if not already done
						if m.typewriterTicker == nil && m.selectedSession != nil {
							// Load custom roasts from ASCII config
							configFileName := asciiConfigFileName(m.selectedSession.Name)
							configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)
							customRoasts := ""
							if asciiConfig, err := loadASCIIConfig(configPath); err == nil {
								customRoasts = asciiConfig.Roasts
							}
							m.typewriterTicker = animations.NewTypewriterTicker(m.selectedSession.Name, customRoasts)
						}
					} else {
						m.selectedBackground = "none"
					}
				case "Print":
					// Print is exclusive - disable other backgrounds/effects
					m.enableFire = false
					if m.selectedBackground != "print" {
						m.selectedBackground = "print"
						// Initialize print effect with current session's ASCII art
						if m.selectedSession != nil {
							// Load ASCII config for current session
							configFileName := asciiConfigFileName(m.selectedSession.Name)

							configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)
							if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
								// Get current ASCII variant
								variantIndex := m.asciiArtIndex
								if variantIndex >= len(asciiConfig.ASCIIVariants) {
									variantIndex = 0
								}
								ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])
								// Fast print speed for main UI (3ms per char = very fast)
								m.printEffect = animations.NewPrintEffect(ascii, time.Millisecond*3)
								if m.config.Debug {
									logDebug("Print effect initialized with ASCII (%d lines)", len(strings.Split(ascii, "\n")))
								}
							} else {
								if m.config.Debug {
									logDebug("Cannot load ASCII config from: %s", configPath)
								}
							}
						}
					} else {
						m.selectedBackground = "none"
						m.printEffect = nil
					}
				case "Beams":
					m.enableFire = false
					if m.selectedBackground != "beams" {
						m.selectedBackground = "beams"
						if m.selectedSession != nil {
							configFileName := asciiConfigFileName(m.selectedSession.Name)

							configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)
							if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
								variantIndex := m.asciiArtIndex
								if variantIndex >= len(asciiConfig.ASCIIVariants) {
									variantIndex = 0
								}
								ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])

								beamColors, finalColors := getThemeColorsForBeams(m.currentTheme)
								if m.config.Debug {
									logDebug("Initializing beams with theme: %s, beamColors: %v, finalColors: %v", m.currentTheme, beamColors, finalColors)
								}

								lines := strings.Split(ascii, "\n")
								asciiHeight := len(lines)
								asciiWidth := 0
								for _, line := range lines {
									if len([]rune(line)) > asciiWidth {
										asciiWidth = len([]rune(line))
									}
								}

								m.beamsEffect = animations.NewBeamsTextEffect(animations.BeamsTextConfig{
									Width:              asciiWidth,
									Height:             asciiHeight,
									Text:               ascii,
									BeamGradientStops:  beamColors,
									FinalGradientStops: finalColors,
								})
								if m.config.Debug {
									logDebug("Beams effect initialized")
								}
							}
						}
					} else {
						m.selectedBackground = "none"
						m.beamsEffect = nil
					}
				case "Pour":
					m.enableFire = false
					if m.selectedBackground != "pour" {
						m.selectedBackground = "pour"
						if m.selectedSession != nil {
							configFileName := asciiConfigFileName(m.selectedSession.Name)

							configPath := fmt.Sprintf("%s/ascii_configs/%s.conf", dataDir, configFileName)
							if asciiConfig, err := loadASCIIConfig(configPath); err == nil && len(asciiConfig.ASCIIVariants) > 0 {
								variantIndex := m.asciiArtIndex
								if variantIndex >= len(asciiConfig.ASCIIVariants) {
									variantIndex = 0
								}
								ascii := asciiTextForEffect(asciiConfig, asciiConfig.ASCIIVariants[variantIndex])

								pourColors := getThemeColorsForPour(m.currentTheme)
								if m.config.Debug {
									logDebug("Initializing pour with theme: %s, colors: %v", m.currentTheme, pourColors)
								}

								lines := strings.Split(ascii, "\n")
								asciiHeight := len(lines)
								asciiWidth := 0
								for _, line := range lines {
									if len([]rune(line)) > asciiWidth {
										asciiWidth = len([]rune(line))
									}
								}

								m.pourEffect = animations.NewPourEffect(animations.PourConfig{
									Width:                  asciiWidth,
									Height:                 asciiHeight,
									Text:                   ascii,
									PourDirection:          "down",
									PourSpeed:              1,    // Slower: 1 char per frame instead of 3
									MovementSpeed:          0.05, // Much slower movement
									Gap:                    2,    // Longer gap between rows
									StartingColor:          "#ffffff",
									FinalGradientStops:     pourColors,
									FinalGradientSteps:     12,
									FinalGradientFrames:    5,
									FinalGradientDirection: "horizontal",
								})
								if m.config.Debug {
									logDebug("Pour effect initialized")
								}
							}
						}
					} else {
						m.selectedBackground = "none"
						m.pourEffect = nil
					}

				}

				// Save preference
				m.savePreferences()
				// Refresh menu to update checkboxes
				newModel, cmd := m.navigateToASCIIEffectsSubmenu()
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
					logDebug("Switching to password mode")
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
					m.passwordInput.SetValue("")
					m.authNotice = ""
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

func (m model) passwordInputActive() bool {
	return m.mode == ModePassword && m.focusState == FocusPassword
}

func (m model) View() tea.View {
	view := m.buildView()
	m.frameCache.track(view.Layer)
	return view
}

// Return tea.View with BackgroundColor set
func (m model) buildView() tea.View {
	// Use full terminal dimensions
	termWidth := m.width
	termHeight := m.height
	if termWidth == 0 {
		termWidth = 80
	}
	if termHeight == 0 {
		termHeight = 24
	}
	view := tea.View{AltScreen: m.altScreen, UniformKeyLayout: true}
	if m.mode == ModeScreensaver {
		view.MouseMode = tea.MouseModeAllMotion
	} else if m.altScreen && !m.config.TestMode {
		view.MouseMode = tea.MouseModeCellMotion
	}
	if canvas, ok := m.borderedLoginCanvas(termWidth, termHeight); ok {
		view.Layer = canvas
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	}

	var content string
	switch m.mode {
	case ModePower:
		// Fixed missing power menu rendering
		content = m.renderPowerView(termWidth, termHeight)
	case ModeMenu, ModeThemesSubmenu, ModeBordersSubmenu, ModeBackgroundsSubmenu, ModeWallpaperSubmenu, ModeASCIIEffectsSubmenu:
		// Removed ModeVideoWallpapersSubmenu from rendering
		content = m.renderMenuView(termWidth, termHeight)
	case ModeReleaseNotes:
		content = m.renderReleaseNotesView(termWidth, termHeight)
	case ModeScreensaver:
		// CHANGED 2025-10-10 - Added screensaver rendering
		content = renderScreensaverView(m, termWidth, termHeight)
	default:
		content = m.renderMainView(termWidth, termHeight)
	}

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
		backgroundContent := m.addFireEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: fire at bottom, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, fireY, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if hasRainBackground {
		// CHANGED 2025-10-08 - Render ascii-rain as full background
		// Render rain as full background
		backgroundContent := m.addAsciiRain("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: rain as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "matrix" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render matrix as full background
		backgroundContent := m.addMatrixEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: matrix as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "fireworks" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render fireworks as full background
		backgroundContent := m.addFireworksEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: fireworks as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "sonar" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render sonar as full background
		backgroundContent := m.addSonarEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: sonar as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "cracktro" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render cracktro as full background
		backgroundContent := m.addCracktroEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: cracktro as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "plasma" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render plasma as full background
		backgroundContent := m.addPlasmaEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: plasma as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "aquarium" && (m.mode == ModeLogin || m.mode == ModePassword) {
		// Render aquarium as full background
		backgroundContent := m.addAquariumEffect("")

		// Center the UI content
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		// Create canvas with two layers: aquarium as background, UI centered on top
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	} else if m.selectedBackground == "ak-47" && (m.mode == ModeLogin || m.mode == ModePassword) {
		backgroundContent := m.addAK47Effect("")

		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2

		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	}

	if m.selectedBackground == "l2a1-sar" && (m.mode == ModeLogin || m.mode == ModePassword) {
		backgroundContent := m.addL2A1SAREffect("")
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	}

	if isRotatingModelBackground(m.selectedBackground) && (m.mode == ModeLogin || m.mode == ModePassword) {
		backgroundContent := m.addRotatingModelEffect("")
		contentWidth := lipgloss.Width(content)
		contentHeight := lipgloss.Height(content)
		uiX := (termWidth - contentWidth) / 2
		uiY := (termHeight - contentHeight) / 2
		view.Layer = m.composeLoginLayer(backgroundContent, 0, 0, content, uiX, uiY, termWidth, termHeight)
		view.BackgroundColor = m.viewBackgroundColor()
		return view
	}

	// CHANGED 2025-10-06 - Use X/Y positioning instead of Place() to avoid ghosting
	// Calculate center position manually (CRUSH approach)
	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)
	x := (termWidth - contentWidth) / 2
	y := (termHeight - contentHeight) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Keep a full-terminal backing layer so small popups clear stale cells even
	// when Bubble Tea is running without alt-screen.
	view.Layer = m.composeLayer(blankTerminalContent(termWidth, termHeight), 0, 0, content, x, y, termWidth, termHeight)
	view.BackgroundColor = m.viewBackgroundColor()
	return view
}

func blankTerminalContent(width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	line := strings.Repeat(" ", width)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
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
	m.width, m.height = termWidth, termHeight
	// ponytail: legacy ASCII frames omit some authentication feedback; use the
	// shared form for these states until those frames also reuse renderMainForm.
	needsFeedback := strings.HasPrefix(m.selectedBorderStyle, "ascii") &&
		(m.mode == ModeLoading || m.errorMessage != "" || m.failedAttempts > 0 || m.capsLockOn || m.authPromptReply != nil)
	compactHeader := ""
	if termHeight < 26 && m.selectedSession != nil && m.selectedBorderStyle != "ascii1" && m.selectedBorderStyle != "ascii2" {
		if art, hasInlineColors := m.getSessionASCIIWithColors(); hasInlineColors {
			compactHeader = art
		}
	}
	// Decorative frames assume at least 80 columns and 24 rows.
	if !needsFeedback && termWidth >= 80 && termHeight >= 24 {
		content := m.renderDualBorderLayout(termWidth, termHeight)
		if compactHeader == "" && lipgloss.Width(content) <= termWidth && lipgloss.Height(content) <= termHeight {
			return content
		}
	}
	width := min(72, max(15, termWidth-6))
	form := m.renderMainForm(width)
	var lines []string
	for _, line := range strings.Split(form, "\n") {
		if strings.TrimSpace(stripAnsi(line)) != "" {
			lines = append(lines, line)
		}
	}
	help := "Enter Continue • Tab Focus • F1 Menu"
	if m.authPromptReply != nil {
		help = "Enter Submit • Esc Cancel"
	} else if m.sessionDropdownOpen {
		help = "↑↓ Navigate • Enter Select • Esc Close"
	} else if m.mode == ModeLoading {
		help = "Please wait..."
	} else if m.mode == ModePassword {
		help = "Enter Login • Esc Back • Tab Focus • F1 Menu"
	}
	form = lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, "\n"), "", lipgloss.NewStyle().Foreground(FgMuted).Width(width).Render(help))
	frame := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BorderDefault).Background(BgBase).Padding(0, 2).Render(form)
	if m.selectedSession != nil {
		if compactHeader == "" {
			compactHeader = m.getSessionArt(m.selectedSession.Name)
		}
		if compactHeader != "" {
			compact := lipgloss.JoinVertical(lipgloss.Center, compactHeader, "", frame)
			// ponytail: show the full ANSI art only when it fits; clipping needs per-line ANSI-aware handling.
			if lipgloss.Width(compact) <= termWidth && lipgloss.Height(compact) <= termHeight {
				return compact
			}
		}
	}
	if lipgloss.Width(frame) <= termWidth && lipgloss.Height(frame) <= termHeight {
		return frame
	}
	if m.renderAmbientRow(width) != "" {
		m.hideAmbient = true
		return m.renderMainView(termWidth, termHeight)
	}
	return form
}

// Border rendering functions moved to borders.go
// Includes: renderDualBorderLayout, renderASCII1/2BorderLayout,
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
		defer func() { password = "" }()
		if m.ipcClient == nil {
			return fmt.Errorf("IPC client not initialized - greeter must be run by greetd")
		}

		if err := m.ipcClient.CreateSession(username); err != nil {
			return err
		}

		cancelWithError := func(cause error) error {
			if err := m.ipcClient.CancelSession(); err != nil {
				return fmt.Errorf("%w; failed to cancel greetd session: %v", cause, err)
			}
			return cause
		}

		firstSecret := true
		terminal, err := m.ipcClient.Converse(func(message ipc.AuthMessage) (*string, error) {
			if m.config.Debug {
				logDebug(" Received auth message type: %s", message.AuthMessageType)
			}
			switch message.AuthMessageType {
			case ipc.AuthMessageTypeInfo, ipc.AuthMessageTypeError:
				if err := m.sendAuthNotice(message.AuthMessage); err != nil {
					return nil, err
				}
				return nil, nil
			case ipc.AuthMessageTypeSecret:
				if firstSecret {
					firstSecret = false
					response := password
					password = ""
					return &response, nil
				}
				fallthrough
			case ipc.AuthMessageTypeVisible:
				return m.requestAuthPrompt(message)
			default:
				return nil, fmt.Errorf("unsupported greetd authentication message type %q", message.AuthMessageType)
			}
		})
		password = ""
		if err != nil {
			return cancelWithError(err)
		}

		switch response := terminal.(type) {
		case ipc.Error:
			return cancelWithError(fmt.Errorf("authentication failed: %s - %s", sanitizeAuthText(response.ErrorType), sanitizeAuthText(response.Description)))
		case ipc.Success:
			if m.selectedSession == nil {
				return cancelWithError(fmt.Errorf("no session selected"))
			}
			execParts := sessionCommand(m.selectedSession.Exec)
			if len(execParts) > 0 && filepath.Base(execParts[0]) == "sway" {
				hasFlag := false
				for _, part := range execParts {
					if part == "--unsupported-gpu" {
						hasFlag = true
						break
					}
				}
				if !hasFlag {
					execParts = append([]string{execParts[0], "--unsupported-gpu"}, execParts[1:]...)
				}
			}
			if err := m.ipcClient.StartSession(execParts, []string{}); err != nil {
				return cancelWithError(err)
			}
			return authSuccessMsg{Username: username, Session: m.selectedSession.Name}
		default:
			return cancelWithError(fmt.Errorf("unexpected greetd authentication response %T", terminal))
		}
	}
}

// Utility helper functions (min, stripAnsi, extractCharsWithAnsi, etc.) moved to utils.go

func main() {
	// CHANGED 2025-10-01 - Removed SetColorProfile - not available in lipgloss v2
	// Color profile is now automatically detected via colorprofile package
	// CHANGED 2025-10-14 - Removed sysc-greet.conf loading - hardcoded sessionPalettes provide all needed palettes

	// Initialize config with defaults
	config := Config{
		RememberUsername: true, // Default: remember username
		Secondary:        secondaryConfig{Enabled: true},
	}

	var screensaverTestMode bool // CHANGED 2025-10-11 - Add screensaver test mode flag
	var showVersion bool
	var startWallpaperDaemon bool
	var widgetConfigPath string
	var followShell bool
	var borderPreviewPath string
	var borderPreviewSizes string
	var secondarySaverStartedAt string
	var secondaryReadyFile string

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.BoolVar(&startWallpaperDaemon, "wallpaper-daemon", false, "Start gSlapper with the cached greeter wallpaper")
	flag.BoolVar(&config.TestMode, "test", false, "Enable test mode (no actual authentication)")
	flag.StringVar(&config.TestSession, "test-session", "", "Select a session profile for visual testing (requires --test)")
	flag.BoolVar(&config.SecondarySaver, "secondary-screensaver", false, "")
	flag.StringVar(&secondarySaverStartedAt, "screensaver-started-at", "", "")
	flag.StringVar(&secondaryReadyFile, "secondary-ready-file", "", "")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug output")
	flag.BoolVar(&screensaverTestMode, "screensaver", false, "Start directly in screensaver mode for testing")
	flag.BoolVar(&followShell, "follow-shell", true, "Follow the installed shell account theme by default")
	flag.StringVar(&config.ShellThemeFile, "shell-theme-file", "", "Override the exported shell theme file")
	flag.StringVar(&config.ThemeName, "theme", "", "Theme name from the Themes menu (overrides saved theme)")
	flag.BoolVar(&config.RememberUsername, "remember-username", true, "Remember last logged in username")
	flag.BoolVar(&config.ShowTime, "time", false, "") // Hidden flag - not shown in help
	flag.StringVar(&config.BorderLayout, "border-layout", "hug", "Outer frame layout: hug keeps fixed margins, fill grows to terminal edges")
	flag.StringVar(&config.FormMode, "form-mode", "auto", "Login form mode: auto, opaque, seethrough or frosted")
	flag.IntVar(&config.FormHalo, "form-halo", 1, "Halo width in columns around form text (0-2)")
	flag.Float64Var(&config.FormDim, "form-dim", 0.55, "Frosted form dim strength (0-1)")
	flag.StringVar(&borderPreviewPath, "border-preview", "", "Render a .ans border around a sample login form (requires --test)")
	flag.StringVar(&borderPreviewSizes, "size", "80x24,132x40,200x55", "Comma-separated sizes for --border-preview")
	flag.StringVar(&widgetConfigPath, "widget-config", "", "Widget JSON configuration file (overrides default paths)")
	flag.StringVar(&config.Ambient.ShellConfig, "weather-shell-config", "", "sysc-shell config path when weather-location is sysc-shell")
	flag.BoolVar(&config.Ambient.Metrics, "metrics", false, "Show CPU/Memory usage in the login panel")
	flag.BoolVar(&config.Ambient.GPU, "gpu", false, "Collect optional GPU statistics for upcoming widgets (no display yet)")
	flag.BoolVar(&config.Secondary.Enabled, "secondary-backgrounds", true, "Enable backgrounds on secondary outputs in supported greeter sessions")
	flag.StringVar(&config.Secondary.Exclude, "secondary-exclude", "", "Comma-separated outputs to exclude from secondary backgrounds")
	flag.StringVar(&config.Secondary.Effect, "secondary-effect", "", "Explicit sysc-terminal effect fallback for secondary outputs")
	flag.StringVar(&config.Ambient.WeatherLocation, "weather-location", "", "Show weather at LAT,LON or reuse sysc-shell coordinates")
	flag.StringVar(&config.Ambient.WeatherUnits, "weather-units", "", "Weather temperature units: celsius or fahrenheit")

	// Add help text
	// CHANGED 2025-10-12 - Updated help text to reflect sysc-greet branding
	// CHANGED 2025-10-14 - Removed sysc-greet.conf references
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "sysc-greet - A terminal greeter for greetd\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintln(os.Stderr, "  -follow-shell=false\n    \tKeep an independent greeter theme")
		fmt.Fprintln(os.Stderr, "  -shell-theme-file PATH\n    \tRead the shell theme selection from PATH")
		fmt.Fprintln(os.Stderr, "  -widget-config PATH\n    \tWidget JSON config (user config, then /etc/sysc-greet/widgets.json)")
		fmt.Fprintln(os.Stderr, "  -weather-shell-config PATH\n    \tsysc-shell config for weather-location=sysc-shell")
		fmt.Fprintln(os.Stderr, "  -metrics\n    \tShow CPU/Memory usage in the login panel")
		fmt.Fprintln(os.Stderr, "  -gpu\n    \tCollect optional GPU statistics for upcoming widgets (no display yet)")
		fmt.Fprintln(os.Stderr, "  -secondary-backgrounds=false\n    \tDisable secondary backgrounds (enabled in supported greeter sessions)")
		fmt.Fprintln(os.Stderr, "  -secondary-exclude OUTPUT,OUTPUT\n    \tExclude outputs from secondary backgrounds")
		fmt.Fprintln(os.Stderr, "  -secondary-effect string\n    \texplicit fallback effect for secondary outputs (default empty)")
		fmt.Fprintln(os.Stderr, "  -weather-location LAT,LON\n    \tShow weather at LAT,LON or reuse sysc-shell coordinates")
		fmt.Fprintln(os.Stderr, "  -weather-units string\n    \tWeather units: celsius (default) or fahrenheit")
		fmt.Fprintln(os.Stderr, "  -border-preview FILE.ans\n    Render a border asset at --size values; requires -test")
		fmt.Fprintln(os.Stderr, "  -size 80x24,132x40,200x55\n    Terminal sizes for -border-preview")
		fmt.Fprintln(os.Stderr, "  -border-layout hug|fill\n    Outer frame layout (default hug)")
		fmt.Fprintln(os.Stderr, "  -form-mode auto|opaque|seethrough|frosted\n    Login form compositing mode")
		fmt.Fprintln(os.Stderr, "  -form-halo 0|1|2\n    Halo width in columns around form text")
		fmt.Fprintln(os.Stderr, "  -form-dim 0..1\n    Frosted form dim strength")
		// Manually print flags (excluding hidden ones)
		fmt.Fprintf(os.Stderr, "  -debug\n")
		fmt.Fprintf(os.Stderr, "    	Enable debug output\n")
		fmt.Fprintf(os.Stderr, "  -screensaver\n")
		fmt.Fprintf(os.Stderr, "    	Start directly in screensaver mode for testing\n")
		fmt.Fprintf(os.Stderr, "  -test\n")
		fmt.Fprintf(os.Stderr, "    	Enable test mode (no actual authentication)\n")
		fmt.Fprintf(os.Stderr, "  -test-session string\n")
		fmt.Fprintf(os.Stderr, "    \tSelect a session profile for visual testing (requires -test)\n")
		fmt.Fprintf(os.Stderr, "  -theme string\n")
		fmt.Fprintf(os.Stderr, "    	Theme name (dracula, gruvbox, material, nord, tokyo-night, catppuccin, solarized, monochrome, transishardjob, eldritch)\n")
		fmt.Fprintf(os.Stderr, "  -v	Show version information (shorthand)\n")
		fmt.Fprintf(os.Stderr, "  -version\n")
		fmt.Fprintf(os.Stderr, "    	Show version information\n")
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  ASCII configs: %s/ascii_configs/\n", dataDir)
		fmt.Fprintf(os.Stderr, "\nKey Bindings:\n")
		fmt.Fprintf(os.Stderr, "  Tab       Cycle focus between elements\n")
		fmt.Fprintf(os.Stderr, "  ↑↓       Navigate sessions when focused\n")
		fmt.Fprintf(os.Stderr, "  F1        Settings menu\n")
		fmt.Fprintf(os.Stderr, "  F2        Toggle session dropdown\n")
		fmt.Fprintf(os.Stderr, "  F3        Release notes\n")
		fmt.Fprintf(os.Stderr, "  F4        Power menu\n")
		fmt.Fprintf(os.Stderr, "  Enter     Continue to next step\n")
		fmt.Fprintf(os.Stderr, "  Esc       Cancel/go back\n")
		fmt.Fprintf(os.Stderr, "  Ctrl+C    Quit\n")
	}

	flag.Parse()
	if config.TestSession != "" {
		name, err := normalizeTestSessionOption(config.TestMode, config.TestSession)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		config.TestSession = name
	}
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "follow-shell" {
			config.FollowShell = &followShell
		} else if f.Name == "form-mode" {
			config.FormModeExplicit = true
		}
	})

	// Handle version flag
	if showVersion {
		fmt.Printf("sysc-greet %s\n", Version)
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildDate)
		os.Exit(0)
	}
	var parsedSecondarySaverStart time.Time
	if config.SecondarySaver {
		var err error
		parsedSecondarySaverStart, err = parseSecondarySaverStart(secondarySaverStartedAt)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if secondaryReadyFile == "" {
			fmt.Fprintln(os.Stderr, "secondary screensaver requires --secondary-ready-file")
			os.Exit(2)
		}
	}

	if startWallpaperDaemon {
		if err := runStartupWallpaperDaemon(config); err != nil {
			fmt.Fprintf(os.Stderr, "failed to start wallpaper daemon: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if config.BorderLayout != "hug" && config.BorderLayout != "fill" {
		fmt.Fprintf(os.Stderr, "invalid --border-layout %q: expected hug or fill\n", config.BorderLayout)
		os.Exit(2)
	}
	if !validFormMode(config.FormMode) {
		fmt.Fprintf(os.Stderr, "invalid --form-mode %q: expected auto, opaque, seethrough or frosted\n", config.FormMode)
		os.Exit(2)
	}
	if config.FormHalo < 0 || config.FormHalo > 2 {
		fmt.Fprintf(os.Stderr, "invalid --form-halo %d: expected 0, 1 or 2\n", config.FormHalo)
		os.Exit(2)
	}
	if !validFormDim(config.FormDim) {
		fmt.Fprintf(os.Stderr, "invalid --form-dim %v: expected a number from 0 to 1\n", config.FormDim)
		os.Exit(2)
	}

	// SECURITY: Prevent test mode in production environment
	// Test mode bypasses authentication and should only be used for development
	if config.TestMode && os.Getenv("GREETD_SOCK") != "" {
		fmt.Fprintf(os.Stderr, "SECURITY ERROR: Test mode cannot be enabled in production (GREETD_SOCK is set)\n")
		fmt.Fprintf(os.Stderr, "Test mode bypasses authentication and should only be used for development.\n")
		os.Exit(1)
	}
	if borderPreviewPath != "" {
		if !config.TestMode {
			fmt.Fprintln(os.Stderr, "--border-preview requires --test")
			os.Exit(2)
		}
		sizes, err := parseBorderPreviewSizes(borderPreviewSizes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --size: %v\n", err)
			os.Exit(2)
		}
		data, err := loadBorderPreviewFile(borderPreviewPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "border preview: %v\n", err)
			os.Exit(1)
		}
		applyTheme(config.ThemeName, true)
		preview, err := renderBorderPreview(data, sizes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "border preview: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(preview)
		return
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

	// Bubble Tea handles SIGINT/SIGTERM; terminal close also sends SIGHUP.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP)
	defer cancel()
	paths := []string{"/etc/sysc-greet/widgets.json"}
	if dir, err := os.UserConfigDir(); err == nil {
		paths = append([]string{filepath.Join(dir, "sysc-greet", "widgets.json")}, paths...)
	}
	if widgetConfigPath != "" {
		paths = []string{widgetConfigPath}
	}
	overrides := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { overrides[f.Name] = true })
	ambient, err := loadWidgetConfig(paths, config.Ambient, overrides)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Widget configuration warning: %v\n", err)
		logDebug("Widget configuration warning: %v", err)
	}
	config.Ambient = ambient
	var collector *ambientCollector
	if !config.SecondarySaver {
		collector, err = newAmbientCollector(ctx, config.Ambient)
		if err != nil {
			cancel()
			fmt.Fprintf(os.Stderr, "Invalid ambient configuration: %v\n", err)
			os.Exit(1)
		}
	}
	m := initialModel(config, screensaverTestMode || config.SecondarySaver)
	m.kittyGraphics = kittyGraphicsAvailable()
	m.ambient = collector
	m.secondaryReadyFile = secondaryReadyFile
	if config.SecondarySaver {
		m.screensaverStart = parsedSecondarySaverStart
	}
	m.secondaryRuntime = newSecondaryStateBus()
	m.publishSecondaryRuntime()
	// Use fullscreen terminal modes when a controlling TTY is available.
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err != nil {
		if config.Debug {
			logDebug(" No TTY access, running without alt screen")
		}
	} else {
		tty.Close()
		m.altScreen = true
	}

	var p *tea.Program
	m.sendMessage = func(message tea.Msg) {
		if p != nil {
			p.Send(message)
		}
	}
	p = tea.NewProgram(m, tea.WithContext(ctx))
	var secondaryDone <-chan struct{}
	if config.SecondarySaver {
		done := make(chan struct{})
		close(done)
		secondaryDone = done
	} else {
		secondaryDone = startSecondaryManager(ctx, config.Secondary, m.secondaryRuntime, p.Send)
	}

	final, err := p.Run()
	if finalModel, ok := final.(model); ok {
		if sequence := finalModel.kittyImages.clear(); sequence != "" {
			if _, clearErr := io.WriteString(os.Stdout, sequence); clearErr != nil {
				logDebug("Could not clear Kitty mesh background: %v", clearErr)
			}
		}
	}
	cancel()
	<-secondaryDone
	if collector != nil {
		if closeErr := collector.close(); closeErr != nil {
			logDebug("Close GPU sampler: %v", closeErr)
		}
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// ASCII-2 border rendering function moved to borders.go
