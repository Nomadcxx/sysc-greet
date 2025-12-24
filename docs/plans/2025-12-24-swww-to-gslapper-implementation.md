# swww to gSlapper Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace swww with gSlapper for all wallpaper handling (static + video) in sysc-greet.

**Architecture:** New `internal/wallpaper` package handles gSlapper IPC communication via native Go Unix sockets. Theme changes and wallpaper menu selections send IPC commands. Compositor configs start gSlapper instead of swww-daemon. Fallback to swww for backwards compatibility.

**Tech Stack:** Go (net package for Unix sockets), gSlapper IPC protocol, Bubble Tea TUI

---

## Task 1: Create gSlapper IPC Client Package

**Files:**
- Create: `internal/wallpaper/gslapper.go`

**Step 1: Create the wallpaper package directory**

```bash
mkdir -p internal/wallpaper
```

**Step 2: Write the gSlapper IPC client**

```go
// internal/wallpaper/gslapper.go
package wallpaper

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// GSlapperSocket is the path to the greeter's gSlapper IPC socket
const GSlapperSocket = "/tmp/sysc-greet-wallpaper.sock"

// IsGSlapperRunning checks if gSlapper IPC socket exists
func IsGSlapperRunning() bool {
	_, err := os.Stat(GSlapperSocket)
	return err == nil
}

// SendCommand sends a command to gSlapper via Unix socket and returns the response
func SendCommand(cmd string) (string, error) {
	conn, err := net.DialTimeout("unix", GSlapperSocket, 2*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to connect to gSlapper socket: %w", err)
	}
	defer conn.Close()

	// Set read/write deadline
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Send command
	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(string(buf[:n])), nil
}

// ChangeWallpaper changes the current wallpaper with fade transition
func ChangeWallpaper(path string) error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	// Set fade transition
	SendCommand("set-transition fade")
	SendCommand("set-transition-duration 0.5")

	// Change wallpaper
	resp, err := SendCommand("change " + path)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(resp, "OK") {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// PauseVideo pauses video playback
func PauseVideo() error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	resp, err := SendCommand("pause")
	if err != nil {
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// ResumeVideo resumes video playback
func ResumeVideo() error {
	if !IsGSlapperRunning() {
		return fmt.Errorf("gSlapper is not running")
	}

	resp, err := SendCommand("resume")
	if err != nil {
		return err
	}

	if resp != "OK" {
		return fmt.Errorf("gSlapper error: %s", resp)
	}

	return nil
}

// QueryStatus returns current gSlapper status
func QueryStatus() (string, error) {
	if !IsGSlapperRunning() {
		return "", fmt.Errorf("gSlapper is not running")
	}

	return SendCommand("query")
}
```

**Step 3: Verify package compiles**

Run: `go build ./internal/wallpaper/`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/wallpaper/gslapper.go
git commit -m "feat: add gSlapper IPC client package"
```

---

## Task 2: Update theme.go to Use gSlapper with swww Fallback

**Files:**
- Modify: `cmd/sysc-greet/theme.go:188-229`

**Step 1: Add import for wallpaper package**

At the top of `cmd/sysc-greet/theme.go`, add to imports:

```go
"github.com/Nomadcxx/sysc-greet/internal/wallpaper"
```

**Step 2: Replace setThemeWallpaper function**

Replace the entire `setThemeWallpaper` function (lines 188-229) with:

```go
// setThemeWallpaper sets a theme-specific wallpaper using gSlapper (preferred) or swww (fallback)
func setThemeWallpaper(themeName string, testMode bool) {
	// Never run wallpaper commands in test mode to avoid disrupting user's wallpapers
	if testMode {
		return
	}

	// Normalize theme name for filename
	themeFile := strings.ToLower(strings.ReplaceAll(themeName, " ", "-"))
	wallpaperPath := fmt.Sprintf("/usr/share/sysc-greet/wallpapers/sysc-greet-%s.png", themeFile)

	// Check if wallpaper exists
	if _, err := os.Stat(wallpaperPath); err != nil {
		return
	}

	// Try gSlapper first (preferred)
	if wallpaper.IsGSlapperRunning() {
		go func() {
			if err := wallpaper.ChangeWallpaper(wallpaperPath); err != nil {
				logDebug("gSlapper wallpaper change failed: %v", err)
			}
		}()
		return
	}

	// Fallback to swww if available
	if _, err := exec.LookPath("swww"); err != nil {
		// Neither gSlapper nor swww available, skip silently
		return
	}

	// Use goroutine to avoid blocking the UI
	go func() {
		// First ensure swww-daemon is running
		daemonCmd := exec.Command("swww-daemon")
		daemonCmd.Stdout = nil
		daemonCmd.Stderr = nil
		_ = daemonCmd.Start()

		// Give daemon a moment to start if it wasn't running
		time.Sleep(100 * time.Millisecond)

		// Set wallpaper on all monitors
		cmd := exec.Command("swww", "img", wallpaperPath, "--transition-type", "fade", "--transition-duration", "0.5")
		cmd.Stdout = nil
		cmd.Stderr = nil
		_ = cmd.Run()
	}()
}
```

**Step 3: Verify build compiles**

Run: `go build ./cmd/sysc-greet/`
Expected: No errors

**Step 4: Commit**

```bash
git add cmd/sysc-greet/theme.go
git commit -m "feat: use gSlapper for theme wallpapers with swww fallback"
```

---

## Task 3: Update Wallpaper Menu to Support Static Images

**Files:**
- Modify: `cmd/sysc-greet/wallpaper.go`

**Step 1: Read current wallpaper.go to understand structure**

Run: `head -100 cmd/sysc-greet/wallpaper.go`

**Step 2: Update file extension scanning**

Find the function that scans for wallpaper files and update it to include static image extensions. Look for patterns like `*.mp4` and add:

```go
// Video extensions
videoExts := []string{".mp4", ".mkv", ".webm", ".avi", ".mov"}

