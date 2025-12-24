# swww to gSlapper Migration Design

**Date:** 2025-12-24
**Status:** Approved

## Overview

Replace swww with gSlapper for all wallpaper handling (static + video) in sysc-greet.

## Architecture

### Components

1. **Compositor configs** - Start gSlapper with IPC socket at compositor launch
2. **Go IPC client** - Native Go Unix socket communication (with socat fallback)
3. **Wallpaper menu** - Unified menu listing both video and static files
4. **Theme wallpapers** - Auto-set via IPC on theme change

### Socket Path

`/tmp/sysc-greet-wallpaper.sock` (greeter-isolated, avoids collision with user sessions)

### Startup Flow

```
Compositor starts
    └── gSlapper starts with default wallpaper + IPC socket
    └── kitty + sysc-greet starts
            └── Loads cached theme
            └── Sends IPC "change" command to gSlapper
```

## Go IPC Client

**New file:** `internal/wallpaper/gslapper.go`

```go
package wallpaper

import (
    "net"
    "time"
)

const GSlapperSocket = "/tmp/sysc-greet-wallpaper.sock"

// SendCommand sends a command to gSlapper via Unix socket
func SendCommand(cmd string) (string, error) {
    conn, err := net.DialTimeout("unix", GSlapperSocket, 2*time.Second)
    if err != nil {
        return "", err
    }
    defer conn.Close()

    conn.Write([]byte(cmd + "\n"))

    buf := make([]byte, 1024)
    n, _ := conn.Read(buf)
    return string(buf[:n]), nil
}

// ChangeWallpaper changes the current wallpaper with fade transition
func ChangeWallpaper(path string) error {
    SendCommand("set-transition fade")
    SendCommand("set-transition-duration 0.5")
    _, err := SendCommand("change " + path)
    return err
}

// PauseVideo pauses video playback
func PauseVideo() error {
    _, err := SendCommand("pause")
    return err
}

// ResumeVideo resumes video playback
func ResumeVideo() error {
    _, err := SendCommand("resume")
    return err
}
```

**Fallback:** If gSlapper socket doesn't exist, check for `swww` and use existing swww logic.

## Files to Modify

### Compositor Configs

Replace swww-daemon with gSlapper in:
- `config/hyprland-greeter-config.conf`
- `config/niri-greeter-config.kdl`
- `config/sway-greeter-config`

Change from:
```bash
exec-once = swww-daemon
```

To:
```bash
exec-once = gslapper -I /tmp/sysc-greet-wallpaper.sock -o "fill" '*' /usr/share/sysc-greet/wallpapers/sysc-greet-default.png
```

### Go Code

- `cmd/sysc-greet/theme.go` - Replace swww calls with gSlapper IPC in `setThemeWallpaper()`
- `cmd/sysc-greet/wallpaper.go` - Update wallpaper menu to scan for both video and static files, use gSlapper IPC
- New: `internal/wallpaper/gslapper.go` - IPC client

### Package Files

- `PKGBUILD`, `PKGBUILD-hyprland`, `PKGBUILD-sway` - Add `gslapper` as dependency, make `swww` optional
- `.SRCINFO` files - Update accordingly

### Documentation

- `README.md`, `CONFIGURATION.md` - Update wallpaper references

## Fallback Logic

```go
func setThemeWallpaper(themeName string, testMode bool) {
    if testMode {
        return
    }

    wallpaperPath := fmt.Sprintf("/usr/share/sysc-greet/wallpapers/sysc-greet-%s.png", themeName)

    // Try gSlapper first (preferred)
    if _, err := os.Stat(wallpaper.GSlapperSocket); err == nil {
        wallpaper.ChangeWallpaper(wallpaperPath)
        return
    }

    // Fallback to swww if available
    if _, err := exec.LookPath("swww"); err == nil {
        // existing swww logic...
    }
}
```

**Order of preference:**
1. gSlapper IPC socket exists → use gSlapper
2. swww binary exists → use swww (legacy)
3. Neither available → skip silently

## Installer Updates

### Dependency Detection

Check for required gSlapper build dependencies by distro:

**Arch Linux (pacman):**
```
Runtime: gstreamer, gst-plugins-base, gst-plugins-good, gst-plugins-bad, gst-plugins-ugly, gst-libav
Build: meson, ninja, wayland-protocols
```

**Debian/Ubuntu (apt):**
```
Runtime: gstreamer1.0-tools, gstreamer1.0-plugins-base, gstreamer1.0-plugins-good, gstreamer1.0-plugins-bad, gstreamer1.0-plugins-ugly, gstreamer1.0-libav
Build: meson, ninja-build, wayland-protocols, libunwind-dev
```

**Fedora (dnf):**
```
Runtime: gstreamer1-plugins-base, gstreamer1-plugins-good, gstreamer1-plugins-bad-free, gstreamer1-plugins-ugly, gstreamer1-libav
Build: meson, ninja-build, wayland-protocols-devel
```

### Build Flow

1. Check if gSlapper already installed (`which gslapper`)
2. If not, detect distro and package manager
3. Check/install build dependencies
4. Clone, build, install gSlapper

### Uninstall Behavior

- Do NOT uninstall gSlapper (user may use it independently)
- Only remove sysc-greet specific files

## UI Changes

### Wallpaper Menu

- Scan for both video (`*.mp4`, `*.mkv`, `*.webm`) and static (`*.png`, `*.jpg`, `*.jpeg`, `*.webp`) files
- All wallpaper changes go through gSlapper IPC

### Stop Video Wallpaper

- Rename to "Stop Video Wallpaper"
- Sends `pause` command via IPC
- Only relevant for video wallpapers

## Testing Plan

### Manual Checklist

1. **Compositor startup** - gSlapper starts with default wallpaper, IPC socket created
2. **Theme change** - Wallpaper changes via IPC when switching themes in menu
3. **Wallpaper menu** - Lists both `.mp4` and `.png`/`.jpg` files
4. **Video playback** - Video wallpapers loop correctly
5. **Stop Video Wallpaper** - Pauses video, resumes on selection
6. **Fallback** - If gSlapper socket missing but swww installed, falls back to swww
7. **Test mode** - No wallpaper changes in `--test` mode
8. **Installer** - Correctly detects distro, installs deps, builds gSlapper

### Test Environments

- Hyprland compositor
- Niri compositor
- Sway compositor

## Dependencies

- **Hard dependency:** `gslapper`
- **Optional (fallback):** `swww`
