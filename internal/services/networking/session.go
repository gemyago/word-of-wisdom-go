package networking

import (
	"bufio"
	"io"
)

type Session interface {
	ClientID() string
	ReadLine() (string, error)
	WriteLine(data string) error
}

type session struct {
	clientID string
	stream   io.ReadWriter
	reader   *bufio.Reader
}

func (s *session) ClientID() string {
	return s.clientID
}

func (s *session) ReadLine() (string, error) {
	line, _, err := s.reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (s *session) WriteLine(data string) error {
	_, err := s.stream.Write(append([]byte(data), '\n'))
	return err
}

func NewSession(clientID string, stream io.ReadWriter) Session {
	return &session{
		clientID: clientID,
		stream:   stream,
		reader:   bufio.NewReader(stream),
	}
}
