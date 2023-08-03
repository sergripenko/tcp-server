// Package domain - contains domain entities.
package domain

import (
	"fmt"
	"strings"
)

const (
	// MessageTypeQuit - quit message type.
	MessageTypeQuit = "quit"
	// MessageTypeRequestChallenge - request challenge message type (client -> server).
	MessageTypeRequestChallenge = "request_challenge"
	// MessageTypeResponseChallenge - response challenge message type (server -> client).
	MessageTypeResponseChallenge = "response_challenge"
	// MessageTypeRequestResource - request resource message type (client -> server).
	MessageTypeRequestResource = "request_resource"
	// MessageTypeResponseSource - response source message type (server -> client).
	MessageTypeResponseSource = "response_source"
)

// Message - message struct for both server and client.
type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func isMsgTypeValid(msgType string) bool {
	if !strings.EqualFold(msgType, MessageTypeQuit) &&
		!strings.EqualFold(msgType, MessageTypeRequestChallenge) &&
		!strings.EqualFold(msgType, MessageTypeResponseChallenge) &&
		!strings.EqualFold(msgType, MessageTypeRequestResource) &&
		!strings.EqualFold(msgType, MessageTypeResponseSource) {
		return false
	}

	return true
}

// String - stringify message to send it by tcp-connection.
func (m *Message) String() string {
	return fmt.Sprintf("%s|%s\n", m.Type, m.Payload)
}

// ParseMessage - parses Message from str, checks type and payload.
func ParseMessage(str string) (*Message, error) {
	str = strings.TrimSpace(str)

	parts := strings.Split(str, "|")

	if len(parts) < 1 || len(parts) > 2 {
		return nil, fmt.Errorf("invalid format")
	}

	if !isMsgTypeValid(parts[0]) {
		return nil, fmt.Errorf("invalid type")
	}

	msg := Message{
		Type: parts[0],
	}

	if len(parts) == 2 {
		msg.Payload = parts[1]
	}

	return &msg, nil
}
