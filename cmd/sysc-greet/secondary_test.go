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
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func TestSecondaryTargets(t *testing.T) {
	outputs := []secondaryOutput{
		{Name: "DP-1", Active: true},
		{Name: "DP-2", Active: true},
		{Name: "HDMI-A-1", Active: false},
	}
	got := secondaryOutputTargets(outputs, "DP-1", "")
	if len(got) != 1 || got[0] != "DP-2" {
		t.Fatalf("targets=%v", got)
	}
	if len(secondaryOutputTargets(outputs, "DP-1", " DP-2 , unknown ")) != 0 {
		t.Fatal("excluded output selected")
	}
	if len(secondaryOutputTargets(outputs, "missing", "")) != 0 {
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

func TestReconcileSecondaryChildrenStopsSaverAfterPrimaryWake(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	child, err := startBackgroundChild(cmd)
	if err != nil {
		t.Fatal(err)
	}
	child.kind = secondaryChildScreensaver
	defer child.stop()
	children := map[string]*backgroundChild{"DP-1": child}

	if _, err := reconcileSecondaryChildren(children, false, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := children["DP-1"]; ok {
		t.Fatal("secondary saver remained after the primary left screensaver mode")
	}
	select {
	case <-child.done:
	case <-time.After(2 * time.Second):
		t.Fatal("secondary saver was not stopped and reaped")
	}
}

func TestSecondaryPaletteCatalog(t *testing.T) {
	palettes := secondaryPalettes("theme  dracula \ntheme  tokyo-night tokyonight\neffect matrix 0\n")
	if !palettes["dracula"] || !palettes["tokyo-night"] || !palettes["tokyonight"] || palettes["matrix"] || palettes["blue"] {
		t.Fatalf("palettes=%v", palettes)
	}
}

func TestHasExitedSaverOnlyAcceptsReadyCleanExits(t *testing.T) {
	closed := make(chan struct{})
	close(closed)
	for _, tc := range []struct {
		name  string
		child *backgroundChild
		want  bool
	}{
		{name: "ready clean child", child: &backgroundChild{kind: secondaryChildScreensaver, done: closed, ready: true}, want: true},
		{name: "failed child", child: &backgroundChild{kind: secondaryChildScreensaver, done: closed, ready: true, exitErr: errors.New("exit 1")}},
		{name: "not ready", child: &backgroundChild{kind: secondaryChildScreensaver, done: closed}},
		{name: "background child", child: &backgroundChild{kind: secondaryChildBackground, done: closed, ready: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasExitedSaver(map[string]*backgroundChild{"DP-1": tc.child}); got != tc.want {
				t.Fatalf("hasExitedSaver() = %v, want %v", got, tc.want)
			}
		})
	}
}

type testSecondaryCompositor struct {
	mu                          sync.Mutex
	outputs                     []secondaryOutput
	focused                     string
	focuses                     []string
	listCalls                   int
	failListAfter               int
	listFailed                  chan struct{}
	focusErrorFor               string
	focusErrorBeforeChangeFor   string
	focusErrorBeforeChangeCount int
	fullscreenErr               error
	fullscreenCalls             int
}

func (c *testSecondaryCompositor) Name() string { return "test" }

func (c *testSecondaryCompositor) ListOutputs(context.Context) ([]secondaryOutput, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.listCalls++
	if c.failListAfter > 0 && c.listCalls > c.failListAfter {
		if c.listFailed != nil {
			select {
			case c.listFailed <- struct{}{}:
			default:
			}
		}
		return nil, "", errors.New("output IPC unavailable")
	}
	return append([]secondaryOutput(nil), c.outputs...), c.focused, nil
}

func (c *testSecondaryCompositor) FocusOutput(_ context.Context, output string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.focuses = append(c.focuses, output)
	if output == c.focusErrorBeforeChangeFor && c.focusErrorBeforeChangeCount > 0 {
		c.focusErrorBeforeChangeCount--
		return errors.New("focus failed before changing output")
	}
	c.focused = output
	if output == c.focusErrorFor {
		return errors.New("focus changed before the compositor reported an error")
	}
	return nil
}

func (c *testSecondaryCompositor) setOutputs(outputs []secondaryOutput) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.outputs = append([]secondaryOutput(nil), outputs...)
}

func (c *testSecondaryCompositor) focusHistory() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.focuses...)
}

func (c *testSecondaryCompositor) focusedOutput() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.focused
}

func (c *testSecondaryCompositor) prepareSaverFullscreen(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fullscreenCalls++
	return c.fullscreenErr
}

