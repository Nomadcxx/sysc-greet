# Custom Themes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add support for user-defined custom themes via TOML files and per-session ASCII color overrides.

**Architecture:** Layered override system where themes control all UI colors and ASCII configs can optionally override just the ASCII art color. Custom themes are TOML files discovered at startup from `/usr/share/sysc-greet/themes/` and `~/.config/sysc-greet/themes/`.

**Tech Stack:** Go, TOML (BurntSushi/toml already in go.mod), lipgloss v2

---

## Task 1: Update ASCIIConfig - Rename Colors to Color

**Files:**
- Modify: `cmd/sysc-greet/main.go:237`
- Modify: `cmd/sysc-greet/ascii.go:89-90`

**Step 1: Update ASCIIConfig struct**

In `cmd/sysc-greet/main.go`, change line 237:

```go
// Before:
Colors             []string

// After:
Color              string   // Optional hex color override for ASCII art (e.g., "#89b4fa")
```

**Step 2: Update parsing in ascii.go**

In `cmd/sysc-greet/ascii.go`, change lines 89-90:

```go
// Before:
case "colors":
    config.Colors = strings.Split(value, ",")

// After:
case "color":
    config.Color = strings.TrimSpace(value)
```

**Step 3: Build to verify no compile errors**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds (Color field not yet used)

**Step 4: Commit**

```bash
git add cmd/sysc-greet/main.go cmd/sysc-greet/ascii.go
git commit -m "refactor: rename ASCIIConfig.Colors to Color (single hex value)"
```

---

## Task 2: Wire color= Field into ASCII Rendering

**Files:**
- Modify: `cmd/sysc-greet/ascii.go:235-238`

**Step 1: Update getSessionASCII() to use color override**

In `cmd/sysc-greet/ascii.go`, replace lines 235-238:

```go
// Before:
// Apply static primary color to entire ASCII art block
style := lipgloss.NewStyle().Foreground(Primary).Background(BgBase)
return style.Render(currentASCII)

// After:
// Determine ASCII color: use config override if set, otherwise theme Primary
var asciiColor color.Color = Primary
if asciiConfig.Color != "" {
    asciiColor = lipgloss.Color(asciiConfig.Color)
}

// Apply color to entire ASCII art block
style := lipgloss.NewStyle().Foreground(asciiColor).Background(BgBase)
return style.Render(currentASCII)
```

**Step 2: Add import for color package**

In `cmd/sysc-greet/ascii.go`, ensure import exists:

```go
import (
    "image/color"
    // ... other imports
)
```

**Step 3: Build to verify**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 4: Test manually**

Run: `./sysc-greet --test`
Expected: ASCII art displays (using theme Primary since no color= set yet)

**Step 5: Commit**

```bash
git add cmd/sysc-greet/ascii.go
git commit -m "feat: wire color= field into ASCII rendering"
```

---

## Task 3: Add Custom Theme Scanner

**Files:**
- Modify: `internal/themes/colors.go`

**Step 1: Add imports and package-level custom themes map**

At top of `internal/themes/colors.go`, after existing imports:

```go
import (
    "image/color"
    "os"
    "path/filepath"
    "strings"

    "github.com/BurntSushi/toml"
    "github.com/charmbracelet/lipgloss/v2"
)

// CustomThemes holds loaded custom theme configurations
var CustomThemes = make(map[string]ThemeColors)
```

**Step 2: Add TOML config struct**

After the imports, add:

```go
// CustomThemeConfig represents the TOML structure for custom theme files
type CustomThemeConfig struct {
    Name   string `toml:"name"`
    Colors struct {
        BgBase      string `toml:"bg_base"`
        BgActive    string `toml:"bg_active"`
        Primary     string `toml:"primary"`
        Secondary   string `toml:"secondary"`
        Accent      string `toml:"accent"`
        Warning     string `toml:"warning"`
        Danger      string `toml:"danger"`
        FgPrimary   string `toml:"fg_primary"`
        FgSecondary string `toml:"fg_secondary"`
        FgMuted     string `toml:"fg_muted"`
        BorderFocus string `toml:"border_focus"`
    } `toml:"colors"`
}
```

