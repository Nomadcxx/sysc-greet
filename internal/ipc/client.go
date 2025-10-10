package ipc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

// Message types for requests and responses
type RequestType string
type ResponseType string

const (
	CreateSessionRequest           RequestType = "create_session"
	PostAuthMessageResponseRequest RequestType = "post_auth_message_response"
	StartSessionRequest            RequestType = "start_session"
	CancelSessionRequest           RequestType = "cancel_session"

	SuccessResponse     ResponseType = "success"
	ErrorResponse       ResponseType = "error"
	AuthMessageResponse ResponseType = "auth_message"
)

// Request structures
type CreateSession struct {
	Type     RequestType `json:"type"`
	Username string      `json:"username"`
}

type PostAuthMessageResponse struct {
	Type     RequestType `json:"type"`
	Response *string     `json:"response"`
}

type StartSession struct {
	Type RequestType `json:"type"`
	Cmd  []string    `json:"cmd"`
	Env  []string    `json:"env"`
}

type CancelSession struct {
	Type RequestType `json:"type"`
}

// Response structures
type Success struct {
	Type ResponseType `json:"type"`
}

type Error struct {
	Type        ResponseType `json:"type"`
	ErrorType   string       `json:"error_type"`
	Description string       `json:"description"`
}

type AuthMessage struct {
	Type            ResponseType `json:"type"`
	AuthMessageType string       `json:"auth_message_type"`
	AuthMessage     string       `json:"auth_message"`
}

// Client for communicating with greetd
type Client struct {
	conn net.Conn
}

// NewClient creates a new IPC client connected to greetd
// CHANGED 2025-10-02 03:00 - Use GREETD_SOCK environment variable - Problem: greetd sets socket path dynamically
func NewClient() (*Client, error) {
	socketPath := os.Getenv("GREETD_SOCK")
	if socketPath == "" {
		return nil, fmt.Errorf("GREETD_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to greetd socket at %s: %v", socketPath, err)
	}

	return &Client{conn: conn}, nil
}

// Close closes the connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// SendRequest sends a request to greetd
func (c *Client) SendRequest(req interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBytes, length)

	_, err = c.conn.Write(lengthBytes)
	if err != nil {
		return fmt.Errorf("failed to write length: %v", err)
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	return nil
}

// ReceiveResponse receives a response from greetd
func (c *Client) ReceiveResponse() (interface{}, error) {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(c.conn, lengthBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read length: %v", err)
	}

	length := binary.LittleEndian.Uint32(lengthBytes)
	data := make([]byte, length)
	_, err = io.ReadFull(c.conn, data)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %v", err)
	}

	// Determine response type from JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	respType, ok := raw["type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response: missing type")
	}

	switch ResponseType(respType) {
	case SuccessResponse:
		var resp Success
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		return resp, nil
	case ErrorResponse:
		var resp Error
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		return resp, nil
	case AuthMessageResponse:
		var resp AuthMessage
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown response type: %s", respType)
	}
}

// CreateSession creates a new session for the given username
func (c *Client) CreateSession(username string) error {
	req := CreateSession{
		Type:     CreateSessionRequest,
		Username: username,
	}
	return c.SendRequest(req)
}

// PostAuthMessageResponse sends a response to an auth message
func (c *Client) PostAuthMessageResponse(response *string) error {
	req := PostAuthMessageResponse{
		Type:     PostAuthMessageResponseRequest,
		Response: response,
	}
	return c.SendRequest(req)
}

// StartSession starts the session with the given command and environment
// CHANGED 2025-10-09 21:00 - Wait for greetd's response to StartSession - Problem: Infinite loop on slow machines
// According to greetd protocol, we must wait for greetd to confirm session start before greeter exits
func (c *Client) StartSession(cmd []string, env []string) error {
	req := StartSession{
		Type: StartSessionRequest,
		Cmd:  cmd,
		Env:  env,
	}

	// Send the start session request
	if err := c.SendRequest(req); err != nil {
		return err
	}

	// CRITICAL: Wait for greetd to respond with success or error
	// This ensures greetd has properly initialized the session before greeter exits
	// Without this, on slow hardware, greeter exits before greetd finishes setup -> infinite loop
	resp, err := c.ReceiveResponse()
	if err != nil {
		return fmt.Errorf("failed to receive start session response: %v", err)
	}

	// Check if session started successfully
	if _, ok := resp.(Success); ok {
		return nil // Session successfully started
	}

	if errResp, ok := resp.(Error); ok {
		return fmt.Errorf("failed to start session: %s - %s", errResp.ErrorType, errResp.Description)
	}

	return fmt.Errorf("unexpected response to start_session: %T", resp)
}

// CancelSession cancels the current session
func (c *Client) CancelSession() error {
	req := CancelSession{
		Type: CancelSessionRequest,
	}
	return c.SendRequest(req)
}
