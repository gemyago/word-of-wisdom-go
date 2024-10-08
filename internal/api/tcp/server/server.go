package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"strings"
	"time"
	"word-of-wisdom-go/internal/diag"
	"word-of-wisdom-go/internal/services"

	"go.uber.org/dig"
)

type commandHandler interface {
	Handle(ctx context.Context, session *services.SessionIO) error
}

type ListenerDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Port               int           `name:"config.tcpServer.port"`
	MaxSessionDuration time.Duration `name:"config.tcpServer.maxSessionDuration"`

	// components
	Handler commandHandler

	// services
	services.UUIDGenerator
}

type Listener struct {
	logger             *slog.Logger
	listener           net.Listener
	commandHandler     commandHandler
	port               int
	maxSessionDuration time.Duration
	listeningSignal    chan struct{}
	uuidGenerator      services.UUIDGenerator
}

func NewListener(deps ListenerDeps) *Listener {
	return &Listener{
		port:               deps.Port,
		maxSessionDuration: deps.MaxSessionDuration,
		commandHandler:     deps.Handler,
		logger:             deps.RootLogger.WithGroup("tcp.server"),
		listeningSignal:    make(chan struct{}),
		uuidGenerator:      deps.UUIDGenerator,
	}
}

func extractHost(addr string) string {
	sepIndex := strings.Index(addr, ":")
	if sepIndex >= 0 {
		return addr[:sepIndex]
	}
	return addr
}

func (l *Listener) processAcceptedConnection(ctx context.Context, c net.Conn) {
	// This can be transformed into a middleware like approach
	ctx = diag.SetLogAttributesToContext(
		ctx, diag.LogAttributes{CorrelationID: slog.StringValue(l.uuidGenerator())},
	)
	defer func() {
		if rvr := recover(); rvr != nil {
			l.logger.ErrorContext(
				ctx,
				"Unhandled panic",
				slog.Any("panic", rvr),
				slog.String("stack", string(debug.Stack())),
			)
			c.Close()
		}
	}()
	deadline := time.Now().Add(l.maxSessionDuration)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	remoteAddr := c.RemoteAddr().String()
	if err := c.SetDeadline(deadline); err != nil {
		l.logger.ErrorContext(ctx,
			"Failed to set connection deadline",
			diag.ErrAttr(err),
			slog.String("remoteAddr", remoteAddr),
		)
		return
	}

	l.logger.InfoContext(ctx, "Connection accepted", slog.String("remoteAddr", remoteAddr))
	defer c.Close()
	session := services.NewSessionIO(extractHost(remoteAddr), c)
	if err := l.commandHandler.Handle(ctx, session); err != nil {
		l.logger.ErrorContext(ctx,
			"Failed processing command",
			diag.ErrAttr(err),
			slog.String("remoteAddr", remoteAddr),
		)
	}
}

func (l *Listener) Start(ctx context.Context) error {
	l.logger.InfoContext(ctx, "Starting tcp listener", slog.Int("port", l.port))
	var err error
	l.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", l.port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	close(l.listeningSignal)

	for {
		c, acceptErr := l.listener.Accept()
		if acceptErr != nil {
			// Server stopped
			if errors.Is(acceptErr, net.ErrClosed) {
				return nil
			}

			// Not sure if it worth shutting down the server. Logging for now
			// Ideally we add a health check that will prove that the server is alive
			l.logger.ErrorContext(ctx, "failed to accept connection", diag.ErrAttr(acceptErr))
		} else {
			go l.processAcceptedConnection(ctx, c)
		}
	}
}

func (l *Listener) WaitListening() {
	<-l.listeningSignal
}

func (l *Listener) Close() error {
	if l.listener == nil {
		return nil
	}
	return l.listener.Close()
}
