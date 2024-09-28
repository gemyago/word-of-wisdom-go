package main

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"
	"testing"
	"word-of-wisdom-go/pkg/app/challenges"
	"word-of-wisdom-go/pkg/diag"
	"word-of-wisdom-go/pkg/services/networking"

	"github.com/go-faker/faker/v4"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWow(t *testing.T) {
	newMockDeps := func(t *testing.T) WOWCommandDeps {
		return WOWCommandDeps{
			RootLogger: diag.RootTestLogger(),
			Challenges: challenges.NewMockChallenges(t),
		}
	}

	t.Run("should perform command without challenge", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])
		wantWow := faker.Sentence()
		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		mockSession.MockSendLine("WOW: " + wantWow)

		cmdRes := <-cmdResCh
		require.NoError(t, cmdRes.B)
		assert.Equal(t, wantWow, cmdRes.A)
	})

	t.Run("should handle command sending err", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])
		wantWow := faker.Sentence()
		wantErr := errors.New(faker.Sentence())
		mockSession.MockSetNextError(wantErr)
		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		mockSession.MockSendLine("WOW: " + wantWow)

		cmdRes := <-cmdResCh
		assert.ErrorIs(t, cmdRes.B, wantErr)
	})

	t.Run("should perform command with challenge", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])
		wantWow := faker.Sentence()
		wantChallenge := faker.Sentence()
		wantComplexity := rand.IntN(10)
		wantSolution := faker.Sentence()

		mockChallenges, _ := deps.Challenges.(*challenges.MockChallenges)
		mockChallenges.EXPECT().SolveChallenge(ctx, wantComplexity, wantChallenge).Return(wantSolution, nil)

		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		gotSolutionResult := mockSession.MockSendLineAndWaitResult(
			"CHALLENGE_REQUIRED: " + wantChallenge + ";" + strconv.Itoa(wantComplexity),
		)
		assert.Equal(t, "CHALLENGE_RESULT: "+wantSolution, gotSolutionResult)
		mockSession.MockSendLine("WOW: " + wantWow)

		cmdRes := <-cmdResCh
		require.NoError(t, cmdRes.B)
		assert.Equal(t, wantWow, cmdRes.A)
	})

	t.Run("should handle unexpected no challenge state", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])

		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		mockSession.MockSendLine(faker.Word())
		cmdRes := <-cmdResCh
		assert.ErrorContains(t, cmdRes.B, "unexpected challenge requirement")
	})

	t.Run("should handle no wow after challenge", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])

		mockChallenges, _ := deps.Challenges.(*challenges.MockChallenges)
		mockChallenges.EXPECT().SolveChallenge(ctx, mock.Anything, mock.Anything).Return(faker.Sentence(), nil)

		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		mockSession.MockSendLineAndWaitResult(
			"CHALLENGE_REQUIRED: " + faker.Word() + ";" + strconv.Itoa(rand.Int()),
		)
		mockSession.MockSendLine(faker.Word())

		cmdRes := <-cmdResCh
		assert.ErrorContains(t, cmdRes.B, "got unexpected WOW response")
	})

	t.Run("should handle bad complexity", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])

		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		mockSession.MockSendLine(
			"CHALLENGE_REQUIRED: " + faker.Word() + ";" + faker.Word(),
		)

		cmdRes := <-cmdResCh
		assert.ErrorContains(t, cmdRes.B, "invalid syntax")
	})

	t.Run("should fail if challenge not solved", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		mockSession := networking.NewMockSession()
		cmdResCh := make(chan lo.Tuple2[string, error])

		mockChallenges, _ := deps.Challenges.(*challenges.MockChallenges)
		wantErr := errors.New(faker.Sentence())
		mockChallenges.EXPECT().SolveChallenge(ctx, mock.Anything, mock.Anything).Return("", wantErr)

		go func() {
			res, err := cmd.Process(ctx, mockSession)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := mockSession.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		mockSession.MockSendLine(
			"CHALLENGE_REQUIRED: " + faker.Word() + ";" + strconv.Itoa(rand.Int()),
		)

		cmdRes := <-cmdResCh
		assert.ErrorIs(t, cmdRes.B, wantErr)
	})
}
