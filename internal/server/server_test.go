package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/stretchr/testify/assert"

	"tcp-server/internal/config"
	"tcp-server/internal/domain"
	"tcp-server/pkg/hashcash"
)

func TestServer_HandleRequestChallenge(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         4,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	msg := domain.Message{Type: domain.MessageTypeRequestChallenge}
	client := "client-123"

	resp, err := server.handleRequestChallenge(msg, client)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, domain.MessageTypeResponseChallenge, resp.Type)
}

func TestServer_HandleRequestResource(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         3,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	server.cache.Set(123460, nil, cache.WithExpiration(time.Second*time.Duration(server.hashcashDuration)))

	hashcashData := &hashcash.Data{
		ZerosCount: 3,
		Date:       time.Now().Unix(),
		Client:     "client1",
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
	}

	// Imitate hashcash.ComputeHashcash() on client side.
	hashcashData, err := hashcashData.ComputeHashcash(1000000)
	assert.NoError(t, err)

	dataBytes, err := json.Marshal(hashcashData)
	assert.NoError(t, err)

	msg := domain.Message{Type: domain.MessageTypeRequestResource, Payload: string(dataBytes)}
	client := "client1"

	resp, err := server.handleRequestResource(msg, client)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, domain.MessageTypeResponseSource, resp.Type)
	assert.Contains(t, quotes, resp.Payload)
}

func TestServer_HandleRequestResourceWithWrongClient(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         3,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	server.cache.Set(123460, nil, cache.WithExpiration(time.Second*time.Duration(server.hashcashDuration)))

	hashcashData := &hashcash.Data{
		ZerosCount: 3,
		Date:       time.Now().Unix(),
		Client:     "client1",
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
	}

	dataBytes, err := json.Marshal(hashcashData)
	assert.NoError(t, err)

	msg := &domain.Message{Type: domain.MessageTypeRequestResource, Payload: string(dataBytes)}

	msg, err = server.handleRequestResource(*msg, "client2")
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, errInvalidHashcashClient, err)
}

func TestServer_HandleRequestResourceWithZeroCounter(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         3,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	server.cache.Set(123460, nil, cache.WithExpiration(time.Second*time.Duration(server.hashcashDuration)))

	hashcashData := &hashcash.Data{
		ZerosCount: 3,
		Date:       time.Now().Unix(),
		Client:     "client1",
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
		Counter:    0,
	}

	dataBytes, err := json.Marshal(hashcashData)
	assert.NoError(t, err)

	msg := &domain.Message{Type: domain.MessageTypeRequestResource, Payload: string(dataBytes)}

	msg, err = server.handleRequestResource(*msg, "client1")
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.ErrorIs(t, err, errInvalidHashcash)
}

func TestServer_HandleRequestResourceWithCacheExpired(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         3,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	hashcashData := &hashcash.Data{
		ZerosCount: 3,
		Date:       time.Now().Unix(),
		Client:     "client1",
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", 123460))),
	}

	dataBytes, err := json.Marshal(hashcashData)
	assert.NoError(t, err)

	msg := &domain.Message{Type: domain.MessageTypeRequestResource, Payload: string(dataBytes)}

	msg, err = server.handleRequestResource(*msg, "client1")
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.ErrorIs(t, err, errValueNotFoundInCache)
}

func TestServer_ProcessMessage_Quit(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         4,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	msg := domain.Message{Type: domain.MessageTypeQuit}
	client := "client-123"

	resp, err := server.processMessage(msg, client)
	assert.Nil(t, resp)
	assert.Equal(t, errQuit, err)
}

func TestServer_ProcessMessage_Invalid(t *testing.T) {
	server := NewServer(&config.Config{
		ServerHost:            "localhost",
		ServerPort:            8080,
		HashcashZeros:         4,
		HashcashDuration:      30,
		HashcashMaxIterations: 1000000,
	})

	msg := domain.Message{Type: "invalid_type"}
	client := "client-123"

	resp, err := server.processMessage(msg, client)
	assert.Nil(t, resp)
	assert.Equal(t, errInvalidMessageType, err)
}
