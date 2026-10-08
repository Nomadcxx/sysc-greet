package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWidgetConfig(t *testing.T) {
	dir := t.TempDir()
	shell := filepath.Join(dir, "shell.json")
	if err := os.WriteFile(shell, []byte(`{"weather":{"latitude":0,"longitude":144.9631,"unit":"fahrenheit"},"other":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "widgets.json")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"metrics":true,"weather_location":"sysc-shell","shell_config":"` + shell + `"}`)
	defaults := ambientConfig{}
	got, err := loadWidgetConfig([]string{filepath.Join(dir, "missing"), path}, defaults, nil)
	if err != nil || !got.Metrics || got.WeatherLocation != "0,144.9631" || got.WeatherUnits != "fahrenheit" {
		t.Fatalf("config=%+v err=%v", got, err)
	}
	cli := ambientConfig{WeatherLocation: "-37,145", WeatherUnits: "celsius"}
	got, err = loadWidgetConfig([]string{path}, cli, map[string]bool{"metrics": true, "weather-location": true, "weather-units": true})
	if err != nil || got.Metrics || got.WeatherLocation != cli.WeatherLocation || got.WeatherUnits != "celsius" {
		t.Fatalf("override=%+v err=%v", got, err)
	}
	write(`{"weather_location":"sysc-shell","shell_config":"` + shell + `"}`)
	got, err = loadWidgetConfig([]string{path}, ambientConfig{WeatherUnits: "celsius"}, map[string]bool{"weather-units": true})
	if err != nil || got.WeatherUnits != "celsius" {
		t.Fatalf("units=%+v err=%v", got, err)
	}
	for _, body := range []string{`{"unknown":true}`, `{`, `{"metrics":true} {}`, `{"weather_location":"sysc-shell","shell_config":"/missing/config.json"}`} {
		write(body)
		if _, err := loadWidgetConfig([]string{path}, defaults, nil); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
	for _, body := range []string{`{}`, `{"weather":{"latitude":0}}`, `{"weather":{"latitude":91,"longitude":0}}`, `{"weather":{"latitude":0,"longitude":0,"unit":"kelvin"}}`} {
		if err := os.WriteFile(shell, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		write(`{"weather_location":"sysc-shell","shell_config":"` + shell + `"}`)
		got, err := loadWidgetConfig([]string{path}, defaults, nil)
		if err == nil {
			collector, e := newAmbientCollector(context.Background(), got)
			err = e
			if collector != nil {
				collector.close()
			}
		}
		if err == nil {
			t.Fatalf("accepted shell %s", body)
		}
	}
	write(`{"weather_location":""}`)
	got, err = loadWidgetConfig([]string{path}, defaults, nil)
	if err != nil || strings.TrimSpace(got.WeatherLocation) != "" {
		t.Fatalf("disabled=%+v %v", got, err)
	}
}

func TestWidgetShellFailureKeepsAuthentication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "widgets.json")
	if err := os.WriteFile(path, []byte(`{"metrics":true,"weather_location":"sysc-shell","shell_config":"/missing/config.json"}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := loadWidgetConfig([]string{path}, ambientConfig{}, nil)
	if err == nil || !got.Metrics || got.WeatherLocation != "" {
		t.Fatalf("fallback=%+v err=%v", got, err)
	}
	collector, err := newAmbientCollector(context.Background(), got)
	if err != nil {
		t.Fatal(err)
	}
	defer collector.close()
	m := layoutModel(80, 24)
	m.config.Ambient = got
	if !strings.Contains(m.renderMainView(80, 24), "Username:") {
		t.Fatal("weather failure removed authentication")
	}
}
