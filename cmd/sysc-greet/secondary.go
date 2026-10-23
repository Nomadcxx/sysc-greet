package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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

type secondaryWakeMsg struct{}

type secondaryChildKind uint8

const (
	secondaryChildBackground secondaryChildKind = iota
	secondaryChildScreensaver
)

const secondarySaverWindowClass = "sysc-greet-screensaver"

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

func secondaryEffectCatalog(catalog string) map[string]bool {
	effects := make(map[string]bool)
	for _, line := range strings.Split(catalog, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "effect" {
			effects[fields[1]] = true
		}
	}
	return effects
}

func secondaryEffectID(background string, effects map[string]bool) (string, bool) {
	if background == "ascii-rain" {
		background = "rain"
	}
	if !effects[background] {
		return "", false
	}
	return background, true
}

type backgroundChild struct {
	cmd           *exec.Cmd
	done          chan struct{}
	stopOnce      sync.Once
	stopRequested bool
	exitErr       error
	kind          secondaryChildKind
	socket        string
	readyFile     string
	ready         bool
	effect        string
	theme         string
	launchKey     string
	started       time.Time
}

func startBackgroundChild(cmd *exec.Cmd) (*backgroundChild, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	child := &backgroundChild{cmd: cmd, done: make(chan struct{}), started: time.Now()}
	go func() {
		child.exitErr = cmd.Wait()
		close(child.done)
	}()
	return child, nil
}

func (c *backgroundChild) stop() {
	c.stopOnce.Do(func() {
		c.stopRequested = true
		defer func() {
			if c.socket != "" {
				_ = os.Remove(c.socket)
			}
			if c.readyFile != "" {
				_ = os.Remove(c.readyFile)
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

func launchSecondaryBackground(ctx context.Context, stateBus *secondaryStateBus, expected secondaryRuntimeState, output, binary, socket, effect, theme string) (*backgroundChild, error) {
	if stateBus.load() != expected {
		return nil, context.Canceled
	}
	cmd := exec.Command(binary, "--output", output, "--effect", effect, "--theme", theme, "--ipc-socket", socket)
	cmd.Env = secondaryChildEnvironment()
	child, err := startBackgroundChild(cmd)
	if err != nil {
		return nil, err
	}
	child.kind, child.socket, child.effect, child.theme = secondaryChildBackground, socket, effect, theme
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		response, err := backgroundRequest(ctx, socket, "query")
		if err == nil && strings.HasPrefix(response, "STATUS:") {
			child.ready = true
			return child, nil
		}
		select {
		case <-ctx.Done():
			child.stop()
			return nil, fmt.Errorf("background readiness: %w", ctx.Err())
		case <-stateBus.changed:
			if stateBus.load() != expected {
				child.stop()
				return nil, context.Canceled
			}
		case <-child.done:
			child.stop()
			return nil, fmt.Errorf("background helper exited before readiness")
		case <-ticker.C:
		}
	}
}

func secondaryChildEnvironment() []string {
	var env []string
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		switch name {
		case "GREETD_SOCK", "KITTY_WINDOW_ID", "KITTY_LISTEN_ON", "KITTY_PID", "KITTY_PUBLIC_KEY":
			continue
		}
		env = append(env, entry)
	}
	return env
}

func launchSecondarySaver(ctx context.Context, stateBus *secondaryStateBus, expected secondaryRuntimeState, output, kitty, executable, configFile, theme, startedAt, readyFile string) (*backgroundChild, error) {
	cmd := exec.Command(kitty, secondarySaverCommandArgs(configFile, executable, theme, startedAt, readyFile)...)
	cmd.Env = secondaryChildEnvironment()
	child, err := startBackgroundChild(cmd)
	if err != nil {
		return nil, err
	}
	child.kind, child.readyFile, child.theme = secondaryChildScreensaver, readyFile, theme
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if info, err := os.Stat(readyFile); err == nil && info.Mode().IsRegular() {
			child.ready = true
			_ = os.Remove(readyFile)
			child.readyFile = ""
			return child, nil
		}
		select {
		case <-ctx.Done():
			child.stop()
			return nil, fmt.Errorf("screensaver readiness on %s: %w", output, ctx.Err())
		case <-stateBus.changed:
			if stateBus.load() != expected {
				child.stop()
				return nil, context.Canceled
			}
		case <-child.done:
			child.stop()
			return nil, fmt.Errorf("screensaver on %s exited before readiness", output)
		case <-ticker.C:
		}
	}
}

func secondaryKittyConfig() string {
	const configFile = "/etc/greetd/kitty.conf"
	if info, err := os.Stat(configFile); err == nil && info.Mode().IsRegular() {
		return configFile
	}
	return ""
}

func startSecondaryManager(ctx context.Context, config secondaryConfig, state *secondaryStateBus, send func(tea.Msg)) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		manager, err := detectOutputManager()
		if err != nil {
			sendSecondary(send, secondaryMsg{Err: err})
			return
		}
		runSecondaryManager(ctx, config, state, send, manager)
	}()
	return done
}