**Step 3: Add ScanCustomThemes function**

After GetAvailableThemes(), add:

```go
// ScanCustomThemes scans directories for .toml theme files and loads them
func ScanCustomThemes(dirs []string) []string {
    var names []string
    for _, dir := range dirs {
        if _, err := os.Stat(dir); os.IsNotExist(err) {
            continue // Directory doesn't exist, skip silently
        }

        files, err := filepath.Glob(filepath.Join(dir, "*.toml"))
        if err != nil {
            continue
        }

        for _, f := range files {
            theme, err := loadCustomTheme(f)
            if err != nil {
                // Log warning but continue
                continue
            }

            name := theme.Name
            if name == "" {
                name = strings.TrimSuffix(filepath.Base(f), ".toml")
            }

            CustomThemes[strings.ToLower(name)] = theme
            names = append(names, name)
        }
    }
    return names
}

// loadCustomTheme loads a single custom theme from a TOML file
func loadCustomTheme(path string) (ThemeColors, error) {
    var config CustomThemeConfig
    if _, err := toml.DecodeFile(path, &config); err != nil {
        return ThemeColors{}, err
    }

    name := config.Name
    if name == "" {
        name = strings.TrimSuffix(filepath.Base(path), ".toml")
    }

    return ThemeColors{
        Name:          name,
        BgBase:        lipgloss.Color(config.Colors.BgBase),
        BgElevated:    lipgloss.Color(config.Colors.BgBase), // Same as BgBase
        BgSubtle:      lipgloss.Color(config.Colors.BgBase), // Same as BgBase
        BgActive:      lipgloss.Color(config.Colors.BgActive),
        Primary:       lipgloss.Color(config.Colors.Primary),
        Secondary:     lipgloss.Color(config.Colors.Secondary),
        Accent:        lipgloss.Color(config.Colors.Accent),
        Warning:       lipgloss.Color(config.Colors.Warning),
        Danger:        lipgloss.Color(config.Colors.Danger),
        FgPrimary:     lipgloss.Color(config.Colors.FgPrimary),
        FgSecondary:   lipgloss.Color(config.Colors.FgSecondary),
        FgMuted:       lipgloss.Color(config.Colors.FgMuted),
        FgSubtle:      lipgloss.Color(config.Colors.FgMuted), // Use FgMuted as fallback
        BorderDefault: lipgloss.Color(config.Colors.BgActive),
        BorderFocus:   lipgloss.Color(config.Colors.BorderFocus),
    }, nil
}
```

