# Cache Loading Fix Design

**Date:** 2025-01-08
**Status:** Approved
**Issues:** #34 (Aquarium 10x speed from cache), Gslapper wallpaper cache loading

## Problem Statement

Two bugs affect cached background preferences:

1. **Aquarium 10x speed**: When aquarium effect is loaded from cache, it runs at 10x normal speed. Works correctly when selected from menu.

2. **Gslapper not launching**: Video wallpapers selected via gslapper don't restart after reboot.

### Root Causes

**Aquarium speed issue:**
- Cache loading initializes aquarium with fallback 80x30 dimensions (terminal size unknown at initialization)
- First render detects dimension mismatch and calls `Resize(actual_width, actual_height)`
- `Resize()` → `Reset()` clears `frameCount` and corrupts animation timing state
- Menu selection works because by that point `m.width` and `m.height` are already set to actual values

**Gslapper launch issue:**
- `launchGslapperWallpaper()` called during `initialModel()` before compositor is ready
- gslapper needs active Wayland displays to attach video wallpapers
- Compositor may not be fully initialized during cache load phase

**Additional constraint:**
- `addAquariumEffect()` uses value receiver, so dimension tracking (`m.lastAquariumWidth/Height`) doesn't persist across calls
- Cannot change to pointer receiver without major Bubble Tea architecture refactor

## Solution: Deferred Initialization Pattern

Mirror the menu selection behavior: set `selectedBackground` immediately, but defer effect initialization until compositor is ready and dimensions are known.

### Architecture

**Phase 1: Cache Loading (initialModel)**
- Set `m.selectedBackground = "aquarium"` or `m.selectedBackground = "wallpaper:filename"`
- Leave `m.aquariumEffect = nil`
- Do NOT launch gslapper

**Phase 2: First WindowSizeMsg**
- Detect first window size message: `firstSizeMsg := (m.width == 0 && m.height == 0)`
- At this point: compositor is ready, terminal dimensions are known
- Initialize aquarium if `m.selectedBackground == "aquarium" && m.aquariumEffect == nil`
- Launch gslapper if `m.selectedBackground` starts with `"wallpaper:"`

**Phase 3: Rendering**
- No changes to `addAquariumEffect()` or render flow
- Dimensions match from first render, so no Resize() is triggered
- Animation timing state remains intact

### Why This Works

**Aquarium speed fix:**
- Effect is created with actual terminal dimensions from the start
- `m.lastAquariumWidth/Height` match creation dimensions
- `addAquariumEffect()` dimension check never triggers Resize()
- No Reset() call → animation timing preserved → normal speed

**Gslapper launch fix:**
- Compositor is guaranteed ready by first WindowSizeMsg
- Wayland displays are available for gslapper to attach to
- Same timing as manual wallpaper selection from menu

**Matches existing patterns:**
- Menu selection already uses this pattern (effect is nil until selected)
- WindowSizeMsg is standard Bubble Tea checkpoint for "UI is ready"
- No new architectural concepts introduced

## Implementation Changes

### File: cmd/sysc-greet/main.go

**1. Remove broken pendingCachedBackground approach**
- Delete `pendingCachedBackground string` field from model struct (~line 395)
- Remove previous WindowSizeMsg logic that used this flag

**2. Fix cache loading (~lines 694-702)**

Before:
```go
case "aquarium":
    // Initialize aquarium effect with terminal dimensions
    width := m.width
    height := m.height
    if width == 0 {
        width = 80
    }
    if height == 0 {
        height = 30
    }
    fishColors, ... := getThemeColorsForAquarium(m.currentTheme)
    m.aquariumEffect = animations.NewAquariumEffect(...)
    m.lastAquariumWidth = width
    m.lastAquariumHeight = height
```

After:
```go
case "aquarium":
    // selectedBackground is set by line 589, effect will be initialized in WindowSizeMsg
    // Leave m.aquariumEffect = nil
```

Gslapper case:
```go
default:
    if _, isWallpaper := strings.CutPrefix(m.selectedBackground, "wallpaper:"); isWallpaper {
        // Don't launch yet - wait for compositor to be ready in WindowSizeMsg
    }
```

**3. Add WindowSizeMsg initialization (~line 753)**

```go
case tea.WindowSizeMsg:
    firstSizeMsg := (m.width == 0 && m.height == 0)
    m.width = msg.Width
    m.height = msg.Height
    logDebug("Terminal resized: %dx%d", msg.Width, msg.Height)

    // Initialize cached backgrounds on first size message
    if firstSizeMsg {
        // Aquarium: initialize if selected but not yet created
        if m.selectedBackground == "aquarium" && m.aquariumEffect == nil {
            fishColors, waterColors, seaweedColors, bubbleColor, diverColor, boatColor, mermaidColor, anchorColor := getThemeColorsForAquarium(m.currentTheme)
            m.aquariumEffect = animations.NewAquariumEffect(animations.AquariumConfig{
                Width:         m.width,
                Height:        m.height,
                FishColors:    fishColors,
                WaterColors:   waterColors,
                SeaweedColors: seaweedColors,
                BubbleColor:   bubbleColor,
                DiverColor:    diverColor,
                BoatColor:     boatColor,
                MermaidColor:  mermaidColor,
                AnchorColor:   anchorColor,
            })
            m.lastAquariumWidth = m.width
            m.lastAquariumHeight = m.height
            logDebug("Initialized aquarium from cache: %dx%d", m.width, m.height)
        }

        // Gslapper: launch if wallpaper selected
        if wallpaperFileName, isWallpaper := strings.CutPrefix(m.selectedBackground, "wallpaper:"); isWallpaper {
            launchGslapperWallpaper(wallpaperFileName)
            logDebug("Launched gslapper from cache: %s", wallpaperFileName)
        }
    }

    return m, nil
```

## Testing Plan

### Test Mode
```bash
# Build and test
go build ./cmd/sysc-greet/

# Test 1: Aquarium from cache
./sysc-greet --test
# Select aquarium from menu, exit
# Re-run: ./sysc-greet --test
# Verify: Aquarium loads at normal speed (not 10x)

# Test 2: Gslapper from cache (if applicable in test mode)
# Select video wallpaper, exit, re-run
# Verify: Wallpaper launches
```

### Production Mode
```bash
# Build installer
make installer

# Install and test
sudo ./installer
systemctl restart greetd

# Test 1: Aquarium cache
# Login, select aquarium, logout
# Reboot
# Verify: Aquarium displays at normal speed on greeter

# Test 2: Gslapper cache
# Login, select video wallpaper, logout
# Reboot
# Verify: Video wallpaper plays on greeter
```

## Rollback Plan

If testing fails:
```bash
git revert HEAD
make installer
sudo ./installer
systemctl restart greetd
```

Return to commit db7e4c5 known state (aquarium loads but runs fast).

## Success Criteria

- ✅ Aquarium effect loads from cache at normal speed (not 10x)
- ✅ Gslapper video wallpapers launch from cache after reboot
- ✅ Username caching continues to work
- ✅ Manual menu selection still works for both features
- ✅ No regressions in other background effects (fire, matrix, rain, etc.)
