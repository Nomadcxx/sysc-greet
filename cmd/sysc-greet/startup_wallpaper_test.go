package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/wallpaper"
)

func TestLoadStartupWallpaperPreferencesReportsErrors(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home-file")
	if err := os.WriteFile(home, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previousStderr := os.Stderr
	os.Stderr = writeEnd
	prefs := loadStartupWallpaperPreferences()
	os.Stderr = previousStderr
	if err := writeEnd.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(readEnd)
	if err != nil {
		t.Fatal(err)
	}
	if err := readEnd.Close(); err != nil {
		t.Fatal(err)
	}
	if prefs != nil {
		t.Fatalf("preferences = %#v, want nil after load failure", prefs)
	}
	if !strings.Contains(string(output), "failed to load sysc-greet preferences for startup wallpaper") {
		t.Fatalf("startup preference load error was not reported: %q", output)
	}
}

func TestResolveStartupWallpaperPrefersCachedCustomWallpaper(t *testing.T) {
	home := t.TempDir()
	dataDir := t.TempDir()
	customPath := filepath.Join(home, "Pictures", "wallpapers", "custom.webm")
	themePath := filepath.Join(dataDir, "wallpapers", "sysc-greet-nord.png")
	touchTestFile(t, customPath)
	touchTestFile(t, themePath)

	got := resolveStartupWallpaper(&cache.UserPreferences{
		Theme:     "nord",
		Wallpaper: "custom.webm",
	}, home, dataDir)

	if got.Path != customPath {
		t.Fatalf("expected custom wallpaper %q, got %q", customPath, got.Path)
	}
	if got.Source != "custom" {
		t.Fatalf("expected custom source, got %q", got.Source)
	}
	if !got.IsVideo {
		t.Fatalf("expected custom .webm wallpaper to be treated as video")
	}
}

func TestResolveStartupWallpaperFallsBackToCachedTheme(t *testing.T) {
	home := t.TempDir()
	dataDir := t.TempDir()
	themePath := filepath.Join(dataDir, "wallpapers", "sysc-greet-tokyo-night.png")
	touchTestFile(t, themePath)

	got := resolveStartupWallpaper(&cache.UserPreferences{
		Theme:     "Tokyo Night",
		Wallpaper: "missing.mp4",
	}, home, dataDir)

	if got.Path != themePath {
		t.Fatalf("expected theme wallpaper %q, got %q", themePath, got.Path)
	}
	if got.Source != "theme" {
		t.Fatalf("expected theme source, got %q", got.Source)
	}
	if got.IsVideo {
		t.Fatalf("expected theme wallpaper to be treated as static image")
	}
}

func TestResolveStartupWallpaperFallsBackToDefault(t *testing.T) {
	home := t.TempDir()
	dataDir := t.TempDir()
	defaultPath := filepath.Join(dataDir, "wallpapers", "sysc-greet-default.png")
	touchTestFile(t, defaultPath)

	got := resolveStartupWallpaper(nil, home, dataDir)

	if got.Path != defaultPath {
		t.Fatalf("expected default wallpaper %q, got %q", defaultPath, got.Path)
	}
	if got.Source != "default" {
		t.Fatalf("expected default source, got %q", got.Source)
	}
}

func TestGSlapperStartupArgs(t *testing.T) {
	staticPath := "/tmp/static.png"
	staticArgs := gSlapperStartupArgs(staticPath, false)
	expectedStatic := []string{"-f", "-I", wallpaper.GSlapperSocket, "*", staticPath}
	if !reflect.DeepEqual(staticArgs, expectedStatic) {
		t.Fatalf("static args mismatch:\nexpected %#v\nactual   %#v", expectedStatic, staticArgs)
	}

	videoPath := "/tmp/video.mp4"
	videoArgs := gSlapperStartupArgs(videoPath, true)
	expectedVideo := []string{"-f", "-s", "-I", wallpaper.GSlapperSocket, "-o", "loop", "*", videoPath}
	if !reflect.DeepEqual(videoArgs, expectedVideo) {
		t.Fatalf("video args mismatch:\nexpected %#v\nactual   %#v", expectedVideo, videoArgs)
	}
}

func TestCachedThemeWallpaperSkippedWhenCustomWallpaperExists(t *testing.T) {
	home := t.TempDir()
	touchTestFile(t, filepath.Join(home, "Pictures", "wallpapers", "custom.mp4"))

	withCustom := &cache.UserPreferences{Theme: "nord", Wallpaper: "custom.mp4"}
	if shouldSetCachedThemeWallpaper(withCustom, home) {
		t.Fatalf("expected cached theme wallpaper to be skipped when custom wallpaper is cached")
	}

	themeOnly := &cache.UserPreferences{Theme: "nord"}
	if !shouldSetCachedThemeWallpaper(themeOnly, home) {
		t.Fatalf("expected cached theme wallpaper to be set when no custom wallpaper is cached")
	}

	staleCustom := &cache.UserPreferences{Theme: "nord", Wallpaper: "missing.mp4"}
	if !shouldSetCachedThemeWallpaper(staleCustom, home) {
		t.Fatalf("expected cached theme wallpaper to be set when custom wallpaper cache is stale")
	}
}

func touchTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
