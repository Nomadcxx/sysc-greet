package main

import (
	"os/exec"
	"testing"
	"time"
)

func TestSecondaryTargets(t *testing.T) {
	outputs := map[string]niriOutput{"DP-1": {Logical: &struct{}{}}, "DP-2": {Logical: &struct{}{}}, "HDMI-A-1": {}}
	got := secondaryTargets(outputs, "DP-1", "")
	if len(got) != 1 || got[0] != "DP-2" {
		t.Fatalf("targets=%v", got)
	}
	if len(secondaryTargets(outputs, "DP-1", " DP-2 , unknown ")) != 0 {
		t.Fatal("excluded output selected")
	}
	if len(secondaryTargets(outputs, "missing", "")) != 0 {
		t.Fatal("backgrounds selected without a primary")
	}
}

func TestSecondaryChildCleanup(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	child, err := startBackgroundChild(cmd)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { child.stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("child not stopped and reaped")
	}
	if cmd.ProcessState == nil {
		t.Fatal("child not reaped")
	}
	child.stop()
}

func TestSecondaryPaletteCatalog(t *testing.T) {
	palettes := secondaryPalettes("theme  dracula \ntheme  tokyo-night tokyonight\neffect matrix 0\n")
	if !palettes["dracula"] || !palettes["tokyo-night"] || !palettes["tokyonight"] || palettes["matrix"] || palettes["blue"] {
		t.Fatalf("palettes=%v", palettes)
	}
}
