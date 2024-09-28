package main

import (
	"fmt"
	"net"
	"word-of-wisdom-go/pkg/services/networking"
)

type SessionDialer interface {
	// DialSession establishes new connection and returns session and close function
	DialSession(network, address string) (networking.Session, func() error, error)
}

type SessionDialerFunc func(network, address string) (networking.Session, func() error, error)

func (f SessionDialerFunc) DialSession(network, address string) (networking.Session, func() error, error) {
	return f(network, address)
}

var _ SessionDialer = SessionDialerFunc(nil)

func dialSession(network, address string) (networking.Session, func() error, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to server: %w", err)
	}
	session := networking.NewSession(conn.LocalAddr().String(), conn)
	return session, conn.Close, nil
}
