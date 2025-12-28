package themes

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss/v2"
)

// CustomThemes holds loaded custom theme configurations
var CustomThemes = make(map[string]ThemeColors)

// CustomThemeConfig represents the TOML structure for custom theme files
type CustomThemeConfig struct {
	Name   string `toml:"name"`
	Colors struct {
		BgBase      string `toml:"bg_base"`
		BgActive    string `toml:"bg_active"`
		Primary     string `toml:"primary"`
		Secondary   string `toml:"secondary"`
		Accent      string `toml:"accent"`
		Warning     string `toml:"warning"`
		Danger      string `toml:"danger"`
		FgPrimary   string `toml:"fg_primary"`
		FgSecondary string `toml:"fg_secondary"`
		FgMuted     string `toml:"fg_muted"`
		BorderFocus string `toml:"border_focus"`
	} `toml:"colors"`
}

// ThemeColors holds all colors for a theme
type ThemeColors struct {
	Name string

	// Backgrounds
	BgBase     color.Color
	BgElevated color.Color
	BgSubtle   color.Color
	BgActive   color.Color

	// Primary brand colors
	Primary   color.Color
	Secondary color.Color
	Accent    color.Color
	Warning   color.Color
	Danger    color.Color

	// Text colors
	FgPrimary   color.Color
	FgSecondary color.Color
	FgMuted     color.Color
	FgSubtle    color.Color

	// Border colors
	BorderDefault color.Color
	BorderFocus   color.Color
}

