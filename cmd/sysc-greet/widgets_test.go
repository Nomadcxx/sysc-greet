package main

import (
	"fmt"
	"math"
	"strings"
	"testing"

	metrics "github.com/Nomadcxx/sysc-metrics"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestWidgetMetrics(t *testing.T) {
	m := layoutModel(80, 24)
	if m.renderAmbientRow(72) != "" {
		t.Fatal("disabled widgets rendered")
	}
	m.config.Ambient.Metrics = true
	missing := ansi.Strip(m.renderAmbientRow(72))
	if !strings.Contains(missing, "CPU  --%") || !strings.Contains(missing, "Memory  --%") {
		t.Fatalf("unavailable=%q", missing)
	}
	m.machine = metricsMsg{CPU: metrics.CPUUsage{Valid: true, Fraction: 0}, Memory: metrics.Capacity{UsedBytes: 0, TotalBytes: 100}, MemoryValid: true}
	zero := ansi.Strip(m.renderAmbientRow(72))
	if !strings.Contains(zero, "CPU   0%") || !strings.Contains(zero, "Memory   0%") {
		t.Fatalf("zero=%q", zero)
	}
	if lipgloss.Width(missing) != lipgloss.Width(zero) {
		t.Fatal("values changed row width")
	}
	for _, fraction := range []float64{math.NaN(), math.Inf(1), -1, 1.1} {
		m.machine.CPU.Fraction = fraction
		if !strings.Contains(ansi.Strip(m.renderAmbientRow(72)), "CPU  --%") {
			t.Fatalf("invalid fraction %v displayed", fraction)
		}
	}
	m.machine.Memory = metrics.Capacity{UsedBytes: 1, TotalBytes: 0}
	if !strings.Contains(ansi.Strip(m.renderAmbientRow(72)), "Memory  --%") {
		t.Fatal("invalid memory displayed")
	}
	m.machine.CPU.Fraction = 1
	m.machine.Memory = metrics.Capacity{UsedBytes: math.MaxUint64, TotalBytes: math.MaxUint64}
	full := ansi.Strip(m.renderAmbientRow(72))
	if !strings.Contains(full, "CPU 100%") || !strings.Contains(full, "Memory 100%") {
		t.Fatalf("full=%q", full)
	}
}

func TestWidgetWeather(t *testing.T) {
	m := layoutModel(80, 24)
	m.config.Ambient.WeatherLocation = "0,0"
	missing := m.renderAmbientRow(72)
	if !strings.Contains(ansi.Strip(missing), "--°C") {
		t.Fatal("missing weather not marked unavailable")
	}
	for _, units := range []string{"celsius", "fahrenheit"} {
		m.config.Ambient.WeatherUnits = units
		m.weather = weatherMsg{Valid: true, Temperature: -12, Code: 3}
		row := m.renderAmbientRow(72)
		suffix := "°C"
		if units == "fahrenheit" {
			suffix = "°F"
		}
		if !strings.Contains(ansi.Strip(row), "-12"+suffix) || !strings.Contains(ansi.Strip(row), "Cloudy") {
			t.Fatalf("weather=%q", ansi.Strip(row))
		}
		m.weather.Stale = true
		stale := m.renderAmbientRow(72)
		if !strings.Contains(ansi.Strip(stale), "stale") || lipgloss.Width(stale) != lipgloss.Width(row) || lipgloss.Height(stale) != 3 {
			t.Fatal("stale weather shifted or lost label")
		}
	}
	for _, temperature := range []float64{math.NaN(), math.Inf(1), 1e100} {
		m.weather.Temperature = temperature
		if lipgloss.Width(m.renderAmbientRow(72)) > 72 || !strings.Contains(ansi.Strip(m.renderAmbientRow(72)), "--°F") {
			t.Fatal("unbounded weather value rendered")
		}
	}
	m.config.Ambient.Metrics = true
	if row := ansi.Strip(m.renderAmbientRow(34)); !strings.Contains(row, "CPU") || strings.Contains(row, "Weather") {
		t.Fatalf("narrow row=%q", row)
	}
	if m.renderAmbientRow(10) != "" {
		t.Fatal("row did not yield at very narrow width")
	}
}

func TestWidgetPanelStable(t *testing.T) {
	for _, border := range []string{"classic", "modern", "minimal", "ascii1", "ascii2", "ascii3", "ascii4"} {
		for _, size := range [][2]int{{40, 16}, {80, 24}, {120, 32}, {240, 100}} {
			t.Run(fmt.Sprintf("%s/%dx%d", border, size[0], size[1]), func(t *testing.T) {
				m := layoutModel(size[0], size[1])
				m.selectedBorderStyle = border
				m.config.Ambient = ambientConfig{Metrics: true, WeatherLocation: "0,0"}
				before := m.renderMainView(m.width, m.height)
				m.machine = metricsMsg{CPU: metrics.CPUUsage{Valid: true, Fraction: .42}, Memory: metrics.Capacity{UsedBytes: 35, TotalBytes: 100}, MemoryValid: true}
				m.weather = weatherMsg{Valid: true, Temperature: 14, Code: 0}
				after := m.renderMainView(m.width, m.height)
				if strings.Count(ansi.Strip(before[:strings.Index(before, "Username:")]), "\n") != strings.Count(ansi.Strip(after[:strings.Index(after, "Username:")]), "\n") {
					t.Fatal("snapshot update moved input row")
				}
				if lipgloss.Width(before) != lipgloss.Width(after) || lipgloss.Height(before) != lipgloss.Height(after) {
					t.Fatal("snapshot update moved panel")
				}
				if lipgloss.Width(after) > m.width || lipgloss.Height(after) > m.height {
					t.Fatal("widgets overflow panel")
				}
				plain := ansi.Strip(after)
				if !strings.Contains(plain, "CPU  42%") || !strings.Contains(plain, "Memory  35%") || !strings.Contains(plain, "Username:") {
					t.Fatalf("missing form/status in %q", plain)
				}
				form := m.renderMainForm(72)
				if !strings.Contains(ansi.Strip(form), "CPU") {
					t.Fatal("status outside shared login form")
				}
			})
		}
	}
}

func TestWidgetYieldsForAuthentication(t *testing.T) {
	m := layoutModel(40, 7)
	m.mode = ModePassword
	m.config.Ambient = ambientConfig{Metrics: true, WeatherLocation: "0,0"}
	view := m.renderMainView(m.width, m.height)
	if lipgloss.Width(view) > m.width || lipgloss.Height(view) > m.height {
		t.Fatal("status displaced authentication")
	}
	plain := ansi.Strip(view)
	if strings.Contains(plain, "CPU") || !strings.Contains(plain, "Password:") || !strings.Contains(plain, "Enter Login") {
		t.Fatalf("priority=%q", plain)
	}
	if !m.config.Ambient.Metrics || m.config.Ambient.WeatherLocation != "0,0" {
		t.Fatal("rendering disabled collection in model")
	}
}

func TestWidgetWeatherConditions(t *testing.T) {
	for code, want := range map[int]string{0: "Clear", 1: "Partly cloudy", 2: "Partly cloudy", 3: "Cloudy", 45: "Fog", 48: "Fog", 51: "Drizzle", 57: "Drizzle", 61: "Rain", 67: "Rain", 71: "Snow", 77: "Snow", 80: "Showers", 82: "Showers", 85: "Snow", 86: "Snow", 95: "Storm", 99: "Storm", -1: "--", 100: "--"} {
		if got := weatherCondition(code); got != want || len(got) > 13 {
			t.Fatalf("code %d: %q, want %q", code, got, want)
		}
	}
}

func TestWidgetUpdatesPreserveFocus(t *testing.T) {
	m := layoutModel(80, 24)
	m.config.Ambient = ambientConfig{Metrics: true, WeatherLocation: "0,0"}
	before := m.renderMainView(m.width, m.height)
	next, _ := m.Update(metricsMsg{CPU: metrics.CPUUsage{Valid: true, Fraction: .5}, Memory: metrics.Capacity{UsedBytes: 50, TotalBytes: 100}, MemoryValid: true})
	m = next.(model)
	next, _ = m.Update(weatherMsg{Valid: true, Temperature: 0, Code: 0})
	m = next.(model)
	next, _ = m.Update(weatherMsg{Err: fmt.Errorf("offline")})
	m = next.(model)
	if !m.weather.Stale || m.focusState != FocusUsername || !m.usernameInput.Focused() || m.usernameInput.Value() != "nomadx" || m.hideAmbient {
		t.Fatal("status updates changed focus or input")
	}
	after := m.renderMainView(m.width, m.height)
	if lipgloss.Width(before) != lipgloss.Width(after) || lipgloss.Height(before) != lipgloss.Height(after) {
		t.Fatal("async status changed panel size")
	}
}

func TestWidgetComposition(t *testing.T) {
	m := layoutModel(120, 40)
	m.config.Ambient = ambientConfig{Metrics: true, WeatherLocation: "0,0"}
	m.weather = weatherMsg{Valid: true, Temperature: 16, Code: 2}
	view := ansi.Strip(m.renderAmbientRow(72))
	lines := strings.Split(view, "\n")
	if len(lines) != 3 || strings.Count(lines[0], "╭") != 3 || strings.Count(lines[2], "╰") != 3 {
		t.Fatalf("expected three lightly bordered panels: %q", view)
	}
	for _, label := range []string{"CPU", "Memory", "Weather", "Partly cloudy"} {
		if !strings.Contains(lines[1], label) {
			t.Errorf("missing label %q: %q", label, lines[1])
		}
	}
	if strings.Count(lines[0], "╮  ╭") != 2 || strings.Count(lines[2], "╯  ╰") != 2 {
		t.Fatalf("panels need two-cell gaps: %q", view)
	}
	if lipgloss.Width(view) > 72 {
		t.Fatal("panels exceed their available width")
	}
}
