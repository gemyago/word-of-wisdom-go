package challenges

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"word-of-wisdom-go/pkg/services"

	"go.uber.org/dig"
)

type RecordRequestResult struct {
	ChallengeRequired   bool
	ChallengeComplexity int
}

type RequestRateMonitor interface {
	RecordRequest(ctx context.Context, clientID string) (RecordRequestResult, error)
}

type RequestRateMonitorDeps struct {
	dig.In

	// config

	// WindowDuration specifies max window duration to reset metrics
	WindowDuration time.Duration `name:"config.monitoring.windowDuration"`

	// MaxUnverifiedClientRequests max allowed unverified requests per client within the window
	MaxUnverifiedClientRequests int64 `name:"config.monitoring.maxUnverifiedClientRequests"`

	// MaxUnverifiedRequests max allowed unverified requests globally within the window
	MaxUnverifiedRequests int64 `name:"config.monitoring.maxUnverifiedRequests"`

	// services
	services.TimeProvider
}

type requestRateMonitor struct {
	RequestRateMonitorDeps

	requestRateByClient sync.Map
	globalRequestsCount atomic.Int64

	// timestamp in ms when the window was started
	windowStartedAt atomic.Int64
}

func (m *requestRateMonitor) RecordRequest(_ context.Context, clientID string) (RecordRequestResult, error) {
	// We are not using the context yet, but in a real world system it may be required
	// since we will very likely store counters somewhere

	// This is a naive implementation based on a fixed window algo
	// in a real world system we will need to support distributed scenario
	// and keep this data in something like redis, or use some other replication mechanism
	// and also use some sliding window algo with per client window.

	challengeRequired := false
	complexityRequired := 0

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
	if nextClientCounter > m.MaxUnverifiedClientRequests {
		challengeRequired = true
		complexityRequired = 1
	}

	nextGlobalCount := m.globalRequestsCount.Add(1)
	if !challengeRequired && nextGlobalCount > m.MaxUnverifiedRequests {
		challengeRequired = true
		complexityRequired = 1
	}

	return RecordRequestResult{
		ChallengeRequired:   challengeRequired,
		ChallengeComplexity: complexityRequired,
	}, nil
}

func NewRequestRateMonitor(deps RequestRateMonitorDeps) RequestRateMonitor {
	return &requestRateMonitor{
		RequestRateMonitorDeps: deps,
	}
}
