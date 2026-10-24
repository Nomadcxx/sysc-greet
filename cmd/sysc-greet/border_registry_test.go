package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/Nomadcxx/sysc-greet/internal/ansiart"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestEmbeddedBorderAssets(t *testing.T) {
	want := map[string]struct {
		w, h       int
		left, top  int
		right, bot int
	}{
		"ascii2": {16, 5, 6, 2, 6, 2},
		"nfo01":  {24, 9, 8, 3, 8, 3},
		"nfo02":  {12, 5, 4, 2, 4, 2},
		"nfo03":  {16, 5, 6, 2, 6, 2},
	}
	for id, dims := range want {
		data, err := builtInBorderAssets.ReadFile("assets/borders/" + id + ".ans")
		if err != nil {
			t.Fatalf("read %s: %v", id, err)
		}
		b, err := ansiart.LoadBorder(data)
		if err != nil {
			t.Fatalf("load %s: %v", id, err)
		}
		if b.Name != id || b.Scale != 1 || b.Art.W != dims.w || b.Art.H != dims.h ||
			b.Left != dims.left || b.Top != dims.top || b.Right != dims.right || b.Bottom != dims.bot {
			t.Errorf("%s metadata = %+v, art %dx%d", id, b, b.Art.W, b.Art.H)
		}
		if b.TitleRow != -1 {
			t.Errorf("%s TitleRow = %d, want no title cap (-1)", id, b.TitleRow)
		}
		titleFound := false
		for _, comment := range bComments(t, data) {
			if strings.Contains(comment, "title=-1") {
				titleFound = true
			}
			if strings.Contains(comment, "title=") && !strings.Contains(comment, "title=-1") || strings.Contains(comment, "ramp=") {
				t.Errorf("%s carries an unapproved title or palette ramp: %q", id, comment)
			}
		}
		if !titleFound {
			t.Errorf("%s metadata must declare title=-1", id)
		}
	}
	registry := newBorderRegistry("")
	for id := range want {
		if !registry.has(id) {
			t.Errorf("registry omitted %s", id)
		}
	}
}

func TestBordersMenuUsesRegisteredStylesAndEffects(t *testing.T) {
	m := model{
		borderRegistry:      newBorderRegistry(""),
		selectedBorderStyle: "minimal",
		borderEffect:        "off",
		formMode:            "auto",
		borderPaletteMode:   "theme",
	}
	updated, _ := m.navigateToBordersSubmenu()
	menu := updated.(model)
	options := strings.Join(menu.menuOptions, "\n")
	for _, label := range []string{"Style: NFO 01: Cyan Bevel", "Style: NFO 02: Shaded Double-Line", "Style: NFO 03: Gradient Bar", "Style: Braille Hairline", "Palette: Theme", "Palette: VGA", "Palette: Terminal", "Effect: Off"} {
		if !strings.Contains(options, label) {
			t.Errorf("border menu omitted %q", label)
		}
	}
	if strings.Contains(options, "Effect: Shimmer") || strings.Contains(options, "Palette cycle") {
		t.Fatalf("minimal menu exposes an unsupported effect: %s", options)
	}

	m.selectedBorderStyle = "nfo01"
	updated, _ = m.navigateToBordersSubmenu()
	if !strings.Contains(strings.Join(updated.(model).menuOptions, "\n"), "Effect: Shimmer") {
		t.Fatal("NFO-01 menu omitted its supported shimmer effect")
	}
}

func TestBordersMenuSpaceAppliesPaletteAndFormWithoutClosing(t *testing.T) {
	m := model{
		config:              Config{TestMode: true},
		borderRegistry:      newBorderRegistry(""),
		selectedBorderStyle: "nfo01",
		borderEffect:        "off",
		borderPaletteMode:   "theme",
		formMode:            "auto",
	}
	updated, _ := m.navigateToBordersSubmenu()
	m = updated.(model)

	for _, tc := range []struct {
		option string
		check  func(model) bool
	}{
		{option: "Palette: VGA", check: func(m model) bool { return m.borderPaletteMode == "vga" }},
		{option: "Form: Frosted", check: func(m model) bool { return m.formMode == "frosted" }},
	} {
		t.Run(tc.option, func(t *testing.T) {
			option := "[ ] " + tc.option
			for i, candidate := range m.menuOptions {
				if candidate == option {
					m.menuIndex = i
					break
				}
			}
			if m.menuOptions[m.menuIndex] != option {
				t.Fatalf("menu omitted %q: %v", option, m.menuOptions)
			}

			updated, _ := m.handleKeyInput(tea.KeyPressMsg{Code: ' '})
			m = updated
			if m.mode != ModeBordersSubmenu {
				t.Fatalf("mode = %s after Space; want Borders submenu to remain open", m.mode)
			}
			if !tc.check(m) {
				t.Fatalf("Space did not apply %q", tc.option)
			}
			if m.menuOptions[m.menuIndex] != "[✓] "+tc.option {
				t.Fatalf("selected row = %q, want checked %q", m.menuOptions[m.menuIndex], tc.option)
			}
		})
	}
}

