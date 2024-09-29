package main

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialer(t *testing.T) {
	makeMockDeps := func() SessionDialerDeps {
		return SessionDialerDeps{
			IOTimeout: 10 * time.Second,
		}
	}

	t.Run("should establish target session", func(t *testing.T) {
		port := 50000 + rand.IntN(15000)
		srv := newMockTCPServer()
		require.NoError(t, srv.Start(port))
		defer srv.stop()

		dialer := newSessionDialer(makeMockDeps())
		session, cleanup, err := dialer.DialSession("tcp", fmt.Sprintf(":%d", port))
		require.NoError(t, err)
		defer cleanup() //nolint: errcheck //server stops sooner
		mockData := faker.Sentence()
		require.NoError(t, session.WriteLine(mockData))
		gotRes, err := session.ReadLine()
		require.NoError(t, err)
		assert.Equal(t, mockData, gotRes)

		gotConnections := srv.getConnections()
		assert.Len(t, gotConnections, 1)
		assert.Equal(t, mockData+"\n", gotConnections[0].message)
	})
	t.Run("should handle connection error", func(t *testing.T) {
		port := 50000 + rand.IntN(15000)

		dialer := newSessionDialer(makeMockDeps())
		_, _, err := dialer.DialSession("tcp", fmt.Sprintf(":%d", port))
		require.Error(t, err)
	})
}
