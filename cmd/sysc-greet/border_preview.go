package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"

	"github.com/Nomadcxx/sysc-greet/internal/ansiart"
)

func parseBorderPreviewSizes(value string) ([][2]int, error) {
	var sizes [][2]int
	parts := strings.Split(value, ",")
	if len(parts) > 16 {
		return nil, fmt.Errorf("preview accepts at most 16 sizes")
	}
	for _, part := range parts {
		wText, hText, ok := strings.Cut(strings.TrimSpace(part), "x")
		if !ok {
			return nil, fmt.Errorf("invalid preview size %q; expected WIDTHxHEIGHT", part)
		}
		w, wErr := strconv.Atoi(wText)
		h, hErr := strconv.Atoi(hText)
		if wErr != nil || hErr != nil || w < 1 || w > 512 || h < 1 || h > 256 {
			return nil, fmt.Errorf("invalid preview size %q; use width 1-512 and height 1-256", part)
		}
		sizes = append(sizes, [2]int{w, h})
	}
	if len(sizes) == 0 {
		return nil, fmt.Errorf("no preview sizes specified")
	}
	return sizes, nil
}

func loadBorderPreviewFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open border asset: %w", err)
	}
	data, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	closeErr := f.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read border asset: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close border asset: %w", closeErr)
	}
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("border asset exceeds 1 MiB")
	}
	return data, nil
}

func renderBorderPreview(data []byte, sizes [][2]int) (string, error) {
	border, err := ansiart.LoadBorder(data)
	if err != nil {
		return "", err
	}
	id := border.Name
	if id == "" {
		id = "preview"
	}
	label := border.Label
	if label == "" {
		label = borderLabel(id)
	}
	renderer := &nineSliceRenderer{id: id, label: label, assets: []*ansiart.Border{border}}
	m := model{
		mode:                ModeLogin,
		config:              Config{BorderLayout: "hug"},
		selectedBorderStyle: id,
		borderPaletteMode:   "theme",
	}
	form := lipgloss.JoinVertical(lipgloss.Left,
		"Session:  niri (Wayland)",
		"",
		"Username:  nomadx",
		"",
		"Password:  ********",
	)
	var out strings.Builder
	for _, size := range sizes {
		w, h := size[0], size[1]
		frame, ok := renderer.Frame(&m, w, h, lipgloss.Width(form), lipgloss.Height(form))
		if !ok {
			return "", fmt.Errorf("border does not fit %dx%d", w, h)
		}
		heading := m.renderSessionsHeading(min(100, w-8))
		help := m.renderMainHelpFit(w - 2)
		canvas := lipgloss.NewCanvas(
			lipgloss.NewLayer(clearTerminalLayer{width: w, height: h}).X(0).Y(0),
			lipgloss.NewLayer(nfoGridLayer{
				Grid: frame.Grid, Palette: borderPalette("theme"), PaletteMode: "theme",
				Profile: colorProfile, Left: frame.Left, Top: frame.Top, Right: frame.Right, Bottom: frame.Bottom,
			}).X(frame.X).Y(frame.Y),
			lipgloss.NewLayer(heading).X(max(0, (w-lipgloss.Width(heading))/2)).Y(frame.HeaderY),
			lipgloss.NewLayer(form).X(frame.FormX).Y(frame.FormY),
			lipgloss.NewLayer(help).X(max(0, (w-lipgloss.Width(help))/2)).Y(frame.HelpY),
		)
		fmt.Fprintf(&out, "--- %s %dx%d ---\n%s\n\n", id, w, h, canvas.Render())
	}
	var profiled strings.Builder
	writer := colorprofile.Writer{Forward: &profiled, Profile: colorProfile}
	if _, err := writer.Write([]byte(out.String())); err != nil {
		return "", fmt.Errorf("format border preview: %w", err)
	}
	return profiled.String(), nil
}