func runSecondaryManager(ctx context.Context, config secondaryConfig, stateBus *secondaryStateBus, send func(tea.Msg), manager compositorOutputManager) {
	dir, err := os.MkdirTemp(os.Getenv("XDG_RUNTIME_DIR"), "sysc-greet-secondary-")
	if err != nil {
		sendSecondary(send, secondaryMsg{Err: fmt.Errorf("secondary runtime directory: %w", err)})
		return
	}
	defer os.RemoveAll(dir)
	children := make(map[string]*backgroundChild)
	defer func() {
		for _, child := range children {
			child.stop()
		}
	}()
	failures := make(secondaryLaunchFailures)

	var (
		primary             string
		wakePending         bool
		focusRestorePending bool
		catalogLoaded       bool
		effects             map[string]bool
		palettes            map[string]bool
		terminalBinary      string
		kittyBinary         string
		greeterBinary       string
		lastStatus          secondaryMsg
		haveStatus          bool
		fullscreenReady     bool
	)
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		runtime := stateBus.load()
		if !runtime.Screensaver {
			wakePending = false
		}
		var childErr error
		wakePending, childErr = reconcileSecondaryChildren(children, runtime.Screensaver, wakePending, send, failures)
		outputs, focused, listErr := manager.ListOutputs(ctx)
		if listErr == nil {
			failures.forgetMissing(outputs)
			primary = stablePrimaryOutput(primary, outputs, focused)
			if primary == "" {
				listErr = fmt.Errorf("no stable primary output is active")
			}
		}
		if listErr != nil {
			var restoreErr error
			focusRestorePending, restoreErr = restorePrimaryFocus(ctx, manager, primary, focusRestorePending)
			publishSecondaryStatus(send, &lastStatus, &haveStatus, secondaryMsg{Primary: primary, Err: errors.Join(listErr, childErr, restoreErr)})
		} else {
			desired := make(map[string]secondaryRuntimeState)
			iterationErr := childErr
			focusAttempted := false
			if runtime.Screensaver {
				if !wakePending {
					if runtime.SaverStart.IsZero() {
						iterationErr = fmt.Errorf("screensaver activation time is missing")
					} else {
						for _, output := range secondaryOutputTargets(outputs, primary, "") {
							desired[output] = runtime
						}
					}
				}
				if len(desired) > 0 && kittyBinary == "" {
					kittyBinary, iterationErr = exec.LookPath("kitty")
					if iterationErr == nil {
						greeterBinary, iterationErr = os.Executable()
					}
				}
				if iterationErr != nil {
					desired = make(map[string]secondaryRuntimeState)
				}
			} else {
				content := secondaryContentFor(runtime.Background, runtime.Wallpaper, false, config.Enabled, config.Effect)
				if content.Mode == secondaryContentBackground {
					var effectErr error
					if terminalBinary == "" {
						terminalBinary, effectErr = exec.LookPath("sysc-terminal")
					}
					if effectErr == nil && !catalogLoaded {
						var catalog []byte
						catalogCtx, cancelCatalog := context.WithTimeout(ctx, time.Second)
						catalog, effectErr = exec.CommandContext(catalogCtx, terminalBinary, "--list").Output()
						cancelCatalog()
						if effectErr == nil {
							effects = secondaryEffectCatalog(string(catalog))
							palettes = secondaryPalettes(string(catalog))
							catalogLoaded = true
						}
					}
					var effect string
					if effectErr == nil {
						var ok bool
						effect, ok = secondaryEffectID(content.Effect, effects)
						if !ok {
							effectErr = fmt.Errorf("secondary effect %q is not available in sysc-terminal", content.Effect)
						}
					}
					var theme string
					if effectErr == nil {
						var themeWarning error
						theme, themeWarning = secondaryTheme(runtime.Theme, palettes)
						if theme == "" {
							effectErr = themeWarning
						} else {
							iterationErr = errors.Join(iterationErr, themeWarning)
						}
					}
					if effectErr != nil {
						iterationErr = errors.Join(iterationErr, effectErr)
					} else {
						for _, output := range secondaryOutputTargets(outputs, primary, config.Exclude) {
							desired[output] = secondaryRuntimeState{Theme: theme, Background: effect}
						}
					}
				}
			}

			for output, child := range children {
				want, exists := desired[output]
				if wakePending || !exists || !desiredMatches(child, want, runtime.Screensaver) {
					child.stop()
					delete(children, output)
					continue
				}
				if child.kind == secondaryChildBackground && (child.effect != want.Background || child.theme != want.Theme) {
					response, changeErr := backgroundRequest(ctx, child.socket, "change effect "+want.Background+" theme "+want.Theme)
					if changeErr != nil || response != "OK" {
						if changeErr == nil {
							changeErr = fmt.Errorf("background change: %s", response)
						}
						iterationErr = errors.Join(iterationErr, changeErr)
						child.stop()
						delete(children, output)
					} else {
						child.effect, child.theme = want.Background, want.Theme
						child.launchKey = secondaryLaunchKey(false, want)
					}
				}
			}
			if wakePending {
				for output, child := range children {
					child.stop()
					delete(children, output)
				}
			}

			if !wakePending {
				if runtime.Screensaver {
					if len(desired) > 0 && !fullscreenReady {
						saver, required := manager.(interface{ prepareSaverFullscreen(context.Context) error })
						if required {
							if err := saver.prepareSaverFullscreen(ctx); err != nil {
								iterationErr = errors.Join(iterationErr, err)
								desired = make(map[string]secondaryRuntimeState)
							} else {
								fullscreenReady = true
							}
						} else {
							fullscreenReady = true
						}
					}
					for _, output := range sortedDesiredOutputs(desired) {
						if stateBus.load() != runtime {
							break
						}
						launchKey := secondaryLaunchKey(true, runtime)
						if children[output] != nil || failures.blocked(output, launchKey) {
							continue
						}
						// A compositor can move focus before its IPC call reports an error.
						focusAttempted = true
						if err := manager.FocusOutput(ctx, output); err != nil {
							failures.record(output, launchKey)
							iterationErr = errors.Join(iterationErr, err)
							continue
						}
						readyFile := filepath.Join(dir, fmt.Sprintf("ready-%d", time.Now().UnixNano()))
						child, startErr := launchSecondarySaver(ctx, stateBus, runtime, output, kittyBinary, greeterBinary, secondaryKittyConfig(), runtime.Theme, runtime.SaverStart.Format(time.RFC3339Nano), readyFile)
						if startErr != nil {
							if stateBus.load() != runtime {
								break
							}
							failures.record(output, launchKey)
							iterationErr = errors.Join(iterationErr, startErr)
							continue
						}
						if stateBus.load() != runtime {
							child.stop()
							break
						}
						child.launchKey = launchKey
						children[output] = child
						if hasExitedSaver(children) {
							wakePending = true
							sendSecondary(send, secondaryWakeMsg{})
							for name, saver := range children {
								saver.stop()
								delete(children, name)
							}
							break
						}
					}
				} else {
					for _, output := range sortedDesiredOutputs(desired) {
						if stateBus.load() != runtime {
							break
						}
						want := desired[output]
						launchKey := secondaryLaunchKey(false, want)
						if children[output] != nil || failures.blocked(output, launchKey) {
							continue
						}
						socket := filepath.Join(dir, fmt.Sprintf("background-%d.sock", time.Now().UnixNano()))
						child, startErr := launchSecondaryBackground(ctx, stateBus, runtime, output, terminalBinary, socket, want.Background, want.Theme)
						if startErr != nil {
							if stateBus.load() != runtime {
								break
							}
							failures.record(output, launchKey)
							iterationErr = errors.Join(iterationErr, fmt.Errorf("background %s: %w", output, startErr))
							continue
						}
						if stateBus.load() != runtime {
							child.stop()
							break
						}
						child.launchKey = launchKey
						children[output] = child
					}
				}
			}
			focusRestorePending = focusRestorePending || focusAttempted
			var restoreErr error
			focusRestorePending, restoreErr = restorePrimaryFocus(ctx, manager, primary, focusRestorePending)
			iterationErr = errors.Join(iterationErr, restoreErr)
			if stateBus.load() != runtime {
				continue
			}
			active := make([]string, 0, len(children))
			for output := range children {
				active = append(active, output)
			}
			sort.Strings(active)
			publishSecondaryStatus(send, &lastStatus, &haveStatus, secondaryMsg{Primary: primary, Outputs: active, Err: iterationErr})
		}

		wait := 2 * time.Second
		if runtime.Screensaver && len(children) > 0 {
			wait = 100 * time.Millisecond
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-stateBus.changed:
			timer.Stop()
		case <-timer.C:
		}
	}
}

