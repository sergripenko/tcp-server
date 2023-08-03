// Package server - implements TCP-server.
package server

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"golang.org/x/exp/slog"

	"tcp-server/internal/config"
	"tcp-server/internal/domain"
	"tcp-server/pkg/hashcash"
)

var (
	errQuit                  = errors.New("client requests to close connection")
	errInvalidHashcashClient = errors.New("invalid hashcash client")
	errInvalidMessageType    = errors.New("invalid message type")
	errInvalidHashcash       = errors.New("invalid hashcash")
	errValueNotFoundInCache  = errors.New("value not found in cache")

	//nolint: lll
	quotes = []string{ //nolint: lll
		`You create your own opportunities. Success doesn’t just come and find you–you have to go out and get it`,
		`Never break your promises. Keep every promise; it makes you credible`,
		`You are never as stuck as you think you are. Success is not final, and failure isn’t fatal`,
		`Happiness is a choice. For every minute you are angry, you lose 60 seconds of your own happiness`,
		`Habits develop into character. Character is the result of our mental attitude and the way we spend our time`,
		`Be happy with who you are. Being happy doesn’t mean everything is perfect but that you have decided to look beyond the imperfections`,
		`Don’t seek happiness–create it. You don’t need life to go your way to be happy`,
		`If you want to be happy, stop complaining. If you want happiness, stop complaining about how your life isn’t what you want and make it into what you do want`,
		`Asking for help is a sign of strength. Don’t let your fear of being judged stop you from asking for help when you need it. Sometimes asking for help is the bravest move you can make. You don’t have to go it alone`,
		`Replace every negative thought with a positive one. A positive mind is stronger than a negative thought`,
	}
)

// Server - main struct of TCP-server.
type Server struct {
	host             string
	port             int64
	zerosCount       int
	hashcashDuration int64
	logger           *slog.Logger
	cache            *cache.Cache[int, any]
}

// NewServer - constructor for Server.
func NewServer(config *config.Config) *Server {
	return &Server{
		host:             config.ServerHost,
		port:             config.ServerPort,
		zerosCount:       config.HashcashZeros,
		hashcashDuration: config.HashcashDuration,
		logger:           slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		cache:            cache.New[int, any](),
	}
}

// Run - main function, launches server to listen on given address and handle new connections.
func (s *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("listen connection: %w", err)
	}

	defer listener.Close() //nolint: errcheck

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	s.logger.Info("New client:", slog.String("address", conn.RemoteAddr().String()))
	defer conn.Close() //nolint: errcheck

	reader := bufio.NewReader(conn)

	for {
		req, err := reader.ReadString('\n')
		if err != nil {
			s.logger.Error("Error read connection:", slog.String("err", err.Error()))

			return
		}

		msg, err := domain.ParseMessage(req)
		if err != nil {
			s.logger.Error("Error parse message:", slog.String("err", err.Error()))

			return
		}

		msg, err = s.processMessage(*msg, conn.RemoteAddr().String())
		if err != nil {
			s.logger.Error("Error process message:", slog.String("err", err.Error()))

			return
		}

		if msg != nil {
			if _, err = conn.Write([]byte(msg.String())); err != nil {
				s.logger.Error("Error write message:", slog.String("err", err.Error()))

				return
			}
		}
	}
}

func (s *Server) processMessage(msg domain.Message, client string) (*domain.Message, error) {
	s.logger.Debug("Process message:",
		slog.String("type", msg.Type),
		slog.String("payload", msg.Payload),
		slog.String("client", client))

	switch msg.Type {
	case domain.MessageTypeQuit:
		return nil, errQuit
	case domain.MessageTypeRequestChallenge:
		return s.handleRequestChallenge(msg, client)
	case domain.MessageTypeRequestResource:
		return s.handleRequestResource(msg, client)
	default:
		return nil, errInvalidMessageType
	}
}

func (s *Server) handleRequestChallenge(msg domain.Message, client string) (*domain.Message, error) {
	randValue, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		s.logger.Error("Error generate rand value", slog.String("err", err.Error()))

		return nil, fmt.Errorf("generate rand value: %w", err)
	}

	// Add rand value with expiration (in seconds) to cache.
	s.cache.Set(int(randValue.Int64()), nil, cache.WithExpiration(time.Second*time.Duration(s.hashcashDuration)))

	hashcashData := hashcash.Data{
		ZerosCount: s.zerosCount,
		Date:       time.Now().Unix(),
		Client:     client,
		Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randValue))),
		Counter:    0,
	}

	hashcashBytes, err := json.Marshal(hashcashData)
	if err != nil {
		s.logger.Error("Error marshal hashcash", slog.String("err", err.Error()))

		return nil, fmt.Errorf("marshal hashcash: %w", err)
	}

	msg = domain.Message{
		Type:    domain.MessageTypeResponseChallenge,
		Payload: string(hashcashBytes),
	}

	return &msg, nil
}

func (s *Server) handleRequestResource(msg domain.Message, client string) (*domain.Message, error) {
	var hashcashData hashcash.Data

	if err := json.Unmarshal([]byte(msg.Payload), &hashcashData); err != nil {
		s.logger.Error("Error unmarshal hashcash", slog.String("err", err.Error()))

		return nil, fmt.Errorf("unmarshal hashcash: %w", err)
	}

	if !strings.EqualFold(hashcashData.Client, client) {
		s.logger.Error("Invalid hashcash client",
			slog.String("request client", client),
			slog.String("hashcash client", hashcashData.Client))

		return nil, errInvalidHashcashClient
	}

	randValueBytes, err := base64.StdEncoding.DecodeString(hashcashData.Rand)
	if err != nil {
		s.logger.Error("Error decode rand value bytes", slog.String("err", err.Error()))

		return nil, fmt.Errorf("decode rand value bytes: %w", err)
	}

	randValue, err := strconv.Atoi(string(randValueBytes))
	if err != nil {
		s.logger.Error("Error convert rand value to int", slog.String("err", err.Error()))

		return nil, fmt.Errorf("convert rand value to int: %w", err)
	}

	// Check if rand value exists in cache.
	if _, ok := s.cache.Get(randValue); !ok {
		s.logger.Error("Rand value not found in cache", slog.Int("rand", randValue))
		return nil, errValueNotFoundInCache
	}

	maxIter := hashcashData.Counter
	if maxIter == 0 {
		maxIter = 1
	}

	_, err = hashcashData.ComputeHashcash(maxIter)
	if err != nil {
		s.logger.Error("Error compute hashcash", slog.String("err", err.Error()))

		return nil, errInvalidHashcash
	}

	randIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(quotes))))
	if err != nil {
		s.logger.Error("Error generate rand value", slog.String("err", err.Error()))

		return nil, fmt.Errorf("generate rand value: %w", err)
	}

	msg = domain.Message{
		Type:    domain.MessageTypeResponseSource,
		Payload: quotes[randIdx.Int64()],
	}

	s.cache.Delete(randValue)

	return &msg, nil
}
