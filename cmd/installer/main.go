package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Theme colors - Monochrome (ASCII style)
var (
	BgBase       = lipgloss.Color("#1a1a1a")
	BgElevated   = lipgloss.Color("#2a2a2a")
	Primary      = lipgloss.Color("#ffffff")
	Secondary    = lipgloss.Color("#cccccc")
	Accent       = lipgloss.Color("#ffffff")
	FgPrimary    = lipgloss.Color("#ffffff")
	FgSecondary  = lipgloss.Color("#cccccc")
	FgMuted      = lipgloss.Color("#666666")
	ErrorColor   = lipgloss.Color("#ffffff")
	WarningColor = lipgloss.Color("#888888")
)

// Styles
var (
	checkMark   = lipgloss.NewStyle().Foreground(Accent).SetString("[OK]")
	failMark    = lipgloss.NewStyle().Foreground(ErrorColor).SetString("[FAIL]")
	skipMark    = lipgloss.NewStyle().Foreground(WarningColor).SetString("[SKIP]")
	headerStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
)

type installStep int

const (
	stepWelcome installStep = iota
	stepCompositorSelect
	stepInstalling
	stepComplete
)

type taskStatus int

const (
	statusPending taskStatus = iota
	statusRunning
	statusComplete
	statusFailed
	statusSkipped
)

type installTask struct {
	name        string
	description string
	execute     func(*model) error
	optional    bool
	status      taskStatus
}

type model struct {
	step               installStep
	tasks              []installTask
	currentTaskIndex   int
	width              int
	height             int
	spinner            spinner.Model
	errors             []string
	packageManager     string
	greetdInstalled    bool
	needsGreetd        bool
	selectedCompositor string
	compositorOptions  []string
	compositorIndex    int
}

type taskCompleteMsg struct {
	index   int
	success bool
	skipped bool
	error   string
}

