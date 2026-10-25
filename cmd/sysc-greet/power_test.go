package main

import (
	"os"
	"os/exec"
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
