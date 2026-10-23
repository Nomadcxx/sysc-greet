package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func TestScreensaverVariantUsesSharedActivationTime(t *testing.T) {
	start := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		now  time.Time
		want int
	}{
		{name: "first variant before cycle", now: start.Add(29 * time.Second), want: 0},
		{name: "next variant at thirty seconds", now: start.Add(30 * time.Second), want: 1},
		{name: "variants wrap", now: start.Add(91 * time.Second), want: 0},
		{name: "clock adjustment before activation", now: start.Add(-time.Second), want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := screensaverVariantIndex(3, start, tc.now); got != tc.want {
				t.Fatalf("screensaverVariantIndex() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParseSecondarySaverStartRequiresRFC3339Timestamp(t *testing.T) {
	want := time.Date(2026, 10, 23, 12, 0, 0, 0, time.FixedZone("AEST", 10*60*60))
	got, err := parseSecondarySaverStart("2026-10-23T12:00:00+10:00")
	if err != nil || !got.Equal(want) {
		t.Fatalf("parseSecondarySaverStart() = %v, %v; want %v", got, err, want)
	}
	for _, invalid := range []string{"", "not-a-time"} {
		if _, err := parseSecondarySaverStart(invalid); err == nil {
			t.Errorf("parseSecondarySaverStart(%q) succeeded", invalid)
		}
	}
}

func TestSecondarySaverCommandUsesGreeterKittyAndReadOnlyMode(t *testing.T) {
	got := secondarySaverCommandArgs("/etc/greetd/kitty.conf", "/usr/bin/sysc-greet", "nord", "2026-10-23T12:00:00Z", "/run/user/1000/ready")
	want := []string{
		"--start-as=fullscreen", "--class=sysc-greet-screensaver", "--single-instance=no", "--config=/etc/greetd/kitty.conf", "-o", "close_on_child_death=yes",
		"/usr/bin/sysc-greet", "--secondary-screensaver", "--follow-shell=false", "--theme", "nord",
		"--screensaver-started-at", "2026-10-23T12:00:00Z", "--secondary-ready-file", "/run/user/1000/ready",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("secondarySaverCommandArgs() = %q, want %q", got, want)
	}
}

func TestSecondaryChildEnvironmentDropsAuthenticationAndParentKittyState(t *testing.T) {
	t.Setenv("GREETD_SOCK", "/run/greetd.sock")
	t.Setenv("KITTY_WINDOW_ID", "7")
	t.Setenv("KITTY_LISTEN_ON", "unix:/tmp/kitty.sock")
	t.Setenv("KITTY_PID", "8")
	t.Setenv("KITTY_PUBLIC_KEY", "secret")
	forbidden := []string{"GREETD_SOCK=", "KITTY_WINDOW_ID=", "KITTY_LISTEN_ON=", "KITTY_PID=", "KITTY_PUBLIC_KEY="}
	for _, entry := range secondaryChildEnvironment() {
		for _, prefix := range forbidden {
			if strings.HasPrefix(entry, prefix) {
				t.Fatalf("secondary environment retained %q", prefix)
			}
		}
	}
}

func TestSecondarySaverReadyFileIsPrivateAndNeverOverwritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ready")
	if err := writeSecondarySaverReady(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("ready file mode = %04o, want 0600", got)
	}
	if err := writeSecondarySaverReady(path); err == nil {
		t.Fatal("ready file was overwritten")
	}
}

func TestSecondarySaverModeDoesNotCreateGreetdIPC(t *testing.T) {
	if shouldCreateGreetdIPC(Config{SecondarySaver: true}) {
		t.Fatal("secondary saver mode must not create a greetd IPC client")
	}
	if !shouldCreateGreetdIPC(Config{}) {
		t.Fatal("normal production mode must create a greetd IPC client")
	}
	if shouldCreateGreetdIPC(Config{TestMode: true}) {
		t.Fatal("test mode must not create a greetd IPC client")
	}
}

func TestSecondarySaverModeDoesNotLoadUserPreferences(t *testing.T) {
	if shouldLoadUserPreferences(Config{SecondarySaver: true}) {
		t.Fatal("secondary saver mode must not load preferences that can trigger wallpaper changes")
	}
	if !shouldLoadUserPreferences(Config{}) {
		t.Fatal("normal production mode must load saved preferences")
	}
	if shouldLoadUserPreferences(Config{TestMode: true}) {
		t.Fatal("test mode must not load saved preferences")
	}
}

func TestInitialScreensaverModePublishesActivationTime(t *testing.T) {
	m := initialModel(Config{TestMode: true}, true)
	m.secondaryRuntime = newSecondaryStateBus()
	m.publishSecondaryRuntime()

	runtime := m.secondaryRuntime.load()
	if !runtime.Screensaver || runtime.SaverStart.IsZero() {
		t.Fatalf("screensaver runtime = %+v; want active saver with activation time", runtime)
	}
}

func TestSecondarySaverInputQuitsWithoutWakingIntoLogin(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  tea.Msg
	}{
		{name: "key", msg: tea.KeyPressMsg{Code: tea.KeyEnter}},
		{name: "mouse motion", msg: tea.MouseMotionMsg{X: 10, Y: 5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := model{config: Config{SecondarySaver: true}, mode: ModeScreensaver}
			updated, cmd := m.Update(tc.msg)
			if updated.(model).mode != ModeScreensaver {
				t.Fatal("secondary saver input changed its mode instead of closing the child")
			}
			if cmd == nil {
				t.Fatal("secondary saver input did not stop its Kitty child")
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatal("secondary saver input command did not request child shutdown")
			}
		})
	}
}

func TestSecondaryWakeLeavesPrimaryScreensaverAndClearsSharedStart(t *testing.T) {
	start := time.Now().Add(-time.Minute)
	m := model{config: Config{}, mode: ModeScreensaver, screensaverStart: start, screensaverActive: true}
	updated, _ := m.Update(secondaryWakeMsg{})
	got := updated.(model)
	if got.mode != ModeLogin || !got.screensaverStart.IsZero() || got.screensaverActive {
		t.Fatalf("wake state = mode %q start %v active %v", got.mode, got.screensaverStart, got.screensaverActive)
	}
}

func TestSecondarySaverNeverWritesPreferences(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	m := model{config: Config{SecondarySaver: true}}
	m.savePreferences()
	if _, err := os.Stat(filepath.Join(home, ".cache", "sysc-greet", "preferences")); !os.IsNotExist(err) {
		t.Fatalf("secondary saver wrote preferences, stat error = %v", err)
	}
}

func TestSecondaryRuntimeBusPublishesSaverActivationAndContent(t *testing.T) {
	bus := newSecondaryStateBus()
	started := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	want := secondaryRuntimeState{
		Theme:       "nord",
		Background:  "pour",
		Wallpaper:   "slr.mp4",
		Screensaver: true,
		SaverStart:  started,
	}
	bus.publish(want)
	if got := bus.load(); got != want {
		t.Fatalf("runtime state = %+v, want %+v", got, want)
	}
	select {
	case <-bus.changed:
	default:
		t.Fatal("state change did not wake the reconciler")
	}
}
