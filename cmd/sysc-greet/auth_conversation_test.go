package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nomadcxx/sysc-greet/internal/ipc"
	"github.com/Nomadcxx/sysc-greet/internal/sessions"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func writeAuthTestFrame(conn net.Conn, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var size [4]byte
	binary.LittleEndian.PutUint32(size[:], uint32(len(data)))
	if _, err := conn.Write(size[:]); err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

func readAuthTestFrame(conn net.Conn, value any) error {
	var size [4]byte
	if _, err := io.ReadFull(conn, size[:]); err != nil {
		return err
	}
	data := make([]byte, binary.LittleEndian.Uint32(size[:]))
	if _, err := io.ReadFull(conn, data); err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func TestAuthenticateHandlesInfoPasswordAndVisibleOTPConversation(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "greetd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	t.Setenv("GREETD_SOCK", socketPath)

	client, err := ipc.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()

		var create ipc.CreateSession
		if err := readAuthTestFrame(conn, &create); err != nil {
			serverDone <- err
			return
		}
		if create.Type != ipc.CreateSessionRequest || create.Username != "test-user" {
			serverDone <- errors.New("unexpected create_session request")
			return
		}

		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeInfo,
			AuthMessage:     "Swipe your finger or enter your password\x1b[2J",
		}); err != nil {
			serverDone <- err
			return
		}
		var infoResponse ipc.PostAuthMessageResponse
		if err := readAuthTestFrame(conn, &infoResponse); err != nil {
			serverDone <- err
			return
		}
		if infoResponse.Type != ipc.PostAuthMessageResponseRequest || infoResponse.Response != nil {
			serverDone <- errors.New("info message did not receive a null response")
			return
		}
		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeError,
			AuthMessage:     "Previous authentication method unavailable",
		}); err != nil {
			serverDone <- err
			return
		}
		var noticeResponse ipc.PostAuthMessageResponse
		if err := readAuthTestFrame(conn, &noticeResponse); err != nil {
			serverDone <- err
			return
		}
		if noticeResponse.Type != ipc.PostAuthMessageResponseRequest || noticeResponse.Response != nil {
			serverDone <- errors.New("error notice did not receive a null response")
			return
		}

		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeSecret,
			AuthMessage:     "Password:",
		}); err != nil {
			serverDone <- err
			return
		}
		var passwordResponse ipc.PostAuthMessageResponse
		if err := readAuthTestFrame(conn, &passwordResponse); err != nil {
			serverDone <- err
			return
		}
		if passwordResponse.Response == nil || *passwordResponse.Response != "typed-password" {
			serverDone <- errors.New("first secret prompt did not receive the typed password")
			return
		}

		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeSecret,
			AuthMessage:     "One-time code:",
		}); err != nil {
			serverDone <- err
			return
		}
		var otpResponse ipc.PostAuthMessageResponse
		if err := readAuthTestFrame(conn, &otpResponse); err != nil {
			serverDone <- err
			return
		}
		if otpResponse.Response == nil || *otpResponse.Response != "123456" {
			serverDone <- errors.New("later secret prompt did not receive the UI response")
			return
		}
		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeVisible,
			AuthMessage:     "Confirm code:",
		}); err != nil {
			serverDone <- err
			return
		}
		var visibleResponse ipc.PostAuthMessageResponse
		if err := readAuthTestFrame(conn, &visibleResponse); err != nil {
			serverDone <- err
			return
		}
		if visibleResponse.Response == nil || *visibleResponse.Response != "visible-response" {
			serverDone <- errors.New("visible prompt did not receive the UI response")
			return
		}

		if err := writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse}); err != nil {
			serverDone <- err
			return
		}
		var start ipc.StartSession
		if err := readAuthTestFrame(conn, &start); err != nil {
			serverDone <- err
			return
		}
		if start.Type != ipc.StartSessionRequest || len(start.Cmd) != 1 || start.Cmd[0] != "/bin/true" {
			serverDone <- errors.New("unexpected start_session request")
			return
		}
		serverDone <- writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse})
	}()

	programMessages := make(chan tea.Msg, 4)
	m := model{
		ipcClient:       client,
		selectedSession: &sessions.Session{Name: "Test Session", Exec: "/bin/true"},
		passwordInput:   textinput.New(),
		sendMessage:     func(msg tea.Msg) { programMessages <- msg },
	}
	commandDone := make(chan tea.Msg, 1)
	go func() { commandDone <- m.authenticate("test-user", "typed-password")() }()

	notice := <-programMessages
	if got, ok := notice.(authNoticeMsg); !ok || string(got) != "Swipe your finger or enter your password" {
		t.Fatalf("notice message = %T %q, want sanitized auth notice", notice, got)
	}
	updated, _ := m.Update(notice)
	m = updated.(model)
	notice = <-programMessages
	if got, ok := notice.(authNoticeMsg); !ok || string(got) != "Previous authentication method unavailable" {
		t.Fatal("error auth message was not displayed as a sanitized notice")
	}
	updated, _ = m.Update(notice)
	m = updated.(model)

	prompt := <-programMessages
	promptMsg, ok := prompt.(authPromptMsg)
	if !ok {
		t.Fatalf("prompt message = %T, want authPromptMsg", prompt)
	}
	updated, _ = m.Update(prompt)
	m = updated.(model)
	if m.authPromptText != "One-time code:" || m.passwordInput.EchoMode != textinput.EchoPassword {
		t.Fatal("secret prompt text or masked echo mode was not applied")
	}
	m.passwordInput.SetValue("123456")
	if strings.Contains(stripAnsi(m.renderMainForm(80)), "123456") {
		t.Fatal("secret prompt input must not appear in the rendered form")
	}
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)

	visiblePrompt := <-programMessages
	visiblePromptMsg, ok := visiblePrompt.(authPromptMsg)
	if !ok {
		t.Fatalf("visible prompt message = %T, want authPromptMsg", visiblePrompt)
	}
	updated, _ = m.Update(visiblePrompt)
	m = updated.(model)
	if m.authPromptText != "Confirm code:" || m.passwordInput.EchoMode != textinput.EchoNormal {
		t.Fatal("visible auth prompt did not apply its label or echo mode")
	}
	m.passwordInput.SetValue("visible-response")
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)

	result := <-commandDone
	if _, ok := result.(authSuccessMsg); !ok {
		t.Fatalf("authentication result = %T, want authSuccessMsg", result)
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("fake greetd: %v", err)
	}
	if promptMsg.reply == nil || visiblePromptMsg.reply == nil {
		t.Fatal("prompt has no reply channel")
	}
}

