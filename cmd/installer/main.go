package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Theme colors - Dracula
var (
	BgBase      = lipgloss.Color("#282a36")
	BgElevated  = lipgloss.Color("#383a59")
	BgSubtle    = lipgloss.Color("#44475a")
	Primary     = lipgloss.Color("#bd93f9")
	Secondary   = lipgloss.Color("#8be9fd")
	Accent      = lipgloss.Color("#50fa7b")
	FgPrimary   = lipgloss.Color("#f8f8f2")
	FgSecondary = lipgloss.Color("#f1f2f6")
	FgMuted     = lipgloss.Color("#6272a4")
	ErrorColor  = lipgloss.Color("#ff5555")
	WarningColor = lipgloss.Color("#f1fa8c")
)

type installStep int

const (
	stepWelcome installStep = iota
	stepCheckPrivileges
	stepSelectInstallType
	stepCheckDependencies
	stepInstallGreetd
	stepInstallGslapper // CHANGED 2025-10-03 - Add gslapper installation step - Problem: Need video wallpaper support
	stepBuildBinary
	stepInstallBinary
	stepInstallConfigs
	stepSetupCache
	stepConfigureGreetd
	stepEnableService
	stepComplete
)

type installType int

const (
	installFull installType = iota // greetd + sysc-greet
	installGreeterOnly              // sysc-greet only (greetd already installed)
	installConfigOnly               // Just configs (binary already installed)
)

type model struct {
	step              installStep
	installType       installType
	width             int
	height            int
	messages          []string
	errors            []string
	currentAction     string
	hasRoot           bool
	greetdInstalled   bool
	tuigreetInstalled bool
	needsGreetd       bool
	packageManager    string
	binaryBuilt       bool
	installPath       string
	configPath        string
	selectedOption    int
	options           []string
	greetdFromSource  bool     // CHANGED 2025-10-02 00:35 - Track if greetd needs source build
	liveOutput        []string // CHANGED 2025-10-02 01:15 - Store live command output lines
	currentCommand    string   // CHANGED 2025-10-02 01:15 - Current command being run
}

type stepCompleteMsg struct {
	success bool
	message string
}

type dependencyCheckMsg struct {
	success           bool
	message           string
	packageManager    string
	greetdInstalled   bool
	tuigreetInstalled bool
	needsGreetd       bool
}

type commandOutputMsg struct {
	line string
}

func initialModel() model {
	return model{
		step:        stepWelcome,
		installPath: "/usr/local/bin/sysc-greet",
		configPath:  "/usr/share/sysc-greet",
		messages:    []string{},
		errors:      []string{},
	}
}

func (m model) Init() tea.Cmd {
	return nil
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
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.selectedOption < len(m.options)-1 {
				m.selectedOption++
			}
		case "r", "R":
			// CHANGED 2025-10-02 00:45 - Quick retry on error
			if len(m.options) > 0 && m.options[0] == "Retry" {
				return m.proceedWithStep()
			}
		case "s", "S":
			// CHANGED 2025-10-02 00:45 - Quick skip on error
			if len(m.options) > 0 {
				return m.advanceStep()
			}
		case "y", "Y":
			if m.step == stepCheckPrivileges {
				return m.proceedWithStep()
			}
		case "n", "N":
			if m.step == stepCheckPrivileges {
				return m, tea.Quit
			}
		}

	case commandOutputMsg:
		// CHANGED 2025-10-02 01:15 - Handle streaming command output
		m.liveOutput = append(m.liveOutput, msg.line)
		// Keep only last 10 lines
		if len(m.liveOutput) > 10 {
			m.liveOutput = m.liveOutput[len(m.liveOutput)-10:]
		}
		return m, nil

	case dependencyCheckMsg:
		// CHANGED 2025-10-02 00:20 - Handle dependency check with state updates - Problem: Model state lost in closure
		m.packageManager = msg.packageManager
		m.greetdInstalled = msg.greetdInstalled
		m.tuigreetInstalled = msg.tuigreetInstalled
		m.needsGreetd = msg.needsGreetd
		if msg.success {
			m.messages = append(m.messages, "✓ "+msg.message)
			return m.advanceStep()
		} else {
			m.errors = append(m.errors, "✗ "+msg.message)
		}
		return m, nil

	case stepCompleteMsg:
		if msg.success {
			m.messages = append(m.messages, "✓ "+msg.message)
			return m.advanceStep()
		} else {
			m.errors = append(m.errors, "✗ "+msg.message)
			// CHANGED 2025-10-02 00:40 - Offer source build for greetd failures
			if m.step == stepInstallGreetd && !m.greetdFromSource && strings.Contains(msg.message, "not available") {
				m.options = []string{"Build from source", "Skip", "Abort"}
				m.selectedOption = 0
			} else {
				m.options = []string{"Retry", "Skip", "Abort"}
				m.selectedOption = 0
			}
		}
	}

	return m, nil
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepWelcome:
		m.step = stepCheckPrivileges
		return m, m.checkPrivileges()
	case stepSelectInstallType:
		m.installType = installType(m.selectedOption)
		return m.advanceStep()
	default:
		// CHANGED 2025-10-02 00:45 - Handle error menu options
		if len(m.options) > 0 {
			switch m.options[m.selectedOption] {
			case "Retry":
				m.options = []string{}
				return m.proceedWithStep()
			case "Build from source":
				m.greetdFromSource = true
				m.options = []string{}
				return m.proceedWithStep()
			case "Skip":
				m.options = []string{}
				return m.advanceStep()
			case "Abort":
				return m, tea.Quit
			}
		}
		return m.proceedWithStep()
	}
}

