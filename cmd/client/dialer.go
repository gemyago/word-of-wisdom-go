package main

import (
	"fmt"
	"net"
	"time"
	"word-of-wisdom-go/internal/services"

	"go.uber.org/dig"
)

type SessionDialer interface {
	// DialSession establishes new connection and returns session and close function
	DialSession(network, address string) (*services.SessionIO, func() error, error)
}

type sessionDialerFunc func(network, address string) (*services.SessionIO, func() error, error)

func (f sessionDialerFunc) DialSession(network, address string) (*services.SessionIO, func() error, error) {
	return f(network, address)
}

var _ SessionDialer = sessionDialerFunc(nil)

type SessionDialerDeps struct {
	dig.In

	// config
	IOTimeout time.Duration `name:"config.client.ioTimeout"`
}

func newSessionDialer(deps SessionDialerDeps) SessionDialer {
	return sessionDialerFunc(func(network, address string) (*services.SessionIO, func() error, error) {
		conn, err := net.Dial(network, address)
		if err != nil {
			return nil, nil, fmt.Errorf("error connecting to server: %w", err)
		}
		if err = conn.SetDeadline(time.Now().Add(deps.IOTimeout)); err != nil { // coverage-ignore // hard to simulate this
			return nil, nil, fmt.Errorf("failed to set deadline: %w", err)
		}
		session := services.NewSessionIO(conn.LocalAddr().String(), conn)
		return session, conn.Close, nil
	})
}
