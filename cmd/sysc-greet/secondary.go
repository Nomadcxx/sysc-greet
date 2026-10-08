package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type secondaryConfig struct {
	Enabled bool
	Exclude string
	Effect  string
}
type secondaryMsg struct {
	Primary string
	Outputs []string
	Err     error
}
type niriOutput struct {
	Logical *struct{} `json:"logical"`
}
type niriWorkspace struct {
	ID     uint64 `json:"id"`
	Output string `json:"output"`
}
type niriWindow struct {
	ID          uint64  `json:"id"`
	AppID       string  `json:"app_id"`
	WorkspaceID *uint64 `json:"workspace_id"`
}

func secondaryPalettes(catalog string) map[string]bool {
	palettes := make(map[string]bool)
	for _, line := range strings.Split(catalog, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "theme" {
			continue
		}
		for _, field := range fields[1:] {
			for _, name := range strings.Split(field, ",") {
				palettes[name] = true
			}
		}
	}
	return palettes
}

func secondaryTargets(outputs map[string]niriOutput, primary, excluded string) []string {
	if outputs[primary].Logical == nil {
		return nil
	}
	skip := map[string]bool{primary: true}
	for _, name := range strings.Split(excluded, ",") {
		skip[strings.TrimSpace(name)] = true
	}
	var targets []string
	for name, output := range outputs {
		if output.Logical != nil && !skip[name] {
			targets = append(targets, name)
		}
	}
	sort.Strings(targets)
	return targets
}

type backgroundChild struct {
	cmd      *exec.Cmd
	done     chan struct{}
	stopOnce sync.Once
	socket   string
	theme    string
}

func startBackgroundChild(cmd *exec.Cmd) (*backgroundChild, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	child := &backgroundChild{cmd: cmd, done: make(chan struct{})}
	go func() { _ = cmd.Wait(); close(child.done) }()
	return child, nil
}

func (c *backgroundChild) stop() {
	c.stopOnce.Do(func() {
		defer func() {
			if c.socket != "" {
				_ = os.Remove(c.socket)
			}
		}()
		_ = syscall.Kill(-c.cmd.Process.Pid, syscall.SIGTERM)
		select {
		case <-c.done:
		case <-time.After(time.Second):
			_ = syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
			<-c.done
		}
	})
}

func niriCommand(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "niri", append([]string{"msg", "--json"}, args...)...).Output()
}

func niriRead(ctx context.Context, value any, args ...string) error {
	data, err := niriCommand(ctx, args...)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func backgroundRequest(ctx context.Context, socket, request string) (string, error) {
	conn, err := (&net.Dialer{Timeout: 200 * time.Millisecond}).DialContext(ctx, "unix", socket)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		return "", err
	}
	if _, err := io.WriteString(conn, request+"\n"); err != nil {
		return "", err
	}
	response, err := bufio.NewReader(io.LimitReader(conn, 4097)).ReadString('\n')
	if len(response) > 4096 {
		return "", fmt.Errorf("background response exceeds 4 KiB")
	}
	return strings.TrimSpace(response), err
}

func launchSecondary(ctx context.Context, output, binary, socket, effect, theme string) (*backgroundChild, error) {
	cmd := exec.Command(binary, "--output", output, "--effect", effect, "--theme", theme, "--ipc-socket", socket)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "GREETD_SOCK=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	child, err := startBackgroundChild(cmd)
	if err != nil {
		return nil, err
	}
	child.socket, child.theme = socket, theme
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		response, err := backgroundRequest(ctx, socket, "query")
		if err == nil && strings.HasPrefix(response, "STATUS:") {
			return child, nil
		}
		select {
		case <-ctx.Done():
			child.stop()
			return nil, fmt.Errorf("background readiness: %w", ctx.Err())
		case <-child.done:
			child.stop()
			return nil, fmt.Errorf("background helper exited before readiness")
		case <-ticker.C:
		}
	}
}