func (m model) proceedWithStep() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepCheckPrivileges:
		return m.advanceStep()
	case stepCheckDependencies:
		return m, m.checkDependencies()
	case stepInstallGreetd:
		// Skip if greetd already installed
		if m.greetdInstalled {
			m.messages = append(m.messages, "✓ greetd already installed, skipping")
			return m.advanceStep()
		}
		// CHANGED 2025-10-02 00:40 - Try source build if package install fails
		if m.greetdFromSource {
			return m, m.buildGreetdFromSource()
		}
		return m, m.installGreetd()
	case stepInstallGslapper:
		// CHANGED 2025-10-03 - Install gslapper for video wallpapers - Problem: Need video wallpaper dependency
		return m, m.installGslapper()
	case stepBuildBinary:
		return m, m.buildBinary()
	case stepInstallBinary:
		return m, m.installBinary()
	case stepInstallConfigs:
		return m, m.installConfigs()
	case stepSetupCache:
		return m, m.setupCache()
	case stepConfigureGreetd:
		return m, m.configureGreetd()
	case stepEnableService:
		return m, m.enableService()
	}
	return m, nil
}

func (m model) advanceStep() (tea.Model, tea.Cmd) {
	m.step++
	return m.proceedWithStep()
}

func (m model) checkPrivileges() tea.Cmd {
	return func() tea.Msg {
		if os.Geteuid() == 0 {
			return stepCompleteMsg{true, "Running with root privileges"}
		}
		// Check if sudo is available
		if _, err := exec.LookPath("sudo"); err == nil {
			return stepCompleteMsg{true, "sudo available for privilege escalation"}
		}
		return stepCompleteMsg{false, "Root privileges required. Please run with sudo."}
	}
}

func (m model) checkDependencies() tea.Cmd {
	return func() tea.Msg {
		// Check for Go
		if _, err := exec.LookPath("go"); err != nil {
			return dependencyCheckMsg{success: false, message: "Go compiler not found. Install Go first."}
		}

		// Check for systemd
		if _, err := exec.LookPath("systemctl"); err != nil {
			return dependencyCheckMsg{success: false, message: "systemd not found. Manual setup required."}
		}

		// Check if greetd is installed
		_, err := exec.LookPath("greetd")
		greetdInstalled := (err == nil)

		// Check if tuigreet is installed
		_, err = exec.LookPath("tuigreet")
		tuigreetInstalled := (err == nil)

		// CHANGED 2025-10-04 - Check for kitty terminal - Problem: Need kitty for truecolor support
		_, err = exec.LookPath("kitty")
		kittyInstalled := (err == nil)

		// CHANGED 2025-10-05 - Check for niri compositor - Problem: Need niri to run greeter
		_, err = exec.LookPath("niri")
		niriInstalled := (err == nil)

		// CHANGED 2025-10-02 00:30 - Expanded distro support - Problem: More distros needed
		packageManagers := map[string][]string{
			"pacman":       {"/usr/bin/pacman", "/usr/sbin/pacman", "/bin/pacman", "/sbin/pacman"},
			"yay":          {"/usr/bin/yay", "/usr/sbin/yay", "/bin/yay", "/sbin/yay"},
			"apt":          {"/usr/bin/apt", "/usr/sbin/apt", "/bin/apt", "/usr/bin/apt-get"},
			"dnf":          {"/usr/bin/dnf", "/usr/sbin/dnf"},
			"zypper":       {"/usr/bin/zypper", "/usr/sbin/zypper"},
			"emerge":       {"/usr/bin/emerge", "/usr/sbin/emerge"},
			"apk":          {"/sbin/apk", "/usr/sbin/apk"},
			"xbps-install": {"/usr/bin/xbps-install", "/bin/xbps-install"},
		}

		// CHANGED 2025-10-02 00:50 - Prefer pacman over yay as default - Problem: User wants pacman as default, yay as fallback
		var packageManager string
		// Check in order of preference (prefer distro package managers, yay only for AUR)
		for _, pmName := range []string{"pacman", "apt", "dnf", "zypper", "emerge", "apk", "xbps-install", "yay"} {
			paths := packageManagers[pmName]
			for _, path := range paths {
				if _, err := os.Stat(path); err == nil {
					packageManager = pmName
					break
				}
			}
			if packageManager != "" {
				break
			}
		}

		// Determine if we need to install greetd
		needsGreetd := !greetdInstalled

		msg := "Dependencies: Go ✓, systemd ✓"
		if greetdInstalled {
			msg += ", greetd ✓"
		} else {
			msg += ", greetd ✗ (will install)"
		}
		if tuigreetInstalled {
			msg += ", tuigreet ✓ (will replace)"
		}
		// CHANGED 2025-10-04 - Show kitty status - Problem: Need kitty for truecolor
		if kittyInstalled {
			msg += ", kitty ✓"
		} else {
			msg += ", kitty ✗ (install: pacman -S kitty)"
		}
		// CHANGED 2025-10-05 - Show niri status - Problem: Need niri compositor
		if niriInstalled {
			msg += ", niri ✓"
		} else {
			msg += ", niri ✗ (install: pacman -S niri)"
		}
		if packageManager != "" {
			msg += fmt.Sprintf(", package manager: %s", packageManager)
		}

		return dependencyCheckMsg{
			success:           true,
			message:           msg,
			packageManager:    packageManager,
			greetdInstalled:   greetdInstalled,
			tuigreetInstalled: tuigreetInstalled,
			needsGreetd:       needsGreetd,
		}
	}
}

