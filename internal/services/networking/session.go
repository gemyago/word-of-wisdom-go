package networking

import (
	"bufio"
	"io"
)

type Session struct {
	clientID string
	stream   io.ReadWriter
	reader   *bufio.Reader
}

func (s *Session) ClientID() string {
	return s.clientID
}

func (s *Session) ReadLine() (string, error) {
	line, _, err := s.reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), nil
}

func (s *Session) WriteLine(data string) error {
	_, err := s.stream.Write(append([]byte(data), '\n'))
	return err
}

func NewSession(clientID string, stream io.ReadWriter) *Session {
	return &Session{
		clientID: clientID,
		stream:   stream,
		reader:   bufio.NewReader(stream),
	}
}
