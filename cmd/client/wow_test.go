package main

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"
	"testing"
	"word-of-wisdom-go/internal/app"
	"word-of-wisdom-go/internal/diag"
	"word-of-wisdom-go/internal/services"

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
			Challenges: app.NewMockChallenges(t),
		}
	}

	t.Run("should perform command without challenge", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])
		wantWow := faker.Sentence()
		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		ctrl.MockSendLine("WOW: " + wantWow)

		cmdRes := <-cmdResCh
		require.NoError(t, cmdRes.B)
		assert.Equal(t, wantWow, cmdRes.A)
	})

	t.Run("should handle command sending err", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])
		wantWow := faker.Sentence()
		wantErr := errors.New(faker.Sentence())
		ctrl.MockSetNextReadError(wantErr)
		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		ctrl.MockSendLine("WOW: " + wantWow)

		cmdRes := <-cmdResCh
		assert.ErrorIs(t, cmdRes.B, wantErr)
	})

	t.Run("should perform command with challenge", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])
		wantWow := faker.Sentence()
		wantChallenge := faker.Sentence()
		wantComplexity := rand.IntN(10)
		wantSolution := faker.Sentence()

		mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
		mockChallenges.EXPECT().SolveChallenge(ctx, wantComplexity, wantChallenge).Return(wantSolution, nil)

		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		gotSolutionResult := ctrl.MockSendLineAndWaitResult(
			"CHALLENGE_REQUIRED: " + wantChallenge + ";" + strconv.Itoa(wantComplexity),
		)
		assert.Equal(t, "CHALLENGE_RESULT: "+wantSolution, gotSolutionResult)
		ctrl.MockSendLine("WOW: " + wantWow)

		cmdRes := <-cmdResCh
		require.NoError(t, cmdRes.B)
		assert.Equal(t, wantWow, cmdRes.A)
	})

	t.Run("should handle unexpected no challenge state", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])

		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		ctrl.MockSendLine(faker.Word())
		cmdRes := <-cmdResCh
		assert.ErrorContains(t, cmdRes.B, "unexpected challenge requirement")
	})

	t.Run("should handle no wow after challenge", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])

		mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
		mockChallenges.EXPECT().SolveChallenge(ctx, mock.Anything, mock.Anything).Return(faker.Sentence(), nil)

		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		ctrl.MockSendLineAndWaitResult(
			"CHALLENGE_REQUIRED: " + faker.Word() + ";" + strconv.Itoa(rand.Int()),
		)
		ctrl.MockSendLine(faker.Word())

		cmdRes := <-cmdResCh
		assert.ErrorContains(t, cmdRes.B, "got unexpected WOW response")
	})

	t.Run("should handle bad complexity", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])

		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		ctrl.MockSendLine(
			"CHALLENGE_REQUIRED: " + faker.Word() + ";" + faker.Word(),
		)

		cmdRes := <-cmdResCh
		assert.ErrorContains(t, cmdRes.B, "invalid syntax")
	})

	t.Run("should fail if challenge not solved", func(t *testing.T) {
		deps := newMockDeps(t)
		cmd := newWOWCommand(deps)

		ctx := context.Background()
		ctrl := services.NewMockSessionIOController()
		cmdResCh := make(chan lo.Tuple2[string, error])

		mockChallenges, _ := deps.Challenges.(*app.MockChallenges)
		wantErr := errors.New(faker.Sentence())
		mockChallenges.EXPECT().SolveChallenge(ctx, mock.Anything, mock.Anything).Return("", wantErr)

		go func() {
			res, err := cmd.Process(ctx, ctrl.Session)
			cmdResCh <- lo.Tuple2[string, error]{A: res, B: err}
		}()
		gotCmd := ctrl.MockWaitResult()
		assert.Equal(t, "GET_WOW", gotCmd)
		ctrl.MockSendLine(
			"CHALLENGE_REQUIRED: " + faker.Word() + ";" + strconv.Itoa(rand.Int()),
		)

		cmdRes := <-cmdResCh
		assert.ErrorIs(t, cmdRes.B, wantErr)
	})
}
