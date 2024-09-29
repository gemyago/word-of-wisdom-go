package challenges

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"word-of-wisdom-go/pkg/services"

	"github.com/samber/lo"
	"go.uber.org/dig"
)

type RecordRequestResult struct {
	ChallengeRequired   bool
	ChallengeComplexity int
}

type RequestRateMonitor interface {
	RecordRequest(ctx context.Context, clientID string) (RecordRequestResult, error)
}

type ChallengeConditionFunc func(nextClientCount, nextGlobalCount int64) RecordRequestResult

type RequestRateMonitorDeps struct {
	dig.In `ignore-unexported:"true"`

	// config

	// WindowDuration specifies max window duration to reset metrics
	WindowDuration time.Duration `name:"config.monitoring.windowDuration"`

	// MaxUnverifiedClientRequests max allowed unverified requests per client within the window
	MaxUnverifiedClientRequests int64 `name:"config.monitoring.maxUnverifiedClientRequests"`

	// MaxUnverifiedRequests max allowed unverified requests globally within the window
	MaxUnverifiedRequests int64 `name:"config.monitoring.maxUnverifiedRequests"`

	// services
	services.TimeProvider

	// local deps
	challengeConditionFunc ChallengeConditionFunc
}

type requestRateMonitor struct {
	RequestRateMonitorDeps

	requestRateByClient sync.Map
	globalRequestsCount atomic.Int64

	// timestamp in ms when the window was started
	windowStartedAt atomic.Int64
}

// challengeCondition defines if challenge will be required and the complexity.
func (m *requestRateMonitor) challengeCondition(nextClientCounter, nextGlobalCount int64) RecordRequestResult {
	challengeRequired := false
	complexityRequired := 0

	if nextClientCounter > m.MaxUnverifiedClientRequests {
		challengeRequired = true

		// We just grow it linearly, at some point (somewhere after 5 or 6)
		// it's just going to become unreasonably complex to proceed
		complexityRequired = int(nextClientCounter / m.MaxUnverifiedClientRequests)
	}

	if !challengeRequired && nextGlobalCount > m.MaxUnverifiedRequests {
		challengeRequired = true
		globalCapacityScale := nextGlobalCount / m.MaxUnverifiedRequests

		// If we're under pressure globally (current rate is 2x more than global threshold)
		// then increase min complexity for all users
		// 2 is reasonably complex, so good actors will not be impacted significantly
		complexityRequired = lo.If(globalCapacityScale >= 2, 2).Else(1)
	}

	// TODO: We should probably always use global complexity if it's above the caller complexity

	return RecordRequestResult{
		ChallengeRequired:   challengeRequired,
		ChallengeComplexity: complexityRequired,
	}
}

func (m *requestRateMonitor) RecordRequest(_ context.Context, clientID string) (RecordRequestResult, error) {
	// We are not using the context yet, but in a real world system it may be required
	// since we will very likely store counters somewhere

	// This is a naive implementation based on a fixed window algo
	// in a real world system we will need to support distributed scenario
	// and keep this data in something like redis, or use some other replication mechanism
	// and also use some sliding window algo with per client window.

	now := m.Now().UnixMilli()
	lastTimestamp := m.windowStartedAt.Load()
	if now-lastTimestamp > m.WindowDuration.Milliseconds() {
		if m.windowStartedAt.CompareAndSwap(lastTimestamp, now) {
			m.globalRequestsCount.Store(0)
			m.requestRateByClient.Clear()
		}
	}

	currentClientCounter, _ := m.requestRateByClient.LoadOrStore(clientID, new(int64))
	nextClientCounter := atomic.AddInt64(currentClientCounter.(*int64), 1)
	nextGlobalCount := m.globalRequestsCount.Add(1)

	return m.challengeConditionFunc(nextClientCounter, nextGlobalCount), nil
}

func NewRequestRateMonitor(deps RequestRateMonitorDeps) RequestRateMonitor {
	m := &requestRateMonitor{
		RequestRateMonitorDeps: deps,
	}
	if m.challengeConditionFunc == nil {
		m.challengeConditionFunc = m.challengeCondition
	}
	return m
}
