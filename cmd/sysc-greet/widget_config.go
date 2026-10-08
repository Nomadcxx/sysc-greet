package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

// The first existing file wins; launch flags override its settings.
func loadWidgetConfig(paths []string, cli ambientConfig, overrides map[string]bool) (ambientConfig, error) {
	config := cli
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return cli, fmt.Errorf("read widget config %s: %w", path, err)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&config); err != nil {
			return cli, fmt.Errorf("widget config %s: %w", path, err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return cli, fmt.Errorf("widget config %s: expected one JSON object", path)
		}
		break
	}
	if overrides["metrics"] {
		config.Metrics = cli.Metrics
	}
	if overrides["gpu"] {
		config.GPU = cli.GPU
	}
	if overrides["weather-location"] {
		config.WeatherLocation = cli.WeatherLocation
	}
	if overrides["weather-units"] {
		config.WeatherUnits = cli.WeatherUnits
	}
	if overrides["weather-shell-config"] {
		config.ShellConfig = cli.ShellConfig
	}
	if config.WeatherLocation != "sysc-shell" {
		return config, nil
	}
	config.WeatherLocation = "" // Optional shell weather must not prevent authentication.
	path := config.ShellConfig
	if path == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return config, fmt.Errorf("find sysc-shell config: %w", err)
		}
		path = filepath.Join(dir, "sysc-shell", "config.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("read sysc-shell weather config %s: %w", path, err)
	}
	var shell struct {
		Weather struct {
			Latitude  *float64 `json:"latitude"`
			Longitude *float64 `json:"longitude"`
			Unit      string   `json:"unit"`
		} `json:"weather"`
	}
	if err := json.Unmarshal(data, &shell); err != nil {
		return config, fmt.Errorf("sysc-shell config %s: %w", path, err)
	}
	if shell.Weather.Latitude == nil || shell.Weather.Longitude == nil {
		return config, fmt.Errorf("sysc-shell config %s needs weather.latitude and weather.longitude; set weather_location to LAT,LON in the greeter config instead", path)
	}
	if math.IsNaN(*shell.Weather.Latitude) || math.IsInf(*shell.Weather.Latitude, 0) || *shell.Weather.Latitude < -90 || *shell.Weather.Latitude > 90 ||
		math.IsNaN(*shell.Weather.Longitude) || math.IsInf(*shell.Weather.Longitude, 0) || *shell.Weather.Longitude < -180 || *shell.Weather.Longitude > 180 {
		return config, fmt.Errorf("sysc-shell config %s: weather coordinates outside latitude -90..90, longitude -180..180", path)
	}
	if shell.Weather.Unit != "" && shell.Weather.Unit != "celsius" && shell.Weather.Unit != "fahrenheit" {
		return config, fmt.Errorf("sysc-shell config %s: weather unit must be celsius or fahrenheit", path)
	}
	config.WeatherLocation = fmt.Sprintf("%g,%g", *shell.Weather.Latitude, *shell.Weather.Longitude)
	if !overrides["weather-units"] && config.WeatherUnits == "" {
		config.WeatherUnits = shell.Weather.Unit
	}
	return config, nil
}