func TestAuthPromptEscapeCancelsWithoutSendingInput(t *testing.T) {
	reply := make(chan authPromptResult, 1)
	m := model{
		mode:            ModePassword,
		focusState:      FocusPassword,
		passwordInput:   textinput.New(),
		authPromptType:  ipc.AuthMessageTypeSecret,
		authPromptText:  "OTP:",
		authPromptReply: reply,
	}
	m.passwordInput.SetValue("partial")

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(model)
	result := <-reply
	if !result.cancelled || result.response != nil {
		t.Fatal("Escape must cancel without returning prompt input")
	}
	if m.mode != ModeLoading || m.passwordInput.Value() != "" || m.authPromptReply != nil {
		t.Fatal("Escape must clear prompt state while authentication cancellation completes")
	}
}

func TestAuthPromptInterruptionsDoNotCountAsFailedAttempts(t *testing.T) {
	for _, interruption := range []error{errAuthPromptCancelled, errAuthPromptTimeout} {
		t.Run(interruption.Error(), func(t *testing.T) {
			m := model{
				mode:            ModePassword,
				passwordInput:   textinput.New(),
				authPromptText:  "One-time code:",
				authPromptReply: make(chan authPromptResult, 1),
			}

			updated, _ := m.Update(interruption)
			if got := updated.(model).failedAttempts; got != 0 {
				t.Fatalf("failed attempts = %d after local prompt interruption, want 0", got)
			}
		})
	}
}

