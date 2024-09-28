package networking

import (
	"bytes"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	t.Run("ReadLine", func(t *testing.T) {
		t.Run("should read next line from the stream", func(t *testing.T) {
			var buffer bytes.Buffer
			wantLine := faker.Sentence()
			lo.Must(buffer.WriteString(wantLine))
			lo.Must(buffer.WriteRune('\n'))

			session := NewSession(faker.UUIDHyphenated(), &buffer)
			assert.Equal(t, wantLine, lo.Must(session.ReadLine()))
			assert.NotEmpty(t, session.ClientID())
		})
		t.Run("should handle read errors", func(t *testing.T) {
			var buffer bytes.Buffer
			session := NewSession(faker.UUIDHyphenated(), &buffer)
			_, err := session.ReadLine()
			assert.Error(t, err)
		})
	})
	t.Run("WriteLine", func(t *testing.T) {
		t.Run("should write line to the stream", func(t *testing.T) {
			var buffer bytes.Buffer
			wantLine := faker.Sentence()

			session := NewSession(faker.UUIDHyphenated(), &buffer)
			lo.Must0(session.WriteLine(wantLine))
			assert.Equal(t, wantLine+"\n", buffer.String())
		})
	})
}