func newModel() model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(Secondary)
	s.Spinner = spinner.Dot

	// Check for pre-selected compositor from environment
	preSelectedCompositor := os.Getenv("SYSC_COMPOSITOR")

	tasks := []installTask{
		{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
		{name: "Check dependencies", description: "Checking system dependencies", execute: checkDependencies, status: statusPending},
		{name: "Install greetd", description: "Installing greetd daemon", execute: installGreetd, optional: true, status: statusPending},
		{name: "Install gslapper", description: "Installing video wallpaper support", execute: installGslapper, optional: true, status: statusPending},
		{name: "Build binary", description: "Building sysc-greet", execute: buildBinary, status: statusPending},
		{name: "Install binary", description: "Installing to system", execute: installBinary, status: statusPending},
		{name: "Install configs", description: "Installing configurations", execute: installConfigs, status: statusPending},
		{name: "Setup cache", description: "Setting up cache and permissions", execute: setupCache, status: statusPending},
		{name: "Configure greetd", description: "Configuring greetd daemon", execute: configureGreetd, status: statusPending},
		{name: "Enable service", description: "Enabling greetd service", execute: enableService, status: statusPending},
	}

	return model{
		step:               stepWelcome,
		tasks:              tasks,
		currentTaskIndex:   -1,
		spinner:            s,
		errors:             []string{},
		compositorOptions:  []string{"niri", "hyprland", "sway"},
		compositorIndex:    0,
		selectedCompositor: preSelectedCompositor,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == stepComplete || m.step == stepWelcome || m.step == stepCompositorSelect {
				return m, tea.Quit
			}
			// Allow cancelling during installation
			if m.step == stepInstalling {
				return m, tea.Quit
			}
		case "enter":
			if m.step == stepWelcome {
				// Move to compositor selection if not pre-selected
				if m.selectedCompositor == "" {
					m.step = stepCompositorSelect
					return m, nil
				} else {
					// Skip compositor selection if pre-selected
					m.step = stepInstalling
					m.currentTaskIndex = 0
					m.tasks[0].status = statusRunning
					return m, tea.Batch(
						m.spinner.Tick,
						executeTask(0, &m),
					)
				}
			} else if m.step == stepCompositorSelect {
				// Set selected compositor and start installation
				m.selectedCompositor = m.compositorOptions[m.compositorIndex]
				m.step = stepInstalling
				m.currentTaskIndex = 0
				m.tasks[0].status = statusRunning
				return m, tea.Batch(
					m.spinner.Tick,
					executeTask(0, &m),
				)
			} else if m.step == stepComplete {
				return m, tea.Quit
			}
		case "up", "k":
			if m.step == stepCompositorSelect {
				m.compositorIndex--
				if m.compositorIndex < 0 {
					m.compositorIndex = len(m.compositorOptions) - 1
				}
				return m, nil
			}
		case "down", "j":
			if m.step == stepCompositorSelect {
				m.compositorIndex++
				if m.compositorIndex >= len(m.compositorOptions) {
					m.compositorIndex = 0
				}
				return m, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case taskCompleteMsg:
		// Handle task completion
		if msg.index >= 0 && msg.index < len(m.tasks) {
			if msg.skipped {
				m.tasks[msg.index].status = statusSkipped
			} else if msg.success {
				m.tasks[msg.index].status = statusComplete
			} else {
				m.tasks[msg.index].status = statusFailed
				m.errors = append(m.errors, fmt.Sprintf("Task %s failed: %s", m.tasks[msg.index].name, msg.error))
			}

			// Move to next task or finish
			nextIndex := msg.index + 1
			if nextIndex < len(m.tasks) {
				m.currentTaskIndex = nextIndex
				m.tasks[nextIndex].status = statusRunning
				return m, tea.Batch(
					m.spinner.Tick,
					executeTask(nextIndex, &m),
				)
			} else {
				// All tasks complete
				m.step = stepComplete
				return m, nil
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// ASCII Header
	headerLines := []string{
		"  █████████  █████ █████  █████████    █████████ ",
		" ███░░░░░███░░███ ░░███  ███░░░░░███  ███░░░░░███",
		"░███    ░░░  ░░███ ███  ░███    ░░░  ███     ░░░ ",
		"░░█████████   ░░█████   ░░█████████ ░███         ",
		" ░░░░░░░░███   ░░███     ░░░░░░░░███░███         ",
		" ███    ░███    ░███     ███    ░███░░███     ███",
		"░░█████████     █████   ░░█████████  ░░█████████ ",
		" ░░░░░░░░░     ░░░░░     ░░░░░░░░░    ░░░░░░░░░  ",
		"//////////////SEE YOU IN SPACE COWBOY//////////  ",
	}

	for _, line := range headerLines {
		content.WriteString(headerStyle.Render(line))
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true).
		Align(lipgloss.Center)
	content.WriteString(titleStyle.Render("sysc-greet installer"))
	content.WriteString("\n\n")

	// Main content based on step
	var mainContent string
	switch m.step {
	case stepWelcome:
		mainContent = m.renderWelcome()
	case stepCompositorSelect:
		mainContent = m.renderCompositorSelect()
	case stepInstalling:
		mainContent = m.renderInstalling()
	case stepComplete:
		mainContent = m.renderComplete()
	}

	// Wrap in border
	mainStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Width(m.width - 4)
	content.WriteString(mainStyle.Render(mainContent))
	content.WriteString("\n")

	// Help text
	helpText := m.getHelpText()
	if helpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(FgMuted).
			Italic(true).
			Align(lipgloss.Center)
		content.WriteString("\n" + helpStyle.Render(helpText))
	}

	// Wrap everything in background with centering
	bgStyle := lipgloss.NewStyle().
		Background(BgBase).
		Foreground(FgPrimary).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Top)

	return bgStyle.Render(content.String())
}

func (m model) renderWelcome() string {
	return `sysc-greet installer

Builds binary, installs to system, configures greetd.
Requires root.

Press Enter to continue`
}

func (m model) renderCompositorSelect() string {
	var b strings.Builder
	b.WriteString("Select compositor for sysc-greet:\n\n")

	for i, comp := range m.compositorOptions {
		prefix := "  "
		if i == m.compositorIndex {
			prefix = "> "
		}
		b.WriteString(prefix + comp + "\n")
	}

	b.WriteString("\nUse ↑↓ to select, Enter to continue")
	return b.String()
}

func (m model) renderInstalling() string {
	var b strings.Builder

	// Render all tasks with their current status
	for i, task := range m.tasks {
		var line string
		switch task.status {
		case statusPending:
			line = lipgloss.NewStyle().Foreground(FgMuted).Render("  " + task.name)
		case statusRunning:
			line = m.spinner.View() + " " + lipgloss.NewStyle().Foreground(Secondary).Render(task.description)
		case statusComplete:
			line = checkMark.String() + " " + task.name
		case statusFailed:
			line = failMark.String() + " " + task.name
		case statusSkipped:
			line = skipMark.String() + " " + task.name
		}

		b.WriteString(line)
		if i < len(m.tasks)-1 {
			b.WriteString("\n")
		}
	}

	// Show errors at bottom if any
	if len(m.errors) > 0 {
		b.WriteString("\n\n")
		for _, err := range m.errors {
			b.WriteString(lipgloss.NewStyle().Foreground(WarningColor).Render(err))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m model) renderComplete() string {
	// Check for critical failures
	hasCriticalFailure := false
	for _, task := range m.tasks {
		if task.status == statusFailed && !task.optional {
			hasCriticalFailure = true
			break
		}
	}

	if hasCriticalFailure {
		return lipgloss.NewStyle().Foreground(ErrorColor).Render(
			"Installation failed.\nCheck errors above.\n\nPress Enter to exit")
	}

	// Success
	return fmt.Sprintf(`Installation complete.
Reboot to see sysc-greet.
Selected compositor: %s

`+lipgloss.NewStyle().Foreground(FgMuted).Render(">see you space cowboy")+`

Press Enter to exit`, m.selectedCompositor)
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome:
		return "Enter: Continue  •  Ctrl+C: Quit"
	case stepCompositorSelect:
		return "↑↓: Navigate  •  Enter: Select  •  Ctrl+C: Quit"
	case stepComplete:
		return "Enter: Exit  •  Ctrl+C: Quit"
	default:
		return "Ctrl+C: Cancel"
	}
}

func executeTask(index int, m *model) tea.Cmd {
	return func() tea.Msg {
		// Check if this is an optional task that should be skipped
		if m.tasks[index].optional {
			// Special handling for greetd installation
			if m.tasks[index].name == "Install greetd" && !m.needsGreetd {
				return taskCompleteMsg{
					index:   index,
					success: true,
					skipped: true,
				}
			}
			// Special handling for gslapper installation
			if m.tasks[index].name == "Install gslapper" {
				if _, err := exec.LookPath("gslapper"); err == nil {
					return taskCompleteMsg{
						index:   index,
						success: true,
						skipped: true,
					}
				}
			}
		}

		// Simulate work delay for visibility
		time.Sleep(200 * time.Millisecond)

		err := m.tasks[index].execute(m)

		if err != nil {
			return taskCompleteMsg{
				index:   index,
				success: false,
				error:   err.Error(),
			}
		}

		return taskCompleteMsg{
			index:   index,
			success: true,
		}
	}
}

func checkPrivileges(m *model) error {
	if os.Geteuid() != 0 {
		if _, err := exec.LookPath("sudo"); err != nil {
			return fmt.Errorf("root privileges required")
		}
	}
	return nil
}

func checkDependencies(m *model) error {
	missing := []string{}

	// Check critical deps
	if _, err := exec.LookPath("go"); err != nil {
		missing = append(missing, "go")
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		missing = append(missing, "systemd")
	}

	// Check compositor if selected
	if m.selectedCompositor != "" {
		if _, err := exec.LookPath(m.selectedCompositor); err != nil {
			missing = append(missing, m.selectedCompositor)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing: %s", strings.Join(missing, ", "))
	}

	// Detect package manager
	packageManagers := map[string][]string{
		"pacman": {"/usr/bin/pacman"},
		"apt":    {"/usr/bin/apt"},
		"dnf":    {"/usr/bin/dnf"},
		"yay":    {"/usr/bin/yay"},
	}

	for pmName, paths := range packageManagers {
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				m.packageManager = pmName
				break
			}
		}
		if m.packageManager != "" {
			break
		}
	}

	// Check if greetd installed
	_, err := exec.LookPath("greetd")
	m.greetdInstalled = (err == nil)
	m.needsGreetd = !m.greetdInstalled

	return nil
}

func installGreetd(m *model) error {
	if m.greetdInstalled {
		return nil // Already installed
	}

	if m.packageManager == "" {
		return fmt.Errorf("no package manager found")
	}

	var cmd *exec.Cmd
	switch m.packageManager {
	case "pacman":
		cmd = exec.Command("pacman", "-S", "--noconfirm", "greetd")
	case "yay":
		cmd = exec.Command("sudo", "-u", os.Getenv("SUDO_USER"), "yay", "-S", "--noconfirm", "greetd")
	case "apt":
		exec.Command("apt-get", "update").Run()
		cmd = exec.Command("apt-get", "install", "-y", "greetd")
	case "dnf":
		cmd = exec.Command("dnf", "install", "-y", "greetd")
	default:
		return fmt.Errorf("unsupported package manager: %s", m.packageManager)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("package install failed")
	}

	return nil
}

func installGslapper(m *model) error {
	// Check if already installed
	if _, err := exec.LookPath("gslapper"); err == nil {
		return nil
	}

	// Try AUR for Arch
	if m.packageManager == "yay" {
		cmd := exec.Command("sudo", "-u", os.Getenv("SUDO_USER"), "yay", "-S", "--noconfirm", "gslapper")
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try building from source
	return buildGslapperFromSource(m)
}

func buildGslapperFromSource(m *model) error {
	// Check for build deps
	deps := []string{"meson", "ninja", "git"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("missing build dependency: %s", dep)
		}
	}

	// Clone repo
	exec.Command("rm", "-rf", "/tmp/gslapper-build").Run()
	cloneCmd := exec.Command("git", "clone", "https://github.com/Nomadcxx/gSlapper", "/tmp/gslapper-build")
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("clone failed")
	}

	// Build
	setupCmd := exec.Command("meson", "setup", "build", "--prefix=/usr/local")
	setupCmd.Dir = "/tmp/gslapper-build"
	if err := setupCmd.Run(); err != nil {
		return fmt.Errorf("build setup failed")
	}

	buildCmd := exec.Command("ninja", "-C", "build")
	buildCmd.Dir = "/tmp/gslapper-build"
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed")
	}

	// Install
	installCmd := exec.Command("ninja", "-C", "build", "install")
	installCmd.Dir = "/tmp/gslapper-build"
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("install failed")
	}

	// Cleanup
	exec.Command("rm", "-rf", "/tmp/gslapper-build").Run()

	return nil
}

