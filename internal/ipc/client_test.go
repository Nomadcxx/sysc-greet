package ipc

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"testing"
)

func writeTestFrame(conn net.Conn, value any) error {
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

func readTestAuthResponse(conn net.Conn) (PostAuthMessageResponse, error) {
	var size [4]byte
	if _, err := io.ReadFull(conn, size[:]); err != nil {
		return PostAuthMessageResponse{}, err
	}
	data := make([]byte, binary.LittleEndian.Uint32(size[:]))
	if _, err := io.ReadFull(conn, data); err != nil {
		return PostAuthMessageResponse{}, err
	}
	var response PostAuthMessageResponse
	err := json.Unmarshal(data, &response)
	return response, err
}

func TestClientConverseAnswersInfoWithNullThenFirstSecret(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	serverErr := make(chan error, 1)
	go func() {
		if err := writeTestFrame(serverConn, AuthMessage{
			Type:            AuthMessageResponse,
			AuthMessageType: "info",
			AuthMessage:     "Swipe your finger or enter your password",
		}); err != nil {
			serverErr <- err
			return
		}
		infoResponse, err := readTestAuthResponse(serverConn)
		if err != nil {
			serverErr <- err
			return
		}
		if infoResponse.Response != nil {
			serverErr <- errUnexpectedResponse(infoResponse.Response)
			return
		}

		if err := writeTestFrame(serverConn, AuthMessage{
			Type:            AuthMessageResponse,
			AuthMessageType: "secret",
			AuthMessage:     "Password:",
		}); err != nil {
			serverErr <- err
			return
		}
		secretResponse, err := readTestAuthResponse(serverConn)
		if err != nil {
			serverErr <- err
			return
		}
		if secretResponse.Response == nil || *secretResponse.Response != "typed-password" {
			serverErr <- errUnexpectedResponse(secretResponse.Response)
			return
		}

		serverErr <- writeTestFrame(serverConn, Success{Type: SuccessResponse})
	}()

	client := &Client{conn: clientConn}
	terminal, err := client.Converse(func(message AuthMessage) (*string, error) {
		switch message.AuthMessageType {
		case "info":
			return nil, nil
		case "secret":
			password := "typed-password"
			return &password, nil
		default:
			t.Fatalf("unexpected auth message type %q", message.AuthMessageType)
			return nil, nil
		}
	})
	if err != nil {
		t.Fatalf("Converse() error = %v", err)
	}
	if _, ok := terminal.(Success); !ok {
		t.Fatalf("Converse() terminal = %T, want Success", terminal)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("fake greetd: %v", err)
	}
}

func errUnexpectedResponse(response *string) error {
	if response == nil {
		return io.ErrUnexpectedEOF
	}
	return &unexpectedResponseError{response: *response}
}

type unexpectedResponseError struct {
	response string
}

func (e *unexpectedResponseError) Error() string {
	return "unexpected auth response: " + e.response
}
