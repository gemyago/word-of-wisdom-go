package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"word-of-wisdom-go/internal/app/challenges"
	"word-of-wisdom-go/internal/services/networking"

	"go.uber.org/dig"
)

type WOWCommand interface {
	Process(ctx context.Context, session networking.Session) (string, error)
}

type WOWCommandFunc func(ctx context.Context, session networking.Session) (string, error)

func (f WOWCommandFunc) Process(ctx context.Context, session networking.Session) (string, error) {
	return f(ctx, session)
}

var _ WOWCommandFunc = WOWCommandFunc(nil)

type WOWCommandDeps struct {
	dig.In

	RootLogger *slog.Logger

	// app layer
	challenges.Challenges
}

func newWOWCommand(deps WOWCommandDeps) WOWCommand {
	logger := deps.RootLogger.WithGroup("client")
	return WOWCommandFunc(func(ctx context.Context, session networking.Session) (string, error) {
		logger.DebugContext(ctx, "Sending GET_WOW request")

		if err := session.WriteLine("GET_WOW"); err != nil {
			return "", fmt.Errorf("failed to write to the server: %w", err)
		}

		line, err := session.ReadLine()
		if err != nil {
			return "", fmt.Errorf("failed to read the response: %w", err)
		}

		logger.DebugContext(ctx, "Got response", slog.String("data", line))

		if strings.Index(line, "WOW:") == 0 {
			logger.DebugContext(ctx, "Got WOW response. No challenge required")
			return strings.Trim(line[4:], " "), nil
		}

		if strings.Index(line, "CHALLENGE_REQUIRED:") != 0 {
			return "", fmt.Errorf("got unexpected challenge requirement response %s: %w", line, err)
		}

		separatorIndex := strings.Index(line, ";")

		challenge := strings.Trim(line[len("CHALLENGE_REQUIRED:"):separatorIndex], " ")
		complexity, err := strconv.Atoi(line[separatorIndex+1:])
		if err != nil {
			return "", err
		}

		solveStartedAt := time.Now()
		solution, err := deps.Challenges.SolveChallenge(ctx, complexity, challenge)
		if err != nil {
			return "", err
		}

		logger.DebugContext(
			ctx,
			"Challenge solved. Sending challenge result",
			slog.Duration("solutionDuration", time.Since(solveStartedAt)),
			slog.String("solution", solution),
		)
		if err = session.WriteLine("CHALLENGE_RESULT: " + solution); err != nil {
			return "", fmt.Errorf("failed to write to the server: %w", err)
		}

		line, err = session.ReadLine()
		if err != nil {
			return "", fmt.Errorf("failed to read the response: %w", err)
		}

		logger.DebugContext(ctx, "Got response", slog.String("data", line))

		if strings.Index(line, "WOW:") == 0 {
			return strings.Trim(line[4:], " "), nil
		}

		return "", fmt.Errorf("got unexpected WOW response %s", line)
	})
}
