package animations

import "strings"

// GetFirePalette returns theme-specific fire colors
func GetFirePalette(themeName string) []string {
	// Case-insensitive matching
	switch strings.ToLower(themeName) {
	case "dracula":
		return []string{
			"#282a36", // Background
			"#44475a", // Current line
			"#6272a4", // Comment
			"#8be9fd", // Cyan
			"#50fa7b", // Green
			"#f1fa8c", // Yellow
			"#ffb86c", // Orange
			"#ff79c6", // Pink
			"#ff5555", // Red (hottest)
		}
	case "catppuccin", "catppuccin-mocha":
		return []string{
			"#1e1e2e", // Base
			"#181825", // Mantle
			"#313244", // Surface0
			"#45475a", // Surface1
			"#f38ba8", // Red
			"#fab387", // Peach
			"#f9e2af", // Yellow
			"#a6e3a1", // Green (hot tip)
		}
	case "nord":
		return []string{
			"#2e3440", // Polar Night
			"#3b4252",
			"#434c5e",
			"#4c566a",
			"#bf616a", // Aurora Red
			"#d08770", // Aurora Orange
			"#ebcb8b", // Aurora Yellow
			"#a3be8c", // Aurora Green
		}
	case "tokyo-night", "tokyonight":
		return []string{
			"#1a1b26", // Background
			"#24283b", // Background Dark
			"#414868", // Foreground Gutter
			"#f7768e", // Red
			"#ff9e64", // Orange
			"#e0af68", // Yellow
			"#9ece6a", // Green
		}
	case "gruvbox":
		return []string{
			"#282828", // Background
			"#3c3836", // BG1
			"#504945", // BG2
			"#cc241d", // Red
			"#d65d0e", // Orange
			"#d79921", // Yellow
			"#fabd2f", // Bright Yellow
			"#b8bb26", // Green (hot)
		}
	case "material":
		return []string{
			"#263238", // Background
			"#37474f", // Lighter bg
			"#546e7a", // Selection
			"#f07178", // Red
			"#f78c6c", // Orange
			"#ffcb6b", // Yellow
			"#c3e88d", // Green
		}
	case "solarized":
		return []string{
			"#002b36", // Base03 - darkest
			"#073642", // Base02
			"#586e75", // Base01
			"#dc322f", // Red
			"#cb4b16", // Orange
			"#b58900", // Yellow
			"#859900", // Green
		}
	case "monochrome":
		return []string{
			"#1a1a1a", // Dark gray
			"#2a2a2a",
			"#3a3a3a",
			"#4a4a4a",
			"#5a5a5a",
			"#7a7a7a",
			"#9a9a9a",
			"#bababa",
			"#dadada", // Light gray (hottest)
		}
	case "transishardjob":
		return []string{
			"#55cdfc", // Trans blue
			"#f7a8b8", // Trans pink
			"#ffffff", // White
			"#f7a8b8", // Pink again
			"#55cdfc", // Blue again
			"#ffffff", // White (hottest)
		}
	default:
		return GetDefaultFirePalette()
	}
}

// GetDefaultFirePalette returns classic DOOM-style fire palette
func GetDefaultFirePalette() []string {
	return []string{
		"#000000", "#1a0000", "#330000", "#4d0000",
		"#660000", "#7f0000", "#990000", "#b30000",
		"#cc0000", "#e60000", "#ff0000", "#ff1a1a",
		"#ff3333", "#ff4d4d", "#ff6600", "#ff7f00",
		"#ff9900", "#ffb300", "#ffcc00", "#ffe600",
		"#ffff00", "#ffff33", "#ffff66", "#ffff99",
		"#ffffcc", "#ffffff",
	}
}

// GetMatrixPalette returns theme-specific matrix rain colors
func GetMatrixPalette(themeName string) []string {
	switch strings.ToLower(themeName) {
	case "dracula":
		return []string{"#282a36", "#44475a", "#6272a4", "#8be9fd", "#50fa7b", "#ff5555"}
	case "catppuccin", "catppuccin-mocha":
		return []string{"#1e1e2e", "#313244", "#45475a", "#89dceb", "#a6e3a1", "#f38ba8"}
	case "nord":
		return []string{"#2e3440", "#3b4252", "#434c5e", "#88c0d0", "#81a1c1", "#bf616a"}
	case "tokyo-night", "tokyonight":
		return []string{"#1a1b26", "#24283b", "#414868", "#7aa2f7", "#9ece6a", "#f7768e"}
	case "gruvbox":
		return []string{"#282828", "#3c3836", "#504945", "#83a598", "#b8bb26", "#fb4934"}
	case "material":
		return []string{"#263238", "#37474f", "#546e7a", "#89ddff", "#c3e88d", "#f07178"}
	case "solarized":
		return []string{"#002b36", "#073642", "#586e75", "#2aa198", "#859900", "#dc322f"}
	case "monochrome":
		return []string{"#1a1a1a", "#3a3a3a", "#5a5a5a", "#7a7a7a", "#9a9a9a", "#bababa"}
	case "transishardjob":
		return []string{"#1a1a1a", "#55cdfc", "#f7a8b8", "#ffffff", "#f7a8b8", "#55cdfc"}
	default:
		return []string{"#001100", "#003300", "#005500", "#007700", "#00aa00", "#00ff00"}
	}
}

