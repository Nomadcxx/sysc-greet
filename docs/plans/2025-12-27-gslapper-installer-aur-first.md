# gSlapper Installer: AUR-First Installation Design

**Date:** 2025-12-27
**Status:** Approved

## Overview

Update the sysc-greet installer to use AUR packages first when installing gSlapper on Arch-based systems, falling back to source builds when AUR is unavailable or fails.

## Current Behavior

The installer always builds gSlapper from source:
1. Clone from GitHub
2. Build with meson
3. Install to `/usr/local/bin`

## New Behavior

### Installation Flow

```
Is Arch-based? ──No──> Source Build
      │
     Yes
      │
Has AUR helper? ──No──> Source Build
      │
     Yes
      │
AUR install (as $SUDO_USER)
      │
  Success? ──No──> Source Build
      │
     Yes
      │
    Done
```

### Uninstall Flow

```
Check pacman -Qi gslapper ──Found──> pacman -R --noconfirm gslapper [gslapper-debug]
      │
  Not found
      │
Remove /usr/local/bin/gslapper and /usr/local/bin/gslapper-holder
```

## Implementation Details

### Helper Functions

```go
// isArchBased checks if running on Arch or Arch-based distro
func isArchBased() bool {
    if _, err := os.Stat("/etc/arch-release"); err == nil {
        return true
    }
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
```

### Privilege Handling

AUR helpers refuse to run as root. Solution: use `sudo -u $SUDO_USER` to run as original user.

```go
originalUser := os.Getenv("SUDO_USER")
if isArch && aurHelper != "" && originalUser != "" {
    cmd := exec.Command("sudo", "-u", originalUser, aurHelper, "-S", "--noconfirm", "gslapper")
    // ...
}
```

### Install Function

```go
func installGslapper(m *model) tea.Cmd {
    return func() tea.Msg {
        taskIndex := findTaskIndex(m, "Install gSlapper")

        // Detect environment
        isArch := isArchBased()
        aurHelper := ""
        originalUser := os.Getenv("SUDO_USER")

        if isArch && originalUser != "" {
            aurHelper = detectAURHelper()
        }

        // Try AUR if applicable
        if isArch && aurHelper != "" && originalUser != "" {
            cmd := exec.Command("sudo", "-u", originalUser, aurHelper, "-S", "--noconfirm", "gslapper")
            if err := runCommand("Install gSlapper (AUR)", cmd, m); err == nil {
                return taskCompleteMsg{index: taskIndex}
            }
            // AUR failed, continue to source build
        }

        // Fallback: Build from source
        return installGslapperFromSource(m, taskIndex)
    }
}
```

### Source Build Function

```go
func installGslapperFromSource(m *model, taskIndex int) tea.Msg {
    tmpDir := "/tmp/gslapper-build"
    os.RemoveAll(tmpDir)

    // Clone
    cmd := exec.Command("git", "clone", "https://github.com/Nomadcxx/gSlapper.git", tmpDir)
    if err := runCommand("Clone gSlapper", cmd, m); err != nil {
        return taskFailMsg{index: taskIndex, err: err}
    }

    // Meson setup
    cmd = exec.Command("meson", "setup", "build", "--prefix=/usr/local")
    cmd.Dir = tmpDir
    if err := runCommand("Meson setup", cmd, m); err != nil {
        return taskFailMsg{index: taskIndex, err: err}
    }

    // Compile
    cmd = exec.Command("meson", "compile", "-C", "build")
    cmd.Dir = tmpDir
    if err := runCommand("Meson compile", cmd, m); err != nil {
        return taskFailMsg{index: taskIndex, err: err}
    }

    // Install
    cmd = exec.Command("meson", "install", "-C", "build")
    cmd.Dir = tmpDir
    if err := runCommand("Meson install", cmd, m); err != nil {
        return taskFailMsg{index: taskIndex, err: err}
    }

    return taskCompleteMsg{index: taskIndex}
}
```

### Uninstall Function

```go
func uninstallGslapper(m *model) tea.Cmd {
    return func() tea.Msg {
        taskIndex := findTaskIndex(m, "Uninstall gSlapper")

        isArch := isArchBased()
        packagesToRemove := []string{}

        if isArch {
            for _, pkg := range []string{"gslapper", "gslapper-debug"} {
                cmd := exec.Command("pacman", "-Qi", pkg)
                if err := cmd.Run(); err == nil {
                    packagesToRemove = append(packagesToRemove, pkg)
                }
            }
        }

        var cmd *exec.Cmd
        if len(packagesToRemove) > 0 {
            args := append([]string{"-R", "--noconfirm"}, packagesToRemove...)
            cmd = exec.Command("pacman", args...)
        } else {
            cmd = exec.Command("rm", "-f", "/usr/local/bin/gslapper", "/usr/local/bin/gslapper-holder")
        }

        if err := runCommand("Uninstall gSlapper", cmd, m); err != nil {
            return taskFailMsg{index: taskIndex, err: err}
        }

        return taskCompleteMsg{index: taskIndex}
    }
}
```

## Sub-task Display

Each operation shows detailed sub-tasks:

**AUR Install (Arch):**
```
Installing gSlapper                    [OK]
  ├─ Detecting distro and AUR helper   [OK]
  ├─ Installing via yay (as nomadx)    [OK]
```

**Source Build (fallback or non-Arch):**
```
Installing gSlapper                    [OK]
  ├─ Detecting distro and AUR helper   [OK]
  ├─ Installing via yay (as nomadx)    [FAIL]
  ├─ Cloning gSlapper from GitHub      [OK]
  ├─ Configuring build (meson)         [OK]
  ├─ Compiling gSlapper                [OK]
  └─ Installing to /usr/local/bin      [OK]
```

## Benefits

1. **Faster installs** - Binary from AUR vs compiling from source
2. **Better package management** - Tracked by pacman, proper uninstall
3. **Automatic updates** - Users can update via AUR helper
4. **Fallback safety** - Non-Arch systems still work via source build

## Files to Modify

- `cmd/installer/main.go` - Add helper functions and update install/uninstall logic
