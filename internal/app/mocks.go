//go:build !release

package app

import "context"

// Mock interfaces are used to generate mock implementations of all of the components
// that will be reused elsewhere in a system. This helps to minimize the amount of
// duplicate mock implementations that need to be written.

type mockRequestRateMonitor interface {
	RecordRequest(ctx context.Context, clientID string) (RecordRequestResult, error)
}

var _ mockRequestRateMonitor = (*RequestRateMonitor)(nil)

type mockChallenges interface {
	GenerateNewChallenge(clientID string) (string, error)
	VerifySolution(complexity int, challenge, solution string) bool
	SolveChallenge(ctx context.Context, complexity int, challenge string) (string, error)
}

var _ mockChallenges = (*Challenges)(nil)

type mockWowQuery interface {
	GetNextWoW(_ context.Context) (string, error)
}

var _ mockWowQuery = (*WowQuery)(nil)
