package main

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	syscgoanimations "github.com/Nomadcxx/sysc-Go/animations"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func defaultANSILoveConfig() syscgoanimations.ANSILoveConfig {
	return syscgoanimations.ANSILoveConfig{Backend: syscgoanimations.ANSILoveQuad, Depth: ansiLoveColorDepth(colorProfile), Fit: syscgoanimations.ANSILoveFitPan, Zoom: 1.5, CellW: 8, CellH: 16, Palette: "vga", Motion: syscgoanimations.ANSILoveMotionSubtle}
}

func ansiLoveColorDepth(profile colorprofile.Profile) syscgoanimations.ANSILoveColorDepth {
	switch profile {
	case colorprofile.TrueColor:
		return syscgoanimations.ANSILoveTrueColor
	case colorprofile.ANSI256:
		return syscgoanimations.ANSILove256
	default:
		return syscgoanimations.ANSILove16
	}
}

// Invalid optional artwork configuration must not prevent login.
func loadANSILoveConfig(path string) (syscgoanimations.ANSILoveConfig, bool, error) {
	defaults := defaultANSILoveConfig()
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return defaults, false, nil
	}
	if err != nil {
		return defaults, false, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 16*1024+1))
	if err != nil {
		return defaults, false, err
	}
	if len(data) > 16*1024 {
		return defaults, false, fmt.Errorf("%s: configuration exceeds 16 KiB", path)
	}
	cfg, cellOverride := defaults, false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, value, found := strings.Cut(text, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		valid := found
		switch key {
		case "ansi_love_backend":
			choices := map[string]syscgoanimations.ANSILoveBackend{"auto": syscgoanimations.ANSILoveAuto, "exact": syscgoanimations.ANSILoveExact, "half": syscgoanimations.ANSILoveHalf, "quad": syscgoanimations.ANSILoveQuad, "sextant": syscgoanimations.ANSILoveSextant, "octant": syscgoanimations.ANSILoveOctant, "braille-two-color": syscgoanimations.ANSILoveBrailleTwoColor, "braille-fg": syscgoanimations.ANSILoveBrailleFG}
			cfg.Backend, valid = choices[value]
		case "ansi_love_fit":
			choices := map[string]syscgoanimations.ANSILoveFit{"pan": syscgoanimations.ANSILoveFitPan, "contain": syscgoanimations.ANSILoveFitContain, "width": syscgoanimations.ANSILoveFitWidth, "exact": syscgoanimations.ANSILoveFitExact}
			cfg.Fit, valid = choices[value]
		case "ansi_love_zoom":
			cfg.Zoom, err = strconv.ParseFloat(value, 64)
			valid = err == nil && !math.IsNaN(cfg.Zoom) && !math.IsInf(cfg.Zoom, 0) && cfg.Zoom >= 1 && cfg.Zoom <= 3
		case "ansi_love_cell":
			width, height, ok := strings.Cut(value, "x")
			var widthErr, heightErr error
			cfg.CellW, widthErr = strconv.Atoi(width)
			cfg.CellH, heightErr = strconv.Atoi(height)
			valid = ok && widthErr == nil && heightErr == nil && cfg.CellW > 0 && cfg.CellH > 0 && cfg.CellW <= 1024 && cfg.CellH <= 1024
			cellOverride = valid
		case "ansi_love_palette":
			valid = value == "vga" || value == "theme"
			cfg.Palette = value
		case "ansi_love_motion":
			choices := map[string]syscgoanimations.ANSILoveMotion{"off": syscgoanimations.ANSILoveMotionOff, "subtle": syscgoanimations.ANSILoveMotionSubtle, "full": syscgoanimations.ANSILoveMotionFull}
			cfg.Motion, valid = choices[value]
		default:
			valid = false
		}
		if !valid || !found {
			return defaults, false, fmt.Errorf("%s:%d: invalid ANSI Love setting %q", path, line, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return defaults, false, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, cellOverride, nil
}

func (m model) newANSILoveEffect() *syscgoanimations.ANSILoveAnimation {
	effect, err := buildANSILoveEffect(m.ansiLoveConfig(), m.ansiLoveThemePalette())
	if err != nil {
		logDebug("ANSI Love construction: %v", err)
		return nil
	}
	return effect
}

func (m model) ansiLoveConfig() syscgoanimations.ANSILoveConfig {
	cfg := defaultANSILoveConfig()
	if m.config.ANSILove != nil {
		cfg = *m.config.ANSILove
	}
	cfg.Width, cfg.Height = m.width, m.height
	if cfg.Width <= 0 {
		cfg.Width = 80
	}
	if cfg.Height <= 0 {
		cfg.Height = 30
	}
	cfg.Theme = strings.ReplaceAll(m.currentTheme, " ", "-")
	if !m.config.ANSILoveCellOverride && m.ansiLoveCellW > 0 && m.ansiLoveCellH > 0 {
		cfg.CellW, cfg.CellH = m.ansiLoveCellW, m.ansiLoveCellH
	}
	return cfg
}

func (m model) ansiLoveThemePalette() [16][3]uint8 {
	roles := [16]color.Color{BgBase, Danger, Accent, Warning, Secondary, Accent, Secondary, FgSecondary, FgMuted, Danger, Accent, Warning, Primary, Accent, Primary, FgPrimary}
	var palette [16][3]uint8
	for i, c := range roles {
		r, g, b, _ := c.RGBA()
		palette[i] = [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
	}
	return palette
}

func buildANSILoveEffect(cfg syscgoanimations.ANSILoveConfig, palette [16][3]uint8) (*syscgoanimations.ANSILoveAnimation, error) {
	effect, err := syscgoanimations.NewANSILoveEffectConfig(cfg)
	if err != nil {
		return nil, err
	}
	effect.SetState(syscgoanimations.ANSILoveLogin)
	if cfg.Palette == "theme" {
		effect.SetThemePalette(palette)
	}
	return effect, nil
}

func (m *model) scheduleANSILoveBuild() tea.Cmd {
	cfg := m.ansiLoveConfig()
	palette := m.ansiLoveThemePalette()
	m.ansiLoveBuildRequest++
	request := m.ansiLoveBuildRequest
	m.ansiLoveBuildPending = true
	m.ansiLoveBuildFailed = false
	m.ansiLoveResizeRequest++
	m.ansiLoveResizePending = false
	return func() tea.Msg {
		effect, err := buildANSILoveEffect(cfg, palette)
		return ansiLoveBuiltMsg{effect: effect, request: request, err: err}
	}
}

func (m *model) cancelANSILoveBuild() {
	m.ansiLoveBuildRequest++
	m.ansiLoveBuildPending = false
	m.ansiLoveBuildFailed = false
	m.ansiLoveResizeRequest++
	m.ansiLoveResizePending = false
}

func (m model) addANSILoveEffect(content string) string {
	if m.ansiLoveEffect == nil {
		return content
	}
	return m.ansiLoveEffect.Render()
}

func (m *model) updateANSILove(now time.Time) {
	if m.ansiLoveEffect == nil {
		return
	}
	state := syscgoanimations.ANSILoveIdle
	if m.mode == ModeLogin || m.mode == ModePassword {
		state = syscgoanimations.ANSILoveLogin
	}
	m.ansiLoveEffect.SetState(state)
	if !m.ansiLoveLastTick.IsZero() {
		elapsed := now.Sub(m.ansiLoveLastTick)
		// ponytail: gaps above the host's 250 ms idle heartbeat freeze this frame;
		// an explicit pause/resume event can replace this guard if host scheduling changes.
		if elapsed > 0 && elapsed < 250*time.Millisecond {
			m.ansiLoveEffect.UpdateElapsed(elapsed)
		}
	}
	m.ansiLoveLastTick = now
}

type ansiLovePreparedMsg struct {
	effect   *syscgoanimations.ANSILoveAnimation
	prepared *syscgoanimations.ANSILovePrepared
	request  uint64
}

type ansiLoveBuiltMsg struct {
	effect  *syscgoanimations.ANSILoveAnimation
	request uint64
	err     error
}

func (m *model) prepareANSILoveResize(width, height int) tea.Cmd {
	effect := m.ansiLoveEffect
	if effect == nil {
		return nil
	}
	cfg := m.ansiLoveConfig()
	return m.prepareANSILoveLayout(width, height, cfg.CellW, cfg.CellH)
}

func (m *model) prepareANSILoveLayout(width, height, cellW, cellH int) tea.Cmd {
	effect := m.ansiLoveEffect
	if effect == nil {
		return nil
	}
	cfg := m.ansiLoveConfig()
	var prepare func() *syscgoanimations.ANSILovePrepared
	if cfg.Palette == "theme" {
		prepare = effect.PrepareThemePaletteLayout(cfg.Theme, m.ansiLoveThemePalette(), width, height, cellW, cellH)
	} else {
		prepare = effect.PrepareLayout(width, height, cellW, cellH)
	}
	return m.prepareANSILove(effect, prepare)
}

func (m *model) prepareANSILove(effect *syscgoanimations.ANSILoveAnimation, prepare func() *syscgoanimations.ANSILovePrepared) tea.Cmd {
	m.ansiLoveResizeRequest++
	m.ansiLoveResizePending = true
	request := m.ansiLoveResizeRequest
	return func() tea.Msg { return ansiLovePreparedMsg{effect, prepare(), request} }
}

func (m model) requestANSILoveCellSize() tea.Cmd {
	if m.selectedBackground != "ansi-love" || m.config.ANSILoveCellOverride || m.ansiLoveCellW > 0 {
		return nil
	}
	return tea.Raw("\x1b[16t")
}

// Keep the previous frame within the new viewport while its replacement bakes.
func (m model) ansiLoveCanvas(content string, width, height int) *lipgloss.Canvas {
	canvas := lipgloss.NewCanvas(lipgloss.NewLayer(blankTerminalContent(width, height)))
	frame := expandANSILoveREP(m.addANSILoveEffect(""))
	lines := strings.Split(frame, "\n")
	artWidth := lipgloss.Width(frame)
	var credit string
	if artWidth >= 30 && width >= 30 {
		credit = ansi.Cut(lines[len(lines)-1], artWidth-30, artWidth)
	}
	lines = lines[:len(lines)-1]
	if len(lines) > height-1 {
		lines = lines[:max(0, height-1)]
	}
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "")
	}
	art := strings.Join(lines, "\n")
	canvas.AddLayers(lipgloss.NewLayer(art).X(max(0, (width-artWidth)/2)).Y(max(0, (height-1-len(lines))/2)))
	if content != "" {
		canvas.AddLayers(lipgloss.NewLayer(content).X(max(0, (width-lipgloss.Width(content))/2)).Y(max(0, (height-1-lipgloss.Height(content))/2)))
	}
	if credit != "" {
		canvas.AddLayers(lipgloss.NewLayer(credit).X(width - 30).Y(height - 1))
	}
	return canvas
}

func expandANSILoveREP(frame string) string {
	var out strings.Builder
	out.Grow(len(frame))
	lastRune := rune(0)
	for i := 0; i < len(frame); {
		if i+1 < len(frame) && frame[i] == '\x1b' && frame[i+1] == '[' {
			end := i + 2
			for end < len(frame) && !(frame[end] >= 0x40 && frame[end] <= 0x7e) {
				end++
			}
			if end == len(frame) {
				out.WriteString(frame[i:])
				break
			}
			if frame[end] == 'b' {
				count, err := strconv.Atoi(frame[i+2 : end])
				if err == nil && count > 0 && count <= 512 && lastRune != 0 {
					for range count {
						out.WriteRune(lastRune)
					}
				} else {
					out.WriteString(frame[i : end+1])
				}
			} else {
				out.WriteString(frame[i : end+1])
			}
			i = end + 1
			continue
		}
		r, size := utf8.DecodeRuneInString(frame[i:])
		out.WriteString(frame[i : i+size])
		if r != utf8.RuneError {
			lastRune = r
		}
		i += size
	}
	return out.String()
}
