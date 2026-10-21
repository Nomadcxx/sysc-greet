package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
)

func TestANSILoveMenu(t *testing.T) {
	m := model{mode: ModeBackgroundsSubmenu, config: Config{TestMode: true}}
	updated, _ := m.navigateToBackgroundsSubmenu()
	m = updated.(model)
	m.menuIndex = slices.Index(m.menuOptions, "[ ] ANSI Love")
	if m.menuIndex < 0 {
		t.Fatal("ANSI Love missing from background menu")
	}
	m, cmd := m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
	if m.selectedBackground != "ansi-love" {
		t.Fatalf("selected %q", m.selectedBackground)
	}
	if m.ansiLoveEffect != nil || cmd == nil {
		t.Fatal("selecting ANSI Love must schedule its bake asynchronously")
	}
	m.mode = ModeLogin
	if !m.backgroundEffectActive() {
		t.Fatal("ANSI Love inactive at login")
	}
	m.mode = ModeScreensaver
	if !m.backgroundEffectActive() {
		t.Fatal("ANSI Love inactive in screensaver")
	}
	m.mode = ModeMenu
	if m.backgroundEffectActive() {
		t.Fatal("ANSI Love advances behind menu")
	}
}

func TestANSILoveStaleBuildDoesNotReplaceNewer(t *testing.T) {
	m := layoutModel(40, 12)
	m.selectedBackground = "ansi-love"
	stale := m.scheduleANSILoveBuild()
	m.currentTheme = "Nord"
	newer := m.scheduleANSILoveBuild()

	staleMsg, newerMsg := stale(), newer()
	got, _ := m.Update(newerMsg)
	m = got.(model)
	want := m.ansiLoveEffect
	if want == nil {
		t.Fatal("latest build was not installed")
	}
	got, _ = m.Update(staleMsg)
	m = got.(model)
	if m.ansiLoveEffect != want {
		t.Fatal("stale build replaced newer effect")
	}
}

