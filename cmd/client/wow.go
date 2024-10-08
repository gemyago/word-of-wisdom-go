package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"word-of-wisdom-go/internal/api/tcp/commands"
	"word-of-wisdom-go/internal/services"

	"go.uber.org/dig"
)

type WOWCommand interface {
	Process(ctx context.Context, session *services.SessionIO) (string, error)
}

type WOWCommandFunc func(ctx context.Context, session *services.SessionIO) (string, error)

func (f WOWCommandFunc) Process(ctx context.Context, session *services.SessionIO) (string, error) {
	return f(ctx, session)
}

var _ WOWCommandFunc = WOWCommandFunc(nil)

type challengesService interface {
	SolveChallenge(ctx context.Context, complexity int, challenge string) (string, error)
}

type WOWCommandDeps struct {
	dig.In

	RootLogger *slog.Logger

	// app layer
	Challenges challengesService
}

func newWOWCommand(deps WOWCommandDeps) WOWCommand {
	logger := deps.RootLogger.WithGroup("client")
	return WOWCommandFunc(func(ctx context.Context, session *services.SessionIO) (string, error) {
		logger.DebugContext(ctx, "Sending GET_WOW request")

		if err := session.WriteLine(commands.CommandGetWow); err != nil {
			return "", fmt.Errorf("failed to write to the server: %w", err)
		}

		line, err := session.ReadLine()
		if err != nil {
			return "", fmt.Errorf("failed to read the response: %w", err)
		}

		logger.DebugContext(ctx, "Got response", slog.String("data", line))

		if strings.Index(line, commands.WowResponsePrefix) == 0 {
			logger.DebugContext(ctx, "Got WOW response. No challenge required")
			return strings.Trim(line[4:], " "), nil
		}

		if strings.Index(line, commands.ChallengeRequiredPrefix) != 0 {
			return "", fmt.Errorf("got unexpected challenge requirement response %s: %w", line, err)
		}

		separatorIndex := strings.Index(line, ";")

		challenge := strings.Trim(line[len(commands.ChallengeRequiredPrefix):separatorIndex], " ")
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
		if err = session.WriteLine(commands.ChallengeResultPrefix + solution); err != nil {
			return "", fmt.Errorf("failed to write to the server: %w", err)
		}

		line, err = session.ReadLine()
		if err != nil {
			return "", fmt.Errorf("failed to read the response: %w", err)
		}

		logger.DebugContext(ctx, "Got response", slog.String("data", line))

		if strings.Index(line, commands.WowResponsePrefix) == 0 {
			return strings.Trim(line[4:], " "), nil
		}

		return "", fmt.Errorf("got unexpected WOW response %s", line)
	})
}
