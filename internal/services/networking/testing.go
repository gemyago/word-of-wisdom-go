//go:build !release

package networking

import "github.com/go-faker/faker/v4"

type MockSession struct {
	clientID    string
	readBuffer  chan string
	writeBuffer chan string
	nextError   error
}

func (m *MockSession) ClientID() string {
	return m.clientID
}

func (m *MockSession) MockSendLine(line string) {
	go func() {
		m.readBuffer <- line
	}()
}

func (m *MockSession) MockSendLineAndWaitResult(line string) string {
	go func() {
		m.readBuffer <- line
	}()
	return <-m.writeBuffer
}

func (m *MockSession) MockWaitResult() string {
	return <-m.writeBuffer
}

func (m *MockSession) MockSetNextError(err error) {
	m.nextError = err
}

func (m *MockSession) ReadLine() (string, error) {
	data := <-m.readBuffer
	return data, m.nextError
}

func (m *MockSession) WriteLine(data string) error {
	m.writeBuffer <- data
	return m.nextError
}

func NewMockSession() *MockSession {
	return &MockSession{
		clientID:    faker.UUIDHyphenated(),
		readBuffer:  make(chan string),
		writeBuffer: make(chan string),
	}
}