func TestANSILoveConfiguration(t *testing.T) {
	oldProfile := colorProfile
	colorProfile = colorprofile.TrueColor
	t.Cleanup(func() { colorProfile = oldProfile })
	defaults, override, err := loadANSILoveConfig(filepath.Join(t.TempDir(), "missing.conf"))
	if err != nil || override || defaults.Backend != syscgoanimations.ANSILoveQuad || defaults.Depth != syscgoanimations.ANSILoveTrueColor || defaults.Zoom != 1.5 || defaults.Palette != "vga" {
		t.Fatalf("defaults %+v, %v, %v", defaults, override, err)
	}
	path := filepath.Join(t.TempDir(), "ansi-love.conf")
	data := "# ANSI Love\nansi_love_backend=braille-fg\nansi_love_fit=contain\nansi_love_zoom=1.5\nansi_love_cell=9x18\nansi_love_palette=theme\nansi_love_motion=off\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, override, err := loadANSILoveConfig(path)
	if err != nil || !override || cfg.Backend != syscgoanimations.ANSILoveBrailleFG || cfg.Fit != syscgoanimations.ANSILoveFitContain || cfg.Zoom != 1.5 || cfg.CellW != 9 || cfg.CellH != 18 || cfg.Palette != "theme" || cfg.Motion != syscgoanimations.ANSILoveMotionOff {
		t.Fatalf("parsed %+v,%v,%v", cfg, override, err)
	}
	for _, bad := range []string{"ansi_love_backend=nope", "ansi_love_zoom=NaN", "ansi_love_zoom=Inf", "ansi_love_zoom=-1", "ansi_love_zoom=0", "ansi_love_zoom=0.5", "ansi_love_zoom=3.1", "ansi_love_cell=9x0", "ansi_love_cell=9x18junk", "ansi_love_cell=1025x16", "ansi_love_motion=spin", "ansi_love_palette=nope", "ansi_love_fit=stretch", "ansi_love_unknown=yes", "not a key", strings.Repeat("x", 17000)} {
		if err := os.WriteFile(path, []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, override, err := loadANSILoveConfig(path)
		if err == nil || override || cfg != defaults {
			t.Errorf("invalid %q: config %+v, override %v, error %v", bad[:min(len(bad), 100)], cfg, override, err)
		}
	}
}

func TestANSILoveUsesDetectedColorDepth(t *testing.T) {
	oldProfile := colorProfile
	t.Cleanup(func() { colorProfile = oldProfile })
	for _, test := range []struct {
		profile colorprofile.Profile
		depth   syscgoanimations.ANSILoveColorDepth
	}{{colorprofile.TrueColor, syscgoanimations.ANSILoveTrueColor}, {colorprofile.ANSI256, syscgoanimations.ANSILove256}, {colorprofile.ANSI, syscgoanimations.ANSILove16}, {colorprofile.Ascii, syscgoanimations.ANSILove16}, {colorprofile.NoTTY, syscgoanimations.ANSILove16}} {
		colorProfile = test.profile
		if got := defaultANSILoveConfig().Depth; got != test.depth {
			t.Errorf("profile %s selected depth %d, want %d", test.profile, got, test.depth)
		}
	}
}

func TestANSILoveREPExpansionForCanvas(t *testing.T) {
	frame := "\x1b[48;2;1;2;3m \x1b[5bX\x1b[m"
	want := "\x1b[48;2;1;2;3m      X\x1b[m"
	if got := expandANSILoveREP(frame); got != want {
		t.Fatalf("REP expansion %q, want %q", got, want)
	}
}

func TestANSILoveCreditAndBorder(t *testing.T) {
	t.Setenv("SYSC_BG", "ansi-love")
	for _, size := range [][2]int{{40, 12}, {80, 24}, {200, 55}} {
		m := initialModel(Config{TestMode: true}, false)
		m.width, m.height = size[0], size[1]
		m.ansiLoveEffect = m.newANSILoveEffect()
		for _, mode := range []ViewMode{ModeLogin, ModePassword, ModeScreensaver} {
			m.mode = mode
			view := m.View()
			canvas, ok := view.Layer.(interface{ Render() string })
			if !ok {
				t.Fatal("background has no rendered layer")
			}
			lines := strings.Split(ansi.Strip(canvas.Render()), "\n")
			if !strings.HasSuffix(lines[len(lines)-1], "Drawn by Warpus, December 2019") {
				t.Fatalf("%v %s lost final-row credit: %q", size, mode, lines[len(lines)-1])
			}
		}
	}
	m := layoutModel(240, 100)
	m.selectedBackground = "ak-47"
	want := m.renderDualBorderLayout(240, 100)
	m.selectedBackground = "ansi-love"
	if m.renderDualBorderLayout(240, 100) != want {
		t.Fatal("ANSI Love did not suppress outer border like scene backgrounds")
	}
}

func TestANSILoveBackgroundTickSchedulesBuild(t *testing.T) {
	m := layoutModel(40, 12)
	m.selectedBackground = "ansi-love"
	got, cmd := m.Update(bgTickMsg(time.Unix(10, 0)))
	m = got.(model)
	if cmd == nil || m.ansiLoveEffect != nil || !m.ansiLoveBuildPending {
		t.Fatal("first background tick must schedule construction asynchronously")
	}
}

func TestANSILoveAsyncResize(t *testing.T) {
	m := layoutModel(80, 24)
	m.selectedBackground = "ansi-love"
	m.ansiLoveEffect = m.newANSILoveEffect()
	got, old := m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	m = got.(model)
	if old == nil || lipgloss.Width(expandANSILoveREP(m.ansiLoveEffect.Render())) != 80 {
		t.Fatal("resize ran synchronously")
	}
	got, newer := m.Update(tea.WindowSizeMsg{Width: 160, Height: 48})
	m = got.(model)
	newest := newer()
	stale := old()
	got, _ = m.Update(newest)
	m = got.(model)
	got, _ = m.Update(stale)
	m = got.(model)
	if lipgloss.Width(expandANSILoveREP(m.ansiLoveEffect.Render())) != 160 {
		t.Fatal("stale size replaced latest size")
	}
	got, pending := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = got.(model)
	stale = pending()
	m.ansiLoveEffect = m.newANSILoveEffect()
	replacement := m.ansiLoveEffect.Render()
	got, _ = m.Update(stale)
	m = got.(model)
	if m.ansiLoveEffect.Render() != replacement {
		t.Fatal("old instance applied after replacement")
	}
	got, pending = m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = got.(model)
	m.selectedBackground = "none"
	before := m.ansiLoveEffect.Render()
	got, _ = m.Update(pending())
	m = got.(model)
	if m.ansiLoveEffect.Render() != before {
		t.Fatal("disabled effect applied resize")
	}
}

func TestANSILoveElapsedAndFreeze(t *testing.T) {
	oldProfile := colorProfile
	colorProfile = colorprofile.TrueColor
	t.Cleanup(func() { colorProfile = oldProfile })
	m := layoutModel(120, 36)
	m.selectedBackground = "ansi-love"
	m.ansiLoveEffect = m.newANSILoveEffect()
	control := m.newANSILoveEffect()
	initial := control.Render()
	start := time.Unix(10, 0)
	for _, offset := range []time.Duration{0, 50 * time.Millisecond, 100 * time.Millisecond, 150 * time.Millisecond} {
		got, _ := m.Update(bgTickMsg(start.Add(offset)))
		m = got.(model)
		if offset > 0 {
			control.UpdateElapsed(50 * time.Millisecond)
		}
	}
	if m.ansiLoveEffect.Render() != control.Render() {
		t.Fatal("host did not forward real active elapsed time")
	}
	before := m.ansiLoveEffect.Render()
	if before == initial {
		t.Fatal("active interval never animated")
	}
	m.mode = ModeMenu
	got, _ := m.Update(bgTickMsg(start.Add(5 * time.Second)))
	m = got.(model)
	m.mode = ModeLogin
	got, _ = m.Update(bgTickMsg(start.Add(10 * time.Second)))
	m = got.(model)
	if m.ansiLoveEffect.Render() != before {
		t.Fatal("menu inactivity advanced animation")
	}
	got, _ = m.Update(bgTickMsg(start.Add(11 * time.Second)))
	m = got.(model)
	if m.ansiLoveEffect.Render() != before {
		t.Fatal("throttled heartbeat advanced animation")
	}
	got, _ = m.Update(bgTickMsg(start.Add(11*time.Second + 50*time.Millisecond)))
	m = got.(model)
	control.UpdateElapsed(50 * time.Millisecond)
	if m.ansiLoveEffect.Render() != control.Render() {
		t.Fatal("resume jumped phase")
	}
}

func TestANSILoveThemeAndCellOverride(t *testing.T) {
	m := layoutModel(120, 36)
	m.config.TestMode = true
	m.selectedBackground = "ansi-love"
	m.ansiLoveEffect = m.newANSILoveEffect()
	m.ansiLoveEffect.UpdateElapsed(190 * time.Millisecond)
	before, original := m.ansiLoveEffect.Render(), m.ansiLoveEffect
	m.mode = ModeThemesSubmenu
	m.menuOptions = []string{"← Back", "Theme: Nord"}
	m.menuIndex = 1
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.ansiLoveEffect != original || m.ansiLoveEffect.Render() != before {
		t.Fatal("VGA theme update changed scene phase or colors")
	}
	cfg := defaultANSILoveConfig()
	cfg.CellW, cfg.CellH = 9, 18
	m.config.ANSILove = &cfg
	m.config.ANSILoveCellOverride = true
	m.ansiLoveEffect = m.newANSILoveEffect()
	before = m.ansiLoveEffect.Render()
	got, _ := m.Update(uv.CellSizeEvent{Width: 16, Height: 16})
	m = got.(model)
	if m.ansiLoveEffect.Render() != before || m.ansiLoveCellW != 0 {
		t.Fatal("cell reply overrode explicit configuration")
	}
	m.config.ANSILoveCellOverride = false
	got, cellCmd := m.Update(uv.CellSizeEvent{Width: 16, Height: 16})
	m = got.(model)
	if m.ansiLoveCellW != 16 || cellCmd == nil || m.ansiLoveEffect.Render() != before {
		t.Fatal("cell reply did not prepare geometry asynchronously")
	}
	got, _ = m.Update(cellCmd())
	m = got.(model)
	if m.ansiLoveEffect.Render() == before {
		t.Fatal("prepared cell reply did not update geometry")
	}
	before = m.ansiLoveEffect.Render()
	got, _ = m.Update(uv.CellSizeEvent{Width: 2000, Height: 0})
	m = got.(model)
	if m.ansiLoveEffect.Render() != before {
		t.Fatal("invalid terminal cell reply changed geometry")
	}
}

func TestANSILoveThemePaletteBakeIsAsync(t *testing.T) {
	applyTheme("dracula", true)
	t.Cleanup(func() { applyTheme("dracula", true) })
	m := layoutModel(40, 12)
	m.config.TestMode = true
	cfg := defaultANSILoveConfig()
	cfg.Palette = "theme"
	m.config.ANSILove = &cfg
	m.selectedBackground = "ansi-love"
	m.ansiLoveEffect = m.newANSILoveEffect()
	m.ansiLoveEffect.UpdateElapsed(450 * time.Millisecond)
	oldEffect, oldFrame := m.ansiLoveEffect, m.ansiLoveEffect.Render()
	m.mode = ModeThemesSubmenu
	m.menuOptions = []string{"← Back", "Theme: Nord"}
	m.menuIndex = 1

	got, cmd := m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = got
	if cmd == nil || !m.ansiLoveResizePending || m.ansiLoveEffect != oldEffect || m.ansiLoveEffect.Render() != oldFrame {
		t.Fatal("theme palette must bake asynchronously while preserving the current frame")
	}
	prepared, ok := cmd().(ansiLovePreparedMsg)
	if !ok {
		t.Fatal("theme palette did not use the prepared-result path")
	}
	updated, _ := m.Update(prepared)
	m = updated.(model)
	if m.ansiLoveEffect != oldEffect {
		t.Fatal("theme palette update replaced the effect")
	}
	if m.ansiLoveResizePending {
		t.Fatal("theme palette prepared result was rejected")
	}
	if m.ansiLoveEffect.Render() == oldFrame {
		t.Fatal("theme palette update did not change the rendered palette")
	}
}

func TestANSILoveThemeResizeUsesLatestPreparedState(t *testing.T) {
	oldProfile := colorProfile
	colorProfile = colorprofile.TrueColor
	t.Cleanup(func() { colorProfile = oldProfile })
	for _, order := range []string{"theme then resize", "resize then theme"} {
		t.Run(order, func(t *testing.T) {
			applyTheme("dracula", true)
			t.Cleanup(func() { applyTheme("dracula", true) })
			m := layoutModel(40, 12)
			m.config.TestMode = true
			cfg := defaultANSILoveConfig()
			cfg.Palette = "theme"
			m.config.ANSILove = &cfg
			m.selectedBackground = "ansi-love"
			m.ansiLoveEffect = m.newANSILoveEffect()
			m.ansiLoveEffect.UpdateElapsed(450 * time.Millisecond)
			oldEffect := m.ansiLoveEffect

			var themeCmd, resizeCmd tea.Cmd
			if order == "theme then resize" {
				m.mode = ModeThemesSubmenu
				m.menuOptions = []string{"← Back", "Theme: Nord"}
				m.menuIndex = 1
				m, themeCmd = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
				updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
				m, resizeCmd = updated.(model), cmd
			} else {
				updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
				m, resizeCmd = updated.(model), cmd
				m.mode = ModeThemesSubmenu
				m.menuOptions = []string{"← Back", "Theme: Nord"}
				m.menuIndex = 1
				m, themeCmd = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
			}
			if themeCmd == nil || resizeCmd == nil || m.ansiLoveEffect != oldEffect {
				t.Fatal("theme and resize should prepare the existing effect asynchronously")
			}

			first, second := themeCmd(), resizeCmd()
			updated, _ := m.Update(first)
			m = updated.(model)
			updated, _ = m.Update(second)
			m = updated.(model)
			if m.ansiLoveEffect != oldEffect || m.ansiLoveResizePending {
				t.Fatal("latest combined preparation was not applied")
			}
			if got := lipgloss.Width(expandANSILoveREP(m.ansiLoveEffect.Render())); got != 120 {
				t.Fatalf("render width = %d, want latest width 120", got)
			}
			want := m.newANSILoveEffect()
			want.UpdateElapsed(450 * time.Millisecond)
			if got, expected := m.ansiLoveEffect.Render(), want.Render(); got != expected {
				t.Fatal("render does not contain both the latest theme and viewport")
			}
		})
	}
}

func TestANSILoveCellSizeEventPreparesAsync(t *testing.T) {
	m := layoutModel(80, 24)
	m.selectedBackground = "ansi-love"
	m.ansiLoveEffect = m.newANSILoveEffect()
	before := m.ansiLoveEffect.Render()
	got, cmd := m.Update(uv.CellSizeEvent{Width: 10, Height: 16})
	m = got.(model)
	if cmd == nil || m.ansiLoveEffect.Render() != before {
		t.Fatal("cell-size reply baked synchronously")
	}
	got, _ = m.Update(cmd())
	m = got.(model)
	if m.ansiLoveEffect.Render() == before {
		t.Fatal("prepared cell-size layout was not applied")
	}
}

func TestANSILoveSelectionPersists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := layoutModel(80, 24)
	got, _ := m.navigateToBackgroundsSubmenu()
	m = got.(model)
	m.menuIndex = slices.Index(m.menuOptions, "[ ] ANSI Love")
	m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeySpace})
	prefs, err := cache.LoadPreferences()
	if err != nil || prefs == nil || prefs.Background != "ansi-love" {
		t.Fatalf("preferences %#v, %v", prefs, err)
	}
}