func restorePrimaryFocus(ctx context.Context, manager compositorOutputManager, primary string, pending bool) (bool, error) {
	if !pending || primary == "" {
		return pending, nil
	}
	if err := manager.FocusOutput(ctx, primary); err != nil {
		return true, fmt.Errorf("restore focus to %s: %w", primary, err)
	}
	return false, nil
}

func reconcileSecondaryChildren(children map[string]*backgroundChild, saverActive, wakePending bool, send func(tea.Msg), failures secondaryLaunchFailures) (bool, error) {
	var childErr error
	for output, child := range children {
		select {
		case <-child.done:
			cleanSaverExit := child.kind == secondaryChildScreensaver && child.exitErr == nil
			if cleanSaverExit && saverActive && !wakePending && !child.stopRequested && child.ready {
				wakePending = true
				sendSecondary(send, secondaryWakeMsg{})
			} else if child.exitErr != nil {
				childErr = errors.Join(childErr, fmt.Errorf("secondary output %s exited: %w", output, child.exitErr))
			}
			// A helper that dies on its own counts toward the relaunch limit.
			if failures != nil && !child.stopRequested && !cleanSaverExit && child.launchKey != "" {
				if time.Since(child.started) >= secondaryStableRun {
					delete(failures, output)
				}
				failures.record(output, child.launchKey)
			}
			child.stop()
			delete(children, output)
			continue
		default:
		}
		if wakePending || (saverActive && child.kind != secondaryChildScreensaver) || (!saverActive && child.kind == secondaryChildScreensaver) {
			child.stop()
			delete(children, output)
		}
	}
	return wakePending, childErr
}

