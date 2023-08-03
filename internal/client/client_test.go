package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"tcp-server/internal/domain"
	"tcp-server/pkg/hashcash"
)

type mockConn struct {
	ReadFunc             func([]byte) (int, error)
	WriteFunc            func([]byte) (int, error)
	CloseFunc            func() error
	LocalAddrFunc        func() net.Addr
	RemoteAddrFunc       func() net.Addr
	SetDeadlineFunc      func(time.Time) error
	SetReadDeadlineFunc  func(time.Time) error
	SetWriteDeadlineFunc func(time.Time) error
}

func (m mockConn) Close() error {
	return m.CloseFunc()
}

func (m mockConn) LocalAddr() net.Addr {
	return m.LocalAddrFunc()
}

func (m mockConn) RemoteAddr() net.Addr {
	return m.RemoteAddrFunc()
}

func (m mockConn) SetDeadline(t time.Time) error {
	return m.SetDeadlineFunc(t)
}

func (m mockConn) SetReadDeadline(t time.Time) error {
	return m.SetReadDeadlineFunc(t)
}

func (m mockConn) SetWriteDeadline(t time.Time) error {
	return m.SetWriteDeadlineFunc(t)
}

func (m mockConn) Read(p []byte) (n int, err error) {
	return m.ReadFunc(p)
}

func (m mockConn) Write(p []byte) (n int, err error) {
	return m.WriteFunc(p)
}

func TestClient_HandleConnectionWriteError(t *testing.T) {
	client := NewClient("localhost:8080", 1000000)

	mock := mockConn{
		WriteFunc: func(p []byte) (int, error) {
			return 0, fmt.Errorf("mock write error")
		},
	}

	_, err := client.handleConnection(mock)
	assert.Error(t, err)
	assert.Equal(t, "send request: mock write error", err.Error())
}

func TestClient_HandleConnectionReadError(t *testing.T) {
	client := NewClient("localhost:8080", 1000000)

	mock := mockConn{
		WriteFunc: func(p []byte) (int, error) {
			return 0, nil
		},
		ReadFunc: func(p []byte) (int, error) {
			return 0, fmt.Errorf("mock read error")
		},
	}

	_, err := client.handleConnection(mock)
	assert.Error(t, err)
	assert.Equal(t, "read msg: mock read error", err.Error())
}

func TestClient_HandleConnectionParseMsgError(t *testing.T) {
	client := NewClient("localhost:8080", 1000000)

	mock := mockConn{
		WriteFunc: func(p []byte) (int, error) {
			return 0, nil
		},
		ReadFunc: func(p []byte) (int, error) {
			return getReadMock("invalid|msg|\n", p), nil
		},
	}

	_, err := client.handleConnection(mock)
	assert.Error(t, err)
	assert.Equal(t, "parse msg: invalid format", err.Error())
}

func TestClient_HandleConnectionInvalidPayload(t *testing.T) {
	client := NewClient("localhost:8080", 1000000)

	mock := mockConn{
		WriteFunc: func(p []byte) (int, error) {
			return 0, nil
		},
		ReadFunc: func(p []byte) (int, error) {
			return getReadMock(fmt.Sprintf("%s|%s\n", domain.MessageTypeResponseChallenge, "ivalid payload"), p), nil
		},
	}

	_, err := client.handleConnection(mock)
	assert.Error(t, err)
	assert.Equal(t, "unmarshal msg", err.Error())
}

func TestClient_HandleConnectionOk(t *testing.T) {
	client := NewClient("localhost:8080", 1000000)

	hashcashData := hashcash.Data{
		ZerosCount: 3,
		Date:       time.Now().Unix(),
		Client:     "client1",
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
	}

	var readCount, writeCount int

	mock := mockConn{
		WriteFunc: func(p []byte) (int, error) {
			if writeCount == 0 {
				writeCount++

				assert.Equal(t, fmt.Sprintf("%s|%s\n", domain.MessageTypeRequestChallenge, ""), string(p))
			} else {
				msg, err := domain.ParseMessage(string(p))
				assert.NoError(t, err)

				err = json.Unmarshal([]byte(msg.Payload), &hashcashData)
				assert.NoError(t, err)

				_, err = hashcashData.ComputeHashcash(0)
				assert.NoError(t, err)
			}

			return 0, nil
		},
		ReadFunc: func(p []byte) (int, error) {
			if readCount == 0 {
				readCount++

				bytes, err := json.Marshal(hashcashData)

				assert.NoError(t, err)

				return getReadMock(fmt.Sprintf("%s|%s\n", domain.MessageTypeResponseChallenge, string(bytes)), p), nil
			} else {
				return getReadMock(fmt.Sprintf("%s|%s\n", domain.MessageTypeResponseChallenge, "mock quote"), p), nil
			}
		},
	}

	resp, err := client.handleConnection(mock)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "mock quote", resp)
}

func getReadMock(str string, b []byte) int {
	dataBytes := []byte(str)
	var count int

	for idx := range dataBytes {
		b[idx] = dataBytes[idx]
		count++

		if count >= len(b) {
			break
		}
	}

	return count
}