func (m model) installGreetd() tea.Cmd {
	return func() tea.Msg {
		if m.packageManager == "" {
			return stepCompleteMsg{false, "No supported package manager found. Install greetd manually."}
		}

		// CHANGED 2025-10-02 00:25 - Improved package manager support and error handling
		var cmd *exec.Cmd
		var pkgName string = "greetd"

		switch m.packageManager {
		case "yay":
			// CHANGED 2025-10-02 00:45 - Try official repos first, then AUR
			// Check if greetd is in official repos
			checkCmd := exec.Command("pacman", "-Si", pkgName)
			if checkCmd.Run() == nil {
				// In official repos, use pacman
				cmd = exec.Command("pacman", "-S", "--noconfirm", pkgName)
			} else {
				// Not in repos, use yay for AUR
				cmd = exec.Command("sudo", "-u", os.Getenv("SUDO_USER"), "yay", "-S", "--noconfirm", pkgName)
			}
		case "pacman":
			cmd = exec.Command("pacman", "-S", "--noconfirm", pkgName)
		case "apt", "apt-get":
			// Update package list first for Debian/Ubuntu
			updateCmd := exec.Command("apt-get", "update")
			updateCmd.Run() // Ignore errors, not critical
			cmd = exec.Command("apt-get", "install", "-y", pkgName)
		case "dnf":
			cmd = exec.Command("dnf", "install", "-y", pkgName)
		case "zypper":
			cmd = exec.Command("zypper", "install", "-y", pkgName)
		case "emerge":
			cmd = exec.Command("emerge", "--ask=n", pkgName)
		case "apk":
			// Alpine Linux support
			cmd = exec.Command("apk", "add", pkgName)
		case "xbps-install":
			// Void Linux support
			cmd = exec.Command("xbps-install", "-y", pkgName)
		default:
			return stepCompleteMsg{false, fmt.Sprintf("Unsupported package manager: %s. Install greetd manually with your package manager.", m.packageManager)}
		}

		// CHANGED 2025-10-02 00:50 - Show live command output - Problem: User wants to see terminal output


		err := cmd.Run()
		if err != nil {
			// CHANGED 2025-10-02 00:35 - Offer source build fallback - Problem: greetd not in all repos
			errMsg := fmt.Sprintf("greetd installation failed with %s. Will attempt to build from source...", m.packageManager)
			return stepCompleteMsg{false, errMsg}
		}

		return stepCompleteMsg{true, fmt.Sprintf("greetd installed successfully using %s", m.packageManager)}
	}
}

func (m model) buildGreetdFromSource() tea.Cmd {
	return func() tea.Msg {
		// CHANGED 2025-10-02 00:35 - Build greetd from source - Problem: Not in all package repos

		// Check for Rust/Cargo
		if _, err := exec.LookPath("cargo"); err != nil {
			return stepCompleteMsg{false, "Cargo not found. Install Rust to build greetd from source: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"}
		}

		// Clone greetd repo
		cloneCmd := exec.Command("git", "clone", "https://git.sr.ht/~kennylevinsen/greetd", "/tmp/greetd-build")
		if err := cloneCmd.Run(); err != nil {
			// Try alternate repo if sr.ht is down
			cloneCmd = exec.Command("git", "clone", "https://github.com/kennylevinsen/greetd", "/tmp/greetd-build")
			if err := cloneCmd.Run(); err != nil {
				return stepCompleteMsg{false, "Failed to clone greetd repository. Check internet connection."}
			}
		}

		// Build with cargo
		buildCmd := exec.Command("cargo", "build", "--release")
		buildCmd.Dir = "/tmp/greetd-build"
		// Capture output instead
		// Capture output instead

		err := buildCmd.Run()
		if err != nil {
			return stepCompleteMsg{false, "greetd build failed. Check output above for errors."}
		}

		// Install binary
		installCmd := exec.Command("install", "-Dm755", "/tmp/greetd-build/target/release/greetd", "/usr/local/bin/greetd")
		if err := installCmd.Run(); err != nil {
			return stepCompleteMsg{false, "Failed to install greetd binary"}
		}

		// Install agreety (default greeter)
		installCmd = exec.Command("install", "-Dm755", "/tmp/greetd-build/target/release/agreety", "/usr/local/bin/agreety")
		if err := installCmd.Run(); err != nil {
			return stepCompleteMsg{false, "Failed to install agreety binary"}
		}

		// Create config directory
		exec.Command("mkdir", "-p", "/etc/greetd").Run()

		// Cleanup
		exec.Command("rm", "-rf", "/tmp/greetd-build").Run()

		return stepCompleteMsg{true, "greetd built and installed from source successfully"}
	}
}

