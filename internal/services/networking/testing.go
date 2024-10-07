//go:build !release

package networking

import (
	"strings"

	"github.com/go-faker/faker/v4"
)

type mockSessionStream struct {
	readBuffer  chan string
	writeBuffer chan string
	nextError   error
}

func (m *mockSessionStream) Read(p []byte) (int, error) {
	line := <-m.readBuffer
	copy(p, line)
	return len(line), m.nextError
}

func (m *mockSessionStream) Write(p []byte) (int, error) {
	line := string(p)
	go func() {
		m.writeBuffer <- line
	}()
	return len(p), m.nextError
}

type MockSessionController struct {
	Session *Session
	stream  *mockSessionStream
}

func (m *MockSessionController) MockSendLine(line string) {
	go func() {
		m.stream.readBuffer <- line + "\n"
	}()
}

func (m *MockSessionController) MockSendLineAndWaitResult(line string) string {
	go func() {
		m.stream.readBuffer <- line + "\n"
	}()
	return m.MockWaitResult()
}

func (m *MockSessionController) MockWaitResult() string {
	result := <-m.stream.writeBuffer
	return strings.TrimSuffix(result, "\n")
}

func (m *MockSessionController) MockSetNextError(err error) {
	m.stream.nextError = err
}

func NewMockSessionController() *MockSessionController {
	stream := &mockSessionStream{
		readBuffer:  make(chan string),
		writeBuffer: make(chan string),
	}
	return &MockSessionController{
		Session: NewSession(faker.UUIDHyphenated(), stream),
		stream:  stream,
	}
}