func buildBinary(m *model) error {
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", "sysc-greet", "./cmd/sysc-greet/")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed")
	}
	return nil
}

func installBinary(m *model) error {
	cmd := exec.Command("install", "-Dm755", "sysc-greet", "/usr/local/bin/sysc-greet")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install failed")
	}
	return nil
}

func installConfigs(m *model) error {
	configPath := "/usr/share/sysc-greet"

	// Create directories
	dirs := []string{
		configPath + "/ascii_configs",
		configPath + "/fonts",
		configPath + "/Assets",
		configPath + "/wallpapers",
	}

	for _, dir := range dirs {
		if err := exec.Command("mkdir", "-p", dir).Run(); err != nil {
			return fmt.Errorf("failed to create %s", dir)
		}
	}

	// Copy files
	copies := map[string]string{
		"ascii_configs/":            configPath + "/",
		"fonts/":                    configPath + "/",
		"config/kitty-greeter.conf": "/etc/greetd/kitty.conf",
	}

	// Optional copies
	if _, err := os.Stat("Assets"); err == nil {
		copies["Assets/"] = configPath + "/"
	}

	for src, dst := range copies {
		if err := exec.Command("cp", "-r", src, dst).Run(); err != nil {
			return fmt.Errorf("failed to copy %s", src)
		}
	}

	// Copy wallpapers if directory exists
	// FIXED 2025-10-17 - Always copy wallpapers directory if it exists
	if _, err := os.Stat("wallpapers"); err == nil {
		if err := exec.Command("cp", "-r", "wallpapers/", configPath+"/").Run(); err != nil {
			return fmt.Errorf("failed to copy wallpapers")
		}
	}

	return nil
}

