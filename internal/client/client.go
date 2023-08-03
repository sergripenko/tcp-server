// Package client - implements TCP-client.
package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/exp/slog"

	"tcp-server/internal/domain"
	"tcp-server/pkg/hashcash"
)

// Client - client for TCP server.
type Client struct {
	address       string
	maxIterations int
	logger        *slog.Logger
}

// NewClient - constructor for Client.
func NewClient(address string, maxIterations int) *Client {
	return &Client{
		address:       address,
		maxIterations: maxIterations,
		logger:        slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

// Run - main function, launches client to connect to server and handle new connections.
func (c *Client) Run() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}

	c.logger.Info("Connected to", slog.String("address", c.address))

	defer conn.Close() //nolint: errcheck

	for {
		message, err := c.handleConnection(conn)
		if err != nil {
			return err
		}

		c.logger.Info("Quote result:", slog.String("quote", message))
		time.Sleep(5 * time.Second)
	}
}

func (c *Client) handleConnection(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)

	// Requesting challenge.
	msg := &domain.Message{
		Type: domain.MessageTypeRequestChallenge,
	}

	_, err := conn.Write([]byte(msg.String()))
	if err != nil {
		c.logger.Error("Write msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("send request: %w", err)
	}

	// Reading and parsing response.
	msgStr, err := reader.ReadString('\n')
	if err != nil {
		c.logger.Error("Read msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("read msg: %w", err)
	}

	msg, err = domain.ParseMessage(msgStr)
	if err != nil {
		c.logger.Error("Parse msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("parse msg: %w", err)
	}

	var hashcashData *hashcash.Data

	err = json.Unmarshal([]byte(msg.Payload), &hashcashData)
	if err != nil {
		c.logger.Error("Unmarshal msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("unmarshal msg")
	}

	hashcashData, err = hashcashData.ComputeHashcash(c.maxIterations)
	if err != nil {
		c.logger.Error("Compute hashcash:", slog.String("err", err.Error()))

		return "", fmt.Errorf("compute hashcash: %w", err)
	}

	bytesData, err := json.Marshal(hashcashData)
	if err != nil {
		c.logger.Error("Marshal hashcash:", slog.String("err", err.Error()))

		return "", fmt.Errorf("marshal hashcash: %w", err)
	}

	msg = &domain.Message{
		Type:    domain.MessageTypeRequestResource,
		Payload: string(bytesData),
	}

	_, err = conn.Write([]byte(msg.String()))
	if err != nil {
		c.logger.Error("Write msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("send request: %w", err)
	}

	// Get result quote.
	msgStr, err = reader.ReadString('\n')
	if err != nil {
		c.logger.Error("Read msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("read msg: %w", err)
	}

	msg, err = domain.ParseMessage(msgStr)
	if err != nil {
		c.logger.Error("Parse msg:", slog.String("err", err.Error()))

		return "", fmt.Errorf("parse msg: %w", err)
	}

	return msg.Payload, nil
}