func TestAuthPromptVisibleInputIsEchoedAndSanitized(t *testing.T) {
	reply := make(chan authPromptResult, 1)
	m := model{mode: ModeLoading, passwordInput: textinput.New()}
	updated, _ := m.Update(authPromptMsg{
		message: ipc.AuthMessage{
			AuthMessageType: ipc.AuthMessageTypeVisible,
			AuthMessage:     "Verification code:\x1b[2J\n",
		},
		reply: reply,
	})
	m = updated.(model)
	if m.mode != ModePassword || m.authPromptText != "Verification code: " || m.passwordInput.EchoMode != textinput.EchoNormal {
		t.Fatal("visible prompt must display sanitized text and echo input")
	}
	if got := m.renderMainHelpFit(80); got != "Enter Submit • Esc Cancel" {
		t.Fatalf("auth prompt help = %q, want only submit and cancel controls", got)
	}
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyF1})
	m = updated.(model)
	if m.mode != ModePassword {
		t.Fatal("function keys must not leave a pending authentication prompt")
	}
	m.passwordInput.SetValue("123456")
	if !strings.Contains(stripAnsi(m.renderMainForm(80)), "123456") {
		t.Fatal("visible prompt input must be echoed")
	}
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(model)
	result := <-reply
	if result.cancelled || result.response == nil || *result.response != "123456" {
		t.Fatal("Enter must return visible prompt input to greetd")
	}
}

func TestAuthPromptTimeoutReturnsWithoutResponse(t *testing.T) {
	programMessages := make(chan tea.Msg, 1)
	m := model{sendMessage: func(msg tea.Msg) { programMessages <- msg }}
	_, err := m.requestAuthPromptWithin(ipc.AuthMessage{
		AuthMessageType: ipc.AuthMessageTypeSecret,
		AuthMessage:     "OTP:",
	}, time.Millisecond)
	if !errors.Is(err, errAuthPromptTimeout) {
		t.Fatal("prompt timeout must return the timeout error")
	}
	if _, ok := (<-programMessages).(authPromptMsg); !ok {
		t.Fatal("timed-out auth prompt was not sent to the UI")
	}
}

func TestAuthenticateAcceptsSuccessWithoutAuthMessages(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "greetd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	t.Setenv("GREETD_SOCK", socketPath)

	client, err := ipc.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()
		var create ipc.CreateSession
		if err := readAuthTestFrame(conn, &create); err != nil {
			serverDone <- err
			return
		}
		if err := writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse}); err != nil {
			serverDone <- err
			return
		}
		var start ipc.StartSession
		if err := readAuthTestFrame(conn, &start); err != nil {
			serverDone <- err
			return
		}
		serverDone <- writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse})
	}()

	m := model{
		ipcClient:       client,
		selectedSession: &sessions.Session{Name: "Fingerprint", Exec: "/bin/true"},
	}
	result := m.authenticate("test-user", "unused-password")()
	if _, ok := result.(authSuccessMsg); !ok {
		t.Fatalf("authentication result = %T, want authSuccessMsg", result)
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("fake greetd: %v", err)
	}
}