func setupCache(m *model) error {
	// Create cache directory
	if err := exec.Command("mkdir", "-p", "/var/cache/sysc-greet").Run(); err != nil {
		return fmt.Errorf("cache dir creation failed")
	}

	// Create greeter home
	if err := exec.Command("mkdir", "-p", "/var/lib/greeter/Pictures/wallpapers").Run(); err != nil {
		return fmt.Errorf("greeter home creation failed")
	}

	// Create greeter user if needed
	// FIXED 2025-10-15 - Add render group for modern Intel/AMD iGPU support
	// Modern Linux uses 'render' group for /dev/dri/renderD* (non-privileged GPU access)
	// Both 'video' and 'render' groups needed for laptop iGPU compatibility
	cmd := exec.Command("id", "greeter")
	if err := cmd.Run(); err != nil {
		// User doesn't exist - create with video,render,input groups
		cmd = exec.Command("useradd", "-M", "-G", "video,render,input", "-s", "/usr/bin/nologin", "greeter")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("greeter user creation failed")
		}
	} else {
		// User exists - ensure they have required groups
		// CRITICAL: This fixes laptops where greeter user exists but lacks render group
		exec.Command("usermod", "-aG", "video,render,input", "greeter").Run()
	}

	// Set ownership
	paths := []string{"/var/cache/sysc-greet", "/var/lib/greeter"}
	for _, path := range paths {
		if err := exec.Command("chown", "-R", "greeter:greeter", path).Run(); err != nil {
			return fmt.Errorf("ownership change failed for %s", path)
		}
	}

	// Set permissions
	if err := exec.Command("chmod", "755", "/var/lib/greeter").Run(); err != nil {
		return fmt.Errorf("permissions change failed")
	}

	return nil
}

