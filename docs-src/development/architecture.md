# Development Architecture

This document describes the architecture and design of sysc-greet.

## Overview

sysc-greet is a graphical greeter for [greetd](https://git.sr.ht/~kennylevinsen/greetd), written in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework (v2).

## Project Structure

```
sysc-greet/
├── cmd/
│   ├── installer/       # Interactive installation wizard
│   └── sysc-greet/    # Main greeter binary
│       ├── main.go       # Application entry point, model, update loop
│       ├── theme.go       # Theme application and wallpaper management
│       ├── ascii.go       # ASCII art loading and parsing
│       ├── wallpaper.go   # Wallpaper menu and gSlapper/swww handling
│       ├── menu.go        # Menu system and navigation
│       ├── screensaver.go # Screensaver mode and idle detection
│       ├── ui_components.go # Reusable UI components
│       ├── utils.go       # Helper functions
│       └── views.go       # View rendering for different modes
├── internal/
│   ├── animations/    # Background and text effects
│   │   ├── fire.go       # DOOM PSX fire effect
│   │   ├── rain.go       # ASCII rain effect
│   │   ├── matrix.go     # Matrix rain effect
│   │   ├── fireworks.go  # Firework particle system
│   │   ├── aquarium.go   # Animated aquarium scene
│   │   ├── ticker.go     # Typewriter and scrolling ticker
│   │   ├── print_effect.go # Print animation for ASCII
│   │   ├── beams_text.go # Beams text effect
│   │   └── pour.go       # Pour text effect
│   ├── cache/          # User preferences persistence
│   ├── ipc/            # greetd IPC client
│   ├── sessions/       # XDG session detection
│   ├── themes/         # Theme definitions (colors.go, themes.go)
│   └── wallpaper/      # gSlapper IPC client
├── ascii_configs/        # Session ASCII art configurations
├── config/              # Compositor configuration templates
└── fonts/               # Figlet font files
```

## Core Components

### Model-View-Update Pattern

sysc-greet follows Bubble Tea's Elm architecture:
- **Model** - Holds all application state
- **View** - Renders the UI based on model state
- **Update** - Handles messages and transitions state
- **Init** - Returns initial commands

### State Management

The `model` struct in `cmd/sysc-greet/main.go` maintains:
- User inputs (username, password)
- Selected session and preferences
- Current mode (login, password, menu, screensaver, etc.)
- Animation state and timers
- IPC client connection

### View Modes

sysc-greet operates in different view modes:

| Mode | Description |
|-------|-------------|
| ModeLogin | Initial login screen with username input |
| ModePassword | Password entry screen |
| ModeMenu | Main settings menu |
| ModeThemesSubmenu | Theme selection |
| ModeBordersSubmenu | Border style selection |
| ModeBackgroundsSubmenu | Background effects selection |
| ModeWallpaperSubmenu | Wallpaper selection |
| ModeASCIIEffectsSubmenu | ASCII text effects selection |
| ModePower | Power menu (shutdown/reboot) |
| ModeReleaseNotes | Release notes display |
| ModeScreensaver | Idle screensaver |

## TTY Compatibility

sysc-greet uses `colorprofile` library to detect terminal capabilities and provide appropriate fallback colors:

- **TrueColor terminals** - Full 24-bit RGB color support
- **ANSI256 terminals** - 256-color palette support
- **Basic TTY** - Falls back to 16 ANSI colors

Color rendering uses `lipgloss.Complete()` function which selects the best available color depth.

## IPC Communication

### greetd Client

sysc-greet communicates with greetd via Unix socket for authentication:

1. Create session with greetd
2. Send create session request with username and session
3. Receive success/failure response
4. On success, start the selected session

### gSlapper IPC

Wallpaper management uses gSlapper IPC protocol via Unix socket at `/tmp/sysc-greet-wallpaper.sock`:

- `IsGSlapperRunning()` - Check if gSlapper socket exists
- `ChangeWallpaper(path)` - Change wallpaper with fade transition
- `PauseVideo()` - Pause video playback
- `ResumeVideo()` - Resume video playback

### Fallback Behavior

If gSlapper is not available, sysc-greet falls back to swww for wallpaper management.

## Configuration System

### User Preferences

Stored in `/var/cache/sysc-greet/`:

- Selected theme
- Selected background effect
- Selected wallpaper
- Selected border style
- Last session and username (if `--remember-username` enabled)
- ASCII variant index

### Session Detection

sysc-greet reads XDG session files from standard paths:

- `/usr/share/xsessions`
- `/usr/share/wayland-sessions`
- `/run/current-system/sw/share/xsessions`
- `/run/current-system/sw/share/wayland-sessions`

Each session is parsed for:
- Name (from `Name=` field)
- Exec command (from `Exec=` field)
- Type (X11 or Wayland, based on path)

### ASCII Config System

Each session can have a `.conf` file in `/usr/share/sysc-greet/ascii_configs/`:

Parsed fields:
- `name` - Display name
- `ascii_1`, `ascii_2`, etc. - Multiple ASCII art variants
- `colors` - Hex colors for rainbow effect
- `roasts` - Custom roast messages separated by `│`

## Animation System

### Background Effects

All background effects implement a common interface:

- `Update(frame)` - Advance animation by one frame
- Initialized with width and height
- Use theme colors for rendering

Effects:
- Fire - PSX DOOM algorithm with particle system
- Matrix - Falling characters with trail effect
- ASCII Rain - Falling ASCII using theme colors
- Fireworks - Particle explosion system
- Aquarium - Swimming fish with bubble particles

### ASCII Effects

- Typewriter - Character-by-character text display
- Print - Line-by-line ASCII rendering
- Beams - Horizontal color beam scanning
- Pour - Character cascade with gradient

## Key Bindings

sysc-greet handles key input centrally:

| Key | Action |
|------|--------|
| F1 | Settings menu |
| F2 | Session selection |
| F3 | Release notes |
| F4 | Power menu |
| Page Up/Down | Cycle ASCII variants |
| Tab | Focus navigation |
| Enter | Submit/Select |
| Esc | Cancel/Back |
| Ctrl+C | Blocked in production mode (security) |

## Security Features

### CAPS LOCK Detection

sysc-greet uses Kitty keyboard protocol to detect CAPS LOCK state:
- Requests `RequestUniformKeyLayout` on init
- Checks `tea.ModCapsLock` in key messages
- Displays warning when CAPS LOCK is on

### Failed Attempt Tracking

Tracks consecutive failed authentication attempts:
- Displays warning message after failed login
- Resets counter on successful authentication
- Returns to username field on failure (not password)

### Test Mode

The `--test` flag bypasses greetd IPC for safe testing:
- Uses mock sessions if none found
- Does not execute user sessions
- Allows Ctrl+C to exit
- Disables wallpaper changes in production

## Performance Considerations

### Non-blocking Operations

Wallpaper changes and IPC calls run in goroutines to avoid blocking the UI:
```go
go func() {
    if err := wallpaper.ChangeWallpaper(path); err != nil {
        logDebug("gSlapper wallpaper change failed: %v", err)
    }
}()
```

### Lazy Initialization

Effects are initialized on first use when terminal dimensions are known:
- Aquarium effect created when first selected with valid dimensions
- gSlapper launched when first needed, not at startup

## Dependencies

- [charmbracelet/bubbletea/v2](https://github.com/charmbracelet/bubbletea) - TUI framework
- [charmbracelet/lipgloss/v2](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [charmbracelet/colorprofile](https://github.com/charmbracelet/colorprofile) - Terminal capability detection
- [mbndr/figlet4go](https://github.com/mbndr/figlet4go) - ASCII art generation

## Build System

- Go 1.25.1 toolchain
- Build tags and commit info injected via ldflags
- Version info available at runtime for debugging
