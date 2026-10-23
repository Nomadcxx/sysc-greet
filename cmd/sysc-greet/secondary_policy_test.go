package main

import "testing"

func TestSecondaryContentPolicy(t *testing.T) {
	tests := []struct {
		name        string
		background  string
		wallpaper   string
		screensaver bool
		enabled     bool
		override    string
		wantMode    secondaryContentMode
		wantEffect  string
	}{
		{name: "no selected effect keeps wallpaper", enabled: true, wantMode: secondaryContentNone},
		{name: "selected effect mirrors to secondary", background: "fighter-rotation", enabled: true, wantMode: secondaryContentBackground, wantEffect: "fighter-rotation"},
		{name: "video wallpaper suppresses effect", background: "matrix", wallpaper: "slr.mp4", enabled: true, wantMode: secondaryContentNone},
		{name: "screensaver takes precedence over video", background: "matrix", wallpaper: "slr.mp4", screensaver: true, enabled: true, wantMode: secondaryContentScreensaver},
		{name: "explicit legacy effect fills none selection", enabled: true, override: "pour", wantMode: secondaryContentBackground, wantEffect: "pour"},
		{name: "selected effect wins over legacy override", background: "matrix", enabled: true, override: "pour", wantMode: secondaryContentBackground, wantEffect: "matrix"},
		{name: "disabled backgrounds do not stop screensaver", screensaver: true, enabled: false, wantMode: secondaryContentScreensaver},
		{name: "disabled backgrounds keep wallpaper", background: "matrix", enabled: false, wantMode: secondaryContentNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := secondaryContentFor(tt.background, tt.wallpaper, tt.screensaver, tt.enabled, tt.override)
			if got.Mode != tt.wantMode || got.Effect != tt.wantEffect {
				t.Fatalf("secondaryContentFor() = %+v, want mode %v effect %q", got, tt.wantMode, tt.wantEffect)
			}
		})
	}
}
