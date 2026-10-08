package main

import (
	"cmp"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SYSC ASCII header lines for beam animation
var asciiHeaderLines = []string{
	"  ██████████████░████░       ████░   ██████████████░  ██████████████░",
	"████████████████░████░       ████░ ████████████████░████████████████░",
	"████░            ██████░   ██████░ ████░            ████░            ",
	"██████████████░    ████████████░   ██████████████░  ████░            ",
	"  ██████████████░    ████████░       ██████████████░████░            ",
	"            ████░      ████░                   ████░████░            ",
	"████████████████░      ████░       ████████████████░████████████████░",
	"██████████████░        ████░       ██████████████░    ██████████████░",
	"           //////////////SEE YOU IN SPACE COWBOY//////////",
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// isArchBased checks if running on Arch or Arch-based distro
func isArchBased() bool {
	// Check for arch-release file
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return true
	}
	// Check for pacman
	if _, err := exec.LookPath("pacman"); err == nil {
		return true
	}
	return false
}

// detectAURHelper finds available AUR helper (yay > paru > none)
func detectAURHelper() string {
	helpers := []string{"yay", "paru"}
	for _, helper := range helpers {
		if path, err := exec.LookPath(helper); err == nil {
			return path
		}
	}
	return ""
}

// binaryToPackage maps binary names to package names for distros where they differ
// Key: package manager, Value: map of binary name to package name
var binaryToPackage = map[string]map[string]string{
	"apt": {
		"ninja": "ninja-build",
	},
	"dnf": {
		"ninja": "ninja-build",
	},
	"yum": {
		"ninja": "ninja-build",
	},
}

// getBinaryPackageName returns the package name for a binary, accounting for distro differences
func getBinaryPackageName(binary, packageManager string) string {
	if pmMap, ok := binaryToPackage[packageManager]; ok {
		if pkg, ok := pmMap[binary]; ok {
			return pkg
		}
	}
	return binary // Default: package name matches binary name
}

// checkPackageInstalled checks if a package is already installed
func checkPackageInstalled(pkg, packageManager string) bool {
	var cmd *exec.Cmd
	switch packageManager {
	case "pacman":
		cmd = exec.Command("pacman", "-Q", pkg)
	case "apt":
		cmd = exec.Command("dpkg-query", "-W", "-f=${Status}", pkg)
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), "install ok installed")
	case "dnf", "yum", "zypper":
		cmd = exec.Command("rpm", "-q", pkg)
	default:
		return false
	}
	return cmd.Run() == nil
}