func (c *testSecondaryCompositor) setFullscreenError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fullscreenErr = err
}

func (c *testSecondaryCompositor) fullscreenAttempts() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fullscreenCalls
}

func TestSecondaryTerminalHelperProcess(t *testing.T) {
	if os.Getenv("SYSC_GREET_TERMINAL_HELPER") != "1" {
		return
	}
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 1 && args[0] == "--list" {
		fmt.Fprintln(os.Stdout, "effect fire 0\neffect rain 0\ntheme dracula nord")
		os.Exit(0)
	}
	values := make(map[string]string)
	for i := 0; i+1 < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			values[args[i]] = args[i+1]
			i++
		}
	}
	output, effect, theme, socket, logPath := values["--output"], values["--effect"], values["--theme"], values["--ipc-socket"], os.Getenv("SYSC_GREET_HELPER_LOG")
	logEvent := func(event, nextEffect, nextTheme string) {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			os.Exit(2)
		}
		_, _ = fmt.Fprintf(file, "%s|%s|%s|%s\n", event, output, nextEffect, nextTheme)
		_ = file.Close()
	}
	logEvent("started", effect, theme)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	stop := func() {
		logEvent("stopped", effect, theme)
		os.Exit(0)
	}
	if effect == os.Getenv("SYSC_GREET_HELPER_DELAY_EFFECT") {
		for {
			if _, err := os.Stat(os.Getenv("SYSC_GREET_HELPER_RELEASE_FILE")); err == nil {
				break
			}
			select {
			case <-signals:
				stop()
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		os.Exit(2)
	}
	go func() {
		<-signals
		stop()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		request, _ := bufio.NewReader(conn).ReadString('\n')
		request = strings.TrimSpace(request)
		response := "OK\n"
		if request == "query" {
			response = "STATUS:running\n"
		} else if strings.HasPrefix(request, "change effect ") {
			fields := strings.Fields(request)
			if len(fields) == 5 {
				effect, theme = fields[2], fields[4]
				logEvent("changed", effect, theme)
			} else {
				response = "ERROR\n"
			}
		}
		_, _ = io.WriteString(conn, response)
		_ = conn.Close()
	}
}

func TestSecondaryManagerMirrorsSaverAndRestoresPrimaryFocus(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kittyLog := filepath.Join(t.TempDir(), "kitty.log")
	kitty := filepath.Join(binDir, "kitty")
	script := "#!/bin/sh\nprintf 'started\\n' >> \"$KITTY_TEST_LOG\"\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--secondary-ready-file\" ]; then\n    shift\n    : > \"$1\"\n  fi\n  shift\ndone\nexec sleep 60\n"
	if err := os.WriteFile(kitty, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("KITTY_TEST_LOG", kittyLog)
	manager := &testSecondaryCompositor{
		outputs: []secondaryOutput{{Name: "DP-1", Active: true}, {Name: "DP-2", Active: true}, {Name: "eDP-1", Active: true}},
		focused: "eDP-1",
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false, Exclude: "DP-2"}, state, nil, manager)
		close(done)
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(kittyLog)
		if strings.Count(string(data), "started") >= 2 {
			focuses := manager.focusHistory()
			if len(focuses) >= 3 && focuses[len(focuses)-1] == "eDP-1" {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	data, err := os.ReadFile(kittyLog)
	if err != nil || strings.Count(string(data), "started") != 2 {
		cancel()
		<-done
		t.Fatalf("started %d saver children, error = %v; want 2", strings.Count(string(data), "started"), err)
	}
	if got := manager.focusHistory(); len(got) < 3 || got[0] != "DP-1" || got[1] != "DP-2" || got[len(got)-1] != "eDP-1" {
		cancel()
		<-done
		t.Fatalf("focus history = %v, want DP-1, DP-2, then primary eDP-1", got)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("secondary manager did not stop and reap Kitty children")
	}
}

func TestSecondaryManagerRestoresPrimaryAfterFailedFocusAttempt(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kitty := filepath.Join(binDir, "kitty")
	if err := os.WriteFile(kitty, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	manager := &testSecondaryCompositor{
		outputs:       []secondaryOutput{{Name: "DP-1", Active: true}, {Name: "eDP-1", Active: true}},
		focused:       "eDP-1",
		focusErrorFor: "DP-1",
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false}, state, nil, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop")
		}
	}()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		focuses := manager.focusHistory()
		for i := 0; i+1 < len(focuses); i++ {
			if focuses[i] == "DP-1" && focuses[i+1] == "eDP-1" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("focus history = %v; want primary restored after failed target focus", manager.focusHistory())
}

func TestSecondaryManagerRetriesFailedPrimaryFocusRestore(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kitty := filepath.Join(binDir, "kitty")
	script := "#!/bin/sh\nready=\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--secondary-ready-file\" ]; then\n    shift\n    ready=$1\n  fi\n  shift\ndone\n: > \"$ready\"\nexec sleep 60\n"
	if err := os.WriteFile(kitty, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	manager := &testSecondaryCompositor{
		outputs:                     []secondaryOutput{{Name: "DP-1", Active: true}, {Name: "eDP-1", Active: true}},
		focused:                     "eDP-1",
		focusErrorBeforeChangeFor:   "eDP-1",
		focusErrorBeforeChangeCount: 1,
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false}, state, nil, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop and reap Kitty child")
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		focuses := manager.focusHistory()
		primaryFocuses := 0
		for _, output := range focuses {
			if output == "eDP-1" {
				primaryFocuses++
			}
		}
		if primaryFocuses >= 2 && manager.focusedOutput() == "eDP-1" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("focus history=%v current=%q; want retry to restore primary", manager.focusHistory(), manager.focusedOutput())
}

func TestSecondaryManagerWakesPrimaryWhenOutputListingFails(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kittyLog := filepath.Join(t.TempDir(), "kitty.log")
	exitFile := filepath.Join(t.TempDir(), "exit-kitty")
	kitty := filepath.Join(binDir, "kitty")
	script := "#!/bin/sh\nprintf 'started\\n' >> \"$KITTY_TEST_LOG\"\nready=\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--secondary-ready-file\" ]; then\n    shift\n    ready=$1\n  fi\n  shift\ndone\n: > \"$ready\"\nwhile [ ! -e \"$KITTY_TEST_EXIT_FILE\" ]; do sleep 0.01; done\n"
	if err := os.WriteFile(kitty, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("KITTY_TEST_LOG", kittyLog)
	t.Setenv("KITTY_TEST_EXIT_FILE", exitFile)
	manager := &testSecondaryCompositor{
		outputs:       []secondaryOutput{{Name: "eDP-1", Active: true}, {Name: "DP-1", Active: true}},
		focused:       "eDP-1",
		failListAfter: 1,
		listFailed:    make(chan struct{}, 1),
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	messages := make(chan tea.Msg, 16)
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false}, state, func(msg tea.Msg) { messages <- msg }, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop and reap Kitty child")
		}
	}()

	deadline := time.After(3 * time.Second)
	for {
		data, _ := os.ReadFile(kittyLog)
		if strings.Contains(string(data), "started") {
			break
		}
		select {
		case <-deadline:
			t.Fatal("secondary Kitty saver did not start")
		case <-time.After(10 * time.Millisecond):
		}
	}
	select {
	case <-manager.listFailed:
	case <-time.After(3 * time.Second):
		t.Fatal("compositor output listing did not fail")
	}
	if err := os.WriteFile(exitFile, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	deadline = time.After(2 * time.Second)
	for {
		select {
		case msg := <-messages:
			if _, ok := msg.(secondaryWakeMsg); ok {
				return
			}
		case <-deadline:
			t.Fatal("secondary saver exit did not wake the primary while output listing was failing")
		}
	}
}

func TestSecondaryManagerRetriesFullscreenPreparationBeforeSaverLaunch(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kittyLog := filepath.Join(t.TempDir(), "kitty.log")
	kitty := filepath.Join(binDir, "kitty")
	script := "#!/bin/sh\nready=\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--secondary-ready-file\" ]; then\n    shift\n    ready=$1\n  fi\n  shift\ndone\nprintf 'started\\n' >> \"$KITTY_TEST_LOG\"\n: > \"$ready\"\nexec sleep 60\n"
	if err := os.WriteFile(kitty, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("KITTY_TEST_LOG", kittyLog)
	manager := &testSecondaryCompositor{
		outputs:       []secondaryOutput{{Name: "eDP-1", Active: true}, {Name: "DP-1", Active: true}},
		focused:       "eDP-1",
		fullscreenErr: errors.New("Hyprland rule unavailable"),
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false}, state, nil, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop and reap Kitty child")
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for manager.fullscreenAttempts() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if manager.fullscreenAttempts() == 0 {
		t.Fatal("secondary manager did not prepare fullscreen saver rules")
	}
	if _, err := os.Stat(kittyLog); !os.IsNotExist(err) {
		t.Fatalf("Kitty saver launched before fullscreen preparation succeeded, stat error = %v", err)
	}

	manager.setFullscreenError(nil)
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(kittyLog); err == nil && strings.Contains(string(data), "started") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("secondary saver did not launch after fullscreen preparation recovered")
}

func TestSecondaryManagerCancelsSaverLaunchWhenPrimaryWakes(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kittyLog := filepath.Join(t.TempDir(), "kitty.log")
	kitty := filepath.Join(binDir, "kitty")
	script := "#!/bin/sh\nready=\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--secondary-ready-file\" ]; then\n    shift\n    ready=$1\n  fi\n  shift\ndone\ntrap 'printf \"stopped\\n\" >> \"$KITTY_TEST_LOG\"; exit 0' TERM\nprintf 'started\\n' >> \"$KITTY_TEST_LOG\"\nwhile [ ! -e \"$KITTY_TEST_RELEASE_FILE\" ]; do sleep 0.01; done\n: > \"$ready\"\nexec sleep 60\n"
	if err := os.WriteFile(kitty, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("KITTY_TEST_LOG", kittyLog)
	t.Setenv("KITTY_TEST_RELEASE_FILE", filepath.Join(t.TempDir(), "release"))
	manager := &testSecondaryCompositor{
		outputs: []secondaryOutput{{Name: "eDP-1", Active: true}, {Name: "DP-1", Active: true}, {Name: "DP-2", Active: true}},
		focused: "eDP-1",
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false}, state, nil, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop")
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(kittyLog); err == nil && strings.Contains(string(data), "started") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if data, err := os.ReadFile(kittyLog); err != nil || !strings.Contains(string(data), "started") {
		t.Fatalf("secondary Kitty saver did not start: %v", err)
	}

	state.publish(secondaryRuntimeState{Theme: "nord"})
	deadline = time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(kittyLog); err == nil && strings.Contains(string(data), "stopped") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("secondary Kitty saver kept launching after the primary woke")
}

func TestSecondaryManagerTracksSaverOutputRemovalAndReappearance(t *testing.T) {
	runtimeDir := t.TempDir()
	binDir := t.TempDir()
	kittyLog := filepath.Join(t.TempDir(), "kitty.log")
	kitty := filepath.Join(binDir, "kitty")
	script := "#!/bin/sh\nready=\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--secondary-ready-file\" ]; then\n    shift\n    ready=$1\n  fi\n  shift\ndone\nprintf 'started %s\\n' \"$$\" >> \"$KITTY_TEST_LOG\"\n: > \"$ready\"\nexec sleep 60\n"
	if err := os.WriteFile(kitty, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	t.Setenv("KITTY_TEST_LOG", kittyLog)
	manager := &testSecondaryCompositor{
		outputs: []secondaryOutput{{Name: "eDP-1", Active: true}, {Name: "DP-1", Active: true}},
		focused: "eDP-1",
	}
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Screensaver: true, SaverStart: time.Now()})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: false}, state, nil, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop and reap Kitty child")
		}
	}()

	readStartedPIDs := func() []int {
		data, err := os.ReadFile(kittyLog)
		if err != nil {
			return nil
		}
		var pids []int
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 || fields[0] != "started" {
				continue
			}
			pid, err := strconv.Atoi(fields[1])
			if err == nil {
				pids = append(pids, pid)
			}
		}
		return pids
	}
	waitForStartedCount := func(count int) []int {
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if pids := readStartedPIDs(); len(pids) >= count {
				return pids
			}
			time.Sleep(10 * time.Millisecond)
		}
		return readStartedPIDs()
	}
	first := waitForStartedCount(1)
	if len(first) != 1 {
		t.Fatalf("started saver PIDs = %v, want one", first)
	}

	manager.setOutputs([]secondaryOutput{{Name: "eDP-1", Active: true}})
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && syscall.Kill(first[0], 0) == nil {
		time.Sleep(10 * time.Millisecond)
	}
	if err := syscall.Kill(first[0], 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("removed-output saver process is still alive: %v", err)
	}

	manager.setOutputs([]secondaryOutput{{Name: "DP-2", Active: true}, {Name: "eDP-1", Active: true}})
	second := waitForStartedCount(2)
	if len(second) != 2 {
		t.Fatalf("started saver PIDs after output reappeared = %v, want two", second)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("secondary manager did not stop and reap reappeared-output saver")
	}
	if err := syscall.Kill(second[1], 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("reappeared-output saver process is still alive: %v", err)
	}
}

