package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const sharedShellTheme = "/var/lib/sysc-greet/shell-theme/theme"

func shellThemePath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		path := filepath.Join(dir, "sysc-shell", "shell-theme")
		if _, err := os.Lstat(path); err == nil {
			return path
		}
	}
	return sharedShellTheme
}

// A theme file carries a bounded name, never a path or executable content.
func shellThemeName(path string, available []string) string {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64 {
		return ""
	}
	data, err := io.ReadAll(io.LimitReader(f, 65))
	if err != nil || len(data) > 64 {
		return ""
	}
	name := strings.TrimSpace(string(data))
	switch name {
	case "tokyo-night", "tokyonight":
		name = "Tokyo Night"
	case "catppuccin-mocha":
		name = "Catppuccin"
	}
	for _, candidate := range available {
		if strings.EqualFold(name, candidate) {
			return candidate
		}
	}
	return ""
}

func (m model) startupTheme(cached string) string {
	requested := m.config.ThemeName
	if requested == "" && m.followShell {
		path := m.config.ShellThemeFile
		if path == "" {
			path = shellThemePath()
		}
		requested = shellThemeName(path, m.availableThemes)
	}
	return startupThemeName(requested, cached, m.availableThemes)
}
