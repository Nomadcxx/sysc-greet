package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseNiriOutputs(t *testing.T) {
	got, err := parseNiriOutputs([]byte(`{"DP-1":{"logical":{"width":1920}},"HDMI-A-1":{"logical":null}}`))
	if err != nil {
		t.Fatal(err)
	}
	if want := []secondaryOutput{{Name: "DP-1", Active: true}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseNiriOutputs() = %#v, want %#v", got, want)
	}
}

func TestParseNiriFocusedOutput(t *testing.T) {
	got, err := parseNiriFocusedOutput([]byte(`Output "DP-1" (DP-1)`))
	if err != nil || got != "DP-1" {
		t.Fatalf("parseNiriFocusedOutput() = %q, %v; want DP-1", got, err)
	}
}

func TestParseSwayOutputs(t *testing.T) {
	got, err := parseSwayOutputs([]byte(`[{"name":"eDP-1","active":true,"focused":true},{"name":"DP-2","active":false,"focused":false}]`))
	if err != nil {
		t.Fatal(err)
	}
	if want := []secondaryOutput{{Name: "eDP-1", Active: true, Focused: true}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSwayOutputs() = %#v, want %#v", got, want)
	}
}

func TestParseHyprlandOutputs(t *testing.T) {
	got, err := parseHyprlandOutputs([]byte(`[{"name":"DP-1","focused":false},{"name":"eDP-1","focused":true}]`))
	if err != nil {
		t.Fatal(err)
	}
	if want := []secondaryOutput{{Name: "DP-1", Active: true}, {Name: "eDP-1", Active: true, Focused: true}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseHyprlandOutputs() = %#v, want %#v", got, want)
	}
}

