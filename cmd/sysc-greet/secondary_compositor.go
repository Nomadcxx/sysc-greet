package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type secondaryOutput struct {
	Name    string
	Active  bool
	Focused bool
}

type compositorOutputManager interface {
	Name() string
	ListOutputs(context.Context) ([]secondaryOutput, string, error)
	FocusOutput(context.Context, string) error
}

type compositorKind uint8

const (
	compositorNiri compositorKind = iota
	compositorSway
	compositorHyprland
	compositorMango
	compositorCagebreak
)

type commandCompositor struct {
	kind compositorKind
}

func detectOutputManager() (compositorOutputManager, error) {
	for _, candidate := range []struct {
		kind compositorKind
		env  string
	}{
		{compositorNiri, "NIRI_SOCKET"}, // sysc-compositor exposes this Niri-compatible IPC too.
		{compositorCagebreak, "CAGEBREAK_SOCKET"},
		{compositorMango, "MANGO_INSTANCE_SIGNATURE"},
		{compositorHyprland, "HYPRLAND_INSTANCE_SIGNATURE"},
		{compositorSway, "SWAYSOCK"},
	} {
		if os.Getenv(candidate.env) != "" {
			return commandCompositor{kind: candidate.kind}, nil
		}
	}
	return nil, fmt.Errorf("no supported greeter compositor found in the environment")
}

func (c commandCompositor) Name() string {
	switch c.kind {
	case compositorNiri:
		return "niri"
	case compositorSway:
		return "sway"
	case compositorHyprland:
		return "hyprland"
	case compositorMango:
		return "mango"
	case compositorCagebreak:
		return "cagebreak"
	default:
		return "unknown"
	}
}

func (c commandCompositor) ListOutputs(ctx context.Context) ([]secondaryOutput, string, error) {
	var (
		outputs []secondaryOutput
		focused string
		err     error
	)
	switch c.kind {
	case compositorNiri:
		var data []byte
		data, err = runOutputCommand(ctx, "niri", "msg", "--json", "outputs")
		if err == nil {
			outputs, err = parseNiriOutputs(data)
		}
		if err == nil {
			data, err = runOutputCommand(ctx, "niri", "msg", "--json", "focused-output")
		}
		if err == nil {
			focused, err = parseNiriFocusedOutput(data)
		}
	case compositorSway:
		var data []byte
		data, err = runOutputCommand(ctx, "swaymsg", "-t", "get_outputs", "-r")
		if err == nil {
			outputs, err = parseSwayOutputs(data)
			focused = focusedFromOutputs(outputs)
		}
	case compositorHyprland:
		var data []byte
		data, err = runOutputCommand(ctx, "hyprctl", "-j", "monitors")
		if err == nil {
			outputs, err = parseHyprlandOutputs(data)
			focused = focusedFromOutputs(outputs)
		}
	case compositorMango:
		var data []byte
		data, err = runOutputCommand(ctx, "mmsg", "get", "all-monitors")
		if err == nil {
			outputs, err = parseMangoOutputs(data)
			focused = focusedFromOutputs(outputs)
		}
	case compositorCagebreak:
		var data []byte
		data, err = cagebreakSocketRequest(ctx, os.Getenv("CAGEBREAK_SOCKET"), "dump", true)
		if err == nil {
			outputs, focused, err = parseCagebreakDump(data)
		}
	default:
		err = fmt.Errorf("unsupported compositor kind %d", c.kind)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list %s outputs: %w", c.Name(), err)
	}
	if len(outputs) == 0 {
		return nil, "", fmt.Errorf("%s reported no active outputs", c.Name())
	}
	if focused == "" {
		return nil, "", fmt.Errorf("%s did not report a focused output", c.Name())
	}
	return outputs, focused, nil
}

func (c commandCompositor) FocusOutput(ctx context.Context, name string) error {
	if c.kind == compositorCagebreak {
		return c.focusCagebreakOutput(ctx, name)
	}
	command, args := compositorFocusCommand(c.kind, name)
	if command == "" {
		return fmt.Errorf("unsupported compositor kind %d", c.kind)
	}
	_, err := runOutputCommand(ctx, command, args...)
	if err != nil {
		return fmt.Errorf("focus %s output %q: %w", c.Name(), name, err)
	}
	return nil
}

func compositorFocusCommand(kind compositorKind, target string) (string, []string) {
	switch kind {
	case compositorNiri:
		return "niri", []string{"msg", "action", "focus-monitor", target}
	case compositorSway:
		return "swaymsg", []string{"focus", "output", target}
	case compositorHyprland:
		return "hyprctl", []string{"dispatch", "focusmonitor", target}
	case compositorMango:
		return "mmsg", []string{"dispatch", "focusmon," + target}
	default:
		return "", nil
	}
}

func (c commandCompositor) focusCagebreakOutput(ctx context.Context, name string) error {
	socket := os.Getenv("CAGEBREAK_SOCKET")
	outputs, focused, err := c.ListOutputs(ctx)
	if err != nil {
		return err
	}
	if !hasOutput(outputs, name) {
		return fmt.Errorf("cagebreak output %q is not active", name)
	}
	for range outputs {
		if focused == name {
			return nil
		}
		if _, err := cagebreakSocketRequest(ctx, socket, "nextscreen", true); err != nil {
			return fmt.Errorf("focus %s output %q: %w", c.Name(), name, err)
		}
		outputs, focused, err = c.ListOutputs(ctx)
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("cagebreak did not focus output %q after cycling its active screens", name)
}

func hasOutput(outputs []secondaryOutput, name string) bool {
	for _, output := range outputs {
		if output.Active && output.Name == name {
			return true
		}
	}
	return false
}

func runOutputCommand(parent context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, time.Second)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).Output()
}

