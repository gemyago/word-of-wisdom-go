//go:build !release

package networking

import "github.com/go-faker/faker/v4"

type MockSession struct {
	clientID    string
	readBuffer  chan string
	writeBuffer chan string
}

func (m *MockSession) ClientID() string {
	return m.clientID
}

func (m *MockSession) MockSendLineAndWaitResult(line string) string {
	go func() {
		m.readBuffer <- line
	}()
	return <-m.writeBuffer
}

func (m *MockSession) ReadLine() (string, error) {
	data := <-m.readBuffer
	return data, nil
}

func (m *MockSession) WriteLine(data string) error {
	m.writeBuffer <- data
	return nil
}

func NewMockSession() *MockSession {
	return &MockSession{
		clientID:    faker.UUIDHyphenated(),
		readBuffer:  make(chan string),
		writeBuffer: make(chan string),
	}
}
