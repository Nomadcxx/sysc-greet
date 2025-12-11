package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
	tea "github.com/charmbracelet/bubbletea/v2"
)

// Created wallpaper.go for wallpaper/gslapper handling

// navigateToWallpaperSubmenu scans wallpapers directory and builds menu
func (m model) navigateToWallpaperSubmenu() (tea.Model, tea.Cmd) {
	// CHANGED 2025-10-06 - Use /var/lib/greeter/Pictures/wallpapers for greeter user - Problem: $HOME is greeter's home in production
	// Try greeter's wallpaper directory first (for production), then fallback to user's home (for testing)
	wallpaperDirs := []string{
		"/var/lib/greeter/Pictures/wallpapers",
		filepath.Join(os.Getenv("HOME"), "Pictures", "wallpapers"),
	}

	m.menuOptions = []string{"← Back", "Stop Wallpaper"}

	// Try each directory until we find one that exists
	for _, wallpaperDir := range wallpaperDirs {
		files, err := os.ReadDir(wallpaperDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".mp4") {
					m.menuOptions = append(m.menuOptions, file.Name())
				}
			}
			// If we found files, save the directory and break
			if len(m.menuOptions) > 1 {
				break
			}
		}
	}

	m.mode = ModeWallpaperSubmenu
	m.menuIndex = 0
	return m, nil
}

// stopGslapper kills any running gslapper process
func stopGslapper() {
	go func() {
		exec.Command("pkill", "-f", "gslapper").Run()
	}()
}

// launchGslapperWallpaper kills existing gslapper and launches new one with selected wallpaper
func launchGslapperWallpaper(wallpaperFilename string) {
	// CHANGED 2025-10-06 - Try multiple wallpaper directories - Problem: greeter user has different $HOME
	wallpaperPaths := []string{
		filepath.Join("/var/lib/greeter/Pictures/wallpapers", wallpaperFilename),
		filepath.Join(os.Getenv("HOME"), "Pictures", "wallpapers", wallpaperFilename),
	}

	// Find the first existing wallpaper path
	var wallpaperPath string
	for _, path := range wallpaperPaths {
		if _, err := os.Stat(path); err == nil {
			wallpaperPath = path
			break
		}
	}

	// If no path found, use first one anyway (will fail gracefully)
	if wallpaperPath == "" {
		wallpaperPath = wallpaperPaths[0]
	}

	// CHANGED 2025-10-04 - Fixed gslapper flags
	go func() {
		// Kill any existing gslapper process
		killCmd := exec.Command("pkill", "-f", "gslapper")
		killErr := killCmd.Run()

		// Start new gslapper with selected wallpaper
		// Correct syntax: gslapper -s -o "loop panscan=1.0" '*' /path/to/video.mp4
		// -s: daemon mode, -o: gstreamer options, '*': all monitors
		cmd := exec.Command("gslapper", "-s", "-o", "loop panscan=1.0", "*", wallpaperPath)
		startErr := cmd.Start()

		// Write status to a debug file so user can check if it's being called
		debugFile := "/tmp/sysc-greet-wallpaper.log" // FIXED 2025-10-15 - Corrected log filename from bubble-greet to sysc-greet
		logMsg := ""
		if killErr != nil {
			logMsg += "pkill gslapper: " + killErr.Error() + "\n"
		} else {
			logMsg += "pkill gslapper: success\n"
		}
		if startErr != nil {
			logMsg += "gslapper start: " + startErr.Error() + "\n"
		} else {
			logMsg += "gslapper started: gslapper -s -o \"loop panscan=1.0\" '*' " + wallpaperPath + "\n"
		}
		os.WriteFile(debugFile, []byte(logMsg), 0644)
	}()
}

// handleWallpaperSelection processes wallpaper menu selection
func (m model) handleWallpaperSelection(selectedOption string) (tea.Model, tea.Cmd) {
	if selectedOption == "Stop Wallpaper" {
		// Kill gslapper and clear wallpaper preference
		stopGslapper()
		m.selectedWallpaper = ""
		m.gslapperLaunched = false

		// Save cleared preference to cache
		if !m.config.TestMode {
			sessionName := ""
			if m.selectedSession != nil {
				sessionName = m.selectedSession.Name
			}
			cache.SavePreferences(cache.UserPreferences{
				Theme:       m.currentTheme,
				Background:  m.selectedBackground,
				Wallpaper:   m.selectedWallpaper, // Now empty
				BorderStyle: m.selectedBorderStyle,
				Session:     sessionName,
			})
		}
	} else if selectedOption != "← Back" {
		// Launch gslapper with selected wallpaper
		launchGslapperWallpaper(selectedOption)

		// Store wallpaper separately from background effect
		m.selectedWallpaper = selectedOption
		m.gslapperLaunched = true

		// Save preference
		if !m.config.TestMode {
			sessionName := ""
			if m.selectedSession != nil {
				sessionName = m.selectedSession.Name
			}
			cache.SavePreferences(cache.UserPreferences{
				Theme:       m.currentTheme,
				Background:  m.selectedBackground,
				Wallpaper:   m.selectedWallpaper,
				BorderStyle: m.selectedBorderStyle,
				Session:     sessionName,
			})
		}
	}

	m.mode = ModeLogin
	return m, nil
}

// CHANGED 2025-10-04 - Add function to launch asset videos for Fireplace/Particle effects
// launchAssetVideo launches a video from Assets directory with gslapper
func launchAssetVideo(filename string) {
	// Check if file exists in Assets directory
	assetPath := filepath.Join("Assets", filename)
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		// File doesn't exist in Assets, try /usr/share/sysc-greet/Assets (production path)
		assetPath = filepath.Join("/usr/share/sysc-greet/Assets", filename)
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			// Asset not found, silently return
			return
		}
	}

	go func() {
		// Kill any existing gslapper process
		exec.Command("pkill", "-f", "gslapper").Run()

		// Start new gslapper with asset video
		cmd := exec.Command("gslapper", "-s", "-o", "loop panscan=1.0", "*", assetPath)
		cmd.Start()
	}()
}
