//go:build !release

package main

import (
	"net"
	"strconv"
	"sync"
)

type connectionData struct {
	message string
}

type mockTCPServer struct {
	listener    net.Listener
	isRunning   bool
	connections []connectionData
	mu          sync.Mutex
}

func newMockTCPServer() *mockTCPServer {
	return &mockTCPServer{}
}

// Start initializes and starts the TCP server on a random port.
func (s *mockTCPServer) Start(port int) error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	s.isRunning = true

	startupSignal := make(chan struct{})
	go func() {
		startupSignal <- struct{}{}
		for {
			var conn net.Conn
			conn, err = s.listener.Accept()
			if err != nil {
				if !s.isRunning {
					return
				}
				continue
			}
			go s.handleConnection(conn)
		}
	}()
	<-startupSignal
	return nil
}

func (s *mockTCPServer) stop() {
	s.isRunning = false
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *mockTCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}
		message := string(buffer[:n])

		s.mu.Lock()
		s.connections = append(s.connections, connectionData{message: message})
		s.mu.Unlock()

		_, err = conn.Write(buffer[:n])
		if err != nil {
			return
		}
	}
}

func (s *mockTCPServer) getConnections() []connectionData {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]connectionData{}, s.connections...) // Return a copy
}
