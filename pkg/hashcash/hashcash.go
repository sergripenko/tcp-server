// Package hashcash - implements Hashcash algorithm.
package hashcash

import (
	"crypto/sha1" //nolint:gosec
	"errors"
	"fmt"
)

var (
	errMaxIterationsExceeded = errors.New("max iterations exceeded")
)

// Data - struct with fields of Hashcash.
type Data struct {
	ZerosCount int
	Date       int64
	Client     string
	Rand       string
	Counter    int
}

func (d *Data) String() string {
	return fmt.Sprintf("%d:%d:%s:%s:%d", d.ZerosCount, d.Date, d.Client, d.Rand, d.Counter)
}

func sha1Hash(data string) string {
	h := sha1.New() //nolint:gosec
	h.Write([]byte(data))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

func isHashCorrect(hash string, zerosCount int) bool {
	if zerosCount > len(hash) {
		return false
	}

	for _, ch := range hash[:zerosCount] {
		if ch != 48 {
			return false
		}
	}

	return true
}

// ComputeHashcash - calculates correct hashcash by bruteforce.
func (d *Data) ComputeHashcash(maxIterations int) (*Data, error) {
	for d.Counter <= maxIterations || maxIterations <= 0 {
		header := d.String()
		hash := sha1Hash(header)

		if isHashCorrect(hash, d.ZerosCount) {
			return d, nil
		}

		d.Counter++
	}

	return d, errMaxIterationsExceeded
}
