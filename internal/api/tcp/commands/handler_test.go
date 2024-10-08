package commands

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"word-of-wisdom-go/internal/app"
	"word-of-wisdom-go/internal/diag"
	"word-of-wisdom-go/internal/services"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommands(t *testing.T) {
	makeMockDeps := func(t *testing.T) CommandHandlerDeps {
		return CommandHandlerDeps{
			RootLogger:         diag.RootTestLogger(),
			RequestRateMonitor: app.NewMockRequestRateMonitor(t),
			Challenges:         app.NewMockChallenges(t),
			Query:              app.NewMockWowQuery(t),
		}
	}

	t.Run("Handle", func(t *testing.T) {
		t.Run("should fail if err getting command", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			wantErr := errors.New(faker.Sentence())
			ctrl := services.NewMockSessionIOController()
			ctrl.MockSetNextReadError(wantErr)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			ctrl.MockSendLine(faker.Word())
			assert.ErrorIs(t, <-handleErr, wantErr)
		})
		t.Run("should fail if unexpected command", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			result := ctrl.MockSendLineAndWaitResult(faker.Word())
			assert.Equal(t, "ERR: BAD_CMD", result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should process GET_WOW if no challenge required", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				app.RecordRequestResult{}, nil,
			)

			wantWow := faker.Sentence()
			mockQuery, _ := deps.Query.(*app.MockWowQuery)
			mockQuery.EXPECT().GetNextWoW(ctx).Return(wantWow, nil)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			result := ctrl.MockSendLineAndWaitResult("GET_WOW")
			assert.Equal(t, "WOW: "+wantWow, result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should process GET_WOW with challenge required", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			monitorResult := app.RecordRequestResult{
				ChallengeRequired:   true,
				ChallengeComplexity: 5 + rand.IntN(10),
			}
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				monitorResult, nil,
			)

			wantChallenge := faker.UUIDHyphenated()
			wantSolution := faker.UUIDHyphenated()
			mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
			mockChallenges.EXPECT().GenerateNewChallenge(ctrl.Session.ClientID()).Return(wantChallenge, nil)
			mockChallenges.EXPECT().VerifySolution(
				monitorResult.ChallengeComplexity,
				wantChallenge,
				wantSolution,
			).Return(true)

			wantWow := faker.Sentence()
			mockQuery, _ := deps.Query.(*app.MockWowQuery)
			mockQuery.EXPECT().GetNextWoW(ctx).Return(wantWow, nil)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			result := ctrl.MockSendLineAndWaitResult("GET_WOW")
			assert.Equal(t, fmt.Sprintf("CHALLENGE_REQUIRED: %s;%d", wantChallenge, monitorResult.ChallengeComplexity), result)

			result = ctrl.MockSendLineAndWaitResult("CHALLENGE_RESULT: " + wantSolution)
			assert.Equal(t, "WOW: "+wantWow, result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should fail if record request error", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				app.RecordRequestResult{}, errors.New(faker.Sentence()),
			)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			result := ctrl.MockSendLineAndWaitResult("GET_WOW")
			assert.Equal(t, "ERR: INTERNAL_ERROR", result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should handle get next wow query error", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				app.RecordRequestResult{}, nil,
			)

			mockQuery, _ := deps.Query.(*app.MockWowQuery)
			wantErr := errors.New(faker.Sentence())
			mockQuery.EXPECT().GetNextWoW(ctx).Return("", wantErr)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			ctrl.MockSendLine("GET_WOW")
			assert.ErrorIs(t, <-handleErr, wantErr)
		})
		t.Run("should handle challenge generation errors", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				app.RecordRequestResult{ChallengeRequired: true}, nil,
			)

			wantErr := errors.New(faker.Sentence())
			mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
			mockChallenges.EXPECT().GenerateNewChallenge(ctrl.Session.ClientID()).Return("", wantErr)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			ctrl.MockSendLine("GET_WOW")
			assert.ErrorIs(t, <-handleErr, wantErr)
		})
		t.Run("should handle challenge verification error", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			monitorResult := app.RecordRequestResult{
				ChallengeRequired:   true,
				ChallengeComplexity: 5 + rand.IntN(10),
			}
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				monitorResult, nil,
			)

			mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
			mockChallenges.EXPECT().GenerateNewChallenge(ctrl.Session.ClientID()).Return(faker.Word(), nil)
			mockChallenges.EXPECT().VerifySolution(
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(false)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			ctrl.MockSendLineAndWaitResult("GET_WOW")

			result := ctrl.MockSendLineAndWaitResult("CHALLENGE_RESULT: " + faker.Word())
			assert.Equal(t, "ERR: CHALLENGE_VERIFICATION_FAILED", result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should handle session errors to write challenge", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			ctrl := services.NewMockSessionIOController()

			mockMonitor, _ := deps.RequestRateMonitor.(*app.MockRequestRateMonitor)
			monitorResult := app.RecordRequestResult{
				ChallengeRequired:   true,
				ChallengeComplexity: 5 + rand.IntN(10),
			}
			mockMonitor.EXPECT().RecordRequest(ctx, ctrl.Session.ClientID()).Return(
				monitorResult, nil,
			)

			mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
			mockChallenges.EXPECT().GenerateNewChallenge(ctrl.Session.ClientID()).Return(faker.Word(), nil)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, ctrl.Session)
			}()

			wantErr := errors.New(faker.Sentence())

			ctrl.MockSetNextWriteError(wantErr)
			ctrl.MockSendLineAndWaitResult("GET_WOW")
			require.ErrorIs(t, <-handleErr, wantErr)
		})
	})
}
