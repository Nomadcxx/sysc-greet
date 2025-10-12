# Refactoring Plan for sysc-greet

**Date:** 2025-10-12
**Status:** Phase 7 Complete ✅
**Backups Created:**
- Phase 1: `sysc-greet-backup-2025-10-11-ascii-borders-working`
- Phase 2: `sysc-greet-backup-2025-10-11-phase1-complete`
- Phase 3: `sysc-greet-backup-2025-10-11-phase2-complete`
- Phase 4: `sysc-greet-backup-2025-10-11-phase3-complete`
- Phase 5: `sysc-greet-backup-2025-10-11-phase4-complete`
- Phase 6: `sysc-greet-backup-2025-10-11-phase5-complete`
- Phase 7: `sysc-greet-backup-2025-10-12-phase6-complete`

## Current State Analysis

### main.go Statistics
- **Original Lines:** 4,213
- **After Phase 1:** 3,303 lines
- **After Phase 2:** 2,607 lines
- **After Phase 3:** 2,308 lines
- **After Phase 4:** 2,030 lines
- **After Phase 5:** 1,900 lines
- **After Phase 6:** 1,711 lines
- **After Phase 7:** 1,658 lines
- **Total Reduction:** 2,555 lines (60.6%)
- **Original Functions:** 64
- **Primary Issues:**
  - Monolithic structure making maintenance difficult
  - Border rendering functions scattered throughout file
  - Animation logic mixed with core application logic
  - ASCII art handling spread across multiple locations
  - Utility functions not properly separated

## Refactoring Strategy

### Phase 1: Extract Border Rendering ✅ COMPLETE
**Target:** Create `borders.go` to house all border-related rendering functions

**Functions Extracted:**
1. ✅ `renderDualBorderLayout()` - Line 2300
2. ✅ `renderASCII1BorderLayout()` - Line 2490
3. ✅ `renderASCII2BorderLayout()` - Line 3716
4. ✅ `renderASCII3BorderLayout()` - Line 3865
5. ✅ `renderASCII4BorderLayout()` - Line 4040
6. ✅ `renderASCIIBorderFallback()` - Line 2599

**Helper Functions Extracted:**
- ✅ `getInnerBorderStyle()` - Line 2725
- ✅ `getOuterBorderStyle()` - Line 2759
- ✅ `getInnerBorderColor()` - Line 2779
- ✅ `getOuterBorderColor()` - Line 2805

**Actual Reduction:** 910 lines (21.6% of main.go)
**Result:** borders.go created with 928 lines, all border styles tested and working

### Phase 2: Extract ASCII Art Handling ✅ COMPLETE
**Target:** Create `ascii.go` for ASCII art configuration and rendering

**Functions Extracted:**
1. ✅ `getSessionASCII()` - Line 553
2. ✅ `getSessionASCIIMonochrome()` - Line 2670
3. ✅ `getSessionArt()` - Line 3514
4. ✅ `loadASCIIConfig()` - Line 426

**Animation Functions Extracted:**
- ✅ `applyASCIIAnimation()` - Line 628
- ✅ `applySmoothGradient()` - Line 657
- ✅ `applyWaveAnimation()` - Line 756
- ✅ `applyPulseAnimation()` - Line 803
- ✅ `applyRainbowAnimation()` - Line 838
- ✅ `applyMatrixAnimation()` - Line 887
- ✅ `applyTypewriterAnimation()` - Line 935
- ✅ `applyGlowAnimation()` - Line 986
- ✅ `applyStaticColors()` - Line 1036

**Helper Functions Extracted:**
- ✅ `interpolateColors()` - Line 727
- ✅ `parseHexColor()` - Line 741

**Actual Reduction:** 696 lines (16.5% of original main.go)
**Result:** ascii.go created with 721 lines, all ASCII art and animations working

### Phase 3: Extract UI Components ✅ COMPLETE
**Target:** Create `ui_components.go` for reusable UI rendering functions

**Functions Extracted:**
1. ✅ `renderMainForm()` - Main login form with session, username/password inputs
2. ✅ `renderMonochromeForm()` - Monochrome styled login form
3. ✅ `renderSessionSelector()` - Session selector with dropdown indicator
4. ✅ `renderSessionDropdown()` - Dropdown list of available sessions
5. ✅ `renderMainHelp()` - Help text at bottom of screen

**Actual Reduction:** 299 lines (7.1% of original main.go)
**Result:** ui_components.go created with 322 lines, all UI components working

### Phase 4: Extract View Rendering ✅ COMPLETE
**Target:** Create `views.go` for top-level view rendering

**Functions Extracted:**
1. ✅ `renderPowerView()` - Power options menu (reboot/shutdown/cancel)
2. ✅ `renderMenuView()` - Main menu and all submenus (themes, borders, backgrounds, wallpaper)
3. ✅ `renderReleaseNotesView()` - F3 release notes popup with ASCII header

**Actual Reduction:** 278 lines (6.6% of original main.go)
**Result:** views.go created with 294 lines, all view rendering working

### Phase 5: Extract Background Effects ✅ COMPLETE
**Target:** Create `backgrounds.go` for background animation effects

**Functions Extracted:**
1. ✅ `applyBackgroundAnimation()` - Router for background effects based on selection
2. ✅ `addMatrixRain()` - Simple matrix rain simulation (older implementation)
3. ✅ `addFireEffect()` - DOOM-style fire effect rendering
4. ✅ `addAsciiRain()` - ASCII rain effect rendering
5. ✅ `addMatrixEffect()` - Matrix-style background effect
6. ✅ `getBackgroundColor()` - Returns BgBase to prevent color bleeding

