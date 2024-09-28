package commands

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"word-of-wisdom-go/pkg/app/challenges"
	"word-of-wisdom-go/pkg/app/wow"
	"word-of-wisdom-go/pkg/diag"
	"word-of-wisdom-go/pkg/services/networking"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestCommands(t *testing.T) {
	makeMockDeps := func(t *testing.T) CommandHandlerDeps {
		return CommandHandlerDeps{
			RootLogger:         diag.RootTestLogger(),
			RequestRateMonitor: challenges.NewMockRequestRateMonitor(t),
			Challenges:         challenges.NewMockChallenges(t),
			Query:              wow.NewMockQuery(t),
		}
	}

	t.Run("Handle", func(t *testing.T) {
		t.Run("should fail if unexpected command", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			session := networking.NewMockSession()

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, session)
			}()

			result := session.MockSendLineAndWaitResult(faker.Word())
			assert.Equal(t, "ERR: BAD_CMD", result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should process GET_WOW if no challenge required", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			session := networking.NewMockSession()

			mockMonitor, _ := deps.RequestRateMonitor.(*challenges.MockRequestRateMonitor)
			mockMonitor.EXPECT().RecordRequest(ctx, session.ClientID()).Return(
				challenges.RecordRequestResult{}, nil,
			)

			wantWow := faker.Sentence()
			mockQuery, _ := deps.Query.(*wow.MockQuery)
			mockQuery.EXPECT().GetNextWoW(ctx).Return(wantWow, nil)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, session)
			}()

			result := session.MockSendLineAndWaitResult("GET_WOW")
			assert.Equal(t, "WOW: "+wantWow, result)
			assert.NoError(t, <-handleErr)
		})
		t.Run("should process GET_WOW with challenge required", func(t *testing.T) {
			ctx := context.Background()
			deps := makeMockDeps(t)
			h := NewHandler(deps)

			session := networking.NewMockSession()

			mockMonitor, _ := deps.RequestRateMonitor.(*challenges.MockRequestRateMonitor)
			monitorResult := challenges.RecordRequestResult{
				ChallengeRequired:   true,
				ChallengeComplexity: 5 + rand.IntN(10),
			}
			mockMonitor.EXPECT().RecordRequest(ctx, session.ClientID()).Return(
				monitorResult, nil,
			)

			wantChallenge := faker.UUIDHyphenated()
			wantSolution := faker.UUIDHyphenated()
			mockChallenges, _ := deps.Challenges.(*challenges.MockChallenges)
			mockChallenges.EXPECT().GenerateNewChallenge(session.ClientID()).Return(wantChallenge, nil)
			mockChallenges.EXPECT().VerifySolution(
				monitorResult.ChallengeComplexity,
				wantChallenge,
				wantSolution,
			).Return(true)

			wantWow := faker.Sentence()
			mockQuery, _ := deps.Query.(*wow.MockQuery)
			mockQuery.EXPECT().GetNextWoW(ctx).Return(wantWow, nil)

			handleErr := make(chan error)
			go func() {
				handleErr <- h.Handle(ctx, session)
			}()

			result := session.MockSendLineAndWaitResult("GET_WOW")
			assert.Equal(t, fmt.Sprintf("CHALLENGE_REQUIRED: %s;%d", wantChallenge, monitorResult.ChallengeComplexity), result)

			result = session.MockSendLineAndWaitResult("CHALLENGE_RESULT: " + wantSolution)
			assert.Equal(t, "WOW: "+wantWow, result)
			assert.NoError(t, <-handleErr)
		})
	})
}
