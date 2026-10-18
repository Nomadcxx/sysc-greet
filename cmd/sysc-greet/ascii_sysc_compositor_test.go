package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/animations"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestLoadSyscCompositorASCIIConfig(t *testing.T) {
	configPath := filepath.Join("..", "..", "ascii_configs", "sysc-compositor.conf")
	config, err := loadASCIIConfig(configPath)
	if err != nil {
		t.Fatalf("loadASCIIConfig(%q): %v", configPath, err)
	}

	if config.Name != "sysc-compositor" {
		t.Fatalf("Name = %q, want sysc-compositor", config.Name)
	}
	if !config.ANSIStyle {
		t.Fatal("expected sysc-compositor to enable ANSI-style rendering")
	}
	if len(config.ASCIIVariants) == 0 {
		t.Fatal("expected at least one ASCII variant")
	}
	if !strings.HasPrefix(config.ASCIIVariants[0], `\033[0;37m`) {
		t.Fatal("first ASCII variant should contain the TAAG Bleach ANSI export")
	}
	if len(config.ASCIIVariants) < 2 || !strings.HasPrefix(config.ASCIIVariants[1], "█▀▀▀▀▀▀") {
		t.Fatal("second ASCII variant should use the sanitized TAAG shell art")
	}

	maxHeight := 0
	for i, variant := range config.ASCIIVariants {
		if strings.TrimSpace(variant) == "" {
			t.Errorf("ASCII variant %d is empty", i+1)
		}
		if strings.Contains(variant, "\x1b") || strings.Contains(variant, "echo -e") {
			t.Errorf("ASCII variant %d contains shell or ANSI escape text", i+1)
		}
		if height := len(strings.Split(variant, "\n")); height > maxHeight {
			maxHeight = height
		}
	}
	if config.MaxASCIIHeight != maxHeight {
		t.Errorf("MaxASCIIHeight = %d, want %d", config.MaxASCIIHeight, maxHeight)
	}
}

func TestGetSessionASCIIUsesEmbeddedANSIAndThemeRoles(t *testing.T) {
	writeASCIIConfigForTest(t, `name=sysc-compositor
ansi_style=true
ascii_1="""
\033[94;44mA\033[31mB \033[49mC\033[39mX\033[0m
\033[38;2;31;32;33mD\033[2J
"""
`)

	oldPrimary, oldSecondary := Primary, Secondary
	oldDanger, oldBgBase, oldFgPrimary := Danger, BgBase, FgPrimary
	Primary = lipgloss.Color("#112233")
	Secondary = lipgloss.Color("#445566")
	Danger = lipgloss.Color("#ff0000")
	BgBase = lipgloss.Color("#000000")
	FgPrimary = lipgloss.Color("#eeeeee")
	t.Cleanup(func() {
		Primary, Secondary = oldPrimary, oldSecondary
		Danger, BgBase, FgPrimary = oldDanger, oldBgBase, oldFgPrimary
	})

	session := sessions.Session{Name: "sysc-compositor"}
	got := (model{selectedSession: &session}).getSessionASCII()
	if plain := ansi.Strip(got); plain != "AB CX\nD" {
		t.Fatalf("plain ANSI art = %q, want %q", plain, "AB CX\nD")
	}
	aligned := lipgloss.NewStyle().Width(20).Align(lipgloss.Center).Render(got)
	if !strings.Contains(aligned, "38;2;17;34;51") {
		t.Fatal("ANSI foreground color was lost through Lipgloss alignment")
	}
	alignedLines := strings.Split(ansi.Strip(aligned), "\n")
	wantLines := []string{"AB CX", "D"}
	if len(alignedLines) != len(wantLines) {
		t.Fatalf("aligned line count = %d, want %d", len(alignedLines), len(wantLines))
	}
	for i, line := range alignedLines {
		if width := lipgloss.Width(line); width != 20 {
			t.Errorf("aligned line %d width = %d, want 20", i+1, width)
		}
		want := wantLines[i]
		if got := strings.TrimSpace(line); got != want {
			t.Errorf("aligned line %d text = %q, want %q", i+1, got, want)
		}
	}
	if strings.Contains(got, `\033`) || strings.Contains(got, "\x1b[2J") {
		t.Fatalf("config escape text or unsupported CSI reached output: %q", got)
	}
	if !strings.Contains(got, "\x1b[38;2;17;34;51;48;2;68;85;102mA") {
		t.Fatalf("bright blue foreground and blue background did not use theme roles: %q", got)
	}
	if !strings.Contains(got, "\x1b[38;2;255;0;0;48;2;68;85;102mB ") {
		t.Fatalf("standard red foreground did not use Danger: %q", got)
	}
	if !strings.Contains(got, "\x1b[38;2;255;0;0;48;2;0;0;0mC") {
		t.Fatalf("background reset did not restore BgBase: %q", got)
	}
	if !strings.Contains(got, "\x1b[38;2;238;238;238;48;2;0;0;0mX") {
		t.Fatalf("foreground reset did not restore FgPrimary: %q", got)
	}
	if !strings.Contains(got, "\x1b[38;2;238;238;238;48;2;0;0;0mD") {
		t.Fatalf("extended-color payload was misread as standard SGR: %q", got)
	}
}