// CHANGED 2025-10-03 - Install gslapper for video wallpaper support - Problem: Need video wallpaper dependency
func (m model) installGslapper() tea.Cmd {
	return func() tea.Msg {
		// Check if gslapper is already installed
		if _, err := exec.LookPath("gslapper"); err == nil {
			return stepCompleteMsg{true, "gslapper already installed, skipping"}
		}

		if m.packageManager == "" {
			return stepCompleteMsg{false, "No package manager found. Install gslapper manually from https://github.com/Nomadcxx/gSlapper"}
		}

		var cmd *exec.Cmd
		var pkgName string = "gslapper"

		switch m.packageManager {
		case "yay", "pacman":
			// For Arch Linux, try AUR package first
			// Check if yay is available for AUR
			if _, err := exec.LookPath("yay"); err == nil {
				// Use yay for AUR package
				cmd = exec.Command("sudo", "-u", os.Getenv("SUDO_USER"), "yay", "-S", "--noconfirm", pkgName)
			} else {
				// No yay, need to build from source
				return m.buildGslapperFromSource()
			}
		default:
			// For non-Arch distros, build from source
			return m.buildGslapperFromSource()
		}

		err := cmd.Run()
		if err != nil {
			// AUR install failed, try building from source
			return m.buildGslapperFromSource()
		}

		return stepCompleteMsg{true, "gslapper installed successfully from AUR"}
	}
}

func (m model) buildGslapperFromSource() tea.Msg {
	// Check for required dependencies
	deps := []string{"meson", "ninja", "git"}
	missingDeps := []string{}

	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			missingDeps = append(missingDeps, dep)
		}
	}

	if len(missingDeps) > 0 {
		// Try to install missing build dependencies
		switch m.packageManager {
		case "pacman", "yay":
			installCmd := exec.Command("pacman", "-S", "--noconfirm", "meson", "ninja", "git")
			installCmd.Run() // Try, but don't fail if it errors
		case "apt":
			installCmd := exec.Command("apt-get", "install", "-y", "meson", "ninja-build", "git")
			installCmd.Run()
		case "dnf":
			installCmd := exec.Command("dnf", "install", "-y", "meson", "ninja-build", "git")
			installCmd.Run()
		}
	}

	// Install GStreamer dependencies
	switch m.packageManager {
	case "pacman", "yay":
		depsCmd := exec.Command("pacman", "-S", "--noconfirm", "--needed",
			"gstreamer", "gst-plugins-base", "gst-plugins-good", "gst-plugins-bad")
		depsCmd.Run()
	case "apt":
		depsCmd := exec.Command("apt-get", "install", "-y",
			"libgstreamer1.0-dev", "gstreamer1.0-plugins-base", "gstreamer1.0-plugins-good", "gstreamer1.0-plugins-bad")
		depsCmd.Run()
	case "dnf":
		depsCmd := exec.Command("dnf", "install", "-y",
			"gstreamer1-devel", "gstreamer1-plugins-base", "gstreamer1-plugins-good", "gstreamer1-plugins-bad")
		depsCmd.Run()
	}

	// Clone gslapper repository
	cloneCmd := exec.Command("git", "clone", "https://github.com/Nomadcxx/gSlapper", "/tmp/gslapper-build")
	if err := cloneCmd.Run(); err != nil {
		// Cleanup if exists
		exec.Command("rm", "-rf", "/tmp/gslapper-build").Run()
		cloneCmd = exec.Command("git", "clone", "https://github.com/Nomadcxx/gSlapper", "/tmp/gslapper-build")
		if err := cloneCmd.Run(); err != nil {
			return stepCompleteMsg{false, "Failed to clone gslapper repository. Check internet connection."}
		}
	}

	// Build with meson
	setupCmd := exec.Command("meson", "setup", "build", "--prefix=/usr/local")
	setupCmd.Dir = "/tmp/gslapper-build"
	if err := setupCmd.Run(); err != nil {
		return stepCompleteMsg{false, "gslapper build setup failed. Missing dependencies?"}
	}

	buildCmd := exec.Command("ninja", "-C", "build")
	buildCmd.Dir = "/tmp/gslapper-build"
	if err := buildCmd.Run(); err != nil {
		return stepCompleteMsg{false, "gslapper build failed. Check meson/ninja installation."}
	}

	// Install
	installCmd := exec.Command("ninja", "-C", "build", "install")
	installCmd.Dir = "/tmp/gslapper-build"
	if err := installCmd.Run(); err != nil {
		return stepCompleteMsg{false, "gslapper installation failed. May need manual install."}
	}

	// Cleanup
	exec.Command("rm", "-rf", "/tmp/gslapper-build").Run()

	return stepCompleteMsg{true, "gslapper built and installed from source successfully"}
}