func parseNiriOutputs(data []byte) ([]secondaryOutput, error) {
	var raw map[string]struct {
		Logical json.RawMessage `json:"logical"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse niri outputs: %w", err)
	}
	outputs := make([]secondaryOutput, 0, len(raw))
	for name, output := range raw {
		if len(output.Logical) == 0 || string(output.Logical) == "null" {
			continue
		}
		outputs = append(outputs, secondaryOutput{Name: name, Active: true})
	}
	sortSecondaryOutputs(outputs)
	return outputs, nil
}

func parseNiriFocusedOutput(data []byte) (string, error) {
	var raw *struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("parse niri focused output: %w", err)
	}
	if raw == nil || raw.Name == "" {
		return "", fmt.Errorf("empty focused output")
	}
	return raw.Name, nil
}

func parseSwayOutputs(data []byte) ([]secondaryOutput, error) {
	var raw []struct {
		Name    string `json:"name"`
		Active  bool   `json:"active"`
		Focused bool   `json:"focused"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse sway outputs: %w", err)
	}
	outputs := make([]secondaryOutput, 0, len(raw))
	for _, output := range raw {
		if output.Active && output.Name != "" {
			outputs = append(outputs, secondaryOutput{Name: output.Name, Active: true, Focused: output.Focused})
		}
	}
	sortSecondaryOutputs(outputs)
	return outputs, nil
}

func parseHyprlandOutputs(data []byte) ([]secondaryOutput, error) {
	var raw []struct {
		Name    string `json:"name"`
		Focused bool   `json:"focused"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse hyprland monitors: %w", err)
	}
	outputs := make([]secondaryOutput, 0, len(raw))
	for _, output := range raw {
		if output.Name != "" {
			outputs = append(outputs, secondaryOutput{Name: output.Name, Active: true, Focused: output.Focused})
		}
	}
	sortSecondaryOutputs(outputs)
	return outputs, nil
}

func parseMangoOutputs(data []byte) ([]secondaryOutput, error) {
	var raw struct {
		Monitors []struct {
			Name    string `json:"name"`
			Focused bool   `json:"active"` // Mango marks its selected monitor with active=true.
		} `json:"monitors"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse mango monitors: %w", err)
	}
	outputs := make([]secondaryOutput, 0, len(raw.Monitors))
	for _, monitor := range raw.Monitors {
		if monitor.Name != "" {
			outputs = append(outputs, secondaryOutput{Name: monitor.Name, Active: true, Focused: monitor.Focused})
		}
	}
	sortSecondaryOutputs(outputs)
	return outputs, nil
}

func parseCagebreakDump(data []byte) ([]secondaryOutput, string, error) {
	start := strings.IndexByte(string(data), '{')
	if start < 0 {
		return nil, "", fmt.Errorf("cagebreak dump has no JSON object")
	}
	end := len(data)
	if nul := strings.IndexByte(string(data[start:]), 0); nul >= 0 {
		end = start + nul
	}
	var raw struct {
		Focused string `json:"curr_output"`
		Outputs map[string]struct {
			Active int `json:"active"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal(data[start:end], &raw); err != nil {
		return nil, "", fmt.Errorf("parse cagebreak dump: %w", err)
	}
	outputs := make([]secondaryOutput, 0, len(raw.Outputs))
	for name, output := range raw.Outputs {
		if output.Active == 0 {
			continue
		}
		outputs = append(outputs, secondaryOutput{Name: name, Active: true, Focused: name == raw.Focused})
	}
	sortSecondaryOutputs(outputs)
	return outputs, raw.Focused, nil
}

func focusedFromOutputs(outputs []secondaryOutput) string {
	for _, output := range outputs {
		if output.Focused {
			return output.Name
		}
	}
	return ""
}

func stablePrimaryOutput(current string, outputs []secondaryOutput, focused string) string {
	for _, output := range outputs {
		if output.Name == current && output.Active {
			return current
		}
	}
	for _, output := range outputs {
		if output.Name == focused && output.Active {
			return focused
		}
	}
	return ""
}

func sortSecondaryOutputs(outputs []secondaryOutput) {
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Name < outputs[j].Name })
}

func secondaryOutputTargets(outputs []secondaryOutput, primary, excluded string) []string {
	if primary == "" {
		return nil
	}
	primaryActive := false
	for _, output := range outputs {
		if output.Name == primary && output.Active {
			primaryActive = true
			break
		}
	}
	if !primaryActive {
		return nil
	}
	skip := map[string]bool{primary: true}
	for _, name := range strings.Split(excluded, ",") {
		skip[strings.TrimSpace(name)] = true
	}
	var targets []string
	for _, output := range outputs {
		if output.Active && output.Name != "" && !skip[output.Name] {
			targets = append(targets, output.Name)
		}
	}
	sort.Strings(targets)
	return targets
}

func cagebreakSocketRequest(ctx context.Context, socket, command string, readResponse bool) ([]byte, error) {
	if socket == "" {
		return nil, fmt.Errorf("CAGEBREAK_SOCKET is not set")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", socket)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(conn, command+"\n"); err != nil {
		return nil, err
	}
	if !readResponse {
		return nil, nil
	}
	data, err := bufio.NewReader(io.LimitReader(conn, 1<<20)).ReadBytes(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("cagebreak returned an empty response to %q", command)
	}
	return data, nil
}
