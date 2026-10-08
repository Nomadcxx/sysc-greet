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
	if m.config.Ambient.Metrics && width >= 18 {
		cpu, ram := "--%", "--%"
		usage := m.machine.CPU
		if usage.Valid && !math.IsNaN(usage.Fraction) && !math.IsInf(usage.Fraction, 0) && usage.Fraction >= 0 && usage.Fraction <= 1 {
			cpu = fmt.Sprintf("%.0f%%", usage.Fraction*100)
		}
		memory := m.machine.Memory
		if m.machine.MemoryValid && memory.TotalBytes > 0 && memory.UsedBytes <= memory.TotalBytes {
			ram = fmt.Sprintf("%.0f%%", float64(memory.UsedBytes)/float64(memory.TotalBytes)*100)
		}
		parts = append(parts, fmt.Sprintf("CPU %4s  RAM %4s", cpu, ram))
	}
	used := lipgloss.Width(strings.Join(parts, "  "))
	if used > 0 {
		used += 2
	}
	if strings.TrimSpace(m.config.Ambient.WeatherLocation) != "" && width >= used+28 {
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
		parts = append(parts, fmt.Sprintf("Weather %4s°%s %-7s %s", temperature, unit, condition, freshness))
	}
	if len(parts) == 0 {
		return ""
	}
	return lipgloss.NewStyle().Foreground(FgSecondary).Render(strings.Join(parts, "  "))
}

func weatherCondition(code int) string {
	switch code {
	case 0:
		return "Clear"
	case 1, 2:
		return "Partly"
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
