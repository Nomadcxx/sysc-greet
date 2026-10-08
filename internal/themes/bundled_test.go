package themes

import (
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

var bundled = []string{"ayu", "amber", "blue", "purple", "green", "orange", "rose-pine", "kanagawa", "noctalia", "eldritch-abyss", "void", "red", "cyan", "coral", "pink"}

var hexRe = regexp.MustCompile(`^#[0-9a-f]{6}$`)

func bundledDir() string { return filepath.Join("..", "..", "themes") }

func channel(h string, i int) float64 {
	v, _ := strconv.ParseUint(h[1+2*i:3+2*i], 16, 8)
	c := float64(v) / 255
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func luminance(h string) float64 {
	return 0.2126*channel(h, 0) + 0.7152*channel(h, 1) + 0.0722*channel(h, 2)
}

func contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestBundledThemesLoad(t *testing.T) {
	for _, name := range bundled {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(bundledDir(), name+".toml")
			th, err := loadCustomTheme(path)
			if err != nil {
				t.Fatalf("loadCustomTheme: %v", err)
			}
			if got := strings.ToLower(th.Name); got != name {
				t.Fatalf("name lowercases to %q, want shell seed %q", got, name)
			}

			var cfg CustomThemeConfig
			if _, err := toml.DecodeFile(path, &cfg); err != nil {
				t.Fatal(err)
			}
			c := cfg.Colors
			for field, v := range map[string]string{
				"bg_base": c.BgBase, "bg_active": c.BgActive, "primary": c.Primary,
				"secondary": c.Secondary, "accent": c.Accent, "warning": c.Warning,
				"danger": c.Danger, "fg_primary": c.FgPrimary, "fg_secondary": c.FgSecondary,
				"fg_muted": c.FgMuted, "border_focus": c.BorderFocus,
			} {
				if !hexRe.MatchString(v) {
					t.Fatalf("%s = %q, want lowercase #rrggbb", field, v)
				}
			}
			if c.BorderFocus != c.Primary {
				t.Errorf("border_focus %s != primary %s", c.BorderFocus, c.Primary)
			}
		})
	}
}

// Guards future edits; today's values pass with margin (see spec).
func TestBundledThemesContrast(t *testing.T) {
	for _, name := range bundled {
		t.Run(name, func(t *testing.T) {
			var cfg CustomThemeConfig
			if _, err := toml.DecodeFile(filepath.Join(bundledDir(), name+".toml"), &cfg); err != nil {
				t.Fatal(err)
			}
			c := cfg.Colors
			for _, v := range []string{c.BgBase, c.FgMuted, c.Primary, c.Secondary, c.Accent, c.Warning, c.Danger} {
				if !hexRe.MatchString(v) {
					t.Fatalf("invalid contrast color %q", v)
				}
			}
			if r := contrast(c.FgMuted, c.BgBase); r < 4.5 {
				t.Errorf("fg_muted on bg_base = %.1f, want >= 4.5", r)
			}
			for field, v := range map[string]string{
				"primary": c.Primary, "secondary": c.Secondary, "accent": c.Accent,
				"warning": c.Warning, "danger": c.Danger,
			} {
				if r := contrast(v, c.BgBase); r < 3 {
					t.Errorf("%s on bg_base = %.1f, want >= 3", field, r)
				}
			}
		})
	}
}

func TestScanBundledThemes(t *testing.T) {
	original := maps.Clone(CustomThemes)
	t.Cleanup(func() { CustomThemes = original })
	got := ScanCustomThemes([]string{bundledDir()})
	have := map[string]bool{}
	for _, n := range got {
		have[strings.ToLower(n)] = true
	}
	for _, n := range bundled {
		if !have[n] {
			t.Errorf("ScanCustomThemes missing %q (got %v)", n, got)
		}
	}
}

// A user directory scanned after the system one overrides a bundled theme.
func TestUserThemeOverridesBundled(t *testing.T) {
	original := maps.Clone(CustomThemes)
	t.Cleanup(func() { CustomThemes = original })
	if _, err := loadCustomTheme(filepath.Join(bundledDir(), "blue.toml")); err != nil {
		t.Fatal(err)
	}
	userDir := t.TempDir()
	const user = `name = "Blue"
[colors]
bg_base = "#000000"
bg_active = "#111111"
primary = "#ff0000"
secondary = "#00ff00"
accent = "#0000ff"
warning = "#ffff00"
danger = "#ff00ff"
fg_primary = "#ffffff"
fg_secondary = "#eeeeee"
fg_muted = "#dddddd"
border_focus = "#ff0000"
`
	if err := os.WriteFile(filepath.Join(userDir, "blue.toml"), []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}
	ScanCustomThemes([]string{bundledDir(), userDir})
	if got := colorToHex(CustomThemes["blue"].Primary); got != "#ff0000" {
		t.Fatalf("blue primary = %s, want user override #ff0000", got)
	}
}

func TestScanBundledThemesSkipsMissingField(t *testing.T) {
	original := maps.Clone(CustomThemes)
	CustomThemes = make(map[string]ThemeColors)
	t.Cleanup(func() { CustomThemes = original })
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join(bundledDir(), "blue.toml"))
	if err != nil {
		t.Fatal(err)
	}
	// Omit border_focus from an otherwise valid bundled theme.
	data = []byte(strings.ReplaceAll(string(data), "border_focus = \"#42a5f5\"", ""))
	path := filepath.Join(dir, "blue.toml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCustomTheme(path); err == nil {
		t.Fatal("loadCustomTheme accepted a missing border_focus")
	}
	if got := ScanCustomThemes([]string{dir}); len(got) != 0 {
		t.Fatalf("ScanCustomThemes returned invalid theme: %v", got)
	}
	if _, ok := CustomThemes["blue"]; ok {
		t.Fatal("invalid theme entered the registry")
	}
}