**Step 4: Build to verify**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/themes/colors.go
git commit -m "feat: add custom theme scanner for TOML files"
```

---

## Task 4: Update GetTheme to Check Custom Themes

**Files:**
- Modify: `internal/themes/colors.go:38-40`

**Step 1: Update GetTheme to check custom themes first**

In `internal/themes/colors.go`, modify the GetTheme function start:

```go
// GetTheme returns theme colors for the given theme name
func GetTheme(themeName string) ThemeColors {
    // Check custom themes first (allows overriding built-ins)
    if theme, ok := CustomThemes[strings.ToLower(themeName)]; ok {
        return theme
    }

    switch strings.ToLower(themeName) {
    // ... rest of existing switch cases
```

**Step 2: Build to verify**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/themes/colors.go
git commit -m "feat: check custom themes in GetTheme"
```

---

## Task 5: Add availableThemes to Model and Scan at Startup

**Files:**
- Modify: `cmd/sysc-greet/main.go` (model struct ~line 323, initialModel ~line 500)

**Step 1: Add availableThemes field to model struct**

In `cmd/sysc-greet/main.go`, after `currentTheme string` (around line 323), add:

```go
currentTheme           string
availableThemes        []string  // Built-in + custom theme names
```

**Step 2: Scan for themes in initialModel**

In `cmd/sysc-greet/main.go`, in the initialModel function (around line 490-510), after loading preferences but before creating the model, add:

```go
// Scan for custom themes
themeDirs := []string{
    "/usr/share/sysc-greet/themes",
    filepath.Join(os.Getenv("HOME"), ".config/sysc-greet/themes"),
}
customThemeNames := themes.ScanCustomThemes(themeDirs)

// Combine built-in and custom themes
availableThemes := themes.GetAvailableThemes()
availableThemes = append(availableThemes, customThemeNames...)
```

**Step 3: Add import for filepath**

Ensure `path/filepath` is imported in main.go.

**Step 4: Set availableThemes in model initialization**

In the model struct initialization (around line 530-560), add:

```go
availableThemes:        availableThemes,
```

**Step 5: Build to verify**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 6: Commit**

```bash
git add cmd/sysc-greet/main.go
git commit -m "feat: scan and store available themes at startup"
```

---

## Task 6: Build Theme Menu Dynamically

**Files:**
- Modify: `cmd/sysc-greet/menu.go:10-26`

**Step 1: Replace hardcoded theme menu**

In `cmd/sysc-greet/menu.go`, replace the entire navigateToThemesSubmenu function:

```go
// navigateToThemesSubmenu switches to the themes submenu
func (m model) navigateToThemesSubmenu() (tea.Model, tea.Cmd) {
    m.menuOptions = []string{"← Back"}
    for _, theme := range m.availableThemes {
        m.menuOptions = append(m.menuOptions, "Theme: "+theme)
    }

    m.mode = ModeThemesSubmenu
    m.menuIndex = 0
    return m, nil
}
```

**Step 2: Build to verify**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add cmd/sysc-greet/menu.go
git commit -m "feat: build theme menu dynamically from availableThemes"
```

---

## Task 7: Update applyTheme for Custom Themes

**Files:**
- Modify: `cmd/sysc-greet/theme.go:21-22`

**Step 1: Add custom theme check at start of applyTheme**

In `cmd/sysc-greet/theme.go`, at the start of applyTheme function (after the function signature), add:

```go
func applyTheme(themeName string, testMode bool) {
    // Check if this is a custom theme
    if theme, ok := themes.CustomThemes[strings.ToLower(themeName)]; ok {
        BgBase = theme.BgBase.(lipgloss.Color)
        BgElevated = BgBase
        BgSubtle = BgBase
        Primary = theme.Primary.(lipgloss.Color)
        Secondary = theme.Secondary.(lipgloss.Color)
        Accent = theme.Accent.(lipgloss.Color)
        FgPrimary = theme.FgPrimary.(lipgloss.Color)
        FgSecondary = theme.FgSecondary.(lipgloss.Color)
        FgMuted = theme.FgMuted.(lipgloss.Color)

        // Set wallpaper for custom theme
        if !testMode {
            setThemeWallpaper(themeName)
        }
        return
    }

    switch strings.ToLower(themeName) {
    // ... rest of existing switch cases
```

**Step 2: Add import for themes package**

In `cmd/sysc-greet/theme.go`, ensure import:

```go
import (
    // ... existing imports
    "github.com/Nomadcxx/sysc-greet/internal/themes"
)
```

**Step 3: Build to verify**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add cmd/sysc-greet/theme.go
git commit -m "feat: apply custom themes in applyTheme"
```

---

## Task 8: Create Example Theme File

**Files:**
- Create: `examples/themes/example.toml`

**Step 1: Create example theme**

```bash
mkdir -p examples/themes
```

Create `examples/themes/example.toml`:

```toml
# Example custom theme for sysc-greet
# Copy to /usr/share/sysc-greet/themes/ or ~/.config/sysc-greet/themes/

name = "Example"

[colors]
bg_base = "#1a1a2e"
bg_active = "#2a2a3e"
primary = "#e94560"
secondary = "#0f3460"
accent = "#16213e"
warning = "#f59e0b"
danger = "#ef4444"
fg_primary = "#ffffff"
fg_secondary = "#cccccc"
fg_muted = "#888888"
border_focus = "#e94560"
```

**Step 2: Commit**

```bash
git add examples/themes/example.toml
git commit -m "docs: add example custom theme file"
```

---

## Task 9: Update Documentation

**Files:**
- Modify: `docs-src/configuration/themes.md`
- Modify: `docs-src/features/ascii-art.md`

**Step 1: Update themes.md**

Replace the "Custom Themes" section in `docs-src/configuration/themes.md`:

```markdown
## Custom Themes

Create custom themes by placing TOML files in:

- `/usr/share/sysc-greet/themes/` (system-wide)
- `~/.config/sysc-greet/themes/` (user)

Custom themes appear in F1 → Themes alongside built-in themes.

### Format

```toml
# my-theme.toml
name = "My Theme"

[colors]
bg_base = "#1a1a2e"
bg_active = "#2a2a3e"
primary = "#e94560"
secondary = "#0f3460"
accent = "#16213e"
warning = "#f59e0b"
danger = "#ef4444"
fg_primary = "#ffffff"
fg_secondary = "#cccccc"
fg_muted = "#888888"
border_focus = "#e94560"
```

All color fields are required. Use hex format (`#RRGGBB`).

An example theme is provided in the repository at `examples/themes/example.toml`.
```

**Step 2: Update ascii-art.md**

Add after the "Format" section in `docs-src/features/ascii-art.md`:

```markdown
## Per-Session Color Override

Override the ASCII art color for a specific session, independent of your selected theme:

```ini
name=Hyprland
color=#89b4fa

ascii_1=
...
```

If `color=` is set, that color is used for ASCII art. If omitted, the theme's primary color is used.
```

**Step 3: Commit**

```bash
git add docs-src/configuration/themes.md docs-src/features/ascii-art.md
git commit -m "docs: document custom themes and color= field"
```

---

## Task 10: Test Implementation

**Step 1: Build**

Run: `go build ./cmd/sysc-greet/`
Expected: Build succeeds

**Step 2: Create test theme**

```bash
mkdir -p ~/.config/sysc-greet/themes
cat > ~/.config/sysc-greet/themes/test.toml << 'EOF'
name = "Test"

[colors]
bg_base = "#1a1a2e"
bg_active = "#2a2a3e"
primary = "#ff0000"
secondary = "#00ff00"
accent = "#0000ff"
warning = "#ffff00"
danger = "#ff00ff"
fg_primary = "#ffffff"
fg_secondary = "#cccccc"
fg_muted = "#888888"
border_focus = "#ff0000"
EOF
```

**Step 3: Run test mode**

Run: `./sysc-greet --test`

Expected:
- F1 → Themes shows "Test" in the list
- Selecting "Test" applies red primary color
- ASCII art displays in red

**Step 4: Test color= override**

Edit an ASCII config to add `color=#00ff00` and verify ASCII displays green regardless of theme.

**Step 5: Clean up test theme**

```bash
rm ~/.config/sysc-greet/themes/test.toml
```

**Step 6: Final commit**

```bash
git add -A
git commit -m "feat: custom themes and per-session ASCII color override

- Add TOML-based custom theme support
- Scan /usr/share/sysc-greet/themes/ and ~/.config/sysc-greet/themes/
- Build theme menu dynamically from discovered themes
- Add color= field to ASCII configs for per-session override
- Update documentation"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Rename Colors to Color in ASCIIConfig | main.go, ascii.go |
| 2 | Wire color= into ASCII rendering | ascii.go |
| 3 | Add custom theme scanner | colors.go |
| 4 | Update GetTheme for custom themes | colors.go |
| 5 | Scan themes at startup | main.go |
| 6 | Dynamic theme menu | menu.go |
| 7 | Apply custom themes | theme.go |
| 8 | Example theme file | examples/themes/example.toml |
| 9 | Documentation | themes.md, ascii-art.md |
| 10 | Test everything | - |