// checkPackageExists checks if a package exists in repos (can be installed)
func checkPackageExists(pkg, packageManager string) bool {
	var cmd *exec.Cmd
	switch packageManager {
	case "pacman":
		cmd = exec.Command("pacman", "-Si", pkg)
	case "apt":
		cmd = exec.Command("apt-cache", "show", pkg)
	case "dnf":
		cmd = exec.Command("dnf", "info", pkg)
	case "yum":
		cmd = exec.Command("yum", "info", pkg)
	case "zypper":
		cmd = exec.Command("zypper", "info", pkg)
	default:
		return false
	}
	// Suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// packageCheckResult holds the categorized packages for installation
type packageCheckResult struct {
	toInstall        []string // Packages that exist and need installation
	alreadyInstalled []string // Packages already installed
	notFound         []string // Packages not found in repos
}

// getPackagesToInstall categorizes packages by their installation status
func getPackagesToInstall(packages []string, packageManager string) packageCheckResult {
	result := packageCheckResult{}
	for _, pkg := range packages {
		if checkPackageInstalled(pkg, packageManager) {
			result.alreadyInstalled = append(result.alreadyInstalled, pkg)
		} else if checkPackageExists(pkg, packageManager) {
			result.toInstall = append(result.toInstall, pkg)
		} else {
			result.notFound = append(result.notFound, pkg)
		}
	}
	return result
}

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

type installSubTask struct {
	name   string
	status taskStatus
}

type errorInfo struct {
	message string // Error message from command
	command string // Command that was run
	logFile string // Path to log file
}

type installTask struct {
	name           string
	description    string
	execute        func(*model) error
	optional       bool
	status         taskStatus
	subTasks       []installSubTask // Sub-tasks for complex operations
	currentSubTask int              // Index of current sub-task being executed
	errorDetails   *errorInfo       // Error details when task fails
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
	uninstallMode      bool
	selectedOption     int      // 0 = Install, 1 = Uninstall
	selectedCompositor string   // a greeterCompositors name, or "hyprland" via SYSC_COMPOSITOR only
	compositorIndex    int      // Current selection in compositor menu; -1 = SYSC_COMPOSITOR=hyprland
	debugMode          bool     // Show verbose output
	logFile            *os.File // Installer log file
	beams              *BeamsTextEffect
}

// greeterCompositor is one row of the compositor picker
type greeterCompositor struct {
	name        string
	label       string
	binaries    []string
	installable bool // installCompositor can install it when missing
}

// greeterCompositors is the picker. Hyprland is deliberately absent: it is
// deprecated and only reachable through SYSC_COMPOSITOR=hyprland.
var greeterCompositors = []greeterCompositor{
	{name: "niri", label: "Tiling compositor with scrollable workspaces (default)", binaries: []string{"niri"}},
	{name: "cagebreak", label: "Minimal tiling kiosk; replaces Hyprland for the greeter", binaries: []string{"cagebreak"}, installable: true},
	{name: "sway", label: "Stable i3-compatible tiling compositor", binaries: []string{"sway"}},
	{name: "mango", label: "dwl-based tiling compositor (Arch, Debian 13, Ubuntu 26.04, Fedora 43/44)", binaries: []string{"mango"}, installable: true},
}

const mangoUnsupported = "no mango package for this distro (Arch, Debian 13, Ubuntu 26.04, Fedora 43/44 only) — use niri, cagebreak, or sway"

// mangoAvailable reports whether installMango can install mango here
func mangoAvailable(packageManager string) bool {
	return packageManager == "pacman" || mangoArtifact() != ""
}

// compositorBinaries returns the executables that indicate name is installed
func compositorBinaries(name string) []string {
	if name == "hyprland" {
		return []string{"Hyprland", "hyprland"}
	}
	for _, c := range greeterCompositors {
		if c.name == name {
			return c.binaries
		}
	}
	return nil
}

func compositorInstalled(name string) bool {
	for _, bin := range compositorBinaries(name) {
		if _, err := exec.LookPath(bin); err == nil {
			return true
		}
	}
	return false
}

type taskCompleteMsg struct {
	index   int
	success bool
	error   string
}

type subTaskUpdateMsg struct {
	parentIndex  int
	subTaskIndex int
	status       taskStatus
}

func newModel(debugMode bool, logFile *os.File) model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(Secondary)
	s.Spinner = spinner.Dot

	tasks := []installTask{
		{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
		{name: "Check packages", description: "Checking for a distro-managed sysc-greet", execute: checkNoDistroPackage, status: statusPending},
		{name: "Check dependencies", description: "Checking system dependencies", execute: checkDependencies, status: statusPending},
		{name: "Install greetd", description: "Installing greetd daemon", execute: installGreetd, optional: false, status: statusPending},
		{name: "Install kitty", description: "Installing kitty terminal", execute: installKitty, optional: false, status: statusPending},
		{name: "Install compositor", description: "Installing Wayland compositor", execute: installCompositor, optional: false, status: statusPending},
		{
			name:        "Install gslapper",
			description: "Installing wallpaper daemon",
			execute:     installGslapper,
			optional:    false,
			status:      statusPending,
			subTasks: []installSubTask{
				{name: "Check existing installation", status: statusPending},
				{name: "AUR install (if pre-installed)", status: statusPending},
				{name: "Install GStreamer dependencies", status: statusPending},
				{name: "Clone repository", status: statusPending},
				{name: "Build from source", status: statusPending},
				{name: "Install binary", status: statusPending},
			},
		},
		{name: "Build binary", description: "Building sysc-greet", execute: buildBinary, status: statusPending},
		{name: "Install binary", description: "Installing to system", execute: installBinary, status: statusPending},
		{name: "Install configs", description: "Installing configurations", execute: installConfigs, status: statusPending},
		{name: "Setup cache", description: "Setting up cache and permissions", execute: setupCache, status: statusPending},
		{name: "Configure greetd", description: "Configuring greetd daemon", execute: configureGreetd, status: statusPending},
		{name: "Enable service", description: "Enabling greetd service", execute: enableService, status: statusPending},
	}

	// Initialize beams animation with ASCII header
	asciiHeader := strings.Join(asciiHeaderLines, "\n")
	beams := NewBeamsTextEffect(80, len(asciiHeaderLines)+2, asciiHeader)

	m := model{
		step:             stepWelcome,
		tasks:            tasks,
		currentTaskIndex: -1,
		spinner:          s,
		errors:           []string{},
		debugMode:        debugMode,
		logFile:          logFile,
		beams:            beams,
	}

	// Detect package manager during initialization (not in async task)
	detectPackageManager(&m)

	// Check for pre-selected compositor from environment variable
	// Preselect its picker row so Enter confirms it rather than resetting to niri
	if comp := os.Getenv("SYSC_COMPOSITOR"); comp != "" {
		m.step = stepCompositorSelect
		m.compositorIndex = -1
		for i, c := range greeterCompositors {
			if c.name == comp {
				m.compositorIndex = i
			}
		}
		if comp == "hyprland" {
			m.selectedCompositor = comp
		} else if m.compositorIndex == -1 {
			m.compositorIndex = 0
			m.errors = append(m.errors, fmt.Sprintf("SYSC_COMPOSITOR=%s is not a supported greeter compositor", comp))
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.beams != nil {
			m.beams.Resize(m.width, len(asciiHeaderLines)+2)
		}
		return m, nil

	case tickMsg:
		if m.beams != nil {
			m.beams.Update()
		}
		return m, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Allow exit from any step except during installation
			if m.step != stepInstalling {
				return m, tea.Quit
			}
			// During installation, Ctrl+C is disabled for safety
		case "up", "k":
			if m.step == stepWelcome && m.selectedOption > 0 {
				m.selectedOption--
			} else if m.step == stepCompositorSelect && m.compositorIndex > 0 {
				m.compositorIndex--
			}
		case "down", "j":
			if m.step == stepWelcome && m.selectedOption < 1 {
				m.selectedOption++
			} else if m.step == stepCompositorSelect && m.compositorIndex < len(greeterCompositors)-1 {
				m.compositorIndex++
			}
		case "enter":
			if m.step == stepWelcome {
				// Set mode based on selection
				m.uninstallMode = (m.selectedOption == 1)

				// Set appropriate tasks
				if m.uninstallMode {
					m.tasks = []installTask{
						{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
						{name: "Check packages", description: "Checking for a distro-managed sysc-greet", execute: checkNoDistroPackage, status: statusPending},
						{name: "Disable service", description: "Disabling greetd service", execute: disableService, status: statusPending},
						{name: "Remove binary", description: "Removing sysc-greet binary", execute: removeBinary, status: statusPending},
						{name: "Remove gslapper", description: "Removing wallpaper daemon", execute: uninstallGslapper, optional: true, status: statusPending},
						{name: "Remove configs", description: "Removing configurations", execute: removeConfigs, status: statusPending},
						{name: "Clean cache", description: "Cleaning cache directories", execute: cleanCache, optional: true, status: statusPending},
					}
					// Skip compositor selection for uninstall
					m.step = stepInstalling
					m.currentTaskIndex = 0
					m.tasks[0].status = statusRunning
					return m, tea.Batch(
						m.spinner.Tick,
						executeTask(0, &m),
					)
				} else {
					// Go to compositor selection
					m.step = stepCompositorSelect
					return m, nil
				}
			} else if m.step == stepCompositorSelect {
				// Index -1 keeps SYSC_COMPOSITOR=hyprland, which has no picker row
				installable := false
				if m.compositorIndex >= 0 {
					c := greeterCompositors[m.compositorIndex]
					m.selectedCompositor = c.name
					installable = c.installable
				}

				// Installable compositors are installed by installCompositor;
				// everything else must be present up front
				if !compositorInstalled(m.selectedCompositor) {
					if m.selectedCompositor == "mango" && !mangoAvailable(m.packageManager) {
						m.errors = append(m.errors, mangoUnsupported)
						return m, nil
					}
					if !installable {
						m.errors = append(m.errors, fmt.Sprintf("%s is not installed - please install it first", m.selectedCompositor))
						// Stay on compositor selection screen
						return m, nil
					}
				}

				// Start installation
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

	case subTaskUpdateMsg:
		// Update sub-task status
		if msg.parentIndex < len(m.tasks) {
			task := &m.tasks[msg.parentIndex]
			if msg.subTaskIndex < len(task.subTasks) {
				task.subTasks[msg.subTaskIndex].status = msg.status
				task.currentSubTask = msg.subTaskIndex
			}
		}
		return m, nil

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

	// ASCII Header with Beams Animation
	if m.beams != nil {
		content.WriteString(m.beams.Render())
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

func (m model) renderCompositorSelect() string {
	var b strings.Builder

	b.WriteString("Select Wayland compositor:\n\n")

	for i, comp := range greeterCompositors {
		prefix := "  "
		if i == m.compositorIndex {
			prefix = lipgloss.NewStyle().Foreground(Primary).Render("▸ ")
		}
		b.WriteString(prefix + comp.name + "\n")
		b.WriteString("    " + comp.label + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("Hyprland greeter support was removed from this installer; use cagebreak or niri"))
	if m.compositorIndex == -1 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(ErrorColor).Render("⚠ SYSC_COMPOSITOR=hyprland: Hyprland is unmaintained and will be removed in a future release. Enter continues anyway; ↓ picks another compositor."))
	}

	// Show errors if any
	if len(m.errors) > 0 {
		b.WriteString("\n\n")
		for _, err := range m.errors {
			b.WriteString(lipgloss.NewStyle().Foreground(ErrorColor).Render("⚠ " + err))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m model) renderInstalling() string {
	var b strings.Builder

	// Render all tasks with their current status
	for _, task := range m.tasks {
		// Render parent task
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
		b.WriteString(line + "\n")

		// Render sub-tasks if present
		if len(task.subTasks) > 0 {
			for j, subTask := range task.subTasks {
				isLast := (j == len(task.subTasks)-1)
				prefix := "  ├─ "
				if isLast {
					prefix = "  └─ "
				}

				var subLine string
				switch subTask.status {
				case statusPending:
					subLine = lipgloss.NewStyle().Foreground(FgMuted).Render(subTask.name)
				case statusRunning:
					subLine = m.spinner.View() + " " + subTask.name
				case statusComplete:
					subLine = checkMark.String() + " " + subTask.name
				case statusFailed:
					subLine = failMark.String() + " " + subTask.name
				case statusSkipped:
					subLine = skipMark.String() + " " + subTask.name
				}

				b.WriteString(prefix + subLine + "\n")
			}
		}

		// Show error details for failed tasks
		if task.status == statusFailed && task.errorDetails != nil {
			err := task.errorDetails
			b.WriteString(lipgloss.NewStyle().Foreground(ErrorColor).Render(
				fmt.Sprintf("  └─ Error: %s\n", err.message)))
			if err.command != "" {
				b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render(
					fmt.Sprintf("  └─ Command: %s\n", err.command)))
			}
			b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render(
				fmt.Sprintf("  └─ See full logs: %s\n", err.logFile)))
		}
	}

	b.WriteString(m.renderErrors())

	return b.String()
}

// renderErrors lists task errors and skipped-task warnings
func (m model) renderErrors() string {
	var b strings.Builder
	if len(m.errors) > 0 {
		b.WriteString("\n")
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

	// The task list (with its errors) is only drawn while installing, so
	// repeat it here or a failure shows no reason
	if hasCriticalFailure {
		failed := "Installation failed."
		if m.uninstallMode {
			failed = "Uninstall failed."
		}
		return m.renderInstalling() + "\n" + lipgloss.NewStyle().Foreground(ErrorColor).Render(
			failed+"\nFull log: /tmp/sysc-greet-installer.log\n\nPress Enter to exit")
	}

	// Success
	if m.uninstallMode {
		return `Uninstall complete.
sysc-greet has been removed.
` + m.renderErrors() + `
` + lipgloss.NewStyle().Foreground(FgMuted).Render(">see you space cowboy") + `

Press Enter to exit`
	}
	return `Installation complete.
Reboot to see sysc-greet.
` + m.renderErrors() + `
` + lipgloss.NewStyle().Foreground(FgMuted).Render(">see you space cowboy") + `

Press Enter to exit`
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome:
		return "↑/↓: Navigate  •  Enter: Continue  •  Ctrl+C: Quit"
	case stepCompositorSelect:
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

		if index == 0 {
			mode := "install"
			if m.uninstallMode {
				mode = "uninstall"
			}
			m.logf("Mode: %s, compositor: %s, package manager: %s", mode, m.selectedCompositor, m.packageManager)
		}
		m.logf("[%s] Started", m.tasks[index].name)

		err := m.tasks[index].execute(m) // Pass pointer so model changes persist

		if err != nil {
			m.logf("[%s] Failed: %v", m.tasks[index].name, err)
			if m.debugMode {
				fmt.Fprintf(os.Stderr, "\n[DEBUG] Task '%s' failed: %v\n", m.tasks[index].name, err)
			}
			return taskCompleteMsg{
				index:   index,
				success: false,
				error:   err.Error(),
			}
		}

		m.logf("[%s] Done", m.tasks[index].name)
		return taskCompleteMsg{
			index:   index,
			success: true,
		}
	}
}

// logf writes a timestamped line to the installer log
func (m *model) logf(format string, args ...any) {
	if m.logFile == nil {
		return
	}
	fmt.Fprintf(m.logFile, "[%s] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	m.logFile.Sync()
}

// updateSubTaskStatus updates the status of a sub-task for the current task
// Note: In a Bubble Tea app, this doesn't immediately update the UI - it just modifies the model.
// The UI will update on the next render cycle.
func updateSubTaskStatus(m *model, subTaskIndex int, status taskStatus) {
	if m.currentTaskIndex >= 0 && m.currentTaskIndex < len(m.tasks) {
		task := &m.tasks[m.currentTaskIndex]
		if subTaskIndex >= 0 && subTaskIndex < len(task.subTasks) {
			task.subTasks[subTaskIndex].status = status
		}
	}
}

// runCommand executes a command and logs it to the installer log file
func runCommand(taskName string, cmd *exec.Cmd, m *model) error {
	if m.logFile != nil {
		// Log command before execution
		cmdStr := cmd.String()
		logEntry := fmt.Sprintf("[%s] [%s] Running: %s\n",
			time.Now().Format("15:04:05"), taskName, cmdStr)
		m.logFile.WriteString(logEntry)
		m.logFile.Sync() // Flush to disk
	}

	// Capture output
	output, err := cmd.CombinedOutput()

	// Log output
	if m.logFile != nil {
		if len(output) > 0 {
			m.logFile.Write(output)
			m.logFile.WriteString("\n")
		}
		if err != nil {
			m.logFile.WriteString(fmt.Sprintf("[%s] [%s] Error: %v\n\n",
				time.Now().Format("15:04:05"), taskName, err))
		} else {
			m.logFile.WriteString(fmt.Sprintf("[%s] [%s] Success\n\n",
				time.Now().Format("15:04:05"), taskName))
		}
		m.logFile.Sync()
	}

	return err
}

// Task execution functions

func checkPrivileges(m *model) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required - run with sudo")
	}
	return nil
}

// checkNoDistroPackage stops before any file is touched when a distro package
// (AUR variant or release .deb/.rpm) owns sysc-greet. Installing or
// uninstalling over it leaves the package database pointing at overwritten or
// deleted files.
func checkNoDistroPackage(m *model) error {
	pkg, remove := installedDistroPackage(m.packageManager)
	if pkg == "" {
		return nil
	}
	return fmt.Errorf("%s is installed as a system package; remove it first with: sudo %s", pkg, remove)
}

// installedDistroPackage returns the installed sysc-greet package and the
// command that removes it, or "" when sysc-greet is not package-managed
func installedDistroPackage(packageManager string) (pkg, remove string) {
	switch packageManager {
	case "pacman":
		out, _ := exec.Command("pacman", "-Qq").Output()
		for _, name := range strings.Fields(string(out)) {
			if (name == "sysc-greet" || strings.HasPrefix(name, "sysc-greet-")) && !strings.HasSuffix(name, "-debug") {
				return name, "pacman -R " + name
			}
		}
	case "apt":
		out, _ := exec.Command("dpkg-query", "-W", "-f=${Status}", "sysc-greet").Output()
		if strings.Contains(string(out), "install ok installed") {
			return "sysc-greet", "apt remove sysc-greet"
		}
	case "dnf", "yum", "zypper":
		if exec.Command("rpm", "-q", "sysc-greet").Run() == nil {
			return "sysc-greet", packageManager + " remove sysc-greet"
		}
	}
	return "", ""
}

func detectPackageManager(m *model) {
	// Detect package manager (order matters - check base PM first, then helpers)
	// Priority: native package managers first, then AUR helpers
	packageManagers := []struct {
		name string
		path string
	}{
		{"pacman", "/usr/bin/pacman"},
		{"apt", "/usr/bin/apt"},
		{"apt", "/usr/bin/apt-get"}, // Fallback for older Debian/Ubuntu
		{"dnf", "/usr/bin/dnf"},
		{"yum", "/usr/bin/yum"},       // Older Fedora/RHEL
		{"zypper", "/usr/bin/zypper"}, // openSUSE
	}

	for _, pm := range packageManagers {
		if _, err := os.Stat(pm.path); err == nil {
			m.packageManager = pm.name
			if m.debugMode {
				fmt.Fprintf(os.Stderr, "[DEBUG] Detected package manager: %s at %s\n", pm.name, pm.path)
			}
			break
		}
	}

	if m.packageManager == "" && m.debugMode {
		fmt.Fprintf(os.Stderr, "[DEBUG] No package manager detected!\n")
	}

	// Check if greetd installed
	_, err := exec.LookPath("greetd")
	m.greetdInstalled = (err == nil)
	m.needsGreetd = !m.greetdInstalled
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

	return nil
}

func installGreetd(m *model) error {
	if m.greetdInstalled {
		return nil // Already installed - task will succeed silently
	}

	if m.packageManager == "" {
		return fmt.Errorf("package manager not detected - install greetd manually")
	}

	var cmd *exec.Cmd
	var updateCmd *exec.Cmd

	switch m.packageManager {
	case "pacman":
		// Try AUR helper first if available, fall back to pacman
		if _, err := exec.LookPath("yay"); err == nil {
			// Use yay for AUR access (greetd might be in AUR)
			// IMPORTANT: yay must NOT be run as root, but needs sudo internally
			if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && sudoUser != "root" {
				// Drop privileges to run yay as the original user
				// yay will ask for sudo password when needed
				cmd = exec.Command("su", "-", sudoUser, "-c", "yay -S --noconfirm greetd")
			} else {
				// Fallback to pacman if we can't determine non-root user
				cmd = exec.Command("pacman", "-S", "--noconfirm", "greetd")
			}
		} else if _, err := exec.LookPath("paru"); err == nil {
			// Alternative AUR helper
			// IMPORTANT: paru must NOT be run as root, but needs sudo internally
			if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "root" && sudoUser != "" {
				cmd = exec.Command("su", "-", sudoUser, "-c", "paru -S --noconfirm greetd")
			} else {
				// Fallback to pacman if we can't determine non-root user
				cmd = exec.Command("pacman", "-S", "--noconfirm", "greetd")
			}
		} else {
			// Standard pacman (official repos only)
			cmd = exec.Command("pacman", "-S", "--noconfirm", "greetd")
		}

	case "apt":
		// Update package list first for apt-based systems
		updateCmd = exec.Command("apt-get", "update")
		updateCmd.Run() // Ignore errors, proceed anyway
		cmd = exec.Command("apt-get", "install", "-y", "greetd")

	case "dnf":
		cmd = exec.Command("dnf", "install", "-y", "greetd")

	case "yum":
		cmd = exec.Command("yum", "install", "-y", "greetd")

	case "zypper":
		cmd = exec.Command("zypper", "install", "-y", "greetd")

	default:
		return fmt.Errorf("unsupported package manager '%s' - install greetd manually", m.packageManager)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install greetd (try: manual installation)")
	}

	return nil
}

func installKitty(m *model) error {
	// Check if already installed
	if _, err := exec.LookPath("kitty"); err == nil {
		return nil // Already installed
	}

	if m.packageManager == "" {
		return fmt.Errorf("package manager not detected - install kitty manually")
	}

	var cmd *exec.Cmd

	switch m.packageManager {
	case "pacman":
		cmd = exec.Command("pacman", "-S", "--noconfirm", "kitty")

	case "apt":
		cmd = exec.Command("apt-get", "install", "-y", "kitty")

	case "dnf":
		cmd = exec.Command("dnf", "install", "-y", "kitty")

	case "yum":
		cmd = exec.Command("yum", "install", "-y", "kitty")

	case "zypper":
		cmd = exec.Command("zypper", "install", "-y", "kitty")

	default:
		return fmt.Errorf("unsupported package manager '%s' - install kitty manually", m.packageManager)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install kitty")
	}

	return nil
}

// installCagebreakArch installs cagebreak on Arch: official repos first, then
// an AUR helper run as the invoking user (AUR helpers refuse to run as root).
func installCagebreakArch() error {
	if err := exec.Command("pacman", "-S", "--noconfirm", "cagebreak").Run(); err == nil {
		return nil
	}
	helperPath := detectAURHelper()
	if helperPath == "" {
		return fmt.Errorf("cagebreak is AUR-only and no AUR helper found — install it first: paru -S cagebreak (or yay -S cagebreak)")
	}
	var cmd *exec.Cmd
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		cmd = exec.Command("sudo", "-u", sudoUser, helperPath, "-S", "--noconfirm", "cagebreak")
	} else {
		cmd = exec.Command(helperPath, "-S", "--noconfirm", "cagebreak")
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cagebreak AUR install failed — install it manually with %s -S cagebreak", helperPath)
	}
	return nil
}

// ensureSocat installs socat, which the cagebreak greeter config needs to quit
// the compositor after login (echo quit | socat - UNIX-CONNECT:$CAGEBREAK_SOCKET).
func ensureSocat(m *model) error {
	if _, err := exec.LookPath("socat"); err == nil {
		return nil
	}
	var cmd *exec.Cmd
	switch m.packageManager {
	case "pacman":
		cmd = exec.Command("pacman", "-S", "--noconfirm", "socat")
	case "apt":
		cmd = exec.Command("apt-get", "install", "-y", "socat")
	case "dnf":
		cmd = exec.Command("dnf", "install", "-y", "socat")
	case "yum":
		cmd = exec.Command("yum", "install", "-y", "socat")
	case "zypper":
		cmd = exec.Command("zypper", "install", "-y", "socat")
	default:
		return fmt.Errorf("socat is required for cagebreak — install it manually")
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install socat (cagebreak cannot quit after login without it)")
	}
	return nil
}

// cagebreakArtifact returns the prebuilt cagebreak package from sysc-greet
// releases matching this distro, or "" when none matches (source build instead).
// Cagebreak pins wlroots per tag, so each artifact is built against one
// distro's wlroots — matrix in docs/superpowers/plans/2026-07-06-cagebreak-distro-packaging-plan.md.
func cagebreakArtifact() string {
	if runtime.GOARCH != "amd64" {
		return ""
	}
	id, version := osRelease()
	switch id {
	case "ubuntu":
		if version == "24.04" {
			return "cagebreak_2.3.1_ubuntu24.04_amd64.deb"
		}
		// Ubuntu 25.x ships wlroots 0.18 under the same package names as trixie
		if strings.HasPrefix(version, "25.") {
			return "cagebreak_2.4.0_debian13_amd64.deb"
		}
	case "debian":
		if version == "13" {
			return "cagebreak_2.4.0_debian13_amd64.deb"
		}
	case "fedora":
		if version == "42" {
			return "cagebreak-3.1.0-1.fedora42.x86_64.rpm"
		}
	}
	return ""
}

// mangoPackageVersion must match MANGO_VERSION in scripts/build-mango.sh
const mangoPackageVersion = "0.17.4"

// mangoArtifact returns the mangowm package attached to sysc-greet releases
// for this distro (built by scripts/build-mango.sh in CI), or "" when none
// exists. Arch uses extra/mangowm instead.
func mangoArtifact() string {
	if runtime.GOARCH != "amd64" {
		return ""
	}
	id, version := osRelease()
	switch {
	case id == "debian" && version == "13", id == "ubuntu" && version == "26.04":
		return fmt.Sprintf("mangowm_%s_%s%s_amd64.deb", mangoPackageVersion, id, version)
	case id == "fedora" && (version == "43" || version == "44"):
		return fmt.Sprintf("mangowm-%s-1.fedora%s.x86_64.rpm", mangoPackageVersion, version)
	}
	return ""
}

// Release codenames of the Ubuntu and Debian versions that have sysc-greet
// release packages, so derivatives can resolve to their base release.
var (
	ubuntuReleases = map[string]string{"noble": "24.04", "plucky": "25.04", "questing": "25.10", "resolute": "26.04"}
	debianReleases = map[string]string{"trixie": "13"}
)

// osRelease returns the distro ID and VERSION_ID from /etc/os-release,
// resolving Ubuntu and Debian derivatives to their base release
func osRelease() (id, version string) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", ""
	}
	return parseOSRelease(string(data))
}

// parseOSRelease maps derivatives (Mint, Pop!_OS, ...) to the Ubuntu or Debian
// release underneath through ID_LIKE and the base codename. Derivatives whose
// base release is unknown keep their own ID and VERSION_ID.
func parseOSRelease(data string) (id, version string) {
	fields := map[string]string{}
	for _, line := range strings.Split(data, "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			fields[k] = strings.Trim(v, `"`)
		}
	}
	like := strings.Fields(fields["ID_LIKE"])
	switch {
	case slices.Contains(like, "ubuntu"):
		if v, ok := ubuntuReleases[fields["UBUNTU_CODENAME"]]; ok {
			return "ubuntu", v
		}
	case slices.Contains(like, "debian"):
		codename := cmp.Or(fields["DEBIAN_CODENAME"], fields["VERSION_CODENAME"])
		if v, ok := debianReleases[codename]; ok {
			return "debian", v
		}
	}
	return fields["ID"], fields["VERSION_ID"]
}

// installReleasePackage downloads a package attached to the latest
// sysc-greet release and installs it with apt or dnf
func installReleasePackage(m *model, artifact string) error {
	var cmd *exec.Cmd
	path := "/tmp/" + artifact
	switch m.packageManager {
	case "apt":
		cmd = exec.Command("apt-get", "install", "-y", path)
	case "dnf":
		cmd = exec.Command("dnf", "install", "-y", path)
	default:
		return fmt.Errorf("release packages need apt or dnf, not %s", m.packageManager)
	}
	url := "https://github.com/Nomadcxx/sysc-greet/releases/latest/download/" + artifact
	if err := downloadFile(url, path); err != nil {
		return fmt.Errorf("download %s: %v", artifact, err)
	}
	defer os.Remove(path)
	return runCommand("Install "+artifact, cmd, m)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// installCagebreakPrebuilt installs the cagebreak package attached to the
// latest sysc-greet release; any miss falls back to building from source.
func installCagebreakPrebuilt(m *model) error {
	artifact := cagebreakArtifact()
	if artifact == "" {
		return buildCagebreakFromSource(m)
	}
	if err := installReleasePackage(m, artifact); err != nil {
		return buildCagebreakFromSource(m)
	}
	return nil
}

// getCagebreakBuildDeps returns distro package names for building cagebreak.
// The apt wlroots dev package is resolved separately (name carries the version
// on trixie: libwlroots-0.18-dev vs plain libwlroots-dev on Ubuntu 24.04).
func getCagebreakBuildDeps(packageManager string) []string {
	switch packageManager {
	case "apt":
		return []string{
			"git", "ca-certificates", "build-essential", "meson", "ninja-build", "pkg-config",
			"libwayland-dev", "wayland-protocols", "libxkbcommon-dev", "libinput-dev",
			"libcairo2-dev", "libpango1.0-dev", "libfontconfig-dev", "libevdev-dev",
			"libudev-dev", "libpixman-1-dev",
		}
	case "dnf", "yum":
		return []string{
			"git", "gcc", "meson", "ninja-build", "pkgconf-pkg-config",
			"wlroots-devel", "wayland-devel", "wayland-protocols-devel", "libxkbcommon-devel",
			"libinput-devel", "cairo-devel", "pango-devel", "fontconfig-devel",
			"libevdev-devel", "systemd-devel", "pixman-devel",
		}
	case "zypper":
		return []string{
			"git", "gcc", "meson", "ninja", "pkg-config",
			"wlroots-devel", "wayland-devel", "wayland-protocols-devel", "libxkbcommon-devel",
			"libinput-devel", "cairo-devel", "pango-devel", "fontconfig-devel",
			"libevdev-devel", "systemd-devel", "pixman-devel",
		}
	}
	return nil
}

// cagebreakTagForWlroots maps the installed wlroots dev headers to the newest
// cagebreak tag that builds against them (no meson subproject fallback upstream).
func cagebreakTagForWlroots() (string, error) {
	probes := []struct{ module, tag string }{
		{"wlroots-0.20", "3.2.1"},
		{"wlroots-0.19", "3.1.0"},
		{"wlroots-0.18", "2.4.0"},
		{"wlroots", "2.3.1"}, // 0.17 uses the unversioned pkg-config name
	}
	for _, p := range probes {
		if exec.Command("pkg-config", "--exists", p.module).Run() != nil {
			continue
		}
		if p.module == "wlroots" {
			out, err := exec.Command("pkg-config", "--modversion", "wlroots").Output()
			if err != nil || !strings.HasPrefix(strings.TrimSpace(string(out)), "0.17") {
				continue
			}
		}
		return p.tag, nil
	}
	return "", fmt.Errorf("no usable wlroots headers found (cagebreak needs wlroots 0.17–0.20)")
}

// buildCagebreakFromSource mirrors buildGslapperFromSource: install build deps,
// pick the cagebreak tag matching the system wlroots, meson build, install.
func buildCagebreakFromSource(m *model) error {
	deps := getCagebreakBuildDeps(m.packageManager)
	if deps == nil {
		return fmt.Errorf("cagebreak not packaged for this distro — build it from https://github.com/project-repo/cagebreak")
	}

	if m.packageManager == "apt" {
		// Resolve the versioned wlroots dev package name
		wlrootsDev := ""
		for _, candidate := range []string{"libwlroots-0.20-dev", "libwlroots-0.19-dev", "libwlroots-0.18-dev", "libwlroots-dev"} {
			if checkPackageExists(candidate, "apt") {
				wlrootsDev = candidate
				break
			}
		}
		if wlrootsDev == "" {
			return fmt.Errorf("no libwlroots dev package in apt repos — cagebreak cannot be built here")
		}
		deps = append(deps, wlrootsDev)
	}

	var cmd *exec.Cmd
	switch m.packageManager {
	case "apt":
		cmd = exec.Command("apt-get", append([]string{"install", "-y"}, deps...)...)
	case "dnf":
		cmd = exec.Command("dnf", append([]string{"install", "-y"}, deps...)...)
	case "yum":
		cmd = exec.Command("yum", append([]string{"install", "-y"}, deps...)...)
	case "zypper":
		cmd = exec.Command("zypper", append([]string{"install", "-y"}, deps...)...)
	}
	if cmd != nil {
		runCommand("Install cagebreak build deps", cmd, m) // Best effort; verified below
	}

	for _, tool := range []string{"meson", "ninja", "git", "pkg-config"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("missing build tool: %s", tool)
		}
	}

	tag, err := cagebreakTagForWlroots()
	if err != nil {
		return err
	}

	exec.Command("rm", "-rf", "/tmp/cagebreak-build").Run()
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "--branch", tag,
		"https://github.com/project-repo/cagebreak", "/tmp/cagebreak-build")
	if err := runCommand("Clone cagebreak", cloneCmd, m); err != nil {
		return fmt.Errorf("cagebreak clone failed")
	}

	setupCmd := exec.Command("meson", "setup", "build")
	setupCmd.Dir = "/tmp/cagebreak-build"
	if err := runCommand("Configure cagebreak build", setupCmd, m); err != nil {
		return fmt.Errorf("cagebreak build setup failed — check wlroots dev packages")
	}

	buildCmd := exec.Command("ninja", "-C", "build")
	buildCmd.Dir = "/tmp/cagebreak-build"
	if err := runCommand("Build cagebreak", buildCmd, m); err != nil {
		return fmt.Errorf("cagebreak build failed")
	}

	installCmd := exec.Command("install", "-Dm755", "build/cagebreak", "/usr/local/bin/cagebreak")
	installCmd.Dir = "/tmp/cagebreak-build"
	if err := runCommand("Install cagebreak", installCmd, m); err != nil {
		return fmt.Errorf("cagebreak install failed")
	}

	exec.Command("rm", "-rf", "/tmp/cagebreak-build").Run()
	return nil
}

// installMango installs mangowm (mango + mmsg): Arch extra, or the package
// attached to sysc-greet releases. The greeter session script quits mango
// with mmsg, so both binaries are required.
func installMango(m *model) error {
	if !compositorInstalled("mango") {
		switch {
		case m.packageManager == "pacman":
			if err := runCommand("Install mangowm", exec.Command("pacman", "-S", "--noconfirm", "mangowm"), m); err != nil {
				return fmt.Errorf("failed to install mangowm - install it manually: pacman -S mangowm")
			}
		case mangoArtifact() != "":
			if err := installReleasePackage(m, mangoArtifact()); err != nil {
				return fmt.Errorf("mangowm package install failed: %v", err)
			}
		default:
			return fmt.Errorf("%s", mangoUnsupported)
		}
	}
	if _, err := exec.LookPath("mmsg"); err != nil {
		return fmt.Errorf("mmsg not found - the greeter cannot quit mango after login (reinstall mangowm)")
	}
	return nil
}

func installCompositor(m *model) error {
	// The cagebreak greeter config quits the compositor via socat, even when
	// cagebreak itself is already installed
	if m.selectedCompositor == "cagebreak" {
		if err := ensureSocat(m); err != nil {
			return err
		}
	}

	if m.selectedCompositor == "mango" {
		return installMango(m)
	}

	if compositorInstalled(m.selectedCompositor) {
		return nil
	}

	if m.packageManager == "" {
		return fmt.Errorf("package manager not detected - install %s manually", m.selectedCompositor)
	}

	var cmd *exec.Cmd

	// Map compositor names to package names per distro
	switch m.packageManager {
	case "pacman":
		// All three compositors available in Arch official/AUR
		switch m.selectedCompositor {
		case "niri":
			// niri is in AUR, try yay/paru first
			if _, err := exec.LookPath("yay"); err == nil {
				if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && sudoUser != "root" {
					cmd = exec.Command("su", "-", sudoUser, "-c", "yay -S --noconfirm niri")
				} else {
					return fmt.Errorf("niri requires AUR access - install manually")
				}
			} else if _, err := exec.LookPath("paru"); err == nil {
				if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && sudoUser != "root" {
					cmd = exec.Command("su", "-", sudoUser, "-c", "paru -S --noconfirm niri")
				} else {
					return fmt.Errorf("niri requires AUR access - install manually")
				}
			} else {
				return fmt.Errorf("niri requires AUR helper (yay/paru) - install manually")
			}
		case "hyprland":
			cmd = exec.Command("pacman", "-S", "--noconfirm", "hyprland")
		case "sway":
			cmd = exec.Command("pacman", "-S", "--noconfirm", "sway")
		case "cagebreak":
			return installCagebreakArch()
		}

	case "apt":
		// Debian/Ubuntu
		switch m.selectedCompositor {
		case "niri":
			return fmt.Errorf("niri not available in apt repos - build from source manually")
		case "hyprland":
			return fmt.Errorf("hyprland not in standard apt repos - see https://hyprland.org for installation")
		case "sway":
			cmd = exec.Command("apt-get", "install", "-y", "sway")
		case "cagebreak":
			return installCagebreakPrebuilt(m)
		}

	case "dnf":
		// Fedora
		switch m.selectedCompositor {
		case "niri":
			return fmt.Errorf("niri not available in dnf repos - build from source manually")
		case "hyprland":
			return fmt.Errorf("hyprland not in standard dnf repos - see https://hyprland.org for installation")
		case "sway":
			cmd = exec.Command("dnf", "install", "-y", "sway")
		case "cagebreak":
			return installCagebreakPrebuilt(m)
		}

	case "yum":
		// RHEL/CentOS
		switch m.selectedCompositor {
		case "sway":
			cmd = exec.Command("yum", "install", "-y", "sway")
		default:
			return fmt.Errorf("%s not available via yum - install manually", m.selectedCompositor)
		}

	case "zypper":
		// openSUSE
		switch m.selectedCompositor {
		case "hyprland":
			return fmt.Errorf("hyprland may require Tumbleweed or manual install")
		case "sway":
			cmd = exec.Command("zypper", "install", "-y", "sway")
		case "cagebreak":
			// Packaged in Tumbleweed; source build covers anything else
			if err := exec.Command("zypper", "install", "-y", "cagebreak").Run(); err != nil {
				return buildCagebreakFromSource(m)
			}
			return nil
		default:
			return fmt.Errorf("%s may not be in zypper repos - install manually", m.selectedCompositor)
		}

	default:
		return fmt.Errorf("unsupported package manager '%s' - install %s manually", m.packageManager, m.selectedCompositor)
	}

	if cmd == nil {
		return fmt.Errorf("%s installation command not configured", m.selectedCompositor)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s - you may need to install manually", m.selectedCompositor)
	}

	return nil
}

func installGslapper(m *model) error {
	// Sub-task 0: Check if already installed
	updateSubTaskStatus(m, 0, statusRunning)

	// Check if gslapper is already installed (via AUR or otherwise)
	if _, err := exec.LookPath("gslapper"); err == nil {
		updateSubTaskStatus(m, 0, statusComplete)
		// Mark remaining sub-tasks as skipped
		for i := 1; i < 6; i++ {
			updateSubTaskStatus(m, i, statusSkipped)
		}
		return nil
	}
	updateSubTaskStatus(m, 0, statusComplete)

	// Sub-task 1: Skip AUR (nested sudo doesn't work reliably)
	// Users can pre-install gslapper from AUR if they prefer
	updateSubTaskStatus(m, 1, statusSkipped)

	// Build from source (works reliably on all distros)
	return buildGslapperFromSource(m)
}

// getGStreamerDeps returns the distro-specific GStreamer package names for building gSlapper
// gSlapper dependencies from meson.build:
//   - gstreamer-1.0, gstreamer-video-1.0, gstreamer-gl-1.0
//   - wayland-client, wayland-egl, wayland-protocols
//   - egl
//
// NOTE: gSlapper does NOT use GTK4 - it uses wlr-layer-shell protocol directly
func getGStreamerDeps(packageManager string) []string {
	switch packageManager {
	case "pacman":
		// Arch Linux - packages in official repos
		return []string{
			"gstreamer",
			"gst-plugins-base",
			"gst-plugins-good",
			"gst-plugins-bad",
			"wayland",
			"wayland-protocols",
			"egl-wayland",
		}
	case "apt":
		// Debian/Ubuntu - note the different naming convention (gstreamer1.0-*)
		return []string{
			"libgstreamer1.0-dev",
			"libgstreamer-plugins-base1.0-dev",
			"gstreamer1.0-plugins-base",
			"gstreamer1.0-plugins-good",
			"gstreamer1.0-plugins-bad",
			"libegl1-mesa-dev",
			"libwayland-dev",
			"libwayland-egl-backend-dev",
			"wayland-protocols",
		}
	case "dnf":
		// Fedora - note the different naming convention (gstreamer1-*)
		return []string{
			"gstreamer1-devel",
			"gstreamer1-plugins-base-devel",
			"gstreamer1-plugins-base",
			"gstreamer1-plugins-good",
			"gstreamer1-plugins-bad-free",
			"mesa-libEGL-devel",
			"wayland-devel",
			"wayland-protocols-devel",
		}
	case "yum":
		// RHEL/CentOS - similar to Fedora
		return []string{
			"gstreamer1-devel",
			"gstreamer1-plugins-base-devel",
			"mesa-libEGL-devel",
			"wayland-devel",
		}
	case "zypper":
		// openSUSE
		return []string{
			"gstreamer-devel",
			"gstreamer-plugins-base-devel",
			"Mesa-libEGL-devel",
			"wayland-devel",
			"wayland-protocols-devel",
		}
	default:
		return []string{}
	}
}

// installGStreamerDeps installs GStreamer dependencies for building gSlapper
func installGStreamerDeps(m *model) error {
	deps := getGStreamerDeps(m.packageManager)
	if len(deps) == 0 {
		return fmt.Errorf("no GStreamer packages defined for %s", m.packageManager)
	}

	// Categorize packages by installation status
	result := getPackagesToInstall(deps, m.packageManager)

	// Log status in debug mode
	if m.debugMode {
		if len(result.alreadyInstalled) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Already installed: %s\n", strings.Join(result.alreadyInstalled, ", "))
		}
		if len(result.notFound) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Not found in repos: %s\n", strings.Join(result.notFound, ", "))
		}
	}

	// Log to file
	if m.logFile != nil && len(result.notFound) > 0 {
		m.logFile.WriteString(fmt.Sprintf("[GStreamer] Warning: packages not found: %s\n", strings.Join(result.notFound, ", ")))
	}

	// Nothing to install - all packages either installed or not found
	if len(result.toInstall) == 0 {
		if len(result.notFound) > 0 && len(result.alreadyInstalled) == 0 {
			// Only missing packages, nothing installed - this is an error
			return fmt.Errorf("no installable packages found (missing: %s)", strings.Join(result.notFound, ", "))
		}
		return nil // All already installed
	}

	// Warn if some packages not found but proceeding with others
	if len(result.notFound) > 0 {
		if m.debugMode {
			fmt.Fprintf(os.Stderr, "[DEBUG] Warning: some packages not found, proceeding with available: %s\n",
				strings.Join(result.notFound, ", "))
		}
		if m.logFile != nil {
			m.logFile.WriteString(fmt.Sprintf("[GStreamer] Warning: proceeding without unavailable packages: %s\n",
				strings.Join(result.notFound, ", ")))
		}
	}

	var cmd *exec.Cmd
	switch m.packageManager {
	case "pacman":
		args := append([]string{"-S", "--noconfirm", "--needed"}, result.toInstall...)
		cmd = exec.Command("pacman", args...)
	case "apt":
		// Update first
		exec.Command("apt-get", "update").Run()
		args := append([]string{"install", "-y"}, result.toInstall...)
		cmd = exec.Command("apt-get", args...)
	case "dnf":
		args := append([]string{"install", "-y"}, result.toInstall...)
		cmd = exec.Command("dnf", args...)
	case "yum":
		args := append([]string{"install", "-y"}, result.toInstall...)
		cmd = exec.Command("yum", args...)
	case "zypper":
		args := append([]string{"install", "-y"}, result.toInstall...)
		cmd = exec.Command("zypper", args...)
	default:
		return fmt.Errorf("unsupported package manager for GStreamer deps")
	}

	if err := runCommand("Install GStreamer deps", cmd, m); err != nil {
		return fmt.Errorf("failed to install GStreamer dependencies: %s", strings.Join(result.toInstall, ", "))
	}
	return nil
}

func buildGslapperFromSource(m *model) error {
	// Sub-task 2: Install GStreamer dependencies
	updateSubTaskStatus(m, 2, statusRunning)
	if err := installGStreamerDeps(m); err != nil {
		if m.debugMode {
			fmt.Fprintf(os.Stderr, "[DEBUG] GStreamer deps install warning: %v\n", err)
		}
		// Continue anyway - deps might already be installed
	}
	updateSubTaskStatus(m, 2, statusComplete)

	// Check for build tools
	buildTools := []string{"meson", "ninja", "git", "pkg-config"}
	missingBinaries := []string{}
	for _, tool := range buildTools {
		if _, err := exec.LookPath(tool); err != nil {
			missingBinaries = append(missingBinaries, tool)
		}
	}

	// Try to install missing build tools (converting binary names to package names)
	if len(missingBinaries) > 0 {
		// Convert binary names to package names for this distro
		missingPackages := make([]string, len(missingBinaries))
		for i, bin := range missingBinaries {
			missingPackages[i] = getBinaryPackageName(bin, m.packageManager)
		}

		if m.debugMode {
			fmt.Fprintf(os.Stderr, "[DEBUG] Installing build tools: %s (packages: %s)\n",
				strings.Join(missingBinaries, ", "), strings.Join(missingPackages, ", "))
		}

		var cmd *exec.Cmd
		switch m.packageManager {
		case "pacman":
			args := append([]string{"-S", "--noconfirm", "--needed"}, missingPackages...)
			cmd = exec.Command("pacman", args...)
		case "apt":
			args := append([]string{"install", "-y"}, missingPackages...)
			cmd = exec.Command("apt-get", args...)
		case "dnf":
			args := append([]string{"install", "-y"}, missingPackages...)
			cmd = exec.Command("dnf", args...)
		case "yum":
			args := append([]string{"install", "-y"}, missingPackages...)
			cmd = exec.Command("yum", args...)
		case "zypper":
			args := append([]string{"install", "-y"}, missingPackages...)
			cmd = exec.Command("zypper", args...)
		default:
			updateSubTaskStatus(m, 3, statusFailed)
			return fmt.Errorf("missing build tools: %s", strings.Join(missingBinaries, ", "))
		}
		if cmd != nil {
			runCommand("Install build tools", cmd, m) // Best effort
		}
	}

	// Verify build tools are now available
	for _, tool := range []string{"meson", "ninja", "git"} {
		if _, err := exec.LookPath(tool); err != nil {
			updateSubTaskStatus(m, 3, statusFailed)
			return fmt.Errorf("missing build tool: %s", tool)
		}
	}

	// Sub-task 3: Clone repository
	updateSubTaskStatus(m, 3, statusRunning)
	exec.Command("rm", "-rf", "/tmp/gslapper-build").Run()
	cloneCmd := exec.Command("git", "clone", "https://github.com/Nomadcxx/gSlapper", "/tmp/gslapper-build")
	if err := runCommand("Clone gSlapper", cloneCmd, m); err != nil {
		updateSubTaskStatus(m, 3, statusFailed)
		return fmt.Errorf("clone failed")
	}
	updateSubTaskStatus(m, 3, statusComplete)

	// Sub-task 4: Build from source
	updateSubTaskStatus(m, 4, statusRunning)
	setupCmd := exec.Command("meson", "setup", "build", "--prefix=/usr/local")
	setupCmd.Dir = "/tmp/gslapper-build"
	if err := runCommand("Configure build", setupCmd, m); err != nil {
		updateSubTaskStatus(m, 4, statusFailed)
		return fmt.Errorf("build setup failed - check GStreamer dependencies")
	}

	buildCmd := exec.Command("ninja", "-C", "build")
	buildCmd.Dir = "/tmp/gslapper-build"
	if err := runCommand("Build gSlapper", buildCmd, m); err != nil {
		updateSubTaskStatus(m, 4, statusFailed)
		return fmt.Errorf("build failed")
	}
	updateSubTaskStatus(m, 4, statusComplete)

	// Sub-task 5: Install binary
	updateSubTaskStatus(m, 5, statusRunning)
	installCmd := exec.Command("ninja", "-C", "build", "install")
	installCmd.Dir = "/tmp/gslapper-build"
	if err := runCommand("Install gSlapper", installCmd, m); err != nil {
		updateSubTaskStatus(m, 5, statusFailed)
		return fmt.Errorf("install failed")
	}
	updateSubTaskStatus(m, 5, statusComplete)

	// Cleanup
	exec.Command("rm", "-rf", "/tmp/gslapper-build").Run()

	return nil
}

func buildBinary(m *model) error {
	cmd := exec.Command("go", "build", "-buildvcs=false", "-ldflags", versionLDFlags(), "-o", "sysc-greet", "./cmd/sysc-greet/")
	if err := runCommand("Build binary", cmd, m); err != nil {
		return fmt.Errorf("build failed - see /tmp/sysc-greet-installer.log")
	}
	return nil
}

// versionLDFlags stamps the binary the same way the Makefile does. The
// installer runs as root on a user-owned checkout, so git needs
// safe.directory or it refuses to read the repo.
func versionLDFlags() string {
	git := func(args ...string) string {
		out, err := exec.Command("git", append([]string{"-c", "safe.directory=*"}, args...)...).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	version := git("describe", "--tags", "--always", "--dirty")
	if version == "" {
		version = "dev"
	}
	commit := git("rev-parse", "--short", "HEAD")
	if commit == "" {
		commit = "unknown"
	}
	date := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	return fmt.Sprintf("-X 'main.Version=%s' -X 'main.GitCommit=%s' -X 'main.BuildDate=%s'", version, commit, date)
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
		configPath + "/themes",
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

	// Copy example theme if it exists
	// ADDED 2025-12-28 - Copy example custom theme template
	// FIXED 2025-12-28 - Use install with 644 permissions so greeter user can read it
	if _, err := os.Stat("examples/themes/example.toml"); err == nil {
		if err := exec.Command("install", "-m", "644", "examples/themes/example.toml", configPath+"/themes/").Run(); err != nil {
			return fmt.Errorf("failed to copy example theme")
		}
	}

	// Install bundled themes with permissions readable by the greeter user.
	bundled, _ := filepath.Glob("themes/*.toml")
	for _, theme := range bundled {
		if err := exec.Command("install", "-m", "644", theme, configPath+"/themes/").Run(); err != nil {
			return fmt.Errorf("failed to copy theme %s: %w", filepath.Base(theme), err)
		}
	}

	return nil
}

func setupCache(m *model) error {
	// Create greeter-owned state/cache directories used by kitty, gSlapper,
	// shader caches, and sysc-greet preferences.
	dirs := []string{
		"/var/cache/sysc-greet",
		"/var/lib/greeter/Pictures/wallpapers",
		"/var/lib/greeter/.cache",
		"/var/lib/greeter/.config",
		"/var/lib/greeter/.local/state",
		"/tmp/greeter-cache",
	}
	for _, dir := range dirs {
		if err := exec.Command("mkdir", "-p", dir).Run(); err != nil {
			return fmt.Errorf("directory creation failed for %s", dir)
		}
	}

	if err := exec.Command("chmod", "755", "/var/lib/greeter").Run(); err != nil {
		return fmt.Errorf("permissions change failed")
	}

	// Create greeter user if needed
	// FIXED 2025-10-15 - Add render group for modern Intel/AMD iGPU support
	// Modern Linux uses 'render' group for /dev/dri/renderD* (non-privileged GPU access)
	// Both 'video' and 'render' groups needed for laptop iGPU compatibility
	cmd := exec.Command("id", "greeter")
	if err := cmd.Run(); err != nil {
		// User doesn't exist - create with video,render,input groups
		cmd = exec.Command("useradd", "-M", "-d", "/var/lib/greeter", "-G", "video,render,input", "-s", "/usr/bin/nologin", "greeter")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("greeter user creation failed")
		}
	} else {
		// User exists - ensure they have required groups and a writable home.
		// CRITICAL: This fixes laptops where greeter user exists but lacks render group
		if err := exec.Command("usermod", "-d", "/var/lib/greeter", "-aG", "video,render,input", "greeter").Run(); err != nil {
			return fmt.Errorf("greeter user update failed")
		}
	}

	// Set ownership
	paths := []string{"/var/cache/sysc-greet", "/var/lib/greeter", "/tmp/greeter-cache"}
	for _, path := range paths {
		if err := exec.Command("chown", "-R", "greeter:greeter", path).Run(); err != nil {
			return fmt.Errorf("ownership change failed for %s", path)
		}
	}

	return nil
}

func configureGreetd(m *model) error {
	var compositorConfig string
	var greetdCommand string
	var configPath string

	switch m.selectedCompositor {
	case "niri":
		compositorConfig = `// SYSC-Greet Niri config for greetd greeter session
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

spawn-sh-at-startup "cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state /usr/local/bin/sysc-greet --wallpaper-daemon"

spawn-sh-at-startup "cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"

binds {
}
`
		configPath = "/etc/greetd/niri-greeter-config.kdl"
		greetdCommand = "niri -c /etc/greetd/niri-greeter-config.kdl"

	case "hyprland":
		compositorConfig = `# SYSC-Greet Hyprland config for greetd greeter session
# Monitors auto-detected by Hyprland at runtime

# No animations for faster greeter startup
animations {
    enabled = false
}

# Minimal decorations
decoration {
    rounding = 0
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

# CHANGED 2025-10-18 - Disable Hyprland wallpaper/logo for greeter
misc {
    disable_hyprland_logo = true
    disable_splash_rendering = true
    background_color = rgb(000000)
    # Suppress watchdog warning - greetd doesn't pass fd properly to start-hyprland
    disable_watchdog_warning = true
}

# Suppress annoying update/donation popups
ecosystem {
    no_update_news = true
    no_donation_nag = true
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
windowrule = match:class ^(kitty)$, fullscreen on, opacity 1.0

# Layer rules for wallpaper daemon
layerrule = match:namespace wallpaper, blur on

# Startup applications
exec-once = cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state /usr/local/bin/sysc-greet --wallpaper-daemon
exec-once = cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch exit
`
		configPath = "/etc/greetd/hyprland-greeter-config.conf"
		greetdCommand = "start-hyprland -- -c /etc/greetd/hyprland-greeter-config.conf"

	case "sway":
		compositorConfig = `# SYSC-Greet Sway config for greetd greeter session
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

# Window rules for kitty
for_window [app_id="kitty"] fullscreen enable

# Startup applications
exec "cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state /usr/local/bin/sysc-greet --wallpaper-daemon"
exec "cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; swaymsg exit"
`
		configPath = "/etc/greetd/sway-greeter-config"
		greetdCommand = "sway --unsupported-gpu -c /etc/greetd/sway-greeter-config"

	case "cagebreak":
		compositorConfig = `# sysc-greet cagebreak greeter config — used ONLY by the greetd greeter session
# greetd runs: cagebreak -e -c /etc/greetd/cagebreak-greeter-config
# -e enables the IPC socket used to quit cagebreak after login (requires socat)
# Keyboard layout: set XKB_DEFAULT_LAYOUT/XKB_DEFAULT_VARIANT in greetd's command env

background 0.0 0.0 0.0

# gSlapper wallpaper daemon (wlr-layer-shell client, same as niri/sway setup)
exec cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state /usr/local/bin/sysc-greet --wallpaper-daemon

# Greeter UI; when it exits (login or shutdown), quit cagebreak so greetd can start the session
exec cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; echo quit | socat - UNIX-CONNECT:"$CAGEBREAK_SOCKET"

# No bind/definekey lines on purpose: no compositor keybindings reachable from the greeter
`
		configPath = "/etc/greetd/cagebreak-greeter-config"
		greetdCommand = "cagebreak -e -c /etc/greetd/cagebreak-greeter-config"

	case "mango":
		// Installed from the repo checkout (cwd, as in installConfigs) so
		// config/ stays the only copy of the greeter config and session script
		data, err := os.ReadFile("config/mango-greeter-config.conf")
		if err != nil {
			return fmt.Errorf("mango greeter config missing: %v", err)
		}
		compositorConfig = string(data)
		if err := exec.Command("install", "-Dm755", "config/mango-greeter-session.sh", "/etc/greetd/mango-greeter-session.sh").Run(); err != nil {
			return fmt.Errorf("mango greeter session script install failed")
		}
		configPath = "/etc/greetd/mango-greeter-config.conf"
		greetdCommand = "mango -c /etc/greetd/mango-greeter-config.conf -s /etc/greetd/mango-greeter-session.sh"

	default:
		return fmt.Errorf("unknown compositor: %s", m.selectedCompositor)
	}

	// Write compositor config
	if err := os.WriteFile(configPath, []byte(compositorConfig), 0644); err != nil {
		return fmt.Errorf("compositor config write failed")
	}

	// Write greetd config
	greetdConfig := fmt.Sprintf(`[terminal]
vt = 1

[default_session]
command = "%s"
user = "greeter"

[initial_session]
command = "%s"
user = "greeter"
`, greetdCommand, greetdCommand)

	// Keep the previous config when it differs, as the AUR .install scripts do
	if old, err := os.ReadFile("/etc/greetd/config.toml"); err == nil && string(old) != greetdConfig {
		if err := os.WriteFile("/etc/greetd/config.toml.bak", old, 0644); err != nil {
			return fmt.Errorf("greetd config backup failed")
		}
		m.logf("Backed up /etc/greetd/config.toml to /etc/greetd/config.toml.bak")
	}

	if err := os.WriteFile("/etc/greetd/config.toml", []byte(greetdConfig), 0644); err != nil {
		return fmt.Errorf("greetd config write failed")
	}

	// Install polkit rule to allow greeter user to shutdown/reboot
	polkitRule := `polkit.addRule(function(action, subject) {
    if ((action.id == "org.freedesktop.login1.power-off" ||
         action.id == "org.freedesktop.login1.power-off-multiple-sessions" ||
         action.id == "org.freedesktop.login1.reboot" ||
         action.id == "org.freedesktop.login1.reboot-multiple-sessions") &&
        subject.user == "greeter") {
        return polkit.Result.YES;
    }
});
`
	if err := os.MkdirAll("/etc/polkit-1/rules.d", 0755); err != nil {
		return fmt.Errorf("polkit rules directory creation failed")
	}
	if err := os.WriteFile("/etc/polkit-1/rules.d/85-greeter.rules", []byte(polkitRule), 0644); err != nil {
		return fmt.Errorf("polkit rule write failed")
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
		"/etc/greetd/hyprland-greeter-config.conf",
		"/etc/greetd/sway-greeter-config",
		"/etc/greetd/cagebreak-greeter-config",
		"/etc/greetd/mango-greeter-config.conf",
		"/etc/greetd/mango-greeter-session.sh",
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

func uninstallGslapper(m *model) error {
	// Detect installation method
	isArch := isArchBased()
	packagesToRemove := []string{}

	if isArch {
		// Check for both gslapper and gslapper-debug packages
		for _, pkg := range []string{"gslapper", "gslapper-debug"} {
			cmd := exec.Command("pacman", "-Qi", pkg)
			if err := cmd.Run(); err == nil {
				packagesToRemove = append(packagesToRemove, pkg)
			}
		}
	}

	var cmd *exec.Cmd
	if len(packagesToRemove) > 0 {
		// Uninstall via pacman (already running as root, no sudo needed)
		args := append([]string{"-R", "--noconfirm"}, packagesToRemove...)
		cmd = exec.Command("pacman", args...)
	} else {
		// Manually remove binaries (source install)
		cmd = exec.Command("rm", "-f", "/usr/local/bin/gslapper", "/usr/local/bin/gslapper-holder")
	}

	if err := runCommand("Uninstall gSlapper", cmd, m); err != nil {
		// Not critical if gslapper wasn't installed
		return nil
	}

	return nil
}

func main() {
	// Check for debug flag
	debugMode := false
	for _, arg := range os.Args[1:] {
		if arg == "--debug" || arg == "-d" {
			debugMode = true
			break
		}
	}

	// Append so consecutive runs (e.g. uninstall then install) keep their logs
	logFile, err := os.OpenFile("/tmp/sysc-greet-installer.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Warning: Could not create log file: %v\n", err)
		logFile = nil
	}
	if logFile != nil {
		defer logFile.Close()
		// Write startup info
		logFile.WriteString(fmt.Sprintf("\n=== sysc-greet Installer Log ===\n"))
		logFile.WriteString(fmt.Sprintf("Started: %s\n", time.Now().Format("2006-01-02 15:04:05")))
		logFile.WriteString(fmt.Sprintf("Debug Mode: %v\n\n", debugMode))
	}

	p := tea.NewProgram(newModel(debugMode, logFile))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
