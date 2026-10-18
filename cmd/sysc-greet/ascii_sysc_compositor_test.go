package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Nomadcxx/sysc-greet/internal/sessions"
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
	if len(config.ASCIIVariants) == 0 {
		t.Fatal("expected at least one ASCII variant")
	}
	if !strings.HasPrefix(config.ASCIIVariants[0], "  ▄▀▀█▀▀") {
		t.Fatal("first ASCII variant should be the TAAG Bleach output")
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

func TestGetSessionASCIISelectsSyscCompositorProfile(t *testing.T) {
	originalDataDir := dataDir
	dataDir = filepath.Join("..", "..")
	t.Cleanup(func() { dataDir = originalDataDir })

	session := sessions.Session{Name: "sysc-compositor"}
	got := (model{selectedSession: &session}).getSessionASCII()
	if !strings.Contains(got, "▄▀▀█▀▀") {
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
	if got := m.getSessionASCII(); !strings.Contains(got, "▄▀▀█▀▀") {
		t.Fatalf("test session did not load sysc-compositor ASCII: %q", got)
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
