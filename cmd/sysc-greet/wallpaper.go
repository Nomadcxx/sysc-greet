package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nomadcxx/sysc-greet/internal/cache"
	"github.com/Nomadcxx/sysc-greet/internal/wallpaper"
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
				if !file.IsDir() {
					// Check for video and static image extensions
					ext := strings.ToLower(filepath.Ext(file.Name()))
					validExts := []string{".mp4", ".mkv", ".webm", ".avi", ".mov", ".png", ".jpg", ".jpeg", ".webp", ".gif"}
					for _, validExt := range validExts {
						if ext == validExt {
							m.menuOptions = append(m.menuOptions, file.Name())
							break
						}
					}
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

// launchGslapperWallpaper changes wallpaper via gSlapper IPC, falling back to process restart if needed
func launchGslapperWallpaper(wallpaperFilename string) {
	// CHANGED 2025-12-24 - Use IPC for wallpaper changes (no flicker), fallback to restart
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

	go func() {
		// Try IPC first (preferred - no flicker)
		if wallpaper.IsGSlapperRunning() {
			if err := wallpaper.ChangeWallpaper(wallpaperPath); err == nil {
				return // Success via IPC
			}
		}

		// Fallback: kill and restart gslapper
		exec.Command("pkill", "-f", "gslapper").Run()

		// Determine if video or static image
		ext := strings.ToLower(filepath.Ext(wallpaperFilename))
		var cmd *exec.Cmd
		switch ext {
		case ".mp4", ".mkv", ".webm", ".avi", ".mov":
			// Video: use loop and panscan
			cmd = exec.Command("gslapper", "-s", "-o", "loop panscan=1.0", "*", wallpaperPath)
		default:
			// Static image: use fill mode
			cmd = exec.Command("gslapper", "-o", "fill", "*", wallpaperPath)
		}
		cmd.Start()
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
