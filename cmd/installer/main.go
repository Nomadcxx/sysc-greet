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
	checkMark  = lipgloss.NewStyle().Foreground(Accent).SetString("[OK]")
	failMark   = lipgloss.NewStyle().Foreground(ErrorColor).SetString("[FAIL]")
	skipMark   = lipgloss.NewStyle().Foreground(WarningColor).SetString("[SKIP]")
	headerStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
)

type installStep int

const (
	stepWelcome installStep = iota
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
	step             installStep
	tasks            []installTask
	currentTaskIndex int
	width            int
	height           int
	spinner          spinner.Model
	errors           []string
	packageManager   string
	greetdInstalled  bool
	needsGreetd      bool
	uninstallMode    bool
	selectedOption   int // 0 = Install, 1 = Uninstall
}

type taskCompleteMsg struct {
	index   int
	success bool
	error   string
}

func newModel() model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(Secondary)
	s.Spinner = spinner.Dot

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
		step:             stepWelcome,
		tasks:            tasks,
		currentTaskIndex: -1,
		spinner:          s,
		errors:           []string{},
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
			if m.step == stepComplete || m.step == stepWelcome {
				return m, tea.Quit
			}
		case "up", "k":
			if m.step == stepWelcome && m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.step == stepWelcome && m.selectedOption < 1 {
				m.selectedOption++
			}
		case "enter":
			if m.step == stepWelcome {
				// Set mode based on selection
				m.uninstallMode = (m.selectedOption == 1)

				// Set appropriate tasks
				if m.uninstallMode {
					m.tasks = []installTask{
						{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
						{name: "Disable service", description: "Disabling greetd service", execute: disableService, status: statusPending},
						{name: "Remove binary", description: "Removing sysc-greet binary", execute: removeBinary, status: statusPending},
						{name: "Remove configs", description: "Removing configurations", execute: removeConfigs, status: statusPending},
						{name: "Clean cache", description: "Cleaning cache directories", execute: cleanCache, optional: true, status: statusPending},
					}
				}

				// Start installation/uninstallation
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
		}

	case taskCompleteMsg:
		// Update task status
		if msg.success {
			m.tasks[msg.index].status = statusComplete
		} else {
			if m.tasks[msg.index].optional {
				m.tasks[msg.index].status = statusSkipped
				m.errors = append(m.errors, fmt.Sprintf("%s (skipped): %s", m.tasks[msg.index].name, msg.error))
			} else {
				m.tasks[msg.index].status = statusFailed
				m.errors = append(m.errors, fmt.Sprintf("%s: %s", m.tasks[msg.index].name, msg.error))
				m.step = stepComplete
				return m, nil
			}
		}

		// Move to next task
		m.currentTaskIndex++
		if m.currentTaskIndex >= len(m.tasks) {
			m.step = stepComplete
			return m, nil
		}

		// Start next task
		m.tasks[m.currentTaskIndex].status = statusRunning
		return m, executeTask(m.currentTaskIndex, &m)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
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
	title := "sysc-greet installer"
	if m.uninstallMode {
		title = "sysc-greet uninstaller"
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Main content based on step
	var mainContent string
	switch m.step {
	case stepWelcome:
		mainContent = m.renderWelcome()
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
	var b strings.Builder

	b.WriteString("Select an option:\n\n")

	// Install option
	installPrefix := "  "
	if m.selectedOption == 0 {
		installPrefix = lipgloss.NewStyle().Foreground(Primary).Render("▸ ")
	}
	b.WriteString(installPrefix + "Install sysc-greet\n")
	b.WriteString("    Builds binary, installs to system, configures greetd\n\n")

	// Uninstall option
	uninstallPrefix := "  "
	if m.selectedOption == 1 {
		uninstallPrefix = lipgloss.NewStyle().Foreground(Primary).Render("▸ ")
	}
	b.WriteString(uninstallPrefix + "Uninstall sysc-greet\n")
	b.WriteString("    Removes sysc-greet from your system\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("Requires root privileges"))

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
	if m.uninstallMode {
		return `Uninstall complete.
sysc-greet has been removed.

` + lipgloss.NewStyle().Foreground(FgMuted).Render(">see you space cowboy") + `

Press Enter to exit`
	}
	return `Installation complete.
Reboot to see sysc-greet.

` + lipgloss.NewStyle().Foreground(FgMuted).Render(">see you space cowboy") + `

Press Enter to exit`
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome:
		return "↑/↓: Navigate  •  Enter: Continue  •  Ctrl+C: Quit"
	case stepComplete:
		return "Enter: Exit  •  Ctrl+C: Quit"
	default:
		return "Ctrl+C: Cancel"
	}
}

func executeTask(index int, m *model) tea.Cmd {
	return func() tea.Msg {
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

// Task execution functions

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

func configureGreetd(m *model) error {
	niriConfig := `// SYSC-Greet Niri config for greetd greeter session
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

spawn-sh-at-startup "XDG_CACHE_HOME=/tmp/greeter-cache HOME=/var/lib/greeter kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"

binds {
}
`

	if err := os.WriteFile("/etc/greetd/niri-greeter-config.kdl", []byte(niriConfig), 0644); err != nil {
		return fmt.Errorf("niri config write failed")
	}

	greetdConfig := `[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"

[initial_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
`

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

// Uninstall functions

func disableService(m *model) error {
	// Disable greetd service
	if err := exec.Command("systemctl", "disable", "greetd.service").Run(); err != nil {
		// Not a critical error if it's already disabled
		return nil
	}
	return nil
}

func removeBinary(m *model) error {
	// Remove binary
	if err := exec.Command("rm", "-f", "/usr/local/bin/sysc-greet").Run(); err != nil {
		return fmt.Errorf("failed to remove binary")
	}
	return nil
}

func removeConfigs(m *model) error {
	// Remove configs and data
	paths := []string{
		"/usr/share/sysc-greet",
		"/etc/greetd/kitty.conf",
		"/etc/greetd/niri-greeter-config.kdl",
	}

	for _, path := range paths {
		exec.Command("rm", "-rf", path).Run()
	}

	return nil
}

func cleanCache(m *model) error {
	// Clean cache (optional - user might want to keep preferences)
	paths := []string{
		"/var/cache/sysc-greet",
	}

	for _, path := range paths {
		exec.Command("rm", "-rf", path).Run()
	}

	// Note: We don't remove /var/lib/greeter or the greeter user
	// as they might be used by other greeters

	return nil
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