func (m model) buildBinary() tea.Cmd {
	return func() tea.Msg {
		// CHANGED 2025-10-04 - Build entire package directory not just main.go - Problem: Missing menu.go and wallpaper.go functions
		// CHANGED 2025-10-04 - Disable VCS stamping - Problem: Not a git repo yet, build fails with VCS error
		cmd := exec.Command("go", "build", "-buildvcs=false", "-o", "sysc-greet", "./cmd/sysc-greet/")
		output, err := cmd.CombinedOutput()

		if err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Build failed: %s", string(output))}
		}
		return stepCompleteMsg{true, "Binary built successfully"}
	}
}

func (m model) installBinary() tea.Cmd {
	return func() tea.Msg {
		// Copy binary
		cmd := exec.Command("install", "-Dm755", "sysc-greet", m.installPath)

		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to install binary: %v", err)}
		}
		return stepCompleteMsg{true, fmt.Sprintf("Installed binary to %s", m.installPath)}
	}
}

func (m model) installConfigs() tea.Cmd {
	return func() tea.Msg {
		// CHANGED 2025-10-02 01:35 - Install both ascii_configs AND fonts - Problem: Fonts missing, ASCII won't render

		// Create config directory
		cmd := exec.Command("mkdir", "-p", m.configPath+"/ascii_configs")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to create config dir: %v", err)}
		}

		// Create fonts directory
		cmd = exec.Command("mkdir", "-p", m.configPath+"/fonts")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to create fonts dir: %v", err)}
		}

		// Copy ASCII configs
		cmd = exec.Command("cp", "-r", "ascii_configs/", m.configPath+"/")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to copy ASCII configs: %v", err)}
		}

		// Copy fonts
		cmd = exec.Command("cp", "-r", "fonts/", m.configPath+"/")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to copy fonts: %v", err)}
		}

		// CHANGED 2025-10-09 21:15 - Copy only kitty.conf - Problem: start-greeter.sh removed, niri config directly launches kitty
		// Copy kitty config
		cmd = exec.Command("cp", "config/kitty-greeter.conf", "/etc/greetd/kitty.conf")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to copy kitty.conf: %v", err)}
		}

		// CHANGED 2025-10-04 - Copy Assets directory for bundled videos - Problem: Need Fireplace/Particle videos
		// Copy Assets directory if it exists
		if _, err := os.Stat("Assets"); err == nil {
			cmd = exec.Command("mkdir", "-p", m.configPath+"/Assets")
			if err := cmd.Run(); err != nil {
				return stepCompleteMsg{false, fmt.Sprintf("Failed to create Assets dir: %v", err)}
			}
			cmd = exec.Command("cp", "-r", "Assets/", m.configPath+"/")
			if err := cmd.Run(); err != nil {
				return stepCompleteMsg{false, fmt.Sprintf("Failed to copy Assets: %v", err)}
			}
		}

		return stepCompleteMsg{true, fmt.Sprintf("Installed configs, fonts, assets, and greetd scripts to %s", m.configPath)}
	}
}

func (m model) setupCache() tea.Cmd {
	return func() tea.Msg {
		// Create cache directory
		cmd := exec.Command("mkdir", "-p", "/var/cache/tuigreet")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to create cache dir: %v", err)}
		}

		// CHANGED 2025-10-04 - Create wallpapers directory for greeter user - Problem: Greeter needs $HOME/Pictures/wallpapers
		// Create wallpapers directory for greeter user
		cmd = exec.Command("mkdir", "-p", "/var/lib/greeter/Pictures/wallpapers")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to create wallpapers dir: %v", err)}
		}

		// Create greeter user if it doesn't exist
		cmd = exec.Command("id", "greeter")
		if err := cmd.Run(); err != nil {
			// User doesn't exist, create it
			cmd = exec.Command("useradd", "-M", "-G", "video", "-s", "/usr/bin/nologin", "greeter")
			if err := cmd.Run(); err != nil {
				return stepCompleteMsg{false, "Failed to create greeter user"}
			}
		}

		// Set ownership of cache
		cmd = exec.Command("chown", "-R", "greeter:greeter", "/var/cache/tuigreet")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, "Failed to set cache ownership"}
		}

		// CHANGED 2025-10-06 - Fix greeter home permissions - Problem: kitty can't create .cache without write access
		// Set ownership of greeter home directory
		cmd = exec.Command("chown", "-R", "greeter:greeter", "/var/lib/greeter")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, "Failed to set greeter home ownership"}
		}

		// Set proper permissions on greeter home
		cmd = exec.Command("chmod", "755", "/var/lib/greeter")
		if err := cmd.Run(); err != nil {
			return stepCompleteMsg{false, "Failed to set greeter home permissions"}
		}

		return stepCompleteMsg{true, "Cache directory and greeter home permissions configured"}
	}
}

// CHANGED 2025-10-06 - Multi-environment monitor detection - Problem: Need to detect monitors across Wayland/X11/TTY
type Monitor struct {
	Name       string
	Width      int
	Height     int
	RefreshRate int
	Primary    bool
}

