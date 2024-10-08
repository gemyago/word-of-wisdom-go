package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type runSolveChallengeCommandParams struct {
	dig.In `ignore-unexported:"true"`

	RootLogger *slog.Logger

	Challenges challengesService

	// client params
	challengeToSolve string
	complexity       int

	// Expected in a form host:port
	output io.Writer
}

func runSolveChallengeCommand(
	ctx context.Context,
	params runSolveChallengeCommandParams,
) error {
	startedAt := time.Now()
	solution, err := params.Challenges.SolveChallenge(
		ctx,
		params.complexity,
		params.challengeToSolve,
	)
	if err != nil {
		return fmt.Errorf("failed to solve challenge: %w", err)
	}
	solutionDuration := time.Since(startedAt)

	fmt.Fprintln(params.output, "Challenge solve result")
	fmt.Fprint(params.output, "Complexity: ")
	fmt.Fprintln(params.output, params.complexity)
	fmt.Fprint(params.output, "Solution: ")
	fmt.Fprintln(params.output, solution)
	fmt.Fprint(params.output, "Duration: ")
	fmt.Fprintln(params.output, solutionDuration)
	return nil
}

func newSolveChallengeCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "solve-challenge",
		Short: "Command to test solving challenges logic",
	}
	challenge := "any text can be here"
	complexity := 1
	silent := false
	cmd.Flags().StringVar(&challenge, "challenge", challenge, "Challenge to solve")
	cmd.Flags().IntVarP(&complexity, "complexity", "c", complexity, "Complexity of the solution (e.g leading hash zeros)")
	cmd.Flags().BoolVar(&silent, "silent", silent, "Do not produce any output. Just solve")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		return container.Invoke(func(params runSolveChallengeCommandParams) error {
			params.challengeToSolve = challenge
			params.complexity = complexity
			params.output = lo.If(silent, io.Discard).Else(os.Stdout)
			return runSolveChallengeCommand(ctx, params)
		})
	}
	return cmd
}
