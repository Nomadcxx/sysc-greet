package themes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"
)

// Theme defines colors, borders, and ASCII art for sessions
type Theme struct {
	Name            string
	PrimaryColor    lipgloss.Color
	SecondaryColor  lipgloss.Color
	AccentColor     lipgloss.Color
	BackgroundColor lipgloss.Color
	Border          lipgloss.Border
	ASCIIArt        map[string]string // session name -> ASCII art
	Animation       struct {
		Enabled bool
		Type    string
		Speed   int
	}
	Gradient struct {
		Enabled bool
		Start   string
		End     string
	}
	Font struct {
		Bold      bool
		Italic    bool
		Underline bool
	}
	GradientStart lipgloss.Color
	GradientEnd   lipgloss.Color
}

// DefaultTheme is a basic theme
var DefaultTheme = Theme{
	Name:           "Default",
	PrimaryColor:   lipgloss.Color("39"),  // Blue
	SecondaryColor: lipgloss.Color("86"),  // Cyan
	AccentColor:    lipgloss.Color("202"), // Orange
	Border:         lipgloss.NormalBorder(),
	ASCIIArt: map[string]string{
		"GNOME": `
    _____
   /     \
  /  G N  \
 |   O M   |
  \  E     /
   \_____/
 `,
		"KDE Plasma": `
    _____
   /     \
  /  K D  \
 |   E P   |
  \  L A   /
   \  M   /
    \___/
 `,
		// Add more as needed
	},
}

// DarkTheme for dark mode
var DarkTheme = Theme{
	Name:            "Dark",
	PrimaryColor:    lipgloss.Color("15"), // White
	SecondaryColor:  lipgloss.Color("8"),  // Gray
	AccentColor:     lipgloss.Color("11"), // Yellow
	BackgroundColor: lipgloss.Color("0"),  // Black
	Border:          lipgloss.ThickBorder(),
	ASCIIArt:        DefaultTheme.ASCIIArt, // Same ASCII for now
}

// HighContrastTheme for accessibility
var HighContrastTheme = Theme{
	Name:            "High Contrast",
	PrimaryColor:    lipgloss.Color("15"), // White
	SecondaryColor:  lipgloss.Color("0"),  // Black
	AccentColor:     lipgloss.Color("9"),  // Red
	BackgroundColor: lipgloss.Color("0"),  // Black
	Border:          lipgloss.DoubleBorder(),
	ASCIIArt:        DefaultTheme.ASCIIArt,
}

// NeonTheme for a vibrant look
var NeonTheme = Theme{
	Name:            "Neon",
	PrimaryColor:    lipgloss.Color("201"), // Magenta
	SecondaryColor:  lipgloss.Color("51"),  // Cyan
	AccentColor:     lipgloss.Color("226"), // Green
	BackgroundColor: lipgloss.Color("0"),   // Black
	Border:          lipgloss.RoundedBorder(),
	ASCIIArt:        DefaultTheme.ASCIIArt,
}

// GetASCIIArt returns ASCII art for a session, or the session name if not found
func GetASCIIArt(session string, theme Theme) string {
	if art, ok := theme.ASCIIArt[session]; ok {
		return art
	}
	return session
}

// ApplyTheme applies theme colors and font styles to a style
func ApplyTheme(style lipgloss.Style, theme Theme) lipgloss.Style {
	style = style.Foreground(theme.PrimaryColor).BorderForeground(theme.SecondaryColor)
	if theme.Font.Bold {
		style = style.Bold(true)
	}
	if theme.Font.Italic {
		style = style.Italic(true)
	}
	if theme.Font.Underline {
		style = style.Underline(true)
	}
	return style
}