func TestBordersMenuAppliesBorderOptionsWithEnter(t *testing.T) {
	m := model{
		config:              Config{TestMode: true},
		mode:                ModeMenu,
		menuOptions:         []string{"Close Menu", "Themes", "Borders"},
		menuIndex:           2,
		borderRegistry:      newBorderRegistry(""),
		selectedBorderStyle: "minimal",
		borderEffect:        "off",
	}
	selectOption := func(option string) {
		t.Helper()
		optionIndex := -1
		for i, candidate := range m.menuOptions {
			if candidate == option {
				optionIndex = i
				break
			}
		}
		if optionIndex < 0 {
			t.Fatalf("menu omitted %q: %v", option, m.menuOptions)
		}
		for m.menuIndex < optionIndex {
			m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyDown})
		}
		m, _ = m.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
	}

	selectOption("Borders")
	selectOption("[ ] Style: NFO 01: Cyan Bevel")
	if m.mode != ModeLogin || m.selectedBorderStyle != "nfo01" {
		t.Fatalf("mode = %s, border style = %q; want login with NFO 01 selected", m.mode, m.selectedBorderStyle)
	}

	m = m.navigateToMainMenu()
	selectOption("Borders")
	selectOption("[ ] Effect: Shimmer")
	if m.mode != ModeLogin || m.borderEffect != "shimmer" {
		t.Fatalf("mode = %s, border effect = %q; want login with shimmer applied", m.mode, m.borderEffect)
	}

	m = m.navigateToMainMenu()
	selectOption("Borders")
	selectOption("[ ] Effect: Off")
	if m.borderEffect != "off" {
		t.Fatalf("border effect = %q, want off", m.borderEffect)
	}

	m = m.navigateToMainMenu()
	selectOption("Borders")
	selectOption("[ ] Palette: VGA")
	if m.borderPaletteMode != "vga" {
		t.Fatalf("border palette = %q, want vga", m.borderPaletteMode)
	}

	m = m.navigateToMainMenu()
	selectOption("Borders")
	selectOption("[ ] Form: Frosted")
	if m.formMode != "frosted" {
		t.Fatalf("form mode = %q, want frosted", m.formMode)
	}
}

