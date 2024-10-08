package services

import (
	"bufio"
	"io"
)

type SessionIO struct {
	clientID string
	stream   io.ReadWriter
	reader   *bufio.Reader
}

func (s *SessionIO) ClientID() string {
	return s.clientID
}

func (s *SessionIO) ReadLine() (string, error) {
	line, _, err := s.reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (s *SessionIO) WriteLine(data string) error {
	_, err := s.stream.Write(append([]byte(data), '\n'))
	return err
}

func NewSessionIO(clientID string, stream io.ReadWriter) *SessionIO {
	return &SessionIO{
		clientID: clientID,
		stream:   stream,
		reader:   bufio.NewReader(stream),
	}
}