// ApplyGradient applies a gradient to ASCII art
func ApplyGradient(asciiArt string, theme Theme) string {
	if !theme.Gradient.Enabled {
		return asciiArt
	}
	lines := strings.Split(asciiArt, "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, ApplyForegroundGrad(line, theme.GradientStart, theme.GradientEnd))
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// ApplyForegroundGrad applies a horizontal gradient to a string
func ApplyForegroundGrad(input string, color1, color2 lipgloss.Color) string {
	if input == "" {
		return ""
	}
	var o strings.Builder
	clusters := ForegroundGrad(input, false, color1, color2)
	for _, c := range clusters {
		fmt.Fprint(&o, c)
	}
	return o.String()
}

// ForegroundGrad creates gradient clusters for a string
func ForegroundGrad(input string, bold bool, color1, color2 lipgloss.Color) []string {
	if input == "" {
		return []string{""}
	}
	if len(input) == 1 {
		style := lipgloss.NewStyle().Foreground(color1)
		if bold {
			style.Bold(true)
		}
		return []string{style.Render(input)}
	}
	var clusters []string
	gr := uniseg.NewGraphemes(input)
	for gr.Next() {
		clusters = append(clusters, string(gr.Runes()))
	}

	ramp := blendColors(len(clusters), color1, color2)
	for i, c := range ramp {
		style := lipgloss.NewStyle().Foreground(c)
		if bold {
			style.Bold(true)
		}
		clusters[i] = style.Render(clusters[i])
	}
	return clusters
}

// blendColors blends colors for gradient
func blendColors(size int, stops ...lipgloss.Color) []lipgloss.Color {
	if len(stops) < 2 {
		return nil
	}

	numSegments := len(stops) - 1
	blended := make([]lipgloss.Color, 0, size)

	segmentSizes := make([]int, numSegments)
	baseSize := size / numSegments
	remainder := size % numSegments

	for i := range numSegments {
		segmentSizes[i] = baseSize
		if i < remainder {
			segmentSizes[i]++
		}
	}

	for i := range numSegments {
		c1 := stops[i]
		c2 := stops[i+1]
		segmentSize := segmentSizes[i]

		for j := range segmentSize {
			// For simplicity, just alternate between colors
			if j%2 == 0 {
				blended = append(blended, c1)
			} else {
				blended = append(blended, c2)
			}
		}
	}

	return blended
}

// ThemeConfig represents the TOML structure for theme files
type ThemeConfig struct {
	Colors struct {
		Primary    string `toml:"primary"`
		Secondary  string `toml:"secondary"`
		Accent     string `toml:"accent"`
		Background string `toml:"background"`
	} `toml:"colors"`
	Border struct {
		Style string `toml:"style"`
	} `toml:"border"`
	ASCIIArt  map[string]string `toml:"ascii_art"`
	Animation struct {
		Enabled bool   `toml:"enabled"`
		Type    string `toml:"type"`
		Speed   int    `toml:"speed"`
	} `toml:"animation"`
	Gradient struct {
		Enabled bool   `toml:"enabled"`
		Start   string `toml:"start"`
		End     string `toml:"end"`
	} `toml:"gradient"`
	Font struct {
		Bold      bool `toml:"bold"`
		Italic    bool `toml:"italic"`
		Underline bool `toml:"underline"`
	} `toml:"font"`
}

// LoadThemeFromFile loads a theme from a TOML file
func LoadThemeFromFile(filePath string) (Theme, error) {
	var config ThemeConfig
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return Theme{}, fmt.Errorf("failed to decode theme file %s: %v", filePath, err)
	}

	// Parse colors
	primaryColor := lipgloss.Color(config.Colors.Primary)
	secondaryColor := lipgloss.Color(config.Colors.Secondary)
	accentColor := lipgloss.Color(config.Colors.Accent)
	backgroundColor := lipgloss.Color(config.Colors.Background)

	// Parse border style
	var border lipgloss.Border
	switch config.Border.Style {
	case "thick":
		border = lipgloss.ThickBorder()
	case "double":
		border = lipgloss.DoubleBorder()
	case "rounded":
		border = lipgloss.RoundedBorder()
	case "hidden":
		border = lipgloss.HiddenBorder()
	default:
		border = lipgloss.NormalBorder()
	}

	// Parse gradient colors
	var gradientStart, gradientEnd lipgloss.Color
	if config.Gradient.Enabled {
		gradientStart = lipgloss.Color(config.Gradient.Start)
		gradientEnd = lipgloss.Color(config.Gradient.End)
	}

	theme := Theme{
		PrimaryColor:    primaryColor,
		SecondaryColor:  secondaryColor,
		AccentColor:     accentColor,
		BackgroundColor: backgroundColor,
		Border:          border,
		ASCIIArt:        config.ASCIIArt,
		GradientStart:   gradientStart,
		GradientEnd:     gradientEnd,
	}
	theme.Animation.Enabled = config.Animation.Enabled
	theme.Animation.Type = config.Animation.Type
	theme.Animation.Speed = config.Animation.Speed
	theme.Gradient.Enabled = config.Gradient.Enabled
	theme.Gradient.Start = config.Gradient.Start
	theme.Gradient.End = config.Gradient.End
	theme.Font.Bold = config.Font.Bold
	theme.Font.Italic = config.Font.Italic
	theme.Font.Underline = config.Font.Underline
	return theme, nil
}

// LoadThemesFromDir loads all themes from the themes directory
func LoadThemesFromDir(themesDir string) (map[string]Theme, error) {
	themes := make(map[string]Theme)

	err := filepath.Walk(themesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".conf" {
			themeName := strings.TrimSuffix(filepath.Base(path), ".conf")
			theme, err := LoadThemeFromFile(path)
			if err != nil {
				return fmt.Errorf("failed to load theme %s: %v", themeName, err)
			}
			themes[themeName] = theme
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return themes, nil
}
