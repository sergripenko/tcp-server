package domain

import (
	"fmt"
	"testing"
)

func TestIsMsgTypeValid(t *testing.T) {
	tests := []struct {
		msgType string
		valid   bool
	}{
		{MessageTypeQuit, true},
		{MessageTypeRequestChallenge, true},
		{MessageTypeResponseChallenge, true},
		{MessageTypeRequestResource, true},
		{MessageTypeResponseSource, true},
		{"invalid_type", false},
	}

	for _, test := range tests {
		actual := isMsgTypeValid(test.msgType)

		if actual != test.valid {
			t.Errorf("expected validity of type %s to be %v, got %v", test.msgType, test.valid, actual)
		}
	}
}

func TestMessageString(t *testing.T) {
	msg := Message{
		Type:    MessageTypeRequestChallenge,
		Payload: "payload_data",
	}

	expected := "request_challenge|payload_data\n"
	actual := msg.String()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		input  string
		output *Message
		err    error
	}{
		{
			input: "request_challenge|payload_data",
			output: &Message{
				Type:    MessageTypeRequestChallenge,
				Payload: "payload_data",
			},
			err: nil,
		},
		{
			input:  "invalid|format|",
			output: nil,
			err:    fmt.Errorf("invalid format"),
		},
		{
			input:  "invalid_type|payload_data",
			output: nil,
			err:    fmt.Errorf("invalid type"),
		},
	}

	for _, test := range tests {
		actual, err := ParseMessage(test.input)

		switch {
		case err != nil && test.err == nil:
			t.Errorf("did not expect an error but got one: %v", err)
		case err == nil && test.err != nil:
			t.Errorf("expected an error but got none")
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("expected error %v, got %v", test.err, err)
		}

		if actual != nil && test.output != nil && (*actual != *test.output) {
			t.Errorf("expected %v, got %v", *test.output, *actual)
		}
	}
}
