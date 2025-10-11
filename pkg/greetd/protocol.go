package greetd

// Created greetd IPC protocol implementation in Go
// Based on greetd-ipc protocol specification and tuigreet Rust implementation

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// Request types sent to greetd
type Request struct {
	Type string `json:"type"`
	// CreateSession fields
	Username string `json:"username,omitempty"`
	// PostAuthMessageResponse fields
	Response *string `json:"response,omitempty"`
	// StartSession fields
	Cmd []string `json:"cmd,omitempty"`
	Env []string `json:"env,omitempty"`
}

// Response types received from greetd
type Response struct {
	Type string `json:"type"`
	// AuthMessage fields
	AuthMessageType *string `json:"auth_message_type,omitempty"`
	AuthMessage     *string `json:"auth_message,omitempty"`
	// Error fields
	ErrorType    *string `json:"error_type,omitempty"`
	Description  *string `json:"description,omitempty"`
}

// AuthMessageType constants
const (
	AuthMessageSecret  = "secret"
	AuthMessageVisible = "visible"
	AuthMessageInfo    = "info"
	AuthMessageError   = "error"
)

// ErrorType constants
const (
	ErrorTypeAuthError = "auth_error"
	ErrorTypeError     = "error"
)

// Client manages connection to greetd daemon
type Client struct {
	conn net.Conn
}

// NewClient connects to greetd Unix socket
func NewClient(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to greetd socket: %w", err)
	}
	return &Client{conn: conn}, nil
}

// Close closes the connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Send sends a request to greetd
func (c *Client) Send(req Request) error {
	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write length prefix (4 bytes, native endian)
	length := uint32(len(data))
	if err := binary.Write(c.conn, binary.NativeEndian, length); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	// Write JSON payload
	if _, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// Receive receives a response from greetd
func (c *Client) Receive() (*Response, error) {
	// Read length prefix
	var length uint32
	if err := binary.Read(c.conn, binary.NativeEndian, &length); err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}

	// Read JSON payload
	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal JSON
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Helper functions to create requests

func CreateSessionRequest(username string) Request {
	return Request{
		Type:     "create_session",
		Username: username,
	}
}

func PostAuthMessageResponseRequest(response string) Request {
	return Request{
		Type:     "post_auth_message_response",
		Response: &response,
	}
}

func CancelSessionRequest() Request {
	return Request{
		Type: "cancel_session",
	}
}

func StartSessionRequest(cmd []string, env []string) Request {
	return Request{
		Type: "start_session",
		Cmd:  cmd,
		Env:  env,
	}
}

// Helper functions to check response types

func (r *Response) IsAuthMessage() bool {
	return r.Type == "auth_message"
}

func (r *Response) IsSuccess() bool {
	return r.Type == "success"
}

func (r *Response) IsError() bool {
	return r.Type == "error"
}

func (r *Response) IsSecretRequest() bool {
	return r.IsAuthMessage() && r.AuthMessageType != nil && *r.AuthMessageType == AuthMessageSecret
}

func (r *Response) IsVisibleRequest() bool {
	return r.IsAuthMessage() && r.AuthMessageType != nil && *r.AuthMessageType == AuthMessageVisible
}

func (r *Response) IsAuthError() bool {
	return r.IsError() && r.ErrorType != nil && *r.ErrorType == ErrorTypeAuthError
}
