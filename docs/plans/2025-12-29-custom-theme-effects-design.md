# Custom Theme Effects Support

## Problem

Custom themes define colors in TOML files, but background effects (fire, matrix, rain, fireworks) and ASCII effects use hardcoded palettes. When a user selects a custom theme, effects fall through to `default` case and show generic colors instead of the custom theme's colors.

GitHub Issue: #40 - User reported "background effects don't consistently align with theme selections."

## Solution

Generate effect palettes dynamically from custom theme colors.

## Fallback Strategy (3 layers)

```
Layer 1: Custom Theme Check
  ↓ (if not custom or missing colors)
Layer 2: Built-in Theme Switch
  ↓ (if unknown theme name)
Layer 3: Default Palette (always works)
```

### Error Handling

| Scenario | Fallback |
|----------|----------|
| Custom theme missing a color field | Use `#000000` for that field, palette still generates |
| `GetThemeColorStrings` called before themes loaded | Returns `false`, falls to Layer 2 |
| Unknown theme name (not custom, not built-in) | Falls to Layer 3 (default palette) |
| Generated palette too short | Palette functions pad with repeated colors |
| `colorToHex` gets nil color | Returns `#000000` |

**Key principle:** Every palette function always returns a valid palette. No panics, no empty slices.

## Palette Generation from Custom Colors

Each effect maps custom theme colors in a specific order:

### Fire Effect (cool → hot gradient)
```
BgBase → BgActive → Accent → Warning → Danger → Primary → FgPrimary
```

### Matrix Effect (digital rain)
```
BgBase → BgActive → Accent → Secondary → Primary → FgPrimary
```

### Rain Effect (falling drops)
```
Primary → Secondary → Accent → FgMuted
```

### Fireworks Effect (explosion colors)
```
Primary → Secondary → Accent → Warning → Danger → FgPrimary
```

### Screensaver
```
BgBase → Primary → Secondary → Accent → Warning → FgPrimary
```

## Implementation

### File Changes

1. **`internal/themes/colors.go`** - Add helper functions:
   - `ThemeColorStrings` struct (hex strings for palette generation)
   - `GetThemeColorStrings(themeName)` → returns colors + bool (is custom)
   - `colorToHex(color.Color)` → safe hex conversion with nil check

2. **`internal/animations/palettes.go`** - Update each palette function:
   - Import `themes` package
   - Check custom theme first, then fall through to built-in switch
   - Pattern: custom → built-in → default

### Code Pattern

```go
func GetFirePalette(themeName string) []string {
    // Layer 1: Custom theme
    if colors, ok := themes.GetThemeColorStrings(themeName); ok {
        return []string{
            colors.BgBase, colors.BgActive, colors.Accent,
            colors.Warning, colors.Danger, colors.Primary, colors.FgPrimary,
        }
    }

    // Layer 2: Built-in themes (existing switch)
    switch strings.ToLower(themeName) {
    case "dracula":
        // ... existing code
    }

    // Layer 3: Default (existing)
    return GetDefaultFirePalette()
}
```

### Documentation Updates

- Update `docs-src/configuration/themes.md` to note that custom themes automatically work with all background effects
- Update custom theme TOML example comments if needed

## Testing

1. Create/use Example custom theme
2. Test each effect (fire, matrix, rain, fireworks) with custom theme active
3. Verify colors match theme's primary/secondary/accent
4. Test fallback by using unknown theme name
