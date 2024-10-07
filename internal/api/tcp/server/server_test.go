package server

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"testing"
	"time"
	"word-of-wisdom-go/internal/diag"
	"word-of-wisdom-go/internal/services"
	"word-of-wisdom-go/internal/services/networking"

	"github.com/go-faker/faker/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListener(t *testing.T) {
	newMockDeps := func(t *testing.T) ListenerDeps {
		return ListenerDeps{
			RootLogger:         diag.RootTestLogger(),
			Port:               50000 + rand.IntN(15000),
			MaxSessionDuration: 10 * time.Second,
			Handler:            NewMockCommandHandler(t),
			UUIDGenerator:      services.NewUUIDGenerator(),
		}
	}

	t.Run("extractHost", func(t *testing.T) {
		t.Run("should return host only if host:port", func(t *testing.T) {
			host := faker.Word()
			port := rand.IntN(100)
			got := extractHost(fmt.Sprintf("%s:%d", host, port))
			assert.Equal(t, host, got)
		})
		t.Run("should return host if no port", func(t *testing.T) {
			host := faker.Word()
			got := extractHost(host)
			assert.Equal(t, host, got)
		})
	})

	t.Run("Start", func(t *testing.T) {
		t.Run("should process connections", func(t *testing.T) {
			deps := newMockDeps(t)
			srv := NewListener(deps)
			ctx := context.Background()
			go func() {
				assert.NoError(t, srv.Start(ctx))
			}()
			srv.WaitListening()
			defer srv.Close()

			mockHandler, _ := deps.Handler.(*MockCommandHandler)
			handleSignal := make(chan struct{})
			wantData := faker.Sentence()
			mockHandler.EXPECT().Handle(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, s *networking.Session) error {
					gotData, err := s.ReadLine()
					require.NoError(t, err)
					assert.Equal(t, wantData, gotData)
					close(handleSignal)
					return nil
				},
			)

			client, err := net.Dial("tcp", fmt.Sprintf(":%d", deps.Port))
			lo.Must1(client.Write([]byte(wantData + "\n")))
			<-handleSignal
			require.NoError(t, err)
			defer client.Close()
		})

		t.Run("should log command errors", func(t *testing.T) {
			deps := newMockDeps(t)
			srv := NewListener(deps)
			ctx := context.Background()
			go func() {
				assert.NoError(t, srv.Start(ctx))
			}()
			srv.WaitListening()
			defer srv.Close()

			mockHandler, _ := deps.Handler.(*MockCommandHandler)
			handleSignal := make(chan struct{})
			mockHandler.EXPECT().Handle(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, _ *networking.Session) error {
					close(handleSignal)
					return errors.New(faker.Sentence())
				},
			)

			client, err := net.Dial("tcp", fmt.Sprintf(":%d", deps.Port))
			lo.Must1(client.Write([]byte(faker.Word() + "\n")))
			<-handleSignal
			require.NoError(t, err)
			defer client.Close()
		})

		t.Run("should log command panics", func(t *testing.T) {
			deps := newMockDeps(t)
			srv := NewListener(deps)
			ctx := context.Background()
			go func() {
				assert.NoError(t, srv.Start(ctx))
			}()
			srv.WaitListening()
			defer srv.Close()

			mockHandler, _ := deps.Handler.(*MockCommandHandler)
			handleSignal := make(chan struct{})
			mockHandler.EXPECT().Handle(mock.Anything, mock.Anything).RunAndReturn(
				func(_ context.Context, _ *networking.Session) error {
					close(handleSignal)
					panic(errors.New(faker.Sentence()))
				},
			)

			client, err := net.Dial("tcp", fmt.Sprintf(":%d", deps.Port))
			lo.Must1(client.Write([]byte(faker.Word() + "\n")))
			<-handleSignal
			require.NoError(t, err)
			defer client.Close()
		})

		t.Run("should fail to listen on an occupied port", func(t *testing.T) {
			deps := newMockDeps(t)
			srv1 := NewListener(deps)
			ctx := context.Background()
			go func() {
				assert.NoError(t, srv1.Start(ctx))
			}()
			srv1.WaitListening()
			defer srv1.Close()

			srv2 := NewListener(deps)
			require.Error(t, srv2.Start(ctx))
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("should do nothing if not listening", func(t *testing.T) {
			deps := newMockDeps(t)
			srv := NewListener(deps)
			assert.NotPanics(t, func() {
				require.NoError(t, srv.Close())
			})
		})
	})
}