// GetTheme returns theme colors for the given theme name
func GetTheme(themeName string) ThemeColors {
	// Check custom themes first (allows overriding built-ins)
	if theme, ok := CustomThemes[strings.ToLower(themeName)]; ok {
		return theme
	}

	switch strings.ToLower(themeName) {
	case "gruvbox":
		return ThemeColors{
			Name:          "gruvbox",
			BgBase:        lipgloss.Color("#282828"),
			BgElevated:    lipgloss.Color("#282828"),
			BgSubtle:      lipgloss.Color("#282828"),
			BgActive:      lipgloss.Color("#3c3836"),
			Primary:       lipgloss.Color("#fe8019"),
			Secondary:     lipgloss.Color("#8ec07c"),
			Accent:        lipgloss.Color("#fabd2f"),
			Warning:       lipgloss.Color("#d79921"),
			Danger:        lipgloss.Color("#cc241d"),
			FgPrimary:     lipgloss.Color("#ebdbb2"),
			FgSecondary:   lipgloss.Color("#d5c4a1"),
			FgMuted:       lipgloss.Color("#bdae93"),
			FgSubtle:      lipgloss.Color("#a89984"),
			BorderDefault: lipgloss.Color("#665c54"),
			BorderFocus:   lipgloss.Color("#fe8019"),
		}

	case "material":
		return ThemeColors{
			Name:          "material",
			BgBase:        lipgloss.Color("#263238"),
			BgElevated:    lipgloss.Color("#263238"),
			BgSubtle:      lipgloss.Color("#263238"),
			BgActive:      lipgloss.Color("#37474f"),
			Primary:       lipgloss.Color("#80cbc4"),
			Secondary:     lipgloss.Color("#64b5f6"),
			Accent:        lipgloss.Color("#ffab40"),
			Warning:       lipgloss.Color("#ffb300"),
			Danger:        lipgloss.Color("#f44336"),
			FgPrimary:     lipgloss.Color("#eceff1"),
			FgSecondary:   lipgloss.Color("#cfd8dc"),
			FgMuted:       lipgloss.Color("#90a4ae"),
			FgSubtle:      lipgloss.Color("#546e7a"),
			BorderDefault: lipgloss.Color("#37474f"),
			BorderFocus:   lipgloss.Color("#80cbc4"),
		}

	case "nord":
		return ThemeColors{
			Name:          "nord",
			BgBase:        lipgloss.Color("#2e3440"),
			BgElevated:    lipgloss.Color("#2e3440"),
			BgSubtle:      lipgloss.Color("#2e3440"),
			BgActive:      lipgloss.Color("#3b4252"),
			Primary:       lipgloss.Color("#81a1c1"),
			Secondary:     lipgloss.Color("#88c0d0"),
			Accent:        lipgloss.Color("#8fbcbb"),
			Warning:       lipgloss.Color("#ebcb8b"),
			Danger:        lipgloss.Color("#bf616a"),
			FgPrimary:     lipgloss.Color("#eceff4"),
			FgSecondary:   lipgloss.Color("#e5e9f0"),
			FgMuted:       lipgloss.Color("#d8dee9"),
			FgSubtle:      lipgloss.Color("#4c566a"),
			BorderDefault: lipgloss.Color("#3b4252"),
			BorderFocus:   lipgloss.Color("#81a1c1"),
		}

	case "dracula":
		return ThemeColors{
			Name:          "dracula",
			BgBase:        lipgloss.Color("#282a36"),
			BgElevated:    lipgloss.Color("#282a36"),
			BgSubtle:      lipgloss.Color("#282a36"),
			BgActive:      lipgloss.Color("#44475a"),
			Primary:       lipgloss.Color("#bd93f9"),
			Secondary:     lipgloss.Color("#8be9fd"),
			Accent:        lipgloss.Color("#50fa7b"),
			Warning:       lipgloss.Color("#f1fa8c"),
			Danger:        lipgloss.Color("#ff5555"),
			FgPrimary:     lipgloss.Color("#f8f8f2"),
			FgSecondary:   lipgloss.Color("#f1f2f6"),
			FgMuted:       lipgloss.Color("#6272a4"),
			FgSubtle:      lipgloss.Color("#44475a"),
			BorderDefault: lipgloss.Color("#44475a"),
			BorderFocus:   lipgloss.Color("#bd93f9"),
		}

	case "catppuccin", "catppuccin-mocha":
		return ThemeColors{
			Name:          "catppuccin",
			BgBase:        lipgloss.Color("#1e1e2e"),
			BgElevated:    lipgloss.Color("#1e1e2e"),
			BgSubtle:      lipgloss.Color("#1e1e2e"),
			BgActive:      lipgloss.Color("#313244"),
			Primary:       lipgloss.Color("#cba6f7"),
			Secondary:     lipgloss.Color("#89b4fa"),
			Accent:        lipgloss.Color("#a6e3a1"),
			Warning:       lipgloss.Color("#f9e2af"),
			Danger:        lipgloss.Color("#f38ba8"),
			FgPrimary:     lipgloss.Color("#cdd6f4"),
			FgSecondary:   lipgloss.Color("#bac2de"),
			FgMuted:       lipgloss.Color("#a6adc8"),
			FgSubtle:      lipgloss.Color("#585b70"),
			BorderDefault: lipgloss.Color("#313244"),
			BorderFocus:   lipgloss.Color("#cba6f7"),
		}

	case "tokyo night", "tokyonight", "tokyo-night":
		return ThemeColors{
			Name:          "tokyo-night",
			BgBase:        lipgloss.Color("#1a1b26"),
			BgElevated:    lipgloss.Color("#1a1b26"),
			BgSubtle:      lipgloss.Color("#1a1b26"),
			BgActive:      lipgloss.Color("#24283b"),
			Primary:       lipgloss.Color("#7aa2f7"),
			Secondary:     lipgloss.Color("#bb9af7"),
			Accent:        lipgloss.Color("#9ece6a"),
			Warning:       lipgloss.Color("#e0af68"),
			Danger:        lipgloss.Color("#f7768e"),
			FgPrimary:     lipgloss.Color("#c0caf5"),
			FgSecondary:   lipgloss.Color("#a9b1d6"),
			FgMuted:       lipgloss.Color("#565f89"),
			FgSubtle:      lipgloss.Color("#414868"),
			BorderDefault: lipgloss.Color("#24283b"),
			BorderFocus:   lipgloss.Color("#7aa2f7"),
		}

	case "solarized":
		return ThemeColors{
			Name:          "solarized",
			BgBase:        lipgloss.Color("#002b36"),
			BgElevated:    lipgloss.Color("#002b36"),
			BgSubtle:      lipgloss.Color("#002b36"),
			BgActive:      lipgloss.Color("#073642"),
			Primary:       lipgloss.Color("#268bd2"),
			Secondary:     lipgloss.Color("#2aa198"),
			Accent:        lipgloss.Color("#859900"),
			Warning:       lipgloss.Color("#b58900"),
			Danger:        lipgloss.Color("#dc322f"),
			FgPrimary:     lipgloss.Color("#fdf6e3"),
			FgSecondary:   lipgloss.Color("#eee8d5"),
			FgMuted:       lipgloss.Color("#93a1a1"),
			FgSubtle:      lipgloss.Color("#657b83"),
			BorderDefault: lipgloss.Color("#073642"),
			BorderFocus:   lipgloss.Color("#268bd2"),
		}

	case "monochrome":
		return ThemeColors{
			Name:          "monochrome",
			BgBase:        lipgloss.Color("#1a1a1a"),
			BgElevated:    lipgloss.Color("#1a1a1a"),
			BgSubtle:      lipgloss.Color("#1a1a1a"),
			BgActive:      lipgloss.Color("#2a2a2a"),
			Primary:       lipgloss.Color("#ffffff"),
			Secondary:     lipgloss.Color("#cccccc"),
			Accent:        lipgloss.Color("#888888"),
			Warning:       lipgloss.Color("#aaaaaa"),
			Danger:        lipgloss.Color("#999999"),
			FgPrimary:     lipgloss.Color("#ffffff"),
			FgSecondary:   lipgloss.Color("#cccccc"),
			FgMuted:       lipgloss.Color("#666666"),
			FgSubtle:      lipgloss.Color("#444444"),
			BorderDefault: lipgloss.Color("#333333"),
			BorderFocus:   lipgloss.Color("#ffffff"),
		}

	case "transishardjob":
		return ThemeColors{
			Name:          "transishardjob",
			BgBase:        lipgloss.Color("#1a1a1a"),
			BgElevated:    lipgloss.Color("#1a1a1a"),
			BgSubtle:      lipgloss.Color("#1a1a1a"),
			BgActive:      lipgloss.Color("#2a2a2a"),
			Primary:       lipgloss.Color("#5BCEFA"), // Trans flag light blue
			Secondary:     lipgloss.Color("#F5A9B8"), // Trans flag pink
			Accent:        lipgloss.Color("#FFFFFF"), // Trans flag white
			Warning:       lipgloss.Color("#F5A9B8"),
			Danger:        lipgloss.Color("#ff6b9d"),
			FgPrimary:     lipgloss.Color("#FFFFFF"),
			FgSecondary:   lipgloss.Color("#F5A9B8"),
			FgMuted:       lipgloss.Color("#5BCEFA"),
			FgSubtle:      lipgloss.Color("#999999"),
			BorderDefault: lipgloss.Color("#444444"),
			BorderFocus:   lipgloss.Color("#5BCEFA"),
		}

	case "eldritch":
		// Eldritch - https://github.com/eldritch-theme/eldritch
		return ThemeColors{
			Name:          "eldritch",
			BgBase:        lipgloss.Color("#212337"),
			BgElevated:    lipgloss.Color("#212337"),
			BgSubtle:      lipgloss.Color("#212337"),
			BgActive:      lipgloss.Color("#323449"),
			Primary:       lipgloss.Color("#37f499"),
			Secondary:     lipgloss.Color("#04d1f9"),
			Accent:        lipgloss.Color("#a48cf2"),
			Warning:       lipgloss.Color("#f1fc79"),
			Danger:        lipgloss.Color("#f16c75"),
			FgPrimary:     lipgloss.Color("#ebfafa"),
			FgSecondary:   lipgloss.Color("#ABB4DA"),
			FgMuted:       lipgloss.Color("#7081d0"),
			FgSubtle:      lipgloss.Color("#3b4261"),
			BorderDefault: lipgloss.Color("#3b4261"),
			BorderFocus:   lipgloss.Color("#37f499"),
		}

	case "rama":
		// RAMA theme - Inspired by RAMA keyboard aesthetics
		return ThemeColors{
			Name:          "rama",
			BgBase:        lipgloss.Color("#2b2d42"),
			BgElevated:    lipgloss.Color("#2b2d42"),
			BgSubtle:      lipgloss.Color("#2b2d42"),
			BgActive:      lipgloss.Color("#3b3d52"),
			Primary:       lipgloss.Color("#ef233c"),
			Secondary:     lipgloss.Color("#d90429"),
			Accent:        lipgloss.Color("#ef233c"),
			Warning:       lipgloss.Color("#f59e0b"),
			Danger:        lipgloss.Color("#ef233c"),
			FgPrimary:     lipgloss.Color("#edf2f4"),
			FgSecondary:   lipgloss.Color("#8d99ae"),
			FgMuted:       lipgloss.Color("#8d99ae"),
			FgSubtle:      lipgloss.Color("#6d7a8e"),
			BorderDefault: lipgloss.Color("#3b3d52"),
			BorderFocus:   lipgloss.Color("#ef233c"),
		}

	case "dark":
		// DARK theme - True black and true white minimalism
		return ThemeColors{
			Name:          "dark",
			BgBase:        lipgloss.Color("#000000"),
			BgElevated:    lipgloss.Color("#000000"),
			BgSubtle:      lipgloss.Color("#000000"),
			BgActive:      lipgloss.Color("#1a1a1a"),
			Primary:       lipgloss.Color("#ffffff"),
			Secondary:     lipgloss.Color("#ffffff"),
			Accent:        lipgloss.Color("#808080"),
			Warning:       lipgloss.Color("#aaaaaa"),
			Danger:        lipgloss.Color("#999999"),
			FgPrimary:     lipgloss.Color("#ffffff"),
			FgSecondary:   lipgloss.Color("#cccccc"),
			FgMuted:       lipgloss.Color("#666666"),
			FgSubtle:      lipgloss.Color("#444444"),
			BorderDefault: lipgloss.Color("#333333"),
			BorderFocus:   lipgloss.Color("#ffffff"),
		}

	default: // Fallback to Dracula
		return ThemeColors{
			Name:          "dracula",
			BgBase:        lipgloss.Color("#282a36"),
			BgElevated:    lipgloss.Color("#282a36"),
			BgSubtle:      lipgloss.Color("#282a36"),
			BgActive:      lipgloss.Color("#44475a"),
			Primary:       lipgloss.Color("#bd93f9"),
			Secondary:     lipgloss.Color("#8be9fd"),
			Accent:        lipgloss.Color("#50fa7b"),
			Warning:       lipgloss.Color("#f1fa8c"),
			Danger:        lipgloss.Color("#ff5555"),
			FgPrimary:     lipgloss.Color("#f8f8f2"),
			FgSecondary:   lipgloss.Color("#f1f2f6"),
			FgMuted:       lipgloss.Color("#6272a4"),
			FgSubtle:      lipgloss.Color("#44475a"),
			BorderDefault: lipgloss.Color("#44475a"),
			BorderFocus:   lipgloss.Color("#bd93f9"),
		}
	}
}

