package main

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/v2/textinput"
)

// stubPowerCommand replaces the power request with run and records the
// arguments the greeter asked for.
func stubPowerCommand(t *testing.T, run *exec.Cmd) *[]string {
	t.Helper()
	var asked []string
	saved := powerCommand
	powerCommand = func(args ...string) *exec.Cmd {
		asked = args
		return run
	}
	t.Cleanup(func() { powerCommand = saved })
	return &asked
}

func powerTestModel() model {
	return model{
		mode:          ModePower,
		usernameInput: textinput.New(),
		passwordInput: textinput.New(),
	}
}

func TestParsePowerCommand(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{name: "unset uses the built-in action"},
		{name: "argv", raw: `["sudo","reboot"]`, want: []string{"sudo", "reboot"}},
		{name: "reject shell command", raw: "sudo reboot", wantErr: true},
		{name: "reject empty argv", raw: `[]`, wantErr: true},
		{name: "reject empty executable", raw: `[""]`, wantErr: true},
		{name: "reject whitespace executable", raw: `["   "]`, wantErr: true},
		{name: "reject null argument", raw: `["sudo",null]`, wantErr: true},
		{name: "reject NUL argument", raw: `["sudo","\u0000"]`, wantErr: true},
		{name: "reject null", raw: `null`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePowerCommand(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePowerCommand(%q) error = %v; want error %t", tt.raw, err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("parsePowerCommand(%q) = %v; want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestPowerActionsAskLogin1(t *testing.T) {
	for _, tt := range []struct {
		action string
		verb   string
	}{
		{"Reboot", "reboot"},
		{"Shutdown", "poweroff"},
	} {
		t.Run(tt.action, func(t *testing.T) {
			asked := stubPowerCommand(t, exec.Command("true"))
			next, cmd := powerTestModel().Update(powerSelectedMsg(tt.action))
			if got := next.(model).mode; got != ModeLoading {
				t.Fatalf("mode = %s; want %s while the request runs", got, ModeLoading)
			}
			if cmd == nil {
				t.Fatal("no command issued for the power request")
			}
			if msg := cmd(); msg != nil {
				t.Fatalf("accepted request returned %#v; want nil so the greeter stays up", msg)
			}
			if len(*asked) != 3 || (*asked)[1] != "--no-ask-password" || (*asked)[2] != tt.verb {
				t.Fatalf("power request = %v; want a noninteractive %s request", *asked, tt.verb)
			}
		})
	}
}

func TestPowerActionsUseConfiguredArgv(t *testing.T) {
	tests := []struct {
		action string
		argv   []string
		config Config
	}{
		{
			action: "Reboot",
			argv:   []string{"/usr/bin/sudo", "reboot", "now"},
			config: Config{PowerRebootCommand: []string{"/usr/bin/sudo", "reboot", "now"}},
		},
		{
			action: "Shutdown",
			argv:   []string{"/usr/local/bin/power wrapper", "--action", "power off"},
			config: Config{PowerShutdownCommand: []string{"/usr/local/bin/power wrapper", "--action", "power off"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			asked := stubPowerCommand(t, exec.Command("true"))
			m := powerTestModel()
			m.config = tt.config
			_, cmd := m.Update(powerSelectedMsg(tt.action))
			if msg := cmd(); msg != nil {
				t.Fatalf("accepted custom command returned %#v; want nil", msg)
			}
			if !slices.Equal(*asked, tt.argv) {
				t.Fatalf("power command = %v; want exact argv %v", *asked, tt.argv)
			}
		})
	}
}

func TestSystemdPowerActionsUseSystemctl(t *testing.T) {
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		t.Skip("requires a systemd host; elogind commands are checked in the Void VM")
	}
	asked := stubPowerCommand(t, exec.Command("true"))
	_, cmd := powerTestModel().Update(powerSelectedMsg("Reboot"))
	cmd()
	if len(*asked) == 0 || (*asked)[0] != "systemctl" {
		t.Fatalf("systemd power request = %v; systemd loginctl has no reboot verb", *asked)
	}
}

func TestRefusedPowerRequestShowsTheError(t *testing.T) {
	stubPowerCommand(t, exec.Command("sh", "-c", "echo 'Access denied' >&2; exit 1"))
	next, cmd := powerTestModel().Update(powerSelectedMsg("Reboot"))
	msg := cmd()
	if msg == nil {
		t.Fatal("refused request returned no message")
	}
	next, _ = next.(model).Update(msg)
	m := next.(model)
	if m.mode != ModeLogin {
		t.Fatalf("mode = %s after a refused request; want %s", m.mode, ModeLogin)
	}
	if m.failedAttempts != 0 {
		t.Fatal("a refused power request counted as a failed login")
	}
	if !strings.Contains(m.errorMessage, "Access denied") {
		t.Fatalf("errorMessage = %q; want the loginctl error", m.errorMessage)
	}
}
