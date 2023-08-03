package hashcash

import (
	"errors"
	"testing"
)

func TestComputeHashcash(t *testing.T) {
	data := &Data{
		ZerosCount: 5,
		Date:       1629927217,
		Client:     "client",
		Rand:       "random",
		Counter:    0,
	}

	// This test is for a success case; you might need to adjust maxIterations and zerosCount depending on your use case.
	result, err := data.ComputeHashcash(1000000)
	if err != nil {
		t.Errorf("did not expect an error but got one: %v", err)
	}
	if !isHashCorrect(sha1Hash(result.String()), data.ZerosCount) {
		t.Errorf("resulting hash is not correct")
	}

	// Test for failure due to max iterations exceeded
	data.Counter = 0
	_, err = data.ComputeHashcash(5)

	if !errors.Is(err, errMaxIterationsExceeded) {
		t.Errorf("expected error %v, got %v", errMaxIterationsExceeded, err)
	}
}

func TestSha1Hash(t *testing.T) {
	data := "test_data"
	expected := "4f20c649228a94d3cc4d31e9d12ec593e20c0202" // SHA-1 hash for "test_data"

	actual := sha1Hash(data)
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestIsHashCorrect(t *testing.T) {
	tests := []struct {
		hash       string
		zerosCount int
		correct    bool
	}{
		{"00000abcd1234", 5, true},
		{"000abcd1234", 5, false},
		{"00000abcd1234", 10, false}, // this will cover the condition where zerosCount is greater than the length of the hash
		{"00000abcd1234", 14, false}, // additional case to explicitly cover the condition
	}

	for _, test := range tests {
		actual := isHashCorrect(test.hash, test.zerosCount)

		if actual != test.correct {
			t.Errorf("expected %v for hash %s with zerosCount %d, got %v", test.correct, test.hash, test.zerosCount, actual)
		}
	}
}