func TestASCIIANSIColorSlotsResolveToThemeRoles(t *testing.T) {
	oldBgBase, oldDanger, oldAccent := BgBase, Danger, Accent
	oldWarning, oldSecondary, oldPrimary := Warning, Secondary, Primary
	oldFgSecondary, oldFgMuted, oldFgPrimary := FgSecondary, FgMuted, FgPrimary
	BgBase = lipgloss.Color("#010101")
	Danger = lipgloss.Color("#020202")
	Accent = lipgloss.Color("#030303")
	Warning = lipgloss.Color("#040404")
	Secondary = lipgloss.Color("#050505")
	Primary = lipgloss.Color("#060606")
	FgSecondary = lipgloss.Color("#070707")
	FgMuted = lipgloss.Color("#080808")
	FgPrimary = lipgloss.Color("#090909")
	t.Cleanup(func() {
		BgBase, Danger, Accent = oldBgBase, oldDanger, oldAccent
		Warning, Secondary, Primary = oldWarning, oldSecondary, oldPrimary
		FgSecondary, FgMuted, FgPrimary = oldFgSecondary, oldFgMuted, oldFgPrimary
	})

	want := []color.Color{
		BgBase, Danger, Accent, Warning, Secondary, Accent, Secondary, FgSecondary,
		FgMuted, Danger, Accent, Warning, Primary, Accent, Primary, FgPrimary,
	}
	for slot, expected := range want {
		if got := asciiANSIColor(slot); got != expected {
			t.Errorf("ANSI color slot %d resolved to %v, want %v", slot, got, expected)
		}
	}
}

func TestGetSessionASCIIUnmarkedConfigRetainsColorOverride(t *testing.T) {
	writeASCIIConfigForTest(t, `name=sysc-compositor
color=#abcdef
ascii_1="""
LOGO
"""
`)

	session := sessions.Session{Name: "sysc-compositor"}
	got := (model{selectedSession: &session}).getSessionASCII()
	if plain := ansi.Strip(got); plain != "LOGO" {
		t.Fatalf("plain art = %q, want LOGO", plain)
	}
	if !strings.Contains(got, "\x1b[38;2;171;205;239mLOGO") {
		t.Fatalf("unmarked profile did not keep its configured color: %q", got)
	}
}

func TestGetSessionASCIIRendersPlainVariantWithLegacyThemeStyle(t *testing.T) {
	originalDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	t.Cleanup(func() { dataDir = originalDataDir })

	oldPrimary, oldBgBase := Primary, BgBase
	Primary = lipgloss.Color("#112233")
	BgBase = lipgloss.Color("#000000")
	t.Cleanup(func() { Primary, BgBase = oldPrimary, oldBgBase })

	session := sessions.Session{Name: "sysc-compositor"}
	got := (model{selectedSession: &session, asciiArtIndex: 1}).getSessionASCII()
	if plain := ansi.Strip(got); !strings.HasPrefix(plain, "█▀▀▀▀▀▀") {
		t.Fatalf("plain variant was not preserved: %q", plain)
	}
	if !strings.Contains(got, "\x1b[38;2;17;34;51") {
		t.Fatalf("plain variant did not keep the existing theme color path: %q", got)
	}
}