func TestAuthenticateEscapeSendsCancelSession(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "greetd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	t.Setenv("GREETD_SOCK", socketPath)

	client, err := ipc.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()
		var create ipc.CreateSession
		if err := readAuthTestFrame(conn, &create); err != nil {
			serverDone <- err
			return
		}
		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeSecret,
			AuthMessage:     "Password:",
		}); err != nil {
			serverDone <- err
			return
		}
		var passwordResponse ipc.PostAuthMessageResponse
		if err := readAuthTestFrame(conn, &passwordResponse); err != nil {
			serverDone <- err
			return
		}
		if passwordResponse.Response == nil || *passwordResponse.Response != "typed-password" {
			serverDone <- errors.New("first secret prompt did not receive the typed password")
			return
		}
		if err := writeAuthTestFrame(conn, ipc.AuthMessage{
			Type:            ipc.AuthMessageResponse,
			AuthMessageType: ipc.AuthMessageTypeSecret,
			AuthMessage:     "One-time code:",
		}); err != nil {
			serverDone <- err
			return
		}
		var cancel ipc.CancelSession
		if err := readAuthTestFrame(conn, &cancel); err != nil {
			serverDone <- err
			return
		}
		if cancel.Type != ipc.CancelSessionRequest {
			serverDone <- errors.New("Escape did not send cancel_session")
			return
		}
		serverDone <- writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse})
	}()

	programMessages := make(chan tea.Msg, 1)
	m := model{
		ipcClient:       client,
		selectedSession: &sessions.Session{Name: "Test Session", Exec: "/bin/true"},
		passwordInput:   textinput.New(),
		sendMessage:     func(msg tea.Msg) { programMessages <- msg },
	}
	commandDone := make(chan tea.Msg, 1)
	go func() { commandDone <- m.authenticate("test-user", "typed-password")() }()
	prompt := <-programMessages
	updated, _ := m.Update(prompt)
	m = updated.(model)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(model)

	result := <-commandDone
	if err, ok := result.(error); !ok || !strings.Contains(err.Error(), "cancelled") {
		t.Fatal("Escape should return an authentication-cancelled error")
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("fake greetd: %v", err)
	}
}

func TestAuthenticateCanRetryAfterTerminalGreetdError(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "greetd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	t.Setenv("GREETD_SOCK", socketPath)

	client, err := ipc.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()
		for attempt, expectedPassword := range []string{"wrong-password", "right-password"} {
			var create ipc.CreateSession
			if err := readAuthTestFrame(conn, &create); err != nil {
				serverDone <- err
				return
			}
			if err := writeAuthTestFrame(conn, ipc.AuthMessage{
				Type:            ipc.AuthMessageResponse,
				AuthMessageType: ipc.AuthMessageTypeSecret,
				AuthMessage:     "Password:",
			}); err != nil {
				serverDone <- err
				return
			}
			var response ipc.PostAuthMessageResponse
			if err := readAuthTestFrame(conn, &response); err != nil {
				serverDone <- err
				return
			}
			if response.Response == nil || *response.Response != expectedPassword {
				serverDone <- errors.New("retry password response did not match the attempt")
				return
			}
			if attempt == 0 {
				if err := writeAuthTestFrame(conn, ipc.Error{
					Type:        ipc.ErrorResponse,
					ErrorType:   "auth_error",
					Description: "incorrect password",
				}); err != nil {
					serverDone <- err
					return
				}
				var cancel ipc.CancelSession
				if err := readAuthTestFrame(conn, &cancel); err != nil {
					serverDone <- err
					return
				}
				if cancel.Type != ipc.CancelSessionRequest {
					serverDone <- errors.New("failed authentication did not cancel its session")
					return
				}
				if err := writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse}); err != nil {
					serverDone <- err
					return
				}
				continue
			}
			if err := writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse}); err != nil {
				serverDone <- err
				return
			}
			var start ipc.StartSession
			if err := readAuthTestFrame(conn, &start); err != nil {
				serverDone <- err
				return
			}
			serverDone <- writeAuthTestFrame(conn, ipc.Success{Type: ipc.SuccessResponse})
			return
		}
	}()

	m := model{
		ipcClient:       client,
		selectedSession: &sessions.Session{Name: "Test Session", Exec: "/bin/true"},
	}
	first := m.authenticate("test-user", "wrong-password")()
	if _, ok := first.(error); !ok {
		t.Fatal("wrong password must return an authentication error")
	}
	second := m.authenticate("test-user", "right-password")()
	if _, ok := second.(authSuccessMsg); !ok {
		t.Fatalf("retry result = %T, want authSuccessMsg", second)
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("fake greetd: %v", err)
	}
}