func TestSecondaryManagerUpdatesEffectAndStopsForVideoWallpaper(t *testing.T) {
	runtimeDir, err := os.MkdirTemp("/tmp", "sg-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(runtimeDir) })
	binDir := t.TempDir()
	helperLog := filepath.Join(t.TempDir(), "terminal.log")
	terminal := filepath.Join(binDir, "sysc-terminal")
	script := "#!/bin/sh\nexec \"$SYSC_GREET_TEST_BINARY\" -test.run=TestSecondaryTerminalHelperProcess -- \"$@\"\n"
	if err := os.WriteFile(terminal, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SYSC_GREET_TEST_BINARY", executable)
	t.Setenv("SYSC_GREET_TERMINAL_HELPER", "1")
	t.Setenv("SYSC_GREET_HELPER_LOG", helperLog)
	t.Setenv("SYSC_GREET_HELPER_DELAY_EFFECT", "fire")
	t.Setenv("SYSC_GREET_HELPER_RELEASE_FILE", filepath.Join(t.TempDir(), "release"))
	// Race-instrumented helper processes otherwise sleep for one second at exit and hit the catalog deadline.
	raceOptions := append(strings.Fields(os.Getenv("GORACE")), "atexit_sleep_ms=0")
	t.Setenv("GORACE", strings.Join(raceOptions, " "))
	manager := &testSecondaryCompositor{
		outputs: []secondaryOutput{{Name: "eDP-1", Active: true}, {Name: "DP-1", Active: true}},
		focused: "eDP-1",
	}
	managerMessages := make(chan tea.Msg, 8)
	state := newSecondaryStateBus()
	state.publish(secondaryRuntimeState{Theme: "nord", Background: "fire"})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runSecondaryManager(ctx, secondaryConfig{Enabled: true}, state, func(msg tea.Msg) { managerMessages <- msg }, manager)
		close(done)
	}()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("secondary manager did not stop and reap sysc-terminal helper")
		}
	}()

	waitForLogWithin := func(want string, timeout time.Duration) bool {
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			data, err := os.ReadFile(helperLog)
			if err == nil && strings.Contains(string(data), want) {
				return true
			}
			time.Sleep(10 * time.Millisecond)
		}
		return false
	}
	waitForLog := func(want string) bool { return waitForLogWithin(want, 3*time.Second) }
	waitForActiveOutput := func(output string, timeout time.Duration) bool {
		deadline := time.NewTimer(timeout)
		defer deadline.Stop()
		for {
			select {
			case msg := <-managerMessages:
				status, ok := msg.(secondaryMsg)
				if !ok || status.Err != nil {
					continue
				}
				for _, active := range status.Outputs {
					if active == output {
						return true
					}
				}
			case <-deadline.C:
				return false
			}
		}
	}
	if !waitForLog("started|DP-1|fire|nord") {
		t.Fatal("secondary fire helper did not start with the selected theme")
	}
	state.publish(secondaryRuntimeState{Theme: "dracula", Background: "rain"})
	if !waitForLogWithin("started|DP-1|rain|dracula", time.Second) {
		data, _ := os.ReadFile(helperLog)
		t.Fatalf("secondary helper did not reconcile a selection change during startup; log=%q", data)
	}
	if !waitForActiveOutput("DP-1", time.Second) {
		data, _ := os.ReadFile(helperLog)
		t.Fatalf("secondary manager did not register the ready helper before the next state change; log=%q", data)
	}
	data, err := os.ReadFile(helperLog)
	if err != nil {
		t.Fatal(err)
	}
	stoppedFire := strings.Index(string(data), "stopped|DP-1|fire|nord")
	startedRain := strings.Index(string(data), "started|DP-1|rain|dracula")
	if stoppedFire < 0 || startedRain < 0 || stoppedFire > startedRain {
		t.Fatal("stale effect helper was not stopped before the updated helper started")
	}
	state.publish(secondaryRuntimeState{Theme: "nord", Background: "fire"})
	if !waitForLog("changed|DP-1|fire|nord") {
		data, _ := os.ReadFile(helperLog)
		t.Fatalf("ready secondary helper did not receive the selected effect and theme; log=%q", data)
	}
	state.publish(secondaryRuntimeState{Theme: "dracula", Background: "fire", Wallpaper: "gslapper.mp4"})
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, _ = os.ReadFile(helperLog)
		if strings.Count(string(data), "stopped|DP-1|fire|nord") >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if strings.Count(string(data), "stopped|DP-1|fire|nord") < 2 {
		t.Fatal("video wallpaper selection did not stop the secondary effect")
	}
	time.Sleep(200 * time.Millisecond)
	data, err = os.ReadFile(helperLog)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "started|"); got != 2 {
		t.Fatalf("secondary effect restarted under video wallpaper %d times", got)
	}
}
