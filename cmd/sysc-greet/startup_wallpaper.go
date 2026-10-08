package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/themes"
	"github.com/Nomadcxx/sysc-greet/internal/wallpaper"
)

type startupWallpaper struct {
	Path    string
	IsVideo bool
	Source  string
}

func resolveStartupWallpaper(prefs *cache.UserPreferences, home string, resourcesDir string) startupWallpaper {
	if path, ok := cachedCustomWallpaperPath(prefs, home); ok {
		return startupWallpaper{
			Path:    path,
			IsVideo: isVideoWallpaper(path),
			Source:  "custom",
		}
	}

	if prefs != nil && prefs.Theme != "" {
		path := themeWallpaperPath(resourcesDir, prefs.Theme)
		if fileExists(path) {
			return startupWallpaper{
				Path:    path,
				IsVideo: false,
				Source:  "theme",
			}
		}
	}

	path := filepath.Join(resourcesDir, "wallpapers", "sysc-greet-default.png")
	return startupWallpaper{
		Path:    path,
		IsVideo: false,
		Source:  "default",
	}
}

func shouldSetCachedThemeWallpaper(prefs *cache.UserPreferences, home string) bool {
	_, customExists := cachedCustomWallpaperPath(prefs, home)
	return !customExists
}

func cachedCustomWallpaperPath(prefs *cache.UserPreferences, home string) (string, bool) {
	if prefs == nil || prefs.Wallpaper == "" {
		return "", false
	}
	for _, path := range customWallpaperPaths(home, prefs.Wallpaper) {
		if fileExists(path) {
			return path, true
		}
	}
	return "", false
}

func customWallpaperPaths(home string, filename string) []string {
	return []string{
		filepath.Join("/var/lib/greeter/Pictures/wallpapers", filename),
		filepath.Join(home, "Pictures", "wallpapers", filename),
	}
}

func themeWallpaperPath(resourcesDir string, themeName string) string {
	themeFile := strings.ToLower(strings.ReplaceAll(themeName, " ", "-"))
	return filepath.Join(resourcesDir, "wallpapers", fmt.Sprintf("sysc-greet-%s.png", themeFile))
}

func isVideoWallpaper(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".mkv", ".webm", ".avi", ".mov":
		return true
	default:
		return false
	}
}

func gSlapperStartupArgs(path string, isVideo bool) []string {
	if isVideo {
		return []string{"-f", "-s", "-I", wallpaper.GSlapperSocket, "-o", "loop", "*", path}
	}
	return []string{"-f", "-I", wallpaper.GSlapperSocket, "*", path}
}

func runStartupWallpaperDaemon(config Config) error {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "/var/lib/greeter"
	}

	prefs, err := cache.LoadPreferences()
	if err != nil {
		prefs = nil
	}

	available := append(themes.GetAvailableThemes(), themes.ScanCustomThemes([]string{filepath.Join(home, ".config", "sysc-greet", "themes"), filepath.Join(dataDir, "themes")})...)
	prefs = startupThemePreferences(prefs, config, available)
	selected := resolveStartupWallpaper(prefs, home, dataDir)
	if !fileExists(selected.Path) {
		return fmt.Errorf("startup wallpaper not found: %s", selected.Path)
	}

	cmd := exec.Command("gslapper", gSlapperStartupArgs(selected.Path, selected.IsVideo)...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Copy cached preferences so wallpaper startup and login resolve the same theme.
func startupThemePreferences(prefs *cache.UserPreferences, config Config, available []string) *cache.UserPreferences {
	copy := cache.UserPreferences{}
	if prefs != nil {
		copy = *prefs
	}
	follow := config.FollowShell == nil || *config.FollowShell
	if config.FollowShell == nil && copy.FollowShell != nil {
		follow = *copy.FollowShell
	}
	m := model{config: config, followShell: follow, availableThemes: available}
	copy.Theme = m.startupTheme(copy.Theme)
	return &copy
}