// Static image extensions
imageExts := []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}

// Combined extensions
allExts := append(videoExts, imageExts...)
```

**Step 3: Update wallpaper selection to use gSlapper IPC**

Replace any direct gSlapper command execution with IPC calls:

```go
// Instead of exec.Command("gslapper", ...)
if wallpaper.IsGSlapperRunning() {
    wallpaper.ChangeWallpaper(selectedPath)
} else {
    // fallback or error
}
```

**Step 4: Verify build compiles**

Run: `go build ./cmd/sysc-greet/`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/sysc-greet/wallpaper.go
git commit -m "feat: wallpaper menu supports static images via gSlapper IPC"
```

---

## Task 4: Rename "Stop Wallpaper" to "Stop Video Wallpaper"

**Files:**
- Modify: `cmd/sysc-greet/wallpaper.go` or `cmd/sysc-greet/menu.go`

**Step 1: Find the Stop Wallpaper menu option**

Run: `grep -rn "Stop Wallpaper" cmd/sysc-greet/`

**Step 2: Rename to "Stop Video Wallpaper"**

Change the menu option text from "Stop Wallpaper" to "Stop Video Wallpaper"

**Step 3: Update the handler to use gSlapper IPC pause**

```go
// When "Stop Video Wallpaper" is selected
if wallpaper.IsGSlapperRunning() {
    wallpaper.PauseVideo()
}
```

**Step 4: Verify build compiles**

Run: `go build ./cmd/sysc-greet/`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/sysc-greet/wallpaper.go cmd/sysc-greet/menu.go
git commit -m "feat: rename Stop Wallpaper to Stop Video Wallpaper, use IPC pause"
```

---

## Task 5: Update Hyprland Compositor Config

**Files:**
- Modify: `config/hyprland-greeter-config.conf:58`

**Step 1: Replace swww-daemon with gSlapper**

Change line 58 from:
```
exec-once = swww-daemon
```

To:
```
exec-once = gslapper -I /tmp/sysc-greet-wallpaper.sock -o "fill" '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png
```

**Step 2: Commit**

```bash
git add config/hyprland-greeter-config.conf
git commit -m "feat: use gSlapper instead of swww in Hyprland config"
```

---

## Task 6: Update Niri Compositor Config

**Files:**
- Modify: `config/niri-greeter-config.kdl:64`

**Step 1: Replace swww-daemon with gSlapper**

Change line 64 from:
```
spawn-at-startup "swww-daemon"
```

To:
```
spawn-at-startup "gslapper" "-I" "/tmp/sysc-greet-wallpaper.sock" "-o" "fill" "*" "/usr/share/sysc-greet/wallpapers/sysc-greet-default.png"
```

**Step 2: Commit**

```bash
git add config/niri-greeter-config.kdl
git commit -m "feat: use gSlapper instead of swww in Niri config"
```

---

## Task 7: Update Sway Compositor Config

**Files:**
- Modify: `config/sway-greeter-config:33`

**Step 1: Replace swww-daemon with gSlapper**

Change line 33 from:
```
exec swww-daemon
```

To:
```
exec gslapper -I /tmp/sysc-greet-wallpaper.sock -o "fill" '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png
```

**Step 2: Commit**

```bash
git add config/sway-greeter-config
git commit -m "feat: use gSlapper instead of swww in Sway config"
```

---

## Task 8: Update PKGBUILD Dependencies

**Files:**
- Modify: `PKGBUILD`
- Modify: `PKGBUILD-hyprland`
- Modify: `PKGBUILD-sway`

**Step 1: Find current swww dependency lines**

Run: `grep -n "swww" PKGBUILD PKGBUILD-hyprland PKGBUILD-sway`

**Step 2: Update dependencies**

In each PKGBUILD, change the depends array:
- Add `'gslapper'` as a required dependency
- Move `'swww'` to optdepends with description

Example:
```bash
depends=('greetd' 'kitty' 'gslapper' ...)
optdepends=('swww: legacy wallpaper support (fallback)')
```

**Step 3: Commit**

```bash
git add PKGBUILD PKGBUILD-hyprland PKGBUILD-sway
git commit -m "feat: add gSlapper as dependency, make swww optional"
```

---

## Task 9: Update .SRCINFO Files

**Files:**
- Modify: `.SRCINFO`
- Modify: `.SRCINFO-hyprland`
- Modify: `.SRCINFO-sway`

**Step 1: Regenerate .SRCINFO files**

For each PKGBUILD variant:
```bash
makepkg --printsrcinfo > .SRCINFO
```

Or manually update the depends and optdepends lines to match the PKGBUILDs.

**Step 2: Commit**

```bash
git add .SRCINFO .SRCINFO-hyprland .SRCINFO-sway
git commit -m "chore: update .SRCINFO files for gSlapper dependency"
```

---

## Task 10: Update Installer - Add Distro Detection

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Add package manager detection function**

Add after the imports section:

```go
// PackageManager represents a system package manager
type PackageManager struct {
	Name    string
	Install []string // Command to install packages
}