func secondaryTheme(theme string, palettes map[string]bool) (string, error) {
	theme = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(theme)), " ", "-")
	if theme == "" {
		theme = "dracula"
	}
	if palettes[theme] {
		return theme, nil
	}
	if palettes["dracula"] {
		return "dracula", fmt.Errorf("secondary palette %q unavailable; using dracula", theme)
	}
	return "", fmt.Errorf("secondary palette catalog lacks dracula fallback")
}

func hasExitedSaver(children map[string]*backgroundChild) bool {
	for _, child := range children {
		if child.kind != secondaryChildScreensaver || child.stopRequested || !child.ready {
			continue
		}
		select {
		case <-child.done:
			if child.exitErr == nil {
				return true
			}
		default:
		}
	}
	return false
}

func desiredMatches(child *backgroundChild, desired secondaryRuntimeState, saver bool) bool {
	if child == nil {
		return false
	}
	if saver {
		return child.kind == secondaryChildScreensaver && child.theme == desired.Theme
	}
	return desired.Background != "" && child.kind == secondaryChildBackground
}

func sortedDesiredOutputs(desired map[string]secondaryRuntimeState) []string {
	outputs := make([]string, 0, len(desired))
	for output := range desired {
		outputs = append(outputs, output)
	}
	sort.Strings(outputs)
	return outputs
}

func publishSecondaryStatus(send func(tea.Msg), previous *secondaryMsg, havePrevious *bool, next secondaryMsg) {
	if *havePrevious && previous.Primary == next.Primary && strings.Join(previous.Outputs, "\x00") == strings.Join(next.Outputs, "\x00") && errorText(previous.Err) == errorText(next.Err) {
		return
	}
	*previous, *havePrevious = next, true
	sendSecondary(send, next)
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func sendSecondary(send func(tea.Msg), msg tea.Msg) {
	if send != nil {
		send(msg)
	}
}

func (c commandCompositor) prepareSaverFullscreen(ctx context.Context) error {
	if c.kind != compositorHyprland {
		return nil
	}
	rule := "match:class " + secondarySaverWindowClass + ", fullscreen on"
	_, err := runOutputCommand(ctx, "hyprctl", "keyword", "windowrule", rule)
	if err != nil {
		return fmt.Errorf("configure Hyprland fullscreen saver rule: %w", err)
	}
	return nil
}
