package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAmbientConfig(t *testing.T) {
	for _, tc := range []struct {
		location, units string
		valid           bool
	}{
		{"", "celsius", true}, {"0,0", "celsius", true},
		{" -90, 180 ", "fahrenheit", true},
		{"91,0", "celsius", false}, {"0,-181", "celsius", false},
		{"NaN,0", "celsius", false}, {"0,+Inf", "celsius", false},
		{"0", "celsius", false}, {"0,0,0", "celsius", false},
		{"0,0", "kelvin", false},
	} {
		t.Run(tc.location+"/"+tc.units, func(t *testing.T) {
			_, err := newAmbientCollector(context.Background(), ambientConfig{WeatherLocation: tc.location, WeatherUnits: tc.units})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}

func TestAmbientDisabled(t *testing.T) {
	c, err := newAmbientCollector(context.Background(), ambientConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if c.metricsCmd(0) != nil || c.weatherCmd(0) != nil || c.gpuCmd(0) != nil {
		t.Fatal("disabled source scheduled work")
	}
}

func TestAmbientGPU(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := newAmbientCollector(ctx, ambientConfig{GPU: true})
	if err != nil {
		t.Fatal(err)
	}
	result := c.gpuCmd(0)().(gpuMsg)
	if result.Err == nil && result.Snapshot.CollectedAt.IsZero() {
		t.Fatal("GPU snapshot lacks timestamp")
	}
	m := model{ambient: c}
	next, cmd := m.Update(result)
	if cmd == nil || next.(model).gpu.Snapshot.CollectedAt != result.Snapshot.CollectedAt {
		t.Fatal("GPU result not stored/rescheduled")
	}
	cancel()
	if err := c.close(); err != nil {
		t.Fatal(err)
	}
	if c.gpuCmd(0)() != nil {
		t.Fatal("GPU sampling continued after cancellation")
	}
	if err := c.close(); err != nil {
		t.Fatal(err)
	}
}

func TestAmbientMetrics(t *testing.T) {
	c, err := newAmbientCollector(context.Background(), ambientConfig{Metrics: true})
	if err != nil {
		t.Fatal(err)
	}
	first := c.metricsCmd(0)().(metricsMsg)
	if first.CPUErr != nil || first.MemoryErr != nil {
		t.Fatalf("collection: %+v", first)
	}
	if first.CPU.Valid {
		t.Fatal("first CPU sample must be unavailable")
	}
	if !first.MemoryValid || first.Memory.TotalBytes == 0 || first.Memory.UsedBytes > first.Memory.TotalBytes {
		t.Fatalf("invalid RAM: %+v", first)
	}
	time.Sleep(20 * time.Millisecond)
	second := c.metricsCmd(0)().(metricsMsg)
	if !second.CPU.Valid || second.CPU.Fraction < 0 || second.CPU.Fraction > 1 {
		t.Fatalf("invalid second CPU sample: %+v", second)
	}
	m := model{ambient: c}
	next, cmd := m.Update(second)
	if cmd == nil || next.(model).machine != second {
		t.Fatal("metrics result was not stored and rescheduled")
	}
}

func TestWeatherResponse(t *testing.T) {
	valid := `{"current":{"temperature_2m":0,"weather_code":0},"current_units":{"temperature_2m":"°C"}}`
	for _, tc := range []struct {
		name, body string
		status     int
		valid      bool
	}{
		{"zero", valid, 200, true},
		{"missing temperature", `{"current":{"weather_code":0}}`, 200, false},
		{"missing code", `{"current":{"temperature_2m":0}}`, 200, false},
		{"null", `{"current":{"temperature_2m":null,"weather_code":0}}`, 200, false},
		{"unknown code", strings.Replace(valid, `"weather_code":0`, `"weather_code":999`, 1), 200, false},
		{"wrong units", strings.Replace(valid, "°C", "°F", 1), 200, false},
		{"bad JSON", "<html>", 200, false},
		{"trailing JSON", valid + `{}`, 200, false},
		{"oversized", valid + strings.Repeat(" ", 64*1024), 200, false},
		{"server error", valid, 503, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("latitude") != "0" || r.URL.Query().Get("longitude") != "0" || r.URL.Query().Get("temperature_unit") != "celsius" || r.URL.Query().Get("current") != "temperature_2m,weather_code" {
					t.Error("incorrect query", r.URL)
				}
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			result := fetchWeather(context.Background(), srv.Client(), srv.URL, weatherLocation{}, "celsius")
			if result.Valid != tc.valid || (result.Err == nil) != tc.valid {
				t.Fatalf("result: %+v", result)
			}
			if tc.valid && (result.Temperature != 0 || result.Code != 0 || result.CollectedAt.IsZero() || result.Stale) {
				t.Fatalf("zero lost: %+v", result)
			}
		})
	}
}

func TestWeatherStaleRetention(t *testing.T) {
	c, err := newAmbientCollector(context.Background(), ambientConfig{WeatherLocation: "0,0"})
	if err != nil {
		t.Fatal(err)
	}
	fresh := weatherMsg{Temperature: 18, Code: 1, Valid: true, CollectedAt: time.Now()}
	m := model{ambient: c}
	next, cmd := m.Update(fresh)
	if cmd == nil {
		t.Fatal("fresh weather not rescheduled")
	}
	m = next.(model)
	next, cmd = m.Update(weatherMsg{Err: errors.New("offline")})
	m = next.(model)
	if cmd == nil || !m.weather.Valid || !m.weather.Stale || m.weather.Temperature != fresh.Temperature || m.weather.CollectedAt != fresh.CollectedAt || m.weather.Err == nil {
		t.Fatalf("stale weather lost: %+v", m.weather)
	}
	next, _ = m.Update(fresh)
	if next.(model).weather.Stale || next.(model).weather.Err != nil {
		t.Fatal("successful refresh did not clear stale/error")
	}
	next, _ = (model{ambient: c}).Update(weatherMsg{Err: errors.New("offline")})
	if next.(model).weather.Valid {
		t.Fatal("failure without cached weather became valid")
	}
}

func TestAmbientCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := newAmbientCollector(ctx, ambientConfig{Metrics: true, WeatherLocation: "0,0"})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	for _, cmd := range []func() interface{}{
		func() interface{} { return c.metricsCmd(time.Hour)() },
		func() interface{} { return c.weatherCmd(time.Hour)() },
	} {
		done := make(chan interface{}, 1)
		go func() { done <- cmd() }()
		select {
		case result := <-done:
			if result != nil {
				t.Fatal("cancelled command returned a message")
			}
		case <-time.After(time.Second):
			t.Fatal("cancelled timer did not exit")
		}
	}
}

func TestWeatherRequestCancellation(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done() }))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan weatherMsg, 1)
	go func() { done <- fetchWeather(ctx, srv.Client(), srv.URL, weatherLocation{}, "celsius") }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case result := <-done:
		if result.Valid || !errors.Is(result.Err, context.Canceled) {
			t.Fatalf("result: %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not cancel")
	}
}
