//go:build !release

package services

import (
	"strings"
	"time"

	"github.com/go-faker/faker/v4"
)

type MockNow struct {
	value time.Time
}

var _ TimeProvider = &MockNow{}

func (m *MockNow) SetValue(t time.Time) {
	m.value = t
}

func (m *MockNow) Increment(duration time.Duration) {
	m.value = m.value.Add(duration)
}

func (m *MockNow) Now() time.Time {
	return m.value
}

func NewMockNow() *MockNow {
	return &MockNow{
		value: time.UnixMilli(faker.RandomUnixTime()),
	}
}

func MockNowValue(p TimeProvider) time.Time {
	mp, ok := p.(*MockNow)
	if !ok {
		panic("provided TimeProvider is not a MockNow")
	}
	return mp.value
}

type mockSessionIOStream struct {
	readBuffer     chan string
	writeBuffer    chan string
	nextReadError  error
	nextWriteError error
}

func (m *mockSessionIOStream) Read(p []byte) (int, error) {
	line := <-m.readBuffer
	if m.nextReadError != nil {
		return 0, m.nextReadError
	}
	copy(p, line)
	return len(line), nil
}

func (m *mockSessionIOStream) Write(p []byte) (int, error) {
	line := string(p)
	go func() {
		m.writeBuffer <- line
	}()
	if m.nextWriteError != nil {
		return 0, m.nextWriteError
	}
	return len(p), nil
}

type MockSessionIOController struct {
	Session *SessionIO
	stream  *mockSessionIOStream
}

func (m *MockSessionIOController) MockSendLine(line string) {
	go func() {
		m.stream.readBuffer <- line + "\n"
	}()
}

func (m *MockSessionIOController) MockSendLineAndWaitResult(line string) string {
	go func() {
		m.stream.readBuffer <- line + "\n"
	}()
	return m.MockWaitResult()
}

func (m *MockSessionIOController) MockWaitResult() string {
	result := <-m.stream.writeBuffer
	return strings.TrimSuffix(result, "\n")
}

func (m *MockSessionIOController) MockSetNextReadError(err error) {
	m.stream.nextReadError = err
}

func (m *MockSessionIOController) MockSetNextWriteError(err error) {
	m.stream.nextWriteError = err
}

func NewMockSessionIOController() *MockSessionIOController {
	stream := &mockSessionIOStream{
		readBuffer:  make(chan string),
		writeBuffer: make(chan string),
	}
	return &MockSessionIOController{
		Session: NewSessionIO(faker.UUIDHyphenated(), stream),
		stream:  stream,
	}
}