// GetAvailableThemes returns list of all theme names
func GetAvailableThemes() []string {
	return []string{
		"Dracula",
		"Catppuccin",
		"Nord",
		"Tokyo Night",
		"Gruvbox",
		"Material",
		"Solarized",
		"Monochrome",
		"TransIsHardJob",
		"Eldritch",
		"RAMA",
		"Dark",
	}
}

// ScanCustomThemes scans directories for .toml theme files and loads them
func ScanCustomThemes(dirs []string) []string {
	var names []string
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Directory doesn't exist, skip silently
		}

		files, err := filepath.Glob(filepath.Join(dir, "*.toml"))
		if err != nil {
			continue
		}

		for _, f := range files {
			theme, err := loadCustomTheme(f)
			if err != nil {
				// Silently skip invalid theme files
				continue
			}

			name := theme.Name
			if name == "" {
				name = strings.TrimSuffix(filepath.Base(f), ".toml")
			}

			if _, exists := CustomThemes[strings.ToLower(name)]; exists {
				// Theme with this name already loaded, later one wins (no error, just note)
				// This allows user themes to override system themes intentionally
			}
			CustomThemes[strings.ToLower(name)] = theme
			names = append(names, name)
		}
	}
	return names
}

// loadCustomTheme loads a single custom theme from a TOML file
func loadCustomTheme(path string) (ThemeColors, error) {
	var config CustomThemeConfig
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return ThemeColors{}, err
	}

	// Validate required color fields are non-empty
	requiredFields := map[string]string{
		"bg_base":      config.Colors.BgBase,
		"bg_active":    config.Colors.BgActive,
		"primary":      config.Colors.Primary,
		"secondary":    config.Colors.Secondary,
		"accent":       config.Colors.Accent,
		"warning":      config.Colors.Warning,
		"danger":       config.Colors.Danger,
		"fg_primary":   config.Colors.FgPrimary,
		"fg_secondary": config.Colors.FgSecondary,
		"fg_muted":     config.Colors.FgMuted,
		"border_focus": config.Colors.BorderFocus,
	}

	for field, value := range requiredFields {
		if strings.TrimSpace(value) == "" {
			return ThemeColors{}, fmt.Errorf("missing required field: %s", field)
		}
	}

	name := config.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), ".toml")
	}

	return ThemeColors{
		Name:          name,
		BgBase:        lipgloss.Color(config.Colors.BgBase),
		BgElevated:    lipgloss.Color(config.Colors.BgBase), // Same as BgBase
		BgSubtle:      lipgloss.Color(config.Colors.BgBase), // Same as BgBase
		BgActive:      lipgloss.Color(config.Colors.BgActive),
		Primary:       lipgloss.Color(config.Colors.Primary),
		Secondary:     lipgloss.Color(config.Colors.Secondary),
		Accent:        lipgloss.Color(config.Colors.Accent),
		Warning:       lipgloss.Color(config.Colors.Warning),
		Danger:        lipgloss.Color(config.Colors.Danger),
		FgPrimary:     lipgloss.Color(config.Colors.FgPrimary),
		FgSecondary:   lipgloss.Color(config.Colors.FgSecondary),
		FgMuted:       lipgloss.Color(config.Colors.FgMuted),
		FgSubtle:      lipgloss.Color(config.Colors.FgMuted), // Use FgMuted as fallback
		BorderDefault: lipgloss.Color(config.Colors.BgActive),
		BorderFocus:   lipgloss.Color(config.Colors.BorderFocus),
	}, nil
}
