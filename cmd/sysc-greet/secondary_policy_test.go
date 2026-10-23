package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

func TestSecondaryLaunchFailuresBlockAfterLimit(t *testing.T) {
	failures := make(secondaryLaunchFailures)
	fire := secondaryLaunchKey(false, secondaryRuntimeState{Background: "fire", Theme: "nord"})
	for i := 0; i < secondaryLaunchLimit; i++ {
		if failures.blocked("DP-1", fire) {
			t.Fatalf("blocked after %d failures, limit is %d", i, secondaryLaunchLimit)
		}
		failures.record("DP-1", fire)
	}
	if !failures.blocked("DP-1", fire) {
		t.Fatal("output was not blocked after reaching the launch limit")
	}
	if failures.blocked("HDMI-A-1", fire) {
		t.Fatal("a failing output blocked a different output")
	}
	rain := secondaryLaunchKey(false, secondaryRuntimeState{Background: "rain", Theme: "nord"})
	if failures.blocked("DP-1", rain) {
		t.Fatal("a new effect did not get a fresh set of attempts")
	}
	failures.record("DP-1", rain)
	if got := failures["DP-1"].count; got != 1 {
		t.Fatalf("failure count after key change = %d, want 1", got)
	}
	failures.forgetMissing([]secondaryOutput{{Name: "eDP-1", Active: true}})
	if _, ok := failures["DP-1"]; ok {
		t.Fatal("a disconnected output kept its failure history")
	}
}

func TestSecondaryManagerStopsRelaunchingFailingHelper(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	launchLog := filepath.Join(t.TempDir(), "launches.log")
	script := "#!/bin/sh\nif [ \"$1\" = \"--list\" ]; then\n  printf 'effect fire\\ntheme nord,dracula\\n'\n  exit 0\nfi\nprintf 'launch\\n' >> \"$LAUNCH_LOG\"\nexit 1\n"
	if err := os.WriteFile(filepath.Join(binDir, "sysc-terminal"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("LAUNCH_LOG", launchLog)
	manager := &testSecondaryCompositor{
		outputs: []secondaryOutput{{Name: "eDP-1", Active: true}, {Name: "DP-1", Active: true}},
		focused: "eDP-1",
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Background: "fire"})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: true}, state, nil, manager)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
	}()

	launches := func() int {
		data, _ := os.ReadFile(launchLog)
		return strings.Count(string(data), "launch")
	}
	// Three failed attempts take two reconcile intervals; wait one more to prove the fourth never comes.
	deadline := time.Now().Add(time.Duration(secondaryLaunchLimit+1) * 2 * time.Second)
	for time.Now().Before(deadline) && launches() < secondaryLaunchLimit {
		time.Sleep(20 * time.Millisecond)
	}
	if got := launches(); got != secondaryLaunchLimit {
		t.Fatalf("failing helper launched %d times before the deadline, want %d", got, secondaryLaunchLimit)
	}
	time.Sleep(2500 * time.Millisecond)
	if got := launches(); got != secondaryLaunchLimit {
		t.Fatalf("failing helper relaunched after the limit: %d launches", got)
	}

	state.publish(secondaryRuntimeState{Theme: "dracula", Background: "fire"})
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) && launches() == secondaryLaunchLimit {
		time.Sleep(20 * time.Millisecond)
	}
	if got := launches(); got != secondaryLaunchLimit+1 {
		t.Fatalf("theme change did not retry the blocked output: %d launches", got)
	}
}