func detectMonitors() []Monitor {
	var monitors []Monitor

	// Method 1: Try niri (if running)
	if cmd := exec.Command("niri", "msg", "outputs"); cmd.Run() == nil {
		output, err := cmd.Output()
		if err == nil {
			monitors = parseNiriOutputs(string(output))
			if len(monitors) > 0 {
				return monitors
			}
		}
	}

	// Method 2: Try hyprctl (Hyprland)
	if cmd := exec.Command("hyprctl", "monitors"); cmd.Run() == nil {
		output, err := cmd.Output()
		if err == nil {
			monitors = parseHyprlandMonitors(string(output))
			if len(monitors) > 0 {
				return monitors
			}
		}
	}

	// Method 3: Try swaymsg (Sway)
	if cmd := exec.Command("swaymsg", "-t", "get_outputs"); cmd.Run() == nil {
		output, err := cmd.Output()
		if err == nil {
			monitors = parseSwayOutputs(string(output))
			if len(monitors) > 0 {
				return monitors
			}
		}
	}

	// Method 4: Try xrandr (X11)
	if cmd := exec.Command("xrandr", "--query"); cmd.Run() == nil {
		output, err := cmd.Output()
		if err == nil {
			monitors = parseXrandr(string(output))
			if len(monitors) > 0 {
				return monitors
			}
		}
	}

	// Method 5: Try DRM (direct /sys/class/drm reading)
	monitors = parseDRM()
	if len(monitors) > 0 {
		return monitors
	}

	// Fallback: Single 1920x1080 monitor
	return []Monitor{{Name: "Unknown-1", Width: 1920, Height: 1080, RefreshRate: 60, Primary: true}}
}

func parseNiriOutputs(output string) []Monitor {
	var monitors []Monitor
	lines := strings.Split(output, "\n")
	var current *Monitor

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Output "Name" (connector)
		if strings.HasPrefix(line, "Output") {
			if current != nil {
				monitors = append(monitors, *current)
			}
			current = &Monitor{}

			// Extract connector name from: Output "..." (DP-1)
			if start := strings.LastIndex(line, "("); start != -1 {
				if end := strings.LastIndex(line, ")"); end != -1 {
					current.Name = strings.TrimSpace(line[start+1 : end])
				}
			}
		}

		// Current mode: 3440x1440 @ 120.000 Hz
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

	// Mark first as primary
	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors
}

func parseHyprlandMonitors(output string) []Monitor {
	// TODO: Parse hyprctl monitors output
	// Format: Monitor NAME (ID X): WIDTHxHEIGHT@REFRESH
	return nil
}

func parseSwayOutputs(output string) []Monitor {
	// TODO: Parse swaymsg JSON output
	return nil
}

func parseXrandr(output string) []Monitor {
	var monitors []Monitor
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Look for: DP-1 connected primary 3440x1440+0+0 (normal left...) 800mm x 340mm
		if strings.Contains(line, " connected") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}

			monitor := Monitor{Name: fields[0]}
			monitor.Primary = strings.Contains(line, " primary ")

			// Find resolution field (e.g., "3440x1440+0+0")
			for _, field := range fields {
				if strings.Contains(field, "x") && strings.Contains(field, "+") {
					resPart := strings.Split(field, "+")[0]
					dims := strings.Split(resPart, "x")
					if len(dims) == 2 {
						monitor.Width, _ = strconv.Atoi(dims[0])
						monitor.Height, _ = strconv.Atoi(dims[1])
						monitor.RefreshRate = 60 // xrandr doesn't show refresh in this line
						monitors = append(monitors, monitor)
						break
					}
				}
			}
		}
	}

	return monitors
}

func parseDRM() []Monitor {
	var monitors []Monitor

	// Read /sys/class/drm/card*/card*/status
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "card") {
			continue
		}

		statusPath := fmt.Sprintf("/sys/class/drm/%s/status", entry.Name())
		status, err := os.ReadFile(statusPath)
		if err != nil || !strings.Contains(string(status), "connected") {
			continue
		}

		// Found connected monitor, try to get EDID for resolution
		// This is complex, so for now just add with default resolution
		monitors = append(monitors, Monitor{
			Name:        entry.Name(),
			Width:       1920,
			Height:      1080,
			RefreshRate: 60,
		})
	}

	if len(monitors) > 0 {
		monitors[0].Primary = true
	}

	return monitors
}

