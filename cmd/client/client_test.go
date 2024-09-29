package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"word-of-wisdom-go/pkg/diag"
	"word-of-wisdom-go/pkg/services/networking"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Run("runWOWCommand", func(t *testing.T) {
		t.Run("should process wow command", func(t *testing.T) {
			ctx := context.Background()
			mockSession := networking.NewMockSession()
			wantWow := faker.Sentence()
			var output bytes.Buffer
			wantAddress := faker.Word()
			params := runWOWCommandParams{
				serverAddress: wantAddress,
				RootLogger:    diag.RootTestLogger(),
				sessionDialer: sessionDialerFunc(func(network, address string) (networking.Session, func() error, error) {
					assert.Equal(t, "tcp", network)
					assert.Equal(t, wantAddress, address)
					return mockSession, func() error { return nil }, nil
				}),
				WOWCommand: WOWCommandFunc(func(_ context.Context, session networking.Session) (string, error) {
					assert.Equal(t, mockSession, session)
					return wantWow, nil
				}),
				output: &output,
			}
			require.NoError(t, runWOWCommand(ctx, params))
			assert.Equal(t, "Your Word of Wisdom for today:\n"+wantWow+"\n", output.String())
		})
		t.Run("should handle dial errors", func(t *testing.T) {
			ctx := context.Background()
			wantAddress := faker.Word()
			wantDialErr := errors.New(faker.Sentence())
			params := runWOWCommandParams{
				serverAddress: wantAddress,
				RootLogger:    diag.RootTestLogger(),
				sessionDialer: sessionDialerFunc(func(_, _ string) (networking.Session, func() error, error) {
					return nil, nil, wantDialErr
				}),
			}
			assert.ErrorIs(t, wantDialErr, runWOWCommand(ctx, params))
		})
		t.Run("should handle wow command errors", func(t *testing.T) {
			ctx := context.Background()
			wantErr := errors.New(faker.Sentence())
			params := runWOWCommandParams{
				serverAddress: faker.Word(),
				RootLogger:    diag.RootTestLogger(),
				sessionDialer: sessionDialerFunc(func(_, _ string) (networking.Session, func() error, error) {
					return networking.NewMockSession(), func() error { return nil }, nil
				}),
				WOWCommand: WOWCommandFunc(func(_ context.Context, _ networking.Session) (string, error) {
					return "", wantErr
				}),
			}
			assert.ErrorIs(t, wantErr, runWOWCommand(ctx, params))
		})
		t.Run("should log cleanup errors", func(t *testing.T) {
			ctx := context.Background()
			params := runWOWCommandParams{
				serverAddress: faker.Word(),
				RootLogger:    diag.RootTestLogger(),
				sessionDialer: sessionDialerFunc(func(_, _ string) (networking.Session, func() error, error) {
					return networking.NewMockSession(), func() error { return errors.New(faker.Sentence()) }, nil
				}),
				WOWCommand: WOWCommandFunc(func(_ context.Context, _ networking.Session) (string, error) {
					return faker.Sentence(), nil
				}),
				output: &bytes.Buffer{},
			}
			assert.NoError(t, runWOWCommand(ctx, params))
		})
	})
}