func writeASCIIConfigForTest(t *testing.T, config string) {
	t.Helper()
	oldDataDir := dataDir
	dataDir = t.TempDir()
	t.Cleanup(func() { dataDir = oldDataDir })

	configDir := filepath.Join(dataDir, "ascii_configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "sysc-compositor.conf"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGetSessionASCIISelectsSyscCompositorProfile(t *testing.T) {
	originalDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	t.Cleanup(func() { dataDir = originalDataDir })

	session := sessions.Session{Name: "sysc-compositor"}
	got := (model{selectedSession: &session}).getSessionASCII()
	if !strings.Contains(ansi.Strip(got), "▄▀▀█▀▀") {
		t.Fatalf("getSessionASCII() did not select sysc-compositor art: %q", got)
	}
}

func TestNormalizeTestSessionOptionRequiresTestModeAndSafeName(t *testing.T) {
	if _, err := normalizeTestSessionOption(false, "sysc-compositor"); err == nil {
		t.Fatal("expected --test-session to require --test")
	}

	got, err := normalizeTestSessionOption(true, " sysc-compositor ")
	if err != nil {
		t.Fatalf("normalizeTestSessionOption() error: %v", err)
	}
	if got != "sysc-compositor" {
		t.Fatalf("normalized session = %q, want sysc-compositor", got)
	}

	if _, err := normalizeTestSessionOption(true, "../etc"); err == nil {
		t.Fatal("expected path-like session name to be rejected")
	}
}

func TestSelectTestSessionAddsUnavailableSession(t *testing.T) {
	available := []sessions.Session{{Name: "Sway", Exec: "sway", Type: "Wayland"}}
	gotSessions, selected, index := selectTestSession(available, "sysc-compositor")

	if len(gotSessions) != 2 {
		t.Fatalf("session count = %d, want 2", len(gotSessions))
	}
	if index != 1 {
		t.Fatalf("selected index = %d, want 1", index)
	}
	if selected == nil || selected.Name != "sysc-compositor" || selected.Type != "Wayland" {
		t.Fatalf("selected session = %#v, want sysc-compositor Wayland session", selected)
	}
}

func TestSelectTestSessionReusesAvailableSession(t *testing.T) {
	available := []sessions.Session{{Name: "Sysc-Compositor", Exec: "sysc-compositor", Type: "Wayland"}}
	gotSessions, selected, index := selectTestSession(available, "sysc-compositor")

	if len(gotSessions) != 1 || index != 0 {
		t.Fatalf("selection = %d of %d sessions, want 0 of 1", index, len(gotSessions))
	}
	if selected != &gotSessions[0] || selected.Name != "Sysc-Compositor" {
		t.Fatalf("selected session = %#v, want existing session %#v", selected, gotSessions[0])
	}
}

func TestInitialModelSelectsTestSessionASCIIProfile(t *testing.T) {
	originalDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	t.Cleanup(func() { dataDir = originalDataDir })

	m := initialModel(Config{TestMode: true, TestSession: "sysc-compositor"}, false)
	if m.selectedSession == nil || m.selectedSession.Name != "sysc-compositor" {
		t.Fatalf("selected session = %#v, want sysc-compositor", m.selectedSession)
	}
	if got := ansi.Strip(m.getSessionASCII()); !strings.Contains(got, "▄▀▀█▀▀") {
		t.Fatalf("test session did not load sysc-compositor ASCII: %q", got)
	}
}

func TestSyscCompositorASCIIHeaderVisibleInCompactLayout(t *testing.T) {
	originalDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	t.Cleanup(func() { dataDir = originalDataDir })

	for _, size := range [][2]int{{80, 24}, {100, 35}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			m := initialModel(Config{TestMode: true, TestSession: "sysc-compositor"}, false)
			view := m.renderMainView(size[0], size[1])
			if !strings.Contains(ansi.Strip(view), "▄▀▀█▀▀") {
				t.Fatal("login layout dropped the sysc-compositor ASCII header")
			}
			if !strings.Contains(view, "\x1b[38;2;") || !strings.Contains(view, "\x1b[48;2;") {
				t.Fatal("login layout dropped ANSI foreground or background colors")
			}
			if width, height := lipgloss.Width(view), lipgloss.Height(view); width > size[0] || height > size[1] {
				t.Fatalf("layout is %dx%d, exceeds %dx%d", width, height, size[0], size[1])
			}
		})
	}
}

func TestSyscCompositorPrintEffectVisibleInCompactLayout(t *testing.T) {
	originalDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	t.Cleanup(func() { dataDir = originalDataDir })

	m := initialModel(Config{TestMode: true, TestSession: "sysc-compositor"}, false)
	m.selectedBackground = "print"
	m.printEffect = animations.NewPrintEffect("PRINT-FX", 0)
	now := time.Now()
	for i := 1; i <= 5; i++ {
		m.printEffect.Tick(now.Add(time.Duration(i) * time.Second))
	}

	view := m.renderMainView(80, 24)
	if !strings.Contains(ansi.Strip(view), "PRINT") {
		t.Fatal("compact layout dropped the session's active print effect")
	}
}

func TestSessionCommandPassesSyscCompositorExecAsArgv(t *testing.T) {
	execLine := "/usr/local/bin/sysc-compositor -c /home/nomadx/.config/sysc-compositor/config.kdl"
	want := []string{
		"/usr/local/bin/sysc-compositor",
		"-c",
		"/home/nomadx/.config/sysc-compositor/config.kdl",
	}

	if got := sessionCommand(execLine); !reflect.DeepEqual(got, want) {
		t.Fatalf("sessionCommand(%q) = %#v, want %#v", execLine, got, want)
	}
}