func TestParseMangoOutputs(t *testing.T) {
	got, err := parseMangoOutputs([]byte(`{"monitors":[{"name":"eDP-1","active":true},{"name":"DP-1","active":false}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if want := []secondaryOutput{{Name: "DP-1", Active: true}, {Name: "eDP-1", Active: true, Focused: true}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseMangoOutputs() = %#v, want %#v", got, want)
	}
}

func TestParseCagebreakDump(t *testing.T) {
	data := []byte("cg-ipc{\"event_name\":\"dump\",\"curr_output\":\"eDP-1\",\"outputs\":{\"eDP-1\":{\"active\":1,\"curr_workspace\":0,\"workspaces\":[{\"tiles\":[{\"id\":4,\"view\":9}]}]},\"DP-1\":{\"active\":1,\"curr_workspace\":0,\"workspaces\":[{\"tiles\":[{\"id\":8,\"view\":0}]}]}}}\x00")
	got, focused, err := parseCagebreakDump(data)
	if err != nil {
		t.Fatal(err)
	}
	if focused != "eDP-1" {
		t.Fatalf("focused output = %q, want eDP-1", focused)
	}
	if want := []secondaryOutput{{Name: "DP-1", Active: true}, {Name: "eDP-1", Active: true, Focused: true}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCagebreakDump() = %#v, want %#v", got, want)
	}
}

func TestCompositorFocusCommand(t *testing.T) {
	tests := []struct {
		name string
		kind compositorKind
		out  string
		args []string
	}{
		{name: "niri", kind: compositorNiri, out: "niri", args: []string{"msg", "action", "focus-monitor", "DP-2"}},
		{name: "sway", kind: compositorSway, out: "swaymsg", args: []string{"focus", "output", "DP-2"}},
		{name: "hyprland", kind: compositorHyprland, out: "hyprctl", args: []string{"dispatch", "focusmonitor", "DP-2"}},
		{name: "mango", kind: compositorMango, out: "mmsg", args: []string{"dispatch", "focusmon,DP-2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, args := compositorFocusCommand(tt.kind, "DP-2")
			if tt.kind == compositorCagebreak {
				got, args = compositorFocusCommand(tt.kind, "8")
			}
			if got != tt.out || !reflect.DeepEqual(args, tt.args) {
				t.Fatalf("compositorFocusCommand() = %q %v, want %q %v", got, args, tt.out, tt.args)
			}
		})
	}
}

func TestCagebreakFocusOutputSupportsEmptyWorkspace(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "cagebreak.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	commands := make(chan string, 8)
	serverDone := make(chan struct{})
	focused := "eDP-1"
	go func() {
		defer close(serverDone)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			command, _ := bufio.NewReader(conn).ReadString('\n')
			command = strings.TrimSpace(command)
			commands <- command
			var response string
			switch command {
			case "dump":
				response = fmt.Sprintf("cg-ipc{\"event_name\":\"dump\",\"curr_output\":%q,\"outputs\":{\"eDP-1\":{\"active\":1,\"curr_workspace\":0,\"workspaces\":[{\"tiles\":[]}]},\"DP-1\":{\"active\":1,\"curr_workspace\":0,\"workspaces\":[{\"tiles\":[]}]}}}\x00", focused)
			case "nextscreen":
				focused = "DP-1"
				response = "cg-ipc{\"event_name\":\"cycle_outputs\",\"new_output\":\"DP-1\"}\x00"
			}
			_, _ = io.WriteString(conn, response)
			_ = conn.Close()
		}
	}()
	t.Setenv("CAGEBREAK_SOCKET", socket)
	if err := (commandCompositor{kind: compositorCagebreak}).FocusOutput(context.Background(), "DP-1"); err != nil {
		t.Fatal(err)
	}
	closeCommands := func() []string {
		_ = listener.Close()
		<-serverDone
		close(commands)
		var got []string
		for command := range commands {
			got = append(got, command)
		}
		return got
	}
	if got, want := closeCommands(), []string{"dump", "nextscreen", "dump"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Cagebreak focus requests = %v, want %v", got, want)
	}
}

func TestHyprlandSaverFullscreenUsesCurrentRuleSyntax(t *testing.T) {
	binDir := t.TempDir()
	argsFile := filepath.Join(t.TempDir(), "hyprctl-args")
	hyprctl := filepath.Join(binDir, "hyprctl")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$HYPR_TEST_ARGS\"\n"
	if err := os.WriteFile(hyprctl, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HYPR_TEST_ARGS", argsFile)
	if err := (commandCompositor{kind: compositorHyprland}).prepareSaverFullscreen(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	want := "keyword\nwindowrule\nmatch:class sysc-greet-screensaver, fullscreen on\n"
	if string(got) != want {
		t.Fatalf("Hyprland fullscreen rule args = %q, want %q", got, want)
	}
}

func TestSecondaryTargetsAcrossCompositors(t *testing.T) {
	outputs := []secondaryOutput{
		{Name: "DP-2", Active: true},
		{Name: "eDP-1", Active: true},
		{Name: "HDMI-A-1", Active: false},
	}
	if got, want := secondaryOutputTargets(outputs, "eDP-1", ""), []string{"DP-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("secondaryTargets() = %v, want %v", got, want)
	}
	if got := secondaryOutputTargets(outputs, "eDP-1", " DP-2 , unknown "); len(got) != 0 {
		t.Fatalf("excluded targets = %v, want none", got)
	}
	if got := secondaryOutputTargets(outputs, "missing", ""); len(got) != 0 {
		t.Fatalf("targets without a primary = %v, want none", got)
	}
}

func TestStablePrimaryOutputSurvivesTemporaryFocusChanges(t *testing.T) {
	outputs := []secondaryOutput{
		{Name: "DP-2", Active: true, Focused: true},
		{Name: "eDP-1", Active: true},
	}
	if got := stablePrimaryOutput("eDP-1", outputs, "DP-2"); got != "eDP-1" {
		t.Fatalf("stablePrimaryOutput() followed temporary focus to %q", got)
	}
	if got := stablePrimaryOutput("missing", outputs, "DP-2"); got != "DP-2" {
		t.Fatalf("stablePrimaryOutput() = %q after primary removal, want DP-2", got)
	}
}
