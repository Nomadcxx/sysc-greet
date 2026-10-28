package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/sessions"
)

func TestViewLoginWarmAllocations(t *testing.T) {
	m := renderBenchModel(t)
	m.View()

	if got := testing.AllocsPerRun(100, func() { m.View() }); got > 150 {
		t.Fatalf("warm View allocations = %.0f, want <= 150", got)
	}
}

func TestGetSessionASCIIWarmAllocations(t *testing.T) {
	m := renderBenchModel(t)
	m.getSessionASCII()

	if got := testing.AllocsPerRun(100, func() { m.getSessionASCII() }); got > 50 {
		t.Fatalf("warm getSessionASCII allocations = %.0f, want <= 50", got)
	}
}

func BenchmarkViewLoginNiri(b *testing.B) {
	m := renderBenchModel(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.View()
	}
}

func BenchmarkGetSessionASCIINiri(b *testing.B) {
	m := renderBenchModel(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.getSessionASCII()
	}
}

func BenchmarkViewLoginNiriBgOnlyChange(b *testing.B) {
	m := renderBenchModel(b)
	m.selectedBackground = "matrix"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.bgAnimationFrame++
		m.View()
	}
}

func TestViewDoesNotReloadASCIIConfigOnWarmFrames(t *testing.T) {
	m := renderBenchModel(t)
	cache := newASCIIConfigCache()
	load := cache.load
	loads := 0
	cache.load = func(path string) (ASCIIConfig, error) {
		loads++
		return load(path)
	}
	m.asciiConfigCache = cache
	m.View()
	if loads != 1 {
		t.Fatalf("initial View loaded ASCII config %d times, want 1", loads)
	}
	for i := 0; i < 100; i++ {
		m.View()
	}
	if loads != 1 {
		t.Fatalf("warm Views loaded ASCII config %d times, want 1", loads)
	}
}

func TestASCIIConfigCacheRefreshesAfterOneMinute(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	cache := newASCIIConfigCache()
	cache.now = func() time.Time { return now }
	loads, renders := 0, 0
	cache.load = func(string) (ASCIIConfig, error) {
		loads++
		return ASCIIConfig{ASCII: fmt.Sprint(loads)}, nil
	}
	key := asciiRenderKey{path: "session.conf"}
	cache.loadConfig("session.conf")
	render := func() asciiRenderResult {
		renders++
		return asciiRenderResult{text: fmt.Sprint("render", renders)}
	}
	if got := cache.render(key, render).text; got != "render1" {
		t.Fatalf("first render = %q, want render1", got)
	}
	cache.loadConfig("session.conf")
	if got := cache.render(key, render).text; got != "render1" {
		t.Fatalf("warm render = %q, want render1", got)
	}
	if loads != 1 || renders != 1 {
		t.Fatalf("warm cache loads=%d renders=%d, want 1 each", loads, renders)
	}
	now = now.Add(asciiConfigCacheTTL)
	cache.loadConfig("session.conf")
	if got := cache.render(key, render).text; got != "render2" {
		t.Fatalf("refreshed render = %q, want render2", got)
	}
	if loads != 2 || renders != 2 {
		t.Fatalf("refreshed cache loads=%d renders=%d, want 2 each", loads, renders)
	}
}

func TestLoginViewCacheRefreshesEditedASCIIAfterOneMinute(t *testing.T) {
	m := renderBenchModel(t)
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	cache := newASCIIConfigCache()
	cache.now = func() time.Time { return now }
	loads := 0
	cache.load = func(string) (ASCIIConfig, error) {
		loads++
		art := "first"
		if loads > 1 {
			art = "second"
		}
		return ASCIIConfig{ASCIIVariants: []string{art}}, nil
	}
	m.asciiConfigCache = cache
	first := m.renderMainView(60, 20)
	if !strings.Contains(stripAnsi(first), "first") {
		t.Fatal("initial view does not contain cached ASCII art")
	}
	now = now.Add(asciiConfigCacheTTL)
	second := m.renderMainView(60, 20)
	if !strings.Contains(stripAnsi(second), "second") {
		t.Fatal("view did not pick up the edited ASCII art after one minute")
	}
	if loads != 2 {
		t.Fatalf("ASCII config loads = %d, want 2", loads)
	}
}