// One goroutine owns output reconciliation and all child lifetimes.
func startSecondaryBackgrounds(ctx context.Context, config secondaryConfig, currentTheme *atomic.Value, send func(tea.Msg)) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if !config.Enabled {
			return
		}
		if os.Getenv("NIRI_SOCKET") == "" {
			return
		}
		binary, err := exec.LookPath("sysc-terminal")
		if err != nil {
			send(secondaryMsg{Err: fmt.Errorf("secondary backgrounds: %w", err)})
			return
		}
		catalogCtx, cancelCatalog := context.WithTimeout(ctx, time.Second)
		catalog, err := exec.CommandContext(catalogCtx, binary, "--list").Output()
		cancelCatalog()
		if err != nil {
			send(secondaryMsg{Err: fmt.Errorf("background palette catalog: %w", err)})
			return
		}
		palettes := secondaryPalettes(string(catalog))
		if !palettes["dracula"] {
			send(secondaryMsg{Err: fmt.Errorf("background palette catalog lacks dracula fallback")})
			return
		}
		directory, err := os.MkdirTemp(os.Getenv("XDG_RUNTIME_DIR"), "sysc-greet-background-")
		if err != nil {
			send(secondaryMsg{Err: err})
			return
		}
		defer os.RemoveAll(directory)
		if config.Effect == "" {
			config.Effect = "matrix"
		}
		if strings.ContainsAny(config.Effect, " \t\r\n") {
			send(secondaryMsg{Err: fmt.Errorf("secondary effect must be one effect id")})
			return
		}
		children := make(map[string]*backgroundChild)
		defer func() {
			for _, child := range children {
				child.stop()
			}
		}()
		var owner uint64
		var socketID uint64
		var lastUnsupported string
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			theme := strings.ReplaceAll(strings.ToLower(currentTheme.Load().(string)), " ", "-")
			if strings.ContainsAny(theme, " \t\r\n") {
				send(secondaryMsg{Err: fmt.Errorf("secondary theme must be one palette id")})
				return
			}
			if !palettes[theme] {
				if theme != lastUnsupported {
					send(secondaryMsg{Err: fmt.Errorf("background palette %q unavailable; using dracula", theme)})
					lastUnsupported = theme
				}
				theme = "dracula"
			} else {
				lastUnsupported = ""
			}
			var outputs map[string]niriOutput
			var windows []niriWindow
			var workspaces []niriWorkspace
			err := niriRead(ctx, &outputs, "outputs")
			if err == nil {
				err = niriRead(ctx, &windows, "windows")
			}
			if err == nil {
				err = niriRead(ctx, &workspaces, "workspaces")
			}
			primary := ""
			if err == nil {
				for _, window := range windows {
					if owner == 0 && window.AppID == "sysc-greet" {
						owner = window.ID
					}
					if window.ID != owner || window.WorkspaceID == nil {
						continue
					}
					for _, workspace := range workspaces {
						if workspace.ID == *window.WorkspaceID {
							primary = workspace.Output
						}
					}
				}
				if primary == "" {
					err = fmt.Errorf("authentication window output unavailable")
				}
			}
			wanted := secondaryTargets(outputs, primary, config.Exclude)
			keep := make(map[string]bool)
			for _, output := range wanted {
				keep[output] = true
			}
			for output, child := range children {
				select {
				case <-child.done:
					child.stop()
					delete(children, output)
				default:
				}
				if !keep[output] {
					child.stop()
					delete(children, output)
				}
			}
			if err == nil {
				for _, output := range wanted {
					if children[output] != nil {
						child := children[output]
						if child.theme != theme {
							response, changeErr := backgroundRequest(ctx, child.socket, "change effect "+config.Effect+" theme "+theme)
							if changeErr != nil {
								err = changeErr
							} else if response != "OK" {
								err = fmt.Errorf("background theme: %s", response)
							} else {
								child.theme = theme
							}
						}
						continue
					}
					socketID++
					child, launchErr := launchSecondary(ctx, output, binary, filepath.Join(directory, fmt.Sprintf("output-%d.sock", socketID)), config.Effect, theme)
					if launchErr != nil {
						err = fmt.Errorf("background %s: %w", output, launchErr)
						continue
					}
					children[output] = child
				}
			}
			active := make([]string, 0, len(children))
			for output := range children {
				active = append(active, output)
			}
			sort.Strings(active)
			send(secondaryMsg{Primary: primary, Outputs: active, Err: err})
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return done
}