// detectPackageManager detects the system's package manager
func detectPackageManager() *PackageManager {
	// Check for pacman (Arch)
	if _, err := exec.LookPath("pacman"); err == nil {
		return &PackageManager{
			Name:    "pacman",
			Install: []string{"sudo", "pacman", "-S", "--needed", "--noconfirm"},
		}
	}

	// Check for apt (Debian/Ubuntu)
	if _, err := exec.LookPath("apt"); err == nil {
		return &PackageManager{
			Name:    "apt",
			Install: []string{"sudo", "apt", "install", "-y"},
		}
	}

	// Check for dnf (Fedora)
	if _, err := exec.LookPath("dnf"); err == nil {
		return &PackageManager{
			Name:    "dnf",
			Install: []string{"sudo", "dnf", "install", "-y"},
		}
	}

	return nil
}
```

**Step 2: Commit**

```bash
git add cmd/installer/main.go
git commit -m "feat: add package manager detection to installer"
```

---

## Task 11: Update Installer - Add gSlapper Dependency Maps

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Add gSlapper dependency maps**

```go
// gSlapperDeps returns the gSlapper build dependencies for a package manager
func gSlapperDeps(pm *PackageManager) (runtime []string, build []string) {
	switch pm.Name {
	case "pacman":
		runtime = []string{"gstreamer", "gst-plugins-base", "gst-plugins-good", "gst-plugins-bad", "gst-plugins-ugly", "gst-libav"}
		build = []string{"meson", "ninja", "wayland-protocols", "git"}
	case "apt":
		runtime = []string{"gstreamer1.0-tools", "gstreamer1.0-plugins-base", "gstreamer1.0-plugins-good", "gstreamer1.0-plugins-bad", "gstreamer1.0-plugins-ugly", "gstreamer1.0-libav"}
		build = []string{"meson", "ninja-build", "wayland-protocols", "libunwind-dev", "git"}
	case "dnf":
		runtime = []string{"gstreamer1-plugins-base", "gstreamer1-plugins-good", "gstreamer1-plugins-bad-free", "gstreamer1-plugins-ugly", "gstreamer1-libav"}
		build = []string{"meson", "ninja-build", "wayland-protocols-devel", "git"}
	}
	return
}
```

**Step 2: Commit**

```bash
git add cmd/installer/main.go
git commit -m "feat: add gSlapper dependency maps for Arch/Debian/Fedora"
```

---

## Task 12: Update Installer - Add gSlapper Build Function

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Add gSlapper build function**

```go
// installGSlapper installs gSlapper from source
func installGSlapper(m *model) error {
	// Check if already installed
	if _, err := exec.LookPath("gslapper"); err == nil {
		return nil // Already installed
	}

	pm := detectPackageManager()
	if pm == nil {
		return fmt.Errorf("unsupported package manager")
	}

	// Get dependencies
	runtime, build := gSlapperDeps(pm)
	allDeps := append(runtime, build...)

	// Install dependencies
	installCmd := append(pm.Install, allDeps...)
	cmd := exec.Command(installCmd[0], installCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Clone gSlapper
	buildDir := "/tmp/gslapper-build"
	os.RemoveAll(buildDir)

	cloneCmd := exec.Command("git", "clone", "https://github.com/Nomadcxx/gSlapper.git", buildDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone gSlapper: %w", err)
	}

	// Meson setup
	setupCmd := exec.Command("meson", "setup", "build", "--prefix=/usr/local")
	setupCmd.Dir = buildDir
	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr
	if err := setupCmd.Run(); err != nil {
		return fmt.Errorf("meson setup failed: %w", err)
	}

	// Ninja build
	ninjaCmd := exec.Command("ninja", "-C", "build")
	ninjaCmd.Dir = buildDir
	ninjaCmd.Stdout = os.Stdout
	ninjaCmd.Stderr = os.Stderr
	if err := ninjaCmd.Run(); err != nil {
		return fmt.Errorf("ninja build failed: %w", err)
	}

	// Install
	installNinja := exec.Command("sudo", "ninja", "-C", "build", "install")
	installNinja.Dir = buildDir
	installNinja.Stdout = os.Stdout
	installNinja.Stderr = os.Stderr
	if err := installNinja.Run(); err != nil {
		return fmt.Errorf("ninja install failed: %w", err)
	}

	// Cleanup
	os.RemoveAll(buildDir)

	return nil
}
```

**Step 2: Commit**

```bash
git add cmd/installer/main.go
git commit -m "feat: add gSlapper build from source to installer"
```

---

## Task 13: Update Installer - Integrate gSlapper Install into Flow

**Files:**
- Modify: `cmd/installer/main.go`

**Step 1: Find the installation flow**

Run: `grep -n "swww" cmd/installer/main.go`

**Step 2: Replace swww installation with gSlapper**

Find where swww is installed and replace with:

```go
// Install gSlapper (replaces swww)
if err := installGSlapper(m); err != nil {
    return fmt.Errorf("failed to install gSlapper: %w", err)
}
```

**Step 3: Remove swww from uninstall (do NOT uninstall gSlapper)**

Ensure the uninstall function does NOT remove gSlapper - user may use it independently.

**Step 4: Verify build compiles**

Run: `go build ./cmd/installer/`
Expected: No errors

**Step 5: Commit**

```bash
git add cmd/installer/main.go
git commit -m "feat: integrate gSlapper build into installer flow"
```

---

## Task 14: Update Documentation - README.md

**Files:**
- Modify: `README.md`

**Step 1: Find swww references**

Run: `grep -n "swww" README.md`

**Step 2: Update references**

- Replace "swww" with "gSlapper" where appropriate
- Update any installation instructions
- Note gSlapper handles both video and static wallpapers

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README for gSlapper migration"
```

---

## Task 15: Update Documentation - CONFIGURATION.md

**Files:**
- Modify: `CONFIGURATION.md`

**Step 1: Find swww references**

Run: `grep -n "swww" CONFIGURATION.md`

**Step 2: Update references**

- Replace swww references with gSlapper
- Update wallpaper configuration section
- Document IPC socket path `/tmp/sysc-greet-wallpaper.sock`

**Step 3: Commit**

```bash
git add CONFIGURATION.md
git commit -m "docs: update CONFIGURATION for gSlapper migration"
```

---

## Task 16: Build and Test

**Step 1: Build all binaries**

```bash
make build
make installer
```

**Step 2: Test in test mode**

```bash
./sysc-greet --test
```

**Step 3: Verify gSlapper IPC works (if gSlapper running)**

```bash
# In another terminal, start gSlapper manually for testing
gslapper -I /tmp/sysc-greet-wallpaper.sock -o "fill" '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png

# Test IPC
echo "query" | nc -U /tmp/sysc-greet-wallpaper.sock
```

**Step 4: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: address issues found during testing"
```

---

## Task 17: Push to Development Branch

**Step 1: Push all commits**

```bash
git push origin development
```

**Step 2: Review changes**

```bash
git log --oneline -20
```

---

## Summary

Total tasks: 17
Estimated time: 2-3 hours

Key changes:
1. New `internal/wallpaper` package for gSlapper IPC
2. Theme wallpapers use gSlapper with swww fallback
3. Wallpaper menu supports both video and static files
4. All compositor configs use gSlapper
5. Installer builds gSlapper from source with distro-aware deps
6. Documentation updated
