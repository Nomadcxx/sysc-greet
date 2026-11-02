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
	"time"
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
	// Shape of `niri msg --json focused-output` from niri 26.04.
	got, err := parseNiriFocusedOutput([]byte(`{"name":"DP-1","make":"Acer Technologies","model":"Acer X34 P","modes":[{"width":3440,"height":1440,"refresh_rate":59973,"is_preferred":true}],"logical":{"x":0,"y":0,"width":3440,"height":1440}}`))
	if err != nil || got != "DP-1" {
		t.Fatalf("parseNiriFocusedOutput() = %q, %v; want DP-1", got, err)
	}
	if got, err := parseNiriFocusedOutput([]byte("null")); err == nil {
		t.Fatalf("parseNiriFocusedOutput(null) = %q, want error", got)
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
		{name: "niri uses IPC", kind: compositorNiri},
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

func TestNiriOutputManagerUsesIPC(t *testing.T) {
	dir, err := os.MkdirTemp("", "sg-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	socket := filepath.Join(dir, "s")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	t.Setenv("NIRI_SOCKET", socket)
	t.Setenv("PATH", t.TempDir())
	requests := make(chan string, 3)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			scanner := bufio.NewScanner(conn)
			if scanner.Scan() {
				request := scanner.Text()
				requests <- request
				reply := `{"Err":"unexpected request"}`
				switch request {
				case `{"Outputs":null}`:
					reply = `{"Ok":{"Outputs":{"DP-1":{"logical":{"width":1920}},"DP-2":{"logical":{"width":1920}}}}}`
				case `{"FocusedOutput":null}`:
					reply = `{"Ok":{"FocusedOutput":{"name":"DP-1"}}}`
				case `{"Action":{"FocusMonitor":{"output":"DP-2"}}}`:
					reply = `{"Ok":"Handled"}`
				}
				fmt.Fprintln(conn, reply)
			}
			conn.Close()
		}
	}()
	manager := commandCompositor{kind: compositorNiri}
	outputs, focused, err := manager.ListOutputs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if focused != "DP-1" || len(outputs) != 2 {
		t.Fatalf("outputs=%v focus=%s", outputs, focused)
	}
	if err := manager.FocusOutput(context.Background(), "DP-2"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`{"Outputs":null}`, `{"FocusedOutput":null}`, `{"Action":{"FocusMonitor":{"output":"DP-2"}}}`} {
		if got := <-requests; got != want {
			t.Fatalf("request=%s want=%s", got, want)
		}
	}
}

func TestNiriIPCRejectsFailuresAndCancellation(t *testing.T) {
	for _, reply := range []string{`{"Err":"refused"}`, `{"Ok":"Handled"}`, `{"Ok":{"Windows":[]}}`, `{}`, `garbage`} {
		t.Run(reply[:min(len(reply), 16)], func(t *testing.T) {
			dir, err := os.MkdirTemp("", "sg-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			socket := filepath.Join(dir, "s")
			ln, err := net.Listen("unix", socket)
			if err != nil {
				t.Fatal(err)
			}
			defer ln.Close()
			t.Setenv("NIRI_SOCKET", socket)
			go func() {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				bufio.NewReader(conn).ReadString('\n')
				fmt.Fprintln(conn, reply)
			}()
			if _, err := niriSocketRequest(context.Background(), map[string]any{"Outputs": nil}, "Outputs"); err == nil {
				t.Fatal("invalid reply accepted")
			}
		})
	}
	t.Setenv("NIRI_SOCKET", filepath.Join(t.TempDir(), "missing"))
	if _, err := niriSocketRequest(context.Background(), map[string]any{"Outputs": nil}, "Outputs"); err == nil {
		t.Fatal("dial failure accepted")
	}
	dir, err := os.MkdirTemp("", "sg-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	socket := filepath.Join(dir, "s")
	ln, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	t.Setenv("NIRI_SOCKET", socket)
	accepted := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		close(accepted)
		bufio.NewReader(conn).ReadString('\n')
		bufio.NewReader(conn).ReadByte()
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := niriSocketRequest(ctx, map[string]any{"Outputs": nil}, "Outputs"); done <- err }()
	<-accepted
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancel accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not close socket")
	}
}

func TestNiriOutputManagerNamesDetectedCompositor(t *testing.T) {
	manager := commandCompositor{kind: compositorNiri}
	t.Setenv("SYSC_SOCKET", "/run/control.sock")
	t.Setenv("NIRI_SOCKET", "/run/sysc-compositor.wayland-1.7.sock")
	if manager.Name() != "sysc-compositor" {
		t.Fatalf("name=%q", manager.Name())
	}
	t.Setenv("NIRI_SOCKET", "/run/niri.wayland-1.7.sock")
	if manager.Name() != "niri" {
		t.Fatalf("stale control socket changed name=%q", manager.Name())
	}
}