func TestANSIAndNFOBordersApplyPaletteFormAndShimmer(t *testing.T) {
	useBorderGoldenTheme(t)
	registry := newBorderRegistry("")
	for _, id := range []string{"ascii1", "ascii2", "nfo01", "nfo02", "nfo03"} {
		t.Run(id, func(t *testing.T) {
			m := layoutModel(132, 40)
			m.config = Config{BorderLayout: "hug", FormDim: 0.55, FormHalo: 1}
			m.selectedSession = nil
			m.sessions = nil
			m.borderRegistry = registry
			m.selectedBorderStyle = id
			m.borderPaletteMode = "theme"
			m.formMode = "opaque"
			m.borderEffect = "off"
			m.idleTimer = time.Now().Add(-10*time.Second - 25*time.Millisecond)

			render := func() string {
				t.Helper()
				canvas, ok := m.borderedLoginCanvas(m.width, m.height)
				if !ok {
					t.Fatal("bordered login did not fit")
				}
				return canvas.Render()
			}
			theme := render()

			m.borderPaletteMode = "vga"
			if vga := render(); vga == theme {
				t.Fatal("VGA palette did not change the rendered border")
			}

			m.borderPaletteMode = "theme"
			m.formMode = "frosted"
			if frosted := render(); frosted == theme {
				t.Fatal("Frosted form mode did not change the rendered login view")
			}

			m.formMode = "opaque"
			m.borderEffect = "shimmer"
			changed := false
			for frame := 0; frame < 120; frame++ {
				elapsed := time.Duration(frame)*50*time.Millisecond + 25*time.Millisecond
				m.idleTimer = time.Now().Add(-10*time.Second - elapsed)
				if render() != theme {
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("Shimmer did not change the rendered border over 6 seconds of animation")
			}
		})
	}
}

func TestCustomBorderLabelsRemainSelectableWhenDuplicated(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"one", "two"} {
		grid := ansiart.NewGrid(3, 3)
		data := grid.ToCP437ANS(ansiart.Sauce{
			Title: "Custom border", DataType: 1, FileType: 1, TInfoS: "IBM VGA",
			Comments: []string{"nineslice l=1 t=1 r=1 b=1 anchor=start name=" + id + " label=Shared"},
		})
		if err := os.WriteFile(filepath.Join(dir, id+".ans"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	registry := newBorderRegistry(dir)
	first, second := registry.get("one"), registry.get("two")
	if first == nil || second == nil {
		t.Fatal("registry omitted a custom border")
	}
	if first.Label() == second.Label() {
		t.Fatalf("duplicate labels remain ambiguous: %q", first.Label())
	}

	base := model{
		borderRegistry: registry, selectedBorderStyle: "classic",
		borderEffect: "off", borderPaletteMode: "theme", formMode: "auto",
		config: Config{TestMode: true},
	}
	updated, _ := base.navigateToBordersSubmenu()
	menu := updated.(model)
	for _, id := range []string{"two", "one"} {
		label := registry.get(id).Label()
		menu.menuIndex = -1
		for i, option := range menu.menuOptions {
			if strings.Contains(option, "Style: "+label) {
				menu.menuIndex = i
				break
			}
		}
		if menu.menuIndex < 0 {
			t.Fatalf("menu omitted %q", label)
		}
		menu, _ = menu.handleKeyInput(tea.KeyPressMsg{Code: tea.KeyEnter})
		if menu.selectedBorderStyle != id {
			t.Fatalf("selecting %q chose %q, want %q", label, menu.selectedBorderStyle, id)
		}
		menu.mode = ModeBordersSubmenu
		menu.menuIndex = 0
	}
}

func TestBorderedLayoutOmitsSessionsDividerWithTypewriter(t *testing.T) {
	m := layoutModel(80, 24)
	m.selectedBackground = "ticker"
	m.typewriterTicker = animations.NewTypewriterTicker("niri", "test ticker")
	m.borderRegistry = newBorderRegistry("")
	time.Sleep(60 * time.Millisecond)

	canvas, ok := m.borderedLoginCanvas(80, 24)
	if !ok {
		t.Fatal("border did not fit with the typewriter line")
	}
	rendered := stripAnsi(canvas.Render())
	if strings.Contains(rendered, "SESSIONS") {
		t.Fatalf("bordered layout retained the SESSIONS divider: %q", rendered)
	}
	if !strings.Contains(rendered, "t█") {
		t.Fatalf("bordered layout omitted typewriter content: %q", rendered)
	}
}

func TestBorderedCanvasKeepsConfiguredTime(t *testing.T) {
	m := layoutModel(80, 24)
	m.config = Config{BorderLayout: "hug", ShowTime: true}
	m.borderRegistry = newBorderRegistry("")
	date := time.Now().Format("Mon Jan 02, 2006")

	canvas, ok := m.borderedLoginCanvas(80, 24)
	if !ok {
		t.Fatal("border did not fit with the time row")
	}
	if got := stripAnsi(canvas.Render()); !strings.Contains(got, date) {
		t.Fatalf("bordered login omitted configured time %q: %q", date, got)
	}
}

func TestRegisteredBordersFitAt80x24AndOffIsStatic(t *testing.T) {
	registry := newBorderRegistry("")
	for _, renderer := range registry.styles() {
		m := layoutModel(80, 24)
		m.config = Config{BorderLayout: "hug"}
		m.borderRegistry = registry
		m.selectedBorderStyle = renderer.ID()
		m.borderEffect = "off"
		canvas, ok := m.borderedLoginCanvas(80, 24)
		if !ok {
			t.Errorf("%s did not fit at 80x24", renderer.ID())
			continue
		}
		first := canvas.Render()
		if lipgloss.Width(first) > 80 || lipgloss.Height(first) > 24 {
			t.Errorf("%s rendered %dx%d at 80x24", renderer.ID(), lipgloss.Width(first), lipgloss.Height(first))
		}
		if strings.Contains(first, "SESSIONS") || !strings.Contains(first, "Username:") {
			t.Errorf("%s retained the sessions divider or omitted the login field", renderer.ID())
		}
		m.animationFrame = 100
		m.idleTimer = time.Now().Add(-30 * time.Second)
		second, ok := m.borderedLoginCanvas(80, 24)
		if !ok || second.Render() != first {
			t.Errorf("%s changed with Off at 80x24", renderer.ID())
		}
	}
}

func TestBorderedCanvasRetriesWithoutOversizedSessionArt(t *testing.T) {
	oldDataDir := dataDir
	dataDir = t.TempDir()
	t.Cleanup(func() { dataDir = oldDataDir })
	artDir := filepath.Join(dataDir, "ascii_configs")
	if err := os.MkdirAll(artDir, 0o755); err != nil {
		t.Fatal(err)
	}
	var art strings.Builder
	art.WriteString("name=xfce\nascii_1=\"\"\"\n")
	for i := 0; i < 32; i++ {
		fmt.Fprintf(&art, "TALL-ART-ROW-%02d\n", i)
	}
	art.WriteString("\"\"\"\n")
	if err := os.WriteFile(filepath.Join(artDir, "xfce.conf"), []byte(art.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	m := layoutModel(132, 40)
	m.config = Config{BorderLayout: "hug"}
	m.borderRegistry = newBorderRegistry("")
	formWidth := min(96, max(15, m.width-24))
	withArt := lipgloss.JoinVertical(lipgloss.Center, m.getSessionArt("XFCE"), "", "", m.renderMainForm(formWidth))
	if _, ok := m.borderRegistry.get("classic").Frame(&m, m.width, m.height, lipgloss.Width(withArt), lipgloss.Height(withArt)); ok {
		t.Fatal("test setup: oversized session art unexpectedly fits")
	}

	canvas, ok := m.borderedLoginCanvas(m.width, m.height)
	if !ok {
		t.Fatal("border did not retry after the session art exceeded the height budget")
	}
	plain := stripAnsi(canvas.Render())
	if strings.Contains(plain, "TALL-ART-ROW-") {
		t.Fatal("bordered canvas retained session art that exceeded its height budget")
	}
	if !strings.Contains(plain, "Username:") {
		t.Fatal("retry omitted the username field")
	}
}

func TestRegisteredBordersFitAtApprovedSizes(t *testing.T) {
	registry := newBorderRegistry("")
	for _, size := range [][2]int{{80, 24}, {132, 40}, {200, 55}, {480, 135}} {
		for _, renderer := range registry.styles() {
			t.Run(renderer.ID()+"/"+strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(t *testing.T) {
				m := layoutModel(size[0], size[1])
				m.config = Config{BorderLayout: "hug"}
				m.borderRegistry = registry
				m.selectedBorderStyle = renderer.ID()
				m.borderEffect = "off"
				m.formMode = "opaque"
				canvas, ok := m.borderedLoginCanvas(size[0], size[1])
				if !ok {
					t.Fatal("registered border did not fit")
				}
				rendered := canvas.Render()
				if got := lipgloss.Width(rendered); got > size[0] {
					t.Errorf("canvas width = %d, exceeds %d", got, size[0])
				}
				if got := lipgloss.Height(rendered); got > size[1] {
					t.Errorf("canvas height = %d, exceeds %d", got, size[1])
				}
				if strings.Contains(rendered, "SESSIONS") || !strings.Contains(rendered, "Username:") {
					t.Error("canvas retained the sessions heading or omitted the login field")
				}
			})
		}
	}
}

func TestRegisteredBordersFitAt80x24InFillLayout(t *testing.T) {
	registry := newBorderRegistry("")
	for _, renderer := range registry.styles() {
		t.Run(renderer.ID(), func(t *testing.T) {
			m := layoutModel(80, 24)
			m.config = Config{BorderLayout: "fill"}
			m.borderRegistry = registry
			m.selectedBorderStyle = renderer.ID()
			m.formMode = "opaque"
			if _, ok := m.borderedLoginCanvas(80, 24); !ok {
				t.Fatalf("%s did not fit in fill layout at 80x24", renderer.ID())
			}
		})
	}
}

func TestSubcellBorderShimmerChangesPrimaryStrokes(t *testing.T) {
	useBorderGoldenTheme(t)
	for _, tc := range []struct {
		id         string
		subW, subH int
	}{
		{"braille-hairline", 2, 4},
		{"sextant-bevel", 2, 3},
	} {
		t.Run(tc.id, func(t *testing.T) {
			renderer := &subcellBorderRenderer{id: tc.id, label: tc.id, subW: tc.subW, subH: tc.subH}
			m := layoutModel(80, 24)
			m.config = Config{BorderLayout: "hug"}
			m.borderRegistry = &borderRegistry{byID: map[string]borderRenderer{tc.id: renderer}}
			m.selectedBorderStyle = tc.id
			m.borderPaletteMode = "theme"
			m.formMode = "opaque"
			render := func(effect string, idle time.Duration) string {
				t.Helper()
				m.borderEffect = effect
				m.idleTimer = time.Now().Add(-idle)
				canvas, ok := m.borderedLoginCanvas(80, 24)
				if !ok {
					t.Fatal("subcell border did not fit at 80x24")
				}
				return canvas.Render()
			}

			off := render("off", 11*time.Second)
			if got := render("off", 40*time.Second); got != off {
				t.Fatal("Off changed the subcell frame as idle time advanced")
			}
			if got := render("shimmer", 10*time.Second+100*time.Millisecond); got == off {
				t.Fatal("Shimmer did not change the subcell frame")
			}
		})
	}
}

func TestClassicModernAndASCII1ShimmerTheirPrimaryStrokes(t *testing.T) {
	useBorderGoldenTheme(t)
	registry := newBorderRegistry("")
	for _, id := range []string{"classic", "modern", "ascii1"} {
		t.Run(id, func(t *testing.T) {
			m := layoutModel(80, 24)
			m.config = Config{BorderLayout: "hug"}
			frame, ok := registry.get(id).Frame(&m, 80, 24, 40, 3)
			if !ok {
				t.Fatal("border frame did not fit")
			}
			draw := func(shimmer bool) uv.Screen {
				scr := uv.NewScreenBuffer(frame.Grid.W, frame.Grid.H)
				(nfoGridLayer{
					Grid: frame.Grid, Palette: borderPalette("theme"), PaletteMode: "theme", Profile: colorprofile.TrueColor,
					Left: frame.Left, Top: frame.Top, Right: frame.Right, Bottom: frame.Bottom,
					Shimmer: shimmer, ShimmerPrimary: frame.ShimmerPrimary, Speed: frame.Speed,
				}).Draw(scr, scr.Bounds())
				return scr
			}
			static, animated := draw(false), draw(true)
			changed := false
			for y := 0; y < frame.Grid.H && !changed; y++ {
				for x := 0; x < frame.Grid.W; x++ {
					before, after := static.CellAt(x, y), animated.CellAt(x, y)
					if before != nil && after != nil && before.Content != "" && !sameColor(before.Style.Fg, after.Style.Fg) {
						changed = true
						break
					}
				}
			}
			if !changed {
				t.Fatal("shimmer left the primary border color unchanged")
			}
		})
	}
}

func bComments(t *testing.T, data []byte) []string {
	t.Helper()
	_, sauce := ansiart.ParseSAUCE(data)
	if sauce == nil {
		t.Fatal("asset has no SAUCE metadata")
	}
	return sauce.Comments
}

func TestBorderShimmerIdleClock(t *testing.T) {
	for _, tc := range []struct {
		idle   time.Duration
		frame  int
		active bool
	}{
		{0, 0, false},
		{10 * time.Second, 0, true},
		{10*time.Second + 100*time.Millisecond, 2, true},
		{time.Minute, 1000, false},
		{time.Minute + time.Second, 1000, false},
	} {
		frame, active := borderShimmerFrame(tc.idle)
		if frame != tc.frame || active != tc.active {
			t.Errorf("borderShimmerFrame(%s) = (%d, %t), want (%d, %t)", tc.idle, frame, active, tc.frame, tc.active)
		}
	}
}

func TestBorderShimmerStopsFastUITickAfterIdle(t *testing.T) {
	m := model{
		idleTimer:           time.Now().Add(-time.Minute - time.Second),
		mode:                ModeLogin,
		selectedBackground:  "none",
		selectedBorderStyle: "nfo01",
		borderEffect:        "shimmer",
	}
	if m.uiAnimationsActive() {
		t.Fatal("border shimmer kept the fast UI tick active after its idle stop")
	}
}

func TestBrailleDotOrder(t *testing.T) {
	want := [4][2]uint8{{1, 8}, {2, 16}, {4, 32}, {64, 128}}
	for y := range want {
		for x, bit := range want[y] {
			if got := brailleDotBit(x, y); got != bit {
				t.Errorf("braille dot (%d,%d) = %#x, want %#x", x, y, got, bit)
			}
		}
	}
}

func TestBrailleBorderEncodesDotMask(t *testing.T) {
	grid := makeSubcellBorder(8, 4, 2, 4, 14)
	if got := grid.At(2, 0).Ch; got != '\u282d' {
		t.Fatalf("top-left double hairline = U+%04X, want U+282D", got)
	}
}

func TestBorderScaleThresholds(t *testing.T) {
	for _, tc := range []struct {
		w, h, want int
	}{
		{160, 40, 1},
		{319, 79, 1},
		{320, 80, 2},
		{479, 119, 2},
		{480, 120, 3},
	} {
		if got := borderScaleFor(tc.w, tc.h); got != tc.want {
			t.Errorf("borderScaleFor(%d,%d) = %d, want %d", tc.w, tc.h, got, tc.want)
		}
	}
}

func TestTerminalBorderPaletteKeepsANSIColors(t *testing.T) {
	palette := borderPalette("terminal")
	for i, c := range palette {
		got, ok := c.(ansi.BasicColor)
		if !ok || int(got) != i {
			t.Errorf("terminal palette[%d] = %T(%v), want ANSI color %d", i, c, c, i)
		}
	}
}

func TestBorderScaleRespectsHugFrameWidth(t *testing.T) {
	grid := ansiart.NewGrid(17, 7)
	base := newGridBorder("sample", "Sample", grid, 8, 3, 8, 3)
	renderer := &nineSliceRenderer{id: "sample", label: "Sample", assets: []*ansiart.Border{base}}
	m := model{config: Config{BorderLayout: "hug"}}
	if _, ok := renderer.Frame(&m, 320, 80, 72, 8); !ok {
		t.Fatal("hug frame did not fit")
	}
	if renderer.selected.Scale != 1 {
		t.Fatalf("hug frame selected scale %d, want 1", renderer.selected.Scale)
	}

	m.config.BorderLayout = "fill"
	if _, ok := renderer.Frame(&m, 320, 80, 72, 8); !ok {
		t.Fatal("fill frame did not fit")
	}
	if renderer.selected.Scale != 2 {
		t.Fatalf("fill frame selected scale %d, want 2", renderer.selected.Scale)
	}
}

func TestSextantRuneMapsEveryMask(t *testing.T) {
	seen := make(map[rune]uint8, 64)
	for mask := uint8(0); mask < 64; mask++ {
		r := ansiart.SextantRune(mask)
		if previous, ok := seen[r]; ok {
			t.Fatalf("masks %d and %d both map to %q", previous, mask, r)
		}
		seen[r] = mask
		if mask != 0 && mask != 21 && mask != 42 && mask != 63 && (r < 0x1fb00 || r > 0x1fb3b) {
			t.Errorf("mask %d maps outside sextant block: U+%04X", mask, r)
		}
	}
}

func TestSextantBorderUsesOnlyKitty(t *testing.T) {
	for _, tc := range []struct {
		term string
		want bool
	}{
		{"xterm-kitty", true},
		{"XTERM-KITTY", true},
		{"foot", false},
		{"foot-extra", false},
		{"wezterm", false},
		{"xterm-ghostty", false},
		{"xterm-256color", false},
		{"", false},
	} {
		if got := sextantsSupported(tc.term); got != tc.want {
			t.Errorf("sextantsSupported(%q) = %t, want %t", tc.term, got, tc.want)
		}
	}
}

func TestBorderPreviewSizesValidateBounds(t *testing.T) {
	sizes, err := parseBorderPreviewSizes("80x24, 132x40, 200x55")
	if err != nil {
		t.Fatal(err)
	}
	want := [][2]int{{80, 24}, {132, 40}, {200, 55}}
	if len(sizes) != len(want) {
		t.Fatalf("parsed %v, want %v", sizes, want)
	}
	for i := range want {
		if sizes[i] != want[i] {
			t.Errorf("size %d = %v, want %v", i, sizes[i], want[i])
		}
	}
	for _, bad := range []string{"", "80", "0x24", "513x24", "80x257", "80x24,,132x40"} {
		if _, err := parseBorderPreviewSizes(bad); err == nil {
			t.Errorf("accepted invalid size list %q", bad)
		}
	}
	if _, err := parseBorderPreviewSizes(strings.Repeat("80x24,", 16) + "80x24"); err == nil {
		t.Fatal("accepted more than 16 preview sizes")
	}
}

func TestBorderPreviewFileRejectsOversizedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.ans")
	if err := os.WriteFile(path, bytes.Repeat([]byte{'x'}, (1<<20)+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadBorderPreviewFile(path); err == nil {
		t.Fatal("accepted an asset larger than 1 MiB")
	}
}

func TestBorderPreviewRendersFormAndFrame(t *testing.T) {
	data, err := builtInBorderAssets.ReadFile("assets/borders/nfo01.ans")
	if err != nil {
		t.Fatal(err)
	}
	preview, err := renderBorderPreview(data, [][2]int{{80, 24}, {132, 40}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"nfo01 80x24", "nfo01 132x40", "Username:  nomadx", "Password:  ********"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview omitted %q", want)
		}
	}
}

func TestBorderPreviewKeepsFrameVisibleWithoutColor(t *testing.T) {
	data, err := builtInBorderAssets.ReadFile("assets/borders/nfo01.ans")
	if err != nil {
		t.Fatal(err)
	}
	oldProfile := colorProfile
	colorProfile = colorprofile.Ascii
	t.Cleanup(func() { colorProfile = oldProfile })

	preview, err := renderBorderPreview(data, [][2]int{{80, 24}})
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range regexp.MustCompile("\\x1b\\[([0-9;]*)m").FindAllStringSubmatch(preview, -1) {
		for _, param := range strings.Split(match[1], ";") {
			code, err := strconv.Atoi(param)
			if err != nil {
				continue
			}
			if code == 38 || code == 48 || code >= 30 && code <= 37 || code >= 40 && code <= 47 || code >= 90 && code <= 97 || code >= 100 && code <= 107 {
				t.Fatalf("NO_COLOR preview still contains color SGR %q", match[0])
			}
		}
	}
	if strings.Contains(preview, "SESSIONS") || !strings.Contains(preview, "Username:") || !strings.Contains(preview, "▄") {
		t.Fatalf("NO_COLOR preview lost its text or frame glyphs: %q", preview)
	}
}

func TestFailureTintExpiresAt400ms(t *testing.T) {
	grid := ansiart.NewGrid(3, 3)
	grid.Cells[0] = ansiart.Cell{Ch: 'x', Fg: 3, Bg: -1}
	var palette [16]color.Color
	for i := range palette {
		palette[i] = color.RGBA{R: uint8(i), G: uint8(i * 2), B: uint8(i * 3), A: 255}
	}
	base := palette[3]
	draw := func(age time.Duration) color.Color {
		scr := uv.NewScreenBuffer(3, 3)
		(nfoGridLayer{
			Grid: grid, Palette: palette, Profile: colorprofile.TrueColor,
			Left: 1, Top: 1, Right: 1, Bottom: 1, HasFailure: true, FailureAge: age,
		}).Draw(scr, scr.Bounds())
		return scr.CellAt(0, 0).Style.Fg
	}
	if got := draw(399 * time.Millisecond); sameColor(got, base) {
		t.Fatal("border had no failure tint before 400ms")
	}
	if got := draw(400 * time.Millisecond); !sameColor(got, base) {
		t.Fatalf("border remained tinted at 400ms: got %v, want %v", got, base)
	}
}

func TestPrimaryShimmerUsesVGAPaletteInTrueColor(t *testing.T) {
	grid := ansiart.NewGrid(3, 3)
	for i := range grid.Cells {
		grid.Cells[i] = ansiart.Cell{Ch: 'x', Fg: 14, Bg: -1}
	}
	palette := borderPalette("vga")
	oldFgPrimary := FgPrimary
	FgPrimary = color.RGBA{R: 255, A: 255}
	t.Cleanup(func() { FgPrimary = oldFgPrimary })

	scr := uv.NewScreenBuffer(3, 3)
	(nfoGridLayer{
		Grid: grid, Palette: palette, PaletteMode: "vga", Profile: colorprofile.TrueColor,
		Left: 1, Top: 1, Right: 1, Bottom: 1,
		Shimmer: true, ShimmerPrimary: true, Speed: 1,
	}).Draw(scr, scr.Bounds())

	got := scr.CellAt(0, 0).Style.Fg
	want := mixColor(palette[14], palette[15], .6)
	if !sameColor(got, want) {
		t.Fatalf("VGA primary shimmer = %v, want VGA palette highlight %v", got, want)
	}
}

func TestNFOGridLayerDrawDoesNotAllocatePerCell(t *testing.T) {
	grid := ansiart.NewGrid(12, 5)
	for i := range grid.Cells {
		grid.Cells[i] = ansiart.Cell{Ch: 'x', Fg: 7, Bg: 0}
	}
	layer := nfoGridLayer{
		Grid: grid, Palette: borderPalette("vga"), Left: 1, Top: 1, Right: 1, Bottom: 1,
	}
	var screen uv.Screen = uv.NewScreenBuffer(grid.W, grid.H)
	area := screen.Bounds()
	allocs := testing.AllocsPerRun(100, func() { layer.Draw(screen, area) })
	if allocs > 4 {
		t.Fatalf("Draw allocated %.0f objects for %d border cells, want at most 4 per frame", allocs, 2*grid.W+2*grid.H-4)
	}
}

func TestBorderShimmerOverheadWithinBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("performance budgets are skipped by -short")
	}
	// ponytail: The approved storyboard sets frame-time budgets only at 200x55 and 480x135; 80x24 is a fit gate.
	for _, tc := range []struct {
		size   [2]int
		budget time.Duration
	}{
		{[2]int{200, 55}, 25 * time.Microsecond},
		{[2]int{480, 135}, 80 * time.Microsecond},
	} {
		t.Run(strconv.Itoa(tc.size[0])+"x"+strconv.Itoa(tc.size[1]), func(t *testing.T) {
			static := testing.Benchmark(func(b *testing.B) {
				benchmarkNFOGridDraw(b, tc.size, false)
			})
			shimmer := testing.Benchmark(func(b *testing.B) {
				benchmarkNFOGridDraw(b, tc.size, true)
			})
			overhead := time.Duration(shimmer.NsPerOp() - static.NsPerOp())
			if overhead < 0 {
				overhead = 0
			}
			if overhead > tc.budget {
				t.Fatalf("shimmer overhead %s (total %s, static %s) exceeds design budget %s", overhead,
					time.Duration(shimmer.NsPerOp()), time.Duration(static.NsPerOp()), tc.budget)
			}
		})
	}
}

func TestClearTerminalLayerMatchesBlankString(t *testing.T) {
	const width, height = 4, 2
	legacy := lipgloss.NewCanvas(
		lipgloss.NewLayer("ABCD\nEFGH"),
		lipgloss.NewLayer(strings.Repeat(" ", width)+"\n"+strings.Repeat(" ", width)),
	).Render()
	cleared := lipgloss.NewCanvas(
		lipgloss.NewLayer("ABCD\nEFGH"),
		lipgloss.NewLayer(clearTerminalLayer{width: width, height: height}),
	).Render()
	if cleared != legacy {
		t.Fatalf("clear layer output %q differs from blank string output %q", cleared, legacy)
	}
}

var borderBenchmarkView tea.View
var borderBenchmarkGrid *ansiart.Grid

func BenchmarkBorderRender(b *testing.B) {
	registry := newBorderRegistry("")
	for _, size := range [][2]int{{80, 24}, {200, 55}, {480, 135}} {
		for _, renderer := range registry.styles() {
			size, renderer := size, renderer
			b.Run(renderer.ID()+"/"+strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(b *testing.B) {
				m := layoutModel(size[0], size[1])
				m.config = Config{BorderLayout: "hug"}
				m.borderRegistry = registry
				m.selectedBorderStyle = renderer.ID()
				formWidth := min(96, max(15, size[0]-24))
				form := m.renderMainForm(formWidth)
				frame, ok := renderer.Frame(&m, size[0], size[1], lipgloss.Width(form), lipgloss.Height(form))
				if !ok {
					b.Fatal("registered renderer did not fit")
				}
				var render func()
				switch r := renderer.(type) {
				case *nineSliceRenderer:
					border := r.borderFor(&m, size[0], size[1], formWidth)
					render = func() { borderBenchmarkGrid = border.Render(frame.Grid.W, frame.Grid.H) }
				case *subcellBorderRenderer:
					render = func() {
						borderBenchmarkGrid = makeSubcellBorder(frame.Grid.W, frame.Grid.H, r.subW, r.subH, 14)
					}
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					render()
				}
			})
		}
	}
}

func benchmarkNFOGridDraw(b *testing.B, size [2]int, shimmer bool) {
	registry := newBorderRegistry("")
	renderer := registry.get("nfo01")
	m := layoutModel(size[0], size[1])
	form := m.renderMainForm(min(96, max(15, size[0]-24)))
	frame, ok := renderer.Frame(&m, size[0], size[1], lipgloss.Width(form), lipgloss.Height(form))
	if !ok {
		b.Fatal("NFO border did not fit")
	}
	scr := uv.NewScreenBuffer(size[0], size[1])
	layer := nfoGridLayer{
		Grid: frame.Grid, Palette: borderPalette("theme"), PaletteMode: "theme",
		Profile: colorprofile.TrueColor, Left: frame.Left, Top: frame.Top,
		Right: frame.Right, Bottom: frame.Bottom,
		Shimmer: shimmer, Speed: frame.Speed, Rest: 4,
	}
	area := image.Rect(frame.X, frame.Y, frame.X+frame.Grid.W, frame.Y+frame.Grid.H)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layer.Frame = i
		layer.Draw(scr, area)
	}
}

func BenchmarkNFOGridDraw(b *testing.B) {
	for _, size := range [][2]int{{80, 24}, {200, 55}, {480, 135}} {
		size := size
		b.Run(strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(b *testing.B) {
			benchmarkNFOGridDraw(b, size, false)
		})
	}
}

func BenchmarkShimmer(b *testing.B) {
	for _, size := range [][2]int{{80, 24}, {200, 55}, {480, 135}} {
		size := size
		b.Run(strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(b *testing.B) {
			benchmarkNFOGridDraw(b, size, true)
		})
	}
}

func BenchmarkLoginView(b *testing.B) {
	registry := newBorderRegistry("")
	for _, size := range [][2]int{{80, 24}, {200, 55}, {480, 135}} {
		for _, renderer := range registry.styles() {
			size, renderer := size, renderer
			b.Run(renderer.ID()+"/"+strconv.Itoa(size[0])+"x"+strconv.Itoa(size[1]), func(b *testing.B) {
				m := layoutModel(size[0], size[1])
				m.config = Config{BorderLayout: "hug"}
				m.borderRegistry = registry
				m.selectedBorderStyle = renderer.ID()
				m.borderEffect = "off"
				m.formMode = "opaque"
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					view := m.View()
					if view.Layer == nil {
						b.Fatal("registered renderer did not fit")
					}
					borderBenchmarkView = view
				}
			})
		}
	}
}