// GetParticlePalette returns theme-specific particle colors
func GetParticlePalette(themeName string) []string {
	switch strings.ToLower(themeName) {
	case "dracula":
		return []string{"#bd93f9", "#ff79c6", "#8be9fd", "#50fa7b"}
	case "catppuccin", "catppuccin-mocha":
		return []string{"#cba6f7", "#f38ba8", "#89dceb", "#a6e3a1"}
	case "nord":
		return []string{"#88c0d0", "#81a1c1", "#5e81ac", "#8fbcbb"}
	case "tokyo-night", "tokyonight":
		return []string{"#7aa2f7", "#bb9af7", "#7dcfff", "#9ece6a"}
	case "gruvbox":
		return []string{"#d3869b", "#83a598", "#b8bb26", "#fabd2f"}
	case "material":
		return []string{"#89ddff", "#f07178", "#c3e88d", "#ffcb6b"}
	case "solarized":
		return []string{"#268bd2", "#2aa198", "#859900", "#b58900"}
	case "monochrome":
		return []string{"#5a5a5a", "#7a7a7a", "#9a9a9a", "#bababa"}
	case "transishardjob":
		return []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	default:
		return []string{"#ffffff", "#00ffff", "#ff00ff", "#ffff00"}
	}
}

// GetRainPalette returns theme-specific rain colors
func GetRainPalette(themeName string) []string {
	switch strings.ToLower(themeName) {
	case "dracula":
		return []string{"#8be9fd", "#50fa7b", "#ffb86c", "#ff79c6", "#bd93f9"}
	case "catppuccin", "catppuccin-mocha":
		return []string{"#89dceb", "#a6e3a1", "#f9e2af", "#f5c2e7", "#cba6f7"}
	case "nord":
		return []string{"#88c0d0", "#81a1c1", "#5e81ac", "#8fbcbb"}
	case "tokyo-night", "tokyonight":
		return []string{"#7dcfff", "#7aa2f7", "#2ac3de", "#b4f9f8"}
	case "gruvbox":
		return []string{"#83a598", "#8ec07c", "#d3869b", "#fabd2f"}
	case "material":
		return []string{"#89ddff", "#82aaff", "#c3e88d", "#ffcb6b"}
	case "solarized":
		return []string{"#2aa198", "#268bd2", "#6c71c4", "#859900"}
	case "monochrome":
		return []string{"#cccccc", "#aaaaaa", "#888888", "#666666"}
	case "transishardjob":
		return []string{"#55cdfc", "#f7a8b8", "#ffffff"}
	default:
		return []string{"#00ff00", "#00cc00", "#009900", "#006600"}
	}
}

// CHANGED 2025-10-10 - Screensaver palette for theme-aware colors
// GetScreensaverPalette returns theme-specific colors for screensaver elements
// Returns: [background, ascii_primary, ascii_secondary, clock_primary, clock_secondary, date_color]
func GetScreensaverPalette(themeName string) []string {
	switch strings.ToLower(themeName) {
	case "dracula":
		return []string{"#282a36", "#bd93f9", "#8be9fd", "#50fa7b", "#f1fa8c", "#f8f8f2"}
	case "catppuccin", "catppuccin-mocha":
		return []string{"#1e1e2e", "#cba6f7", "#89b4fa", "#a6e3a1", "#f9e2af", "#cdd6f4"}
	case "nord":
		return []string{"#2e3440", "#81a1c1", "#88c0d0", "#8fbcbb", "#d8dee9", "#eceff4"}
	case "tokyo-night", "tokyonight":
		return []string{"#1a1b26", "#7aa2f7", "#bb9af7", "#9ece6a", "#e0af68", "#c0caf5"}
	case "gruvbox":
		return []string{"#282828", "#fe8019", "#8ec07c", "#fabd2f", "#d79921", "#ebdbb2"}
	case "material":
		return []string{"#263238", "#80cbc4", "#64b5f6", "#ffab40", "#ffd54f", "#eceff1"}
	case "solarized":
		return []string{"#002b36", "#268bd2", "#2aa198", "#859900", "#b58900", "#fdf6e3"}
	case "monochrome":
		return []string{"#1a1a1a", "#ffffff", "#cccccc", "#888888", "#666666", "#ffffff"}
	case "transishardjob":
		return []string{"#1a1a1a", "#5BCEFA", "#F5A9B8", "#FFFFFF", "#F5A9B8", "#FFFFFF"}
	default:
		return []string{"#1a1a1a", "#8b5cf6", "#06b6d4", "#10b981", "#f59e0b", "#f8fafc"}
	}
}