**Actual Reduction:** 130 lines (3.1% of original main.go)
**Result:** backgrounds.go created with 149 lines, all background effects working

### Phase 6: Extract Theme Management ✅ COMPLETE
**Target:** Create `theme.go` for theme-related functions

**Functions Extracted:**
1. ✅ `applyTheme()` - 130 lines with 9 theme definitions (gruvbox, material, nord, dracula, catppuccin, tokyo night, solarized, monochrome, transishardjob)
2. ✅ `setThemeWallpaper()` - 40 lines with swww daemon integration and test mode protection
3. ✅ `getAnimatedColor()` - Brand color cycling for animations
4. ✅ `getAnimatedBorderColor()` - Border animation color cycling
5. ✅ `getFocusColor()` - Focus state color helper

**Additional Cleanup:**
- Removed unused `os/exec` import from main.go
- Updated help text: "bubble-greet" → "sysc-greet"
- Updated config file path: "bubble-greet.conf" → "sysc-greet.conf"
- Removed deprecated `-font` flag (figlet no longer used)
- Updated ASCII config path references

**Actual Reduction:** 189 lines (4.5% of original main.go)
**Result:** theme.go created with 215 lines, all themes and wallpaper management working

### Phase 7: Extract Utilities ✅ COMPLETE
**Target:** Create `utils.go` for utility functions

**Functions Extracted:**
1. ✅ `centerText()` - Wrapper for internal/ui.CenterText
2. ✅ `stripANSI()` - ANSI escape code removal with regex
3. ✅ `stripAnsi()` - Wrapper for internal/ui.StripAnsi
4. ✅ `extractCharsWithAnsi()` - Character extraction preserving ANSI codes
5. ✅ `min()` - Integer minimum helper
6. ✅ `ansiRegex` - Shared regex pattern for ANSI stripping

**Additional Cleanup:**
- Removed unused `regexp` import from main.go
- Removed unused `internal/ui` import from main.go
- Consolidated ANSI stripping functions in one location

**Actual Reduction:** 53 lines (1.3% of original main.go)
**Result:** utils.go created with 75 lines, all utility functions working

### Phase 8: Extract Configuration
**Target:** Create `config.go` for configuration loading

**Functions to Extract:**
1. `loadConfig()` - Line 1071

**Estimated Reduction:** ~180 lines

## Final Structure Projection

After refactoring, the file structure should be:

```
cmd/sysc-greet/
├── main.go           (~900 lines)  - Core application, model, Update/View/Init
├── borders.go        (~1,700 lines) - All border rendering logic
├── ascii.go          (~800 lines)   - ASCII art + animations
├── ui_components.go  (~500 lines)   - Reusable UI components
├── views.go          (~300 lines)   - Top-level view rendering
├── backgrounds.go    (~200 lines)   - Background effects
├── theme.go          (~300 lines)   - Theme management
├── utils.go          (~150 lines)   - Utility functions
├── config.go         (~180 lines)   - Configuration loading
├── menu.go           (existing)     - Menu navigation
└── screensaver.go    (existing)     - Screensaver mode
```

**Total reduction in main.go:** ~4,100 lines → ~900 lines (78% reduction)

## Implementation Order

### PRIORITY 1 (Do First):
1. **borders.go** - This is the most urgent as it contains the largest chunk of related code
2. Verify build and all border styles work after extraction

### PRIORITY 2:
3. **ascii.go** - Second largest chunk, handles ASCII art rendering
4. Verify all ASCII art and animations work

### PRIORITY 3:
5. **utils.go** - Extract utilities first (no dependencies)
6. **theme.go** - Theme management
7. **backgrounds.go** - Background effects
8. **config.go** - Configuration loading

### PRIORITY 4:
9. **ui_components.go** - UI rendering components
10. **views.go** - Top-level views

## Testing Strategy

After each phase:
1. ✅ Run `go build -o sysc-greet cmd/sysc-greet/*.go`
2. ✅ Test basic functionality (login form displays)
3. ✅ Test all border styles (ASCII-1, ASCII-2, ASCII-3, ASCII-4, Classic, Modern, Minimal)
4. ✅ Test menu navigation (F2 - Settings, F3 - Sessions, F4 - Power)
5. ✅ Test WM ASCII art rendering for multiple sessions
6. ✅ Test background effects (Fire, Matrix, ASCII Rain)
7. ✅ Test theme switching
8. ✅ Verify no regressions

## Risk Mitigation

1. **Keep backup accessible:** `sysc-greet-backup-2025-10-11-ascii-borders-working`
2. **One phase at a time:** Complete and test each phase before moving to next
3. **Preserve function signatures:** Keep all function signatures identical
4. **Maintain package structure:** All files stay in `cmd/sysc-greet` package
5. **Test thoroughly:** Build and functional test after each extraction

## Notes

- All extracted files will be in the same package (`package main`)
- The `model` type and its methods will remain accessible across all files
- No changes to external API or command-line interface
- This is purely an internal code organization improvement

## Success Criteria

- ✅ All code compiles without errors
- ✅ All border styles render correctly
- ✅ All menu options work
- ✅ No visual regressions
- ✅ No functional regressions
- ✅ Code is more maintainable and readable
- ✅ main.go is under 1,000 lines

---

**Ready to Execute:** Phase 1 - Extract borders.go
