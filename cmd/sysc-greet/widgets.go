package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

// Render only stored snapshots. Slot widths depend on enabled sources, not data.
func (m model) renderAmbientRow(width int) string {
	if m.hideAmbient {
		return ""
	}
	var parts []string
	label := lipgloss.NewStyle().Foreground(FgSecondary)
	value := lipgloss.NewStyle().Foreground(FgPrimary).Bold(true)
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(BorderDefault).Padding(0, 1)
	if m.config.Ambient.Metrics && width >= 29 {
		cpu, ram := "--%", "--%"
		usage := m.machine.CPU
		if usage.Valid && !math.IsNaN(usage.Fraction) && !math.IsInf(usage.Fraction, 0) && usage.Fraction >= 0 && usage.Fraction <= 1 {
			cpu = fmt.Sprintf("%.0f%%", usage.Fraction*100)
		}
		memory := m.machine.Memory
		if m.machine.MemoryValid && memory.TotalBytes > 0 && memory.UsedBytes <= memory.TotalBytes {
			ram = fmt.Sprintf("%.0f%%", float64(memory.UsedBytes)/float64(memory.TotalBytes)*100)
		}
		parts = append(parts,
			panel.Render(label.Render("CPU ")+value.Render(fmt.Sprintf("%4s", cpu))),
			panel.Render(label.Render("Memory ")+value.Render(fmt.Sprintf("%4s", ram))))
	}
	used := 0
	if len(parts) > 0 {
		used = 31 // CPU + Memory panels, their gap, and the gap before weather.
	}
	if strings.TrimSpace(m.config.Ambient.WeatherLocation) != "" && width >= used+38 {
		unit := "C"
		if m.config.Ambient.WeatherUnits == "fahrenheit" {
			unit = "F"
		}
		temperature, condition := "--", "--"
		weather := m.weather
		if weather.Valid && !math.IsNaN(weather.Temperature) && !math.IsInf(weather.Temperature, 0) {
			value := strconv.FormatFloat(weather.Temperature, 'f', 0, 64)
			if len(value) <= 4 {
				temperature, condition = value, weatherCondition(weather.Code)
			}
		}
		freshness := "     "
		if weather.Valid && weather.Stale {
			freshness = lipgloss.NewStyle().Foreground(Warning).Render("stale")
		}
		parts = append(parts, panel.Render(label.Render("Weather ")+
			value.Render(fmt.Sprintf("%4s°%s", temperature, unit))+" "+
			label.Render(fmt.Sprintf("%-13s", condition))+" "+freshness))
	}
	if len(parts) == 0 {
		return ""
	}
	row := parts[0]
	for _, part := range parts[1:] {
		row = lipgloss.JoinHorizontal(lipgloss.Top, row, "  ", part)
	}
	return row
}

func weatherCondition(code int) string {
	switch code {
	case 0:
		return "Clear"
	case 1, 2:
		return "Partly cloudy"
	case 3:
		return "Cloudy"
	case 45, 48:
		return "Fog"
	case 51, 53, 55, 56, 57:
		return "Drizzle"
	case 61, 63, 65, 66, 67:
		return "Rain"
	case 71, 73, 75, 77, 85, 86:
		return "Snow"
	case 80, 81, 82:
		return "Showers"
	case 95, 96, 99:
		return "Storm"
	default:
		return "--"
	}
}