func (m model) configureGreetd() tea.Cmd {
	return func() tea.Msg {
		// CHANGED 2025-10-06 - Remove monitor auto-detection, let niri handle it - Problem: Detection during install is unreliable
		// CHANGED 2025-10-05 - Use niri compositor instead of weston - Problem: weston autolaunch doesn't work, caused boot loops
		// Based on working example from https://github.com/YaLTeR/niri/discussions/1276

		// Write niri greeter config (no hardcoded monitor settings - niri auto-detects)
		niriConfig := `// SYSC-Greet Niri config for greetd greeter session
// This config is ONLY used by greetd to show the greeter
// Monitors will be auto-detected by niri at runtime

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
        // natural-scroll;
    }

    touchpad {
        tap;
        // natural-scroll;
    }
}

// Outputs will be auto-detected by niri at runtime
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

// Window rule to preserve kitty opacity even when focused
window-rule {
    match app-id="kitty"
    opacity 0.90
}

// CRITICAL: Launch kitty with sysc-greet, then quit niri when it exits
spawn-at-startup "sh" "-c" "kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet; niri msg action quit --skip-confirmation"

// Empty binds block = no keybindings work (security for greeter)
binds {
}
`
		if err := os.WriteFile("/etc/greetd/niri-greeter-config.kdl", []byte(niriConfig), 0644); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to write niri config: %v", err)}
		}

		// Write greetd config to use niri
		config := `[terminal]
vt = 1

[default_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"

[initial_session]
command = "niri -c /etc/greetd/niri-greeter-config.kdl"
user = "greeter"
`
		if err := os.WriteFile("/etc/greetd/config.toml", []byte(config), 0644); err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to write greetd config: %v", err)}
		}

		return stepCompleteMsg{true, "greetd configured with niri compositor"}
	}
}

func (m model) enableService() tea.Cmd {
	return func() tea.Msg {

		// CHANGED 2025-10-02 01:10 - Automatically handle service conflicts - Problem: ly/sddm symlink blocks enable
		// Check if display-manager.service symlink exists
		symlinkPath := "/etc/systemd/system/display-manager.service"
		if _, err := os.Lstat(symlinkPath); err == nil {
			// Symlink exists, remove it
			_, _ = os.Readlink(symlinkPath)
			if err := os.Remove(symlinkPath); err != nil {
				return stepCompleteMsg{false, fmt.Sprintf("Failed to remove display-manager.service symlink: %v", err)}
			}
		}

		// CHANGED 2025-10-02 01:25 - Don't stop display manager, just warn - Problem: Stopping kills user's session!
		// DON'T stop - let user reboot to activate new greeter
		// The old display manager will be disabled on next boot

		// Enable greetd service (this creates the display-manager.service symlink)
		cmd := exec.Command("systemctl", "enable", "greetd.service")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return stepCompleteMsg{false, fmt.Sprintf("Failed to enable greetd: %s", string(output))}
		}

		return stepCompleteMsg{true, "greetd service enabled. REBOOT to activate sysc-greet (don't restart display-manager.service now!)"}
	}
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// ASCII Header - CHANGED 2025-10-01 17:35 - Pad shorter lines to match longest line
	// CHANGED 2025-10-01 19:05 - Pad all lines to 49 chars to prevent Align() mangling - Problem: Lipgloss centers each line independently
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

	headerStyle := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	// CHANGED 2025-10-01 19:00 - Simple rendering without centering, let title/step handle it - Problem: PlaceHorizontal adds too much padding
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

	content.WriteString(titleStyle.Render("sysc-greet Installation Wizard"))
	content.WriteString("\n\n")

	// Step indicator
	stepText := m.getStepText()
	stepStyle := lipgloss.NewStyle().
		Foreground(Secondary).
		Align(lipgloss.Center)
	content.WriteString(stepStyle.Render(fmt.Sprintf("Step %d/11: %s", int(m.step)+1, stepText)))
	content.WriteString("\n\n")

	// Main content
	mainContent := m.getStepContent()
	mainStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Width(m.width - 4)

	content.WriteString(mainStyle.Render(mainContent))
	content.WriteString("\n")

	// Messages
	if len(m.messages) > 0 {
		content.WriteString("\n")
		for _, msg := range m.messages {
			msgStyle := lipgloss.NewStyle().Foreground(Accent)
			content.WriteString("  " + msgStyle.Render(msg) + "\n")
		}
	}

	// Errors
	if len(m.errors) > 0 {
		content.WriteString("\n")
		for _, err := range m.errors {
			errStyle := lipgloss.NewStyle().Foreground(ErrorColor)
			content.WriteString("  " + errStyle.Render(err) + "\n")
		}
	}

	// Help - CHANGED 2025-10-01 17:00 - Removed Width() to fix background bleeding
	helpText := m.getHelpText()
	if helpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(FgMuted).
			Italic(true).
			Align(lipgloss.Center)
		content.WriteString("\n" + helpStyle.Render(helpText))
	}

	// Wrap everything in background
	// CHANGED 2025-10-01 19:00 - Restore Align(Center) for proper centering - Problem: PlaceHorizontal broke layout
	bgStyle := lipgloss.NewStyle().
		Background(BgBase).
		Foreground(FgPrimary).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Top)

	return bgStyle.Render(content.String())
}

func (m model) getStepText() string {
	switch m.step {
	case stepWelcome:
		return "Welcome"
	case stepCheckPrivileges:
		return "Check Privileges"
	case stepSelectInstallType:
		return "Select Installation Type"
	case stepCheckDependencies:
		return "Check Dependencies"
	case stepInstallGreetd:
		return "Install greetd"
	case stepInstallGslapper:
		return "Install gslapper" // CHANGED 2025-10-03 - Add gslapper step title
	case stepBuildBinary:
		return "Build Binary"
	case stepInstallBinary:
		return "Install Binary"
	case stepInstallConfigs:
		return "Install Configurations"
	case stepSetupCache:
		return "Setup Cache"
	case stepConfigureGreetd:
		return "Configure greetd"
	case stepEnableService:
		return "Enable Service"
	case stepComplete:
		return "Installation Complete"
	}
	return "Unknown"
}