func getNiriConfig() string {
	return `// SYSC-Greet Niri config for greetd greeter session
// Monitors auto-detected by niri at runtime

hotkey-overlay {
    skip-at-startup
}

input {
    keyboard {
        xkb {
            layout "us"
        }
        repeat-delay 400
        repeat-rate 40
    }

    mouse {
    }

    touchpad {
        tap;
    }
}

layer-rule {
    match namespace="^wallpaper$"
    place-within-backdrop true
}

layout {
    gaps 0
    center-focused-column "never"

    focus-ring {
        off
    }

    border {
        off
    }
}

animations {
    off
}

window-rule {
    match app-id="kitty"
    opacity 0.90
}

spawn-at-startup "swww-daemon"

spawn-sh-at-startup "XDG_CACHE_HOME=/var/cache/sysc-greet HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"

binds {
}
`
}

func getHyprlandConfig() string {
	return `# SYSC-Greet Hyprland config for greetd greeter session
# Monitors auto-detected by Hyprland at runtime

# No animations for faster greeter startup
animations {
    enabled = false
}

# Minimal decorations
decoration {
    rounding = 0
    drop_shadow = false
    blur {
        enabled = false
    }
}

# Greeter doesn't need gaps
general {
    gaps_in = 0
    gaps_out = 0
    border_size = 0
}

# Input configuration
input {
    kb_layout = us
    repeat_delay = 400
    repeat_rate = 40

    touchpad {
        tap-to-click = true
    }
}

# Disable all keybindings (security for greeter)
# No binds = no user control

# Window rules for kitty greeter
windowrulev2 = opacity 0.90, class:^(kitty)$
windowrulev2 = fullscreen, class:^(kitty)$

# Layer rules for wallpaper daemon
layerrule = blur, wallpaper

# Startup applications
exec-once = swww-daemon
exec-once = kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch exit
`
}

