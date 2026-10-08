package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	metrics "github.com/Nomadcxx/sysc-metrics"
	tea "github.com/charmbracelet/bubbletea/v2"
)

const weatherEndpoint = "https://api.open-meteo.com/v1/forecast"

type ambientConfig struct {
	Metrics         bool   `json:"metrics"`
	GPU             bool   `json:"gpu"`
	WeatherLocation string `json:"weather_location"`
	WeatherUnits    string `json:"weather_units"`
	ShellConfig     string `json:"shell_config"`
}

type weatherLocation struct{ latitude, longitude float64 }

type metricsMsg struct {
	CPU         metrics.CPUUsage
	CPUAt       time.Time
	CPUErr      error
	Memory      metrics.Capacity
	MemoryAt    time.Time
	MemoryValid bool
	MemoryErr   error
}

type weatherMsg struct {
	Temperature float64
	Code        int
	CollectedAt time.Time
	Valid       bool
	Stale       bool
	Err         error
}

type gpuMsg struct {
	Snapshot metrics.GPUSnapshot
	Err      error
}

type ambientCollector struct {
	ctx      context.Context
	config   ambientConfig
	cpu      *metrics.CPUSampler
	location *weatherLocation
	client   *http.Client
	gpu      *metrics.GPUSampler
	gpuMu    sync.Mutex // Close can run while Bubble Tea finishes a sampling command.
}

func newAmbientCollector(ctx context.Context, config ambientConfig) (*ambientCollector, error) {
	if config.WeatherUnits == "" {
		config.WeatherUnits = "celsius"
	}
	if config.WeatherUnits != "celsius" && config.WeatherUnits != "fahrenheit" {
		return nil, fmt.Errorf("weather units must be celsius or fahrenheit")
	}
	c := &ambientCollector{ctx: ctx, config: config}
	if strings.TrimSpace(config.WeatherLocation) != "" {
		parts := strings.Split(config.WeatherLocation, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("weather location must be LAT,LON")
		}
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if latErr != nil || lonErr != nil || math.IsNaN(lat) || math.IsInf(lat, 0) || math.IsNaN(lon) || math.IsInf(lon, 0) || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			return nil, fmt.Errorf("weather coordinates must be finite: latitude -90..90, longitude -180..180")
		}
		c.location = &weatherLocation{lat, lon}
		c.client = &http.Client{Timeout: 2 * time.Second}
	}
	if config.Metrics {
		c.cpu = metrics.NewCPUSampler()
	}
	if config.GPU {
		c.gpu = metrics.NewGPUSampler()
	}
	return c, nil
}

func (c *ambientCollector) close() error {
	if c.client != nil {
		c.client.CloseIdleConnections()
	}
	c.gpuMu.Lock()
	defer c.gpuMu.Unlock()
	if c.gpu != nil {
		return c.gpu.Close()
	}
	return nil
}

func (c *ambientCollector) gpuCmd(delay time.Duration) tea.Cmd {
	if c == nil || !c.config.GPU {
		return nil
	}
	return func() tea.Msg {
		if !c.wait(delay) {
			return nil
		}
		c.gpuMu.Lock()
		defer c.gpuMu.Unlock()
		if c.ctx.Err() != nil {
			return nil
		}
		snapshot, err := c.gpu.Sample()
		if c.ctx.Err() != nil {
			return nil
		}
		return gpuMsg{Snapshot: snapshot, Err: err}
	}
}

func (c *ambientCollector) wait(delay time.Duration) bool {
	if delay == 0 {
		return c.ctx.Err() == nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-c.ctx.Done():
		return false
	case <-timer.C:
		return c.ctx.Err() == nil
	}
}

// One result schedules one successor in Update; the CPU sampler has one caller.
func (c *ambientCollector) metricsCmd(delay time.Duration) tea.Cmd {
	if c == nil || !c.config.Metrics {
		return nil
	}
	return func() tea.Msg {
		if !c.wait(delay) {
			return nil
		}
		cpu, cpuErr := c.cpu.Sample()
		memory, memoryErr := metrics.ReadMemory()
		if c.ctx.Err() != nil {
			return nil
		}
		return metricsMsg{CPU: cpu.Usage, CPUAt: cpu.CollectedAt, CPUErr: cpuErr, Memory: memory.Memory, MemoryAt: memory.CollectedAt, MemoryValid: memoryErr == nil, MemoryErr: memoryErr}
	}
}

func (c *ambientCollector) weatherCmd(delay time.Duration) tea.Cmd {
	if c == nil || c.location == nil {
		return nil
	}
	return func() tea.Msg {
		if !c.wait(delay) {
			return nil
		}
		result := fetchWeather(c.ctx, c.client, weatherEndpoint, *c.location, c.config.WeatherUnits)
		if c.ctx.Err() != nil {
			return nil
		}
		return result
	}
}

func fetchWeather(ctx context.Context, client *http.Client, endpoint string, location weatherLocation, units string) weatherMsg {
	query := url.Values{
		"latitude":         {strconv.FormatFloat(location.latitude, 'f', -1, 64)},
		"longitude":        {strconv.FormatFloat(location.longitude, 'f', -1, 64)},
		"current":          {"temperature_2m,weather_code"},
		"temperature_unit": {units},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return weatherMsg{Err: fmt.Errorf("weather request: %w", err)}
	}
	resp, err := client.Do(req)
	if err != nil {
		return weatherMsg{Err: fmt.Errorf("weather request: %w", err)}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return weatherMsg{Err: fmt.Errorf("weather HTTP status %d", resp.StatusCode)}
	}
	const maxBody = 64 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return weatherMsg{Err: fmt.Errorf("weather response: %w", err)}
	}
	if len(body) > maxBody {
		return weatherMsg{Err: fmt.Errorf("weather response exceeds 64 KiB")}
	}
	var payload struct {
		Current struct {
			Temperature *float64 `json:"temperature_2m"`
			Code        *int     `json:"weather_code"`
		} `json:"current"`
		Units struct {
			Temperature string `json:"temperature_2m"`
		} `json:"current_units"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return weatherMsg{Err: fmt.Errorf("weather JSON: %w", err)}
	}
	expectedUnit := "°C"
	if units == "fahrenheit" {
		expectedUnit = "°F"
	}
	if payload.Current.Temperature == nil || payload.Current.Code == nil || payload.Units.Temperature != expectedUnit {
		return weatherMsg{Err: fmt.Errorf("weather response missing temperature/code or incorrect units")}
	}
	temperature := *payload.Current.Temperature
	if math.IsNaN(temperature) || math.IsInf(temperature, 0) {
		return weatherMsg{Err: fmt.Errorf("weather temperature is not finite")}
	}
	switch *payload.Current.Code {
	case 0, 1, 2, 3, 45, 48, 51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 71, 73, 75, 77, 80, 81, 82, 85, 86, 95, 96, 99:
	default:
		return weatherMsg{Err: fmt.Errorf("weather response has unknown WMO code")}
	}
	// ponytail: cache only for this greeter process; persist later if restart reuse is needed.
	return weatherMsg{Temperature: temperature, Code: *payload.Current.Code, CollectedAt: time.Now(), Valid: true}
}