func TestANSILoveResizeDuringConfigChange(t *testing.T) {
	for _, change := range []string{"theme", "cell"} {
		t.Run(change, func(t *testing.T) {
			m := layoutModel(80, 24)
			m.config.TestMode = true
			m.selectedBackground = "ansi-love"
			m.ansiLoveEffect = m.newANSILoveEffect()
			got, pending := m.Update(tea.WindowSizeMsg{Width: 160, Height: 48})
			m = got.(model)
			old := pending()
			var latest tea.Cmd
			if change == "cell" {
				got, latest = m.Update(uv.CellSizeEvent{Width: 10, Height: 16})
				m = got.(model)
			} else {
				m.mode = ModeThemesSubmenu
				m.menuOptions = []string{"← Back", "Theme: Nord"}
				m.menuIndex = 1
				m, latest = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
			}
			if latest == nil {
				t.Fatal("configuration invalidated pending resize without replacement")
			}
			got, _ = m.Update(latest())
			m = got.(model)
			got, _ = m.Update(old)
			m = got.(model)
			if lipgloss.Width(expandANSILoveREP(m.ansiLoveEffect.Render())) != 160 {
				t.Fatal("configuration change lost requested dimensions")
			}
		})
	}
}

func TestANSILoveResizeKeepsFixedCredit(t *testing.T) {
	m := layoutModel(80, 24)
	m.selectedBackground = "ansi-love"
	m.ansiLoveEffect = m.newANSILoveEffect()
	got, _ := m.Update(tea.WindowSizeMsg{Width: 160, Height: 48})
	m = got.(model)
	canvas := m.View().Layer.(interface{ Render() string })
	frame := canvas.Render()
	lines := strings.Split(ansi.Strip(frame), "\n")
	if len(lines) != 48 || lipgloss.Width(frame) != 160 || !strings.HasSuffix(lines[47], "Drawn by Warpus, December 2019") {
		t.Fatal("pending resize did not letterbox old art and fix credit to new viewport")
	}
}
