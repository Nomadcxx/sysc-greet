package main

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nomadcxx/sysc-greet/internal/ipc"
)

const authPromptTimeout = 60 * time.Second

var (
	errAuthPromptCancelled = errors.New("authentication prompt cancelled")
	errAuthPromptTimeout   = errors.New("authentication prompt timed out")
)

type authNoticeMsg string

type authPromptMsg struct {
	message ipc.AuthMessage
	reply   chan authPromptResult
}

type authPromptResult struct {
	response  *string
	cancelled bool
}

func sanitizeAuthText(message string) string {
	message = ansi.Strip(message)
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, message)
}

func (m model) sendAuthNotice(message string) error {
	if m.sendMessage == nil {
		return errors.New("authentication UI is not available")
	}
	m.sendMessage(authNoticeMsg(sanitizeAuthText(message)))
	return nil
}

func (m model) requestAuthPrompt(message ipc.AuthMessage) (*string, error) {
	return m.requestAuthPromptWithin(message, authPromptTimeout)
}

func (m model) requestAuthPromptWithin(message ipc.AuthMessage, timeout time.Duration) (*string, error) {
	if m.sendMessage == nil {
		return nil, errors.New("authentication UI is not available")
	}

	reply := make(chan authPromptResult, 1)
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	m.sendMessage(authPromptMsg{message: message, reply: reply})

	select {
	case result := <-reply:
		if result.cancelled {
			return nil, errAuthPromptCancelled
		}
		if result.response == nil {
			return nil, errors.New("authentication prompt returned no response")
		}
		return result.response, nil
	case <-timer.C:
		return nil, errAuthPromptTimeout
	}
}

func (m model) handleAuthPromptMessage(message authPromptMsg) model {
	m.mode = ModePassword
	m.focusState = FocusPassword
	m.authPromptType = message.message.AuthMessageType
	m.authPromptText = sanitizeAuthText(message.message.AuthMessage)
	m.authPromptReply = message.reply
	m.passwordInput.SetValue("")
	m.passwordInput.EchoMode = textinput.EchoPassword
	if message.message.AuthMessageType == ipc.AuthMessageTypeVisible {
		m.passwordInput.EchoMode = textinput.EchoNormal
	}
	m.passwordInput.Focus()
	m.usernameInput.Blur()
	m.errorMessage = ""
	m.idleTimer = time.Now()
	return m
}

func (m *model) clearAuthPrompt() {
	m.authPromptType = ""
	m.authPromptText = ""
	m.authPromptReply = nil
	m.passwordInput.EchoMode = textinput.EchoPassword
}

func (m model) handleAuthPromptKey(message tea.KeyMsg) (model, tea.Cmd) {
	reply := m.authPromptReply
	switch message.String() {
	case "enter":
		value := m.passwordInput.Value()
		m.passwordInput.SetValue("")
		m.clearAuthPrompt()
		m.mode = ModeLoading
		reply <- authPromptResult{response: &value}
		return m, nil
	case "esc":
		m.passwordInput.SetValue("")
		m.clearAuthPrompt()
		m.mode = ModeLoading
		reply <- authPromptResult{cancelled: true}
		return m, nil
	default:
		var command tea.Cmd
		m.passwordInput, command = m.passwordInput.Update(message)
		return m, command
	}
}
