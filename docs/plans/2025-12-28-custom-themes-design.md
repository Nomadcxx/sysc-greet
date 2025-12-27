# Custom Themes & Per-Session Color Override

**Date:** 2025-12-28
**Status:** Approved

## Overview

Add support for user-defined custom themes via TOML files and per-session ASCII color overrides. This creates a layered system where themes control all UI colors and ASCII configs can optionally override just the ASCII art color.

## Design

### Layered Override System

Priority order (highest to lowest):

1. **ASCII config `color=`** - Overrides ASCII art color only
2. **Custom theme file** - All UI colors
3. **Built-in theme** - All UI colors (fallback)

### Custom Theme Files

**Location:**
- `/usr/share/sysc-greet/themes/` (system-wide)
- `~/.config/sysc-greet/themes/` (user)

**Format:** TOML

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

**Behavior:**
- Scanned at startup
- Added to theme list alongside built-in themes
- Appear in F1 → Themes menu
- Custom theme with same name as built-in overrides it

### Per-Session ASCII Color Override

**Format in ASCII config:**

```ini
# /usr/share/sysc-greet/ascii_configs/hyprland.conf
name=Hyprland
color=#89b4fa

ascii_1=
...
```

**Behavior:**
- If `color=` set → use that color for ASCII art
- If omitted → use theme's Primary color
- Only affects ASCII art, not borders or other UI

### Dynamic Theme Menu

Built-in themes stay hardcoded. Custom themes discovered by scanning directories.

```go
builtInThemes := themes.GetAvailableThemes()
customThemes := themes.ScanCustomThemes(themeDirs)
m.availableThemes = append(builtInThemes, customThemes...)
```

Menu built dynamically from combined list.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Invalid TOML / missing fields | Log warning, skip file |
| Invalid hex color | Log warning, skip file |
| Directory doesn't exist | Silent skip |
| Duplicate theme names | Custom wins over built-in |
| Invalid `color=` in ASCII config | Fall back to theme Primary |

## Implementation

### Files to Modify

| File | Changes |
|------|---------|
| `internal/themes/colors.go` | Add `ScanCustomThemes()`, update `GetTheme()` to check custom themes |
| `cmd/sysc-greet/main.go` | Scan theme dirs at startup, store in model |
| `cmd/sysc-greet/menu.go` | Build theme menu dynamically from `m.availableThemes` |
| `cmd/sysc-greet/ascii.go` | Rename `Colors` → `Color`, use in `getSessionASCII()` if set |
| `cmd/sysc-greet/theme.go` | Update `applyTheme()` to load custom theme if not built-in |

### Documentation Updates

| File | Changes |
|------|---------|
| `docs-src/configuration/themes.md` | Add custom theme section with TOML format |
| `docs-src/features/ascii-art.md` | Document `color=` field |
| `CLAUDE.md` | Update architecture notes |

### New Files

- Example theme: `/usr/share/sysc-greet/themes/example.toml`

### Estimated Scope

- ~100-150 lines of new/modified Go code
- Documentation updates

## Not In Scope

- Hot reload of theme files
- Theme editor/creator UI
- Animated ASCII effects (dormant code stays dormant)