func getSwayConfig() string {
	return `# SYSC-Greet Sway config for greetd greeter session
# Monitors auto-detected by Sway at runtime

# Disable window borders
default_border none
default_floating_border none

# No gaps needed for greeter
gaps inner 0
gaps outer 0

# Input configuration
input * {
    xkb_layout "us"
    repeat_delay 400
    repeat_rate 40
}

input type:touchpad {
    tap enabled
}

# Disable all keybindings (security)
# Empty config = no keys work

# Window rules for kitty (match any)
for_window [app_id="kitty"] opacity 0.90
for_window [app_id="kitty"] fullscreen enable

# Startup applications
exec swaybg -i /usr/share/sysc-greet/wallpapers/sysc-greet-default.png --mode fill
exec "kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; swaymsg exit"
`
}

func configureGreetd(m *model) error {
	var compositorConfig string
	var greetdCommand string

	switch m.selectedCompositor {
	case "niri":
		compositorConfig = getNiriConfig()
		greetdCommand = "niri -c /etc/greetd/niri-greeter-config.kdl"
		if err := os.WriteFile("/etc/greetd/niri-greeter-config.kdl", []byte(compositorConfig), 0644); err != nil {
			return fmt.Errorf("niri config write failed")
		}

	case "hyprland":
		compositorConfig = getHyprlandConfig()
		greetdCommand = "Hyprland -c /etc/greetd/hyprland-greeter-config.conf"
		if err := os.WriteFile("/etc/greetd/hyprland-greeter-config.conf", []byte(compositorConfig), 0644); err != nil {
			return fmt.Errorf("hyprland config write failed")
		}

	case "sway":
		compositorConfig = getSwayConfig()
		greetdCommand = "sway -c /etc/greetd/sway-greeter-config"
		if err := os.WriteFile("/etc/greetd/sway-greeter-config", []byte(compositorConfig), 0644); err != nil {
			return fmt.Errorf("sway config write failed")
		}
	}

	greetdConfig := fmt.Sprintf(`[terminal]
vt = 1

[default_session]
command = "%s"
user = "greeter"

[initial_session]
command = "%s"
user = "greeter"
`, greetdCommand, greetdCommand)

	if err := os.WriteFile("/etc/greetd/config.toml", []byte(greetdConfig), 0644); err != nil {
		return fmt.Errorf("greetd config write failed")
	}

	return nil
}

func enableService(m *model) error {
	// Remove existing display-manager.service symlink
	symlinkPath := "/etc/systemd/system/display-manager.service"
	if _, err := os.Lstat(symlinkPath); err == nil {
		os.Remove(symlinkPath)
	}

	// Enable greetd
	cmd := exec.Command("systemctl", "enable", "greetd.service")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("service enable failed")
	}

	return nil
}

// Monitor detection (simplified from original)
type Monitor struct {
	Name        string
	Width       int
	Height      int
	RefreshRate int
	Primary     bool
}

func parseNiriOutputs(output string) []Monitor {
	var monitors []Monitor
	lines := strings.Split(output, "\n")
	var current *Monitor

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Output") {
			if current != nil {
				monitors = append(monitors, *current)
			}
			current = &Monitor{}

			if start := strings.LastIndex(line, "("); start != -1 {
				if end := strings.LastIndex(line, ")"); end != -1 {
					current.Name = strings.TrimSpace(line[start+1 : end])
				}
			}
		}

		if current != nil && strings.Contains(line, "Current mode:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.Contains(part, "x") {
					dims := strings.Split(part, "x")
					if len(dims) == 2 {
						current.Width, _ = strconv.Atoi(dims[0])
						current.Height, _ = strconv.Atoi(dims[1])
					}
				}
				if part == "@" && i+1 < len(parts) {
					rate := strings.TrimSpace(parts[i+1])
					if f, err := strconv.ParseFloat(rate, 64); err == nil {
						current.RefreshRate = int(f)
					}
				}
			}
		}
	}

	if current != nil {
		monitors = append(monitors, *current)
	}

	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors
}

func main() {
	fmt.Println("Starting installer...")
	p := tea.NewProgram(newModel())
	fmt.Println("Program created, running...")
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Installer finished")
}
