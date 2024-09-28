package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"word-of-wisdom-go/pkg/api/tcp/commands"
	"word-of-wisdom-go/pkg/diag"
	"word-of-wisdom-go/pkg/services/networking"

	"go.uber.org/dig"
)

type ListenerDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	Port int `name:"config.tcpServer.port"`

	// components
	commands.CommandHandler
}

type Listener struct {
	logger         *slog.Logger
	listener       net.Listener
	commandHandler commands.CommandHandler
	port           int
}

func NewListener(deps ListenerDeps) *Listener {
	return &Listener{
		port:           deps.Port,
		commandHandler: deps.CommandHandler,
		logger:         deps.RootLogger.WithGroup("tcp.server"),
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
	remoteAddr := c.RemoteAddr().String()
	l.logger.InfoContext(ctx, "Connection accepted", slog.String("remoteAddr", remoteAddr))
	defer c.Close()
	session := networking.NewSession(extractHost(remoteAddr), c)
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
		return err
	}

	for {
		c, acceptErr := l.listener.Accept()
		if acceptErr != nil {
			// Server stopped
			if errors.Is(acceptErr, net.ErrClosed) {
				return nil
			}

			// TODO: Not sure if it worth shutting down the server. Logging for now
			// Ideally we add a health check that will prove that the server is alive
			l.logger.ErrorContext(ctx, "failed to accept connection", diag.ErrAttr(acceptErr))
		}

		// TODO: Some sort of middleware to inject correlationId
		go l.processAcceptedConnection(ctx, c)
	}
}

func (l *Listener) Close() error {
	if l.listener != nil {
		return nil
	}
	return l.listener.Close()
}
