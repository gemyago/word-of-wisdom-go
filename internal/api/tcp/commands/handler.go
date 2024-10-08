package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"word-of-wisdom-go/internal/app"
	"word-of-wisdom-go/internal/diag"
	"word-of-wisdom-go/internal/services/networking"

	"go.uber.org/dig"
)

// protocol related constants.
const (
	CommandGetWow           = "GET_WOW"
	WowResponsePrefix       = "WOW: "
	challengeRequired       = "CHALLENGE_REQUIRED"
	ChallengeRequiredPrefix = challengeRequired + ": "
	ChallengeResultPrefix   = "CHALLENGE_RESULT: "

	errBadCmdResponse            = "ERR: BAD_CMD"
	errInternalError             = "ERR: INTERNAL_ERROR"
	errUnexpectedChallengeResult = "ERR: UNEXPECTED_CHALLENGE_RESULT"
	errChallengeVerificationFail = "ERR: CHALLENGE_VERIFICATION_FAILED"
)

type wowQuery interface {
	GetNextWoW(_ context.Context) (string, error)
}

type requestRateMonitor interface {
	RecordRequest(ctx context.Context, clientID string) (app.RecordRequestResult, error)
}

type challengesService interface {
	GenerateNewChallenge(clientID string) (string, error)
	VerifySolution(complexity int, challenge, solution string) bool
}

type CommandHandlerDeps struct {
	dig.In

	RootLogger *slog.Logger

	// components
	RequestRateMonitor requestRateMonitor
	Challenges         challengesService
	Query              wowQuery
}

type CommandHandler struct {
	deps   CommandHandlerDeps
	logger *slog.Logger
}

func (h *CommandHandler) trace(ctx context.Context, msg string, args ...any) {
	h.logger.DebugContext(ctx, msg, args...)
}

func (h *CommandHandler) performChallengeVerification(
	ctx context.Context,
	session *networking.Session,
	monitoringResult app.RecordRequestResult,
) (bool, error) {
	var challenge string
	challenge, err := h.deps.Challenges.GenerateNewChallenge(session.ClientID())
	if err != nil {
		return false, fmt.Errorf("failed to generate new challenge: %w", err)
	}

	h.trace(ctx, "Requiring to solve challenge", slog.Int("complexity", monitoringResult.ChallengeComplexity))
	challengeData := fmt.Sprintf(
		"%s%s;%d",
		ChallengeRequiredPrefix, challenge, monitoringResult.ChallengeComplexity)
	if err = session.WriteLine(challengeData); err != nil {
		return false, fmt.Errorf("failed to send challenge: %w", err)
	}
	var cmd string
	if cmd, err = session.ReadLine(); err != nil {
		return false, fmt.Errorf("failed to read challenge result: %w", err)
	}
	if strings.Index(cmd, ChallengeResultPrefix) != 0 {
		h.trace(ctx, "Got unexpected challenge result", slog.String("data", cmd))
		return false, session.WriteLine(errUnexpectedChallengeResult)
	}

	if !h.deps.Challenges.VerifySolution(
		monitoringResult.ChallengeComplexity,
		challenge,
		strings.Trim(cmd[len(ChallengeResultPrefix):], " "),
	) {
		h.trace(ctx, "Challenge verification failed", slog.String("data", cmd))
		return false, session.WriteLine(errChallengeVerificationFail)
	}
	return true, nil
}

func (h *CommandHandler) Handle(ctx context.Context, con *networking.Session) error {
	cmd, err := con.ReadLine()
	if err != nil {
		return err
	}

	// If we need to extend it to support multiple commands
	// then this will need to be refactored roughly as follows:
	// - new Commands component is added that implement all various commands
	// - the HandleCommands will read the command from the connection, and forward
	//   the processing to particular command implementation
	// Keeping it simple for now since we need just a single command.
	if cmd != CommandGetWow {
		h.trace(ctx, "Got bad command", slog.String("cmd", cmd))
		return con.WriteLine(errBadCmdResponse)
	}

	monitoringResult, err := h.deps.RequestRateMonitor.RecordRequest(ctx, con.ClientID())
	if err != nil {
		h.logger.ErrorContext(ctx,
			"Failed to record request",
			slog.String("clientID", con.ClientID()),
			diag.ErrAttr(err),
		)
		return con.WriteLine(errInternalError)
	}

	if monitoringResult.ChallengeRequired {
		var ok bool
		ok, err = h.performChallengeVerification(ctx, con, monitoringResult)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	wow, err := h.deps.Query.GetNextWoW(ctx)
	if err != nil {
		return fmt.Errorf("failed to get next wow: %w", err)
	}

	h.trace(ctx, "Responding with WOW")
	return con.WriteLine(WowResponsePrefix + wow)
}

func NewHandler(deps CommandHandlerDeps) *CommandHandler {
	return &CommandHandler{
		deps:   deps,
		logger: deps.RootLogger.WithGroup("tcp.server.handler"),
	}
}