func TestLoginViewCacheMatchesUncachedAfterStateChanges(t *testing.T) {
	m := renderBenchModel(t)
	t.Cleanup(func() { applyTheme("dracula", true) })
	width, height := 60, 20
	assertMatchesUncached := func(name string) {
		t.Helper()
		got := m.renderMainView(width, height)
		uncached := m
		uncached.loginViewCache = nil
		if want := uncached.renderMainView(width, height); got != want {
			t.Errorf("%s changed cached output", name)
		}
	}
	m.renderMainView(width, height)
	applyTheme("nord", true)
	m.currentTheme = "nord"
	assertMatchesUncached("theme switch")
	m.asciiArtIndex = 1
	assertMatchesUncached("ASCII variant change")
	session := sessions.Session{Name: "Hyprland", Exec: "Hyprland", Type: "Wayland"}
	m.selectedSession = &session
	assertMatchesUncached("session change")
	m.focusState = FocusSession
	assertMatchesUncached("focus change")
	m.errorMessage = "authentication failed"
	assertMatchesUncached("error message change")
	width, height = 68, 22
	assertMatchesUncached("terminal resize")
	m.sessionDropdownOpen = true
	m.sessionIndex = 1
	assertMatchesUncached("session dropdown change")
}

func TestBorderedLoginFormCacheMatchesUncachedAfterStateChanges(t *testing.T) {
	m := renderBenchModel(t)
	applyTheme("dracula", true)
	t.Cleanup(func() { applyTheme("dracula", true) })
	width, height := 213, 54
	assertMatchesUncached := func(name string) {
		t.Helper()
		if _, ok := m.borderedLoginCanvas(width, height); !ok {
			t.Fatal("bordered login canvas unavailable")
		}
		got := m.loginViewCache.form
		uncached := m
		uncached.loginViewCache = &loginViewCache{}
		if _, ok := uncached.borderedLoginCanvas(width, height); !ok {
			t.Fatal("uncached bordered login canvas unavailable")
		}
		if want := uncached.loginViewCache.form; got != want {
			t.Errorf("%s changed cached form", name)
		}
	}
	assertMatchesUncached("initial")
	applyTheme("nord", true)
	m.currentTheme = "nord"
	assertMatchesUncached("theme switch")
	m.asciiArtIndex = 1
	assertMatchesUncached("ASCII variant change")
	session := sessions.Session{Name: "Hyprland", Exec: "Hyprland", Type: "Wayland"}
	m.selectedSession = &session
	assertMatchesUncached("session change")
	m.focusState = FocusSession
	assertMatchesUncached("focus change")
	m.errorMessage = "authentication failed"
	assertMatchesUncached("error message change")
	width, height = 215, 55
	assertMatchesUncached("terminal resize")
	m.sessionDropdownOpen = true
	m.sessionIndex = 1
	assertMatchesUncached("session dropdown change")
}

func renderBenchModel(tb testing.TB) model {
	tb.Helper()
	dataDir = filepath.Join("..", "..")
	applyTheme("dracula", true)
	m := initialModel(Config{TestMode: true}, false)
	s := sessions.Session{Name: "niri", Exec: "niri", Type: "Wayland"}
	h := sessions.Session{Name: "Hyprland", Exec: "Hyprland", Type: "Wayland"}
	m.selectedSession, m.sessions = &s, []sessions.Session{s, h}
	m.sessionIndex, m.width, m.height, m.mode = 0, 213, 54, ModeLogin
	m.selectedBackground = "none"
	m.selectedBorderStyle = "classic"
	return m
}
