# Refactoring Plan for sysc-greet

**Date:** 2025-10-11
**Status:** Phase 1 Complete ✅
**Backup Created:** `sysc-greet-backup-2025-10-11-ascii-borders-working`

## Current State Analysis

### main.go Statistics
- **Total Lines:** 4,213
- **Total Functions:** 64
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

### Phase 2: Extract ASCII Art Handling
**Target:** Create `ascii.go` for ASCII art configuration and rendering

**Functions to Extract:**
1. `getSessionASCII()` - Line 553
2. `getSessionASCIIMonochrome()` - Line 2670
3. `getSessionArt()` - Line 3514
4. `loadASCIIConfig()` - Line 426

**Animation Functions to Extract:**
- `applyASCIIAnimation()` - Line 628
- `applySmoothGradient()` - Line 657
- `applyWaveAnimation()` - Line 756
- `applyPulseAnimation()` - Line 803
- `applyRainbowAnimation()` - Line 838
- `applyMatrixAnimation()` - Line 887
- `applyTypewriterAnimation()` - Line 935
- `applyGlowAnimation()` - Line 986
- `applyStaticColors()` - Line 1036

**Estimated Reduction:** ~800 lines

### Phase 3: Extract UI Components
**Target:** Create `ui_components.go` for reusable UI rendering functions

**Functions to Extract:**
1. `renderMainForm()` - Line 2961
2. `renderMonochromeForm()` - Line 2615
3. `renderSessionSelector()` - Line 3059
4. `renderSessionDropdown()` - Line 3127
5. `renderMainHelp()` - Line 3478

**Estimated Reduction:** ~500 lines

### Phase 4: Extract View Rendering
**Target:** Create `views.go` for top-level view rendering

**Functions to Extract:**
1. `renderPowerView()` - Line 3197
2. `renderMenuView()` - Line 3251
3. `renderReleaseNotesView()` - Line 3381

**Estimated Reduction:** ~300 lines

### Phase 5: Extract Background Effects
**Target:** Create `backgrounds.go` for background animation effects

**Functions to Extract:**
1. `applyBackgroundAnimation()` - Line 2829
2. `addMatrixRain()` - Line 2845
3. `addFireEffect()` - Line 2877
4. `addAsciiRain()` - Line 2904
5. `addMatrixEffect()` - Line 2931
6. `getBackgroundColor()` - Line 2956

**Estimated Reduction:** ~200 lines

### Phase 6: Extract Theme Management
**Target:** Create `theme.go` for theme-related functions

**Functions to Extract:**
1. `applyTheme()` - Line 168
2. `setThemeWallpaper()` - Line 301
3. `getAnimatedColor()` - Line 3495
4. `getAnimatedBorderColor()` - Line 3501
5. `getFocusColor()` - Line 3507

**Estimated Reduction:** ~300 lines

### Phase 7: Extract Utilities
**Target:** Create `utils.go` for utility functions

**Functions to Extract:**
1. `centerText()` - Line 343
2. `stripANSI()` - Line 390
3. `stripAnsi()` - Line 3589 (Note: duplicate, should consolidate)
4. `extractCharsWithAnsi()` - Line 3595
5. `min()` - Line 3580
6. `interpolateColors()` - Line 727
7. `parseHexColor()` - Line 741

**Estimated Reduction:** ~150 lines

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