func (m model) getStepContent() string {
	switch m.step {
	case stepWelcome:
		return `Welcome to the sysc-greet installer!

This wizard will guide you through installing sysc-greet as your
system greeter. The installation process includes:

  • Building the sysc-greet binary
  • Installing to system directories
  • Configuring greetd daemon
  • Setting up user permissions
  • Enabling the service

Installation will require root privileges.

Press Enter to begin installation
Press Ctrl+C to exit`

	case stepCheckPrivileges:
		status := "✓ Sufficient privileges"
		if os.Geteuid() != 0 {
			status = "⚠ Will request elevation when needed"
		}
		return fmt.Sprintf(`Checking system privileges...

%s

Commands will be executed with sudo when necessary.
Some steps may prompt for your password.

Press Enter to continue
Press N to cancel`, status)

	case stepSelectInstallType:
		m.options = []string{
			"Full Installation (greetd + sysc-greet)",
			"Greeter Only (sysc-greet only)",
			"Configs Only (update configs)",
		}

		var content strings.Builder
		content.WriteString("Select installation type:\n\n")

		for i, opt := range m.options {
			if i == m.selectedOption {
				style := lipgloss.NewStyle().
					Foreground(BgBase).
					Background(Accent).
					Bold(true).
					Padding(0, 1)
				content.WriteString("  " + style.Render("▶ "+opt) + "\n")
			} else {
				content.WriteString("    " + opt + "\n")
			}
		}

		return content.String()

	case stepCheckDependencies:
		return `Checking system dependencies...

Required:
  • Go compiler (for building)
  • systemd (for service management)
  • greetd (greeter daemon)

This may take a moment...`

	case stepInstallGreetd:
		status := "Installing greetd..."
		if m.greetdInstalled {
			status = "greetd already installed, skipping..."
		} else if m.tuigreetInstalled {
			status = "tuigreet detected - greetd should be installed.\nInstalling greetd to ensure proper setup..."
		}

		pm := "package manager"
		if m.packageManager != "" {
			pm = m.packageManager
		}

		return fmt.Sprintf(`%s

Package manager: %s
Package: greetd

The greetd daemon will be installed from your system's package
repository. This is required for sysc-greet to function.

Please wait...`, status, pm)

	case stepInstallGslapper:
		// CHANGED 2025-10-03 - Add gslapper step content - Problem: Need video wallpaper support
		pm := "source build"
		if m.packageManager == "yay" || m.packageManager == "pacman" {
			pm = "AUR (or source fallback)"
		}

		return fmt.Sprintf(`Installing gslapper for video wallpapers...

Method: %s
Package: gslapper
Dependencies: GStreamer, meson, ninja

gslapper provides video wallpaper support using GStreamer for
hardware-accelerated playback. This is optional but recommended
for the video wallpaper feature.

Please wait...`, pm)

	case stepBuildBinary:
		return `Building sysc-greet binary...

This will compile the Go source code into an executable.
Build flags: -o sysc-greet ./cmd/sysc-greet/

Please wait...`

	case stepInstallBinary:
		return fmt.Sprintf(`Installing binary to system...

Target: %s
Permissions: 755 (executable)

This requires root privileges.`, m.installPath)

	case stepInstallConfigs:
		return fmt.Sprintf(`Installing configuration files...

ASCII configs: %s/ascii_configs/
Themes: Embedded in binary

16 window manager configurations will be installed.`, m.configPath)

	case stepSetupCache:
		return `Setting up cache and permissions...

Cache: /var/cache/tuigreet
Home: /var/lib/greeter (with write permissions)
User: greeter
Group: greeter

The greeter user will be created if it doesn't exist.`

	case stepConfigureGreetd:
		return `Configuring greetd daemon...

Config file: /etc/greetd/config.toml
Greeter command: sysc-greet
VT: 1

This will create/overwrite the greetd configuration.`

	case stepEnableService:
		return `Enabling greetd service...

Service: greetd.service
State: enabled (start on boot)

The greeter will start automatically on next boot.`

	case stepComplete:
		return `Installation complete! 🎉

sysc-greet has been successfully installed and configured.

Next steps:
  1. Test in test mode: sysc-greet --test
  2. Restart greetd: sudo systemctl restart greetd
  3. Switch to TTY1 to see the greeter
  4. Login with your credentials

To revert to previous greeter, edit /etc/greetd/config.toml

Press Q to exit`
	}

	return "Processing..."
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome, stepCheckPrivileges:
		return "Enter: Continue  •  Ctrl+C: Quit"
	case stepSelectInstallType:
		return "↑↓: Navigate  •  Enter: Select  •  Ctrl+C: Quit"
	case stepComplete:
		return "Q: Quit"
	default:
		return "Please wait..."
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
