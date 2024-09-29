package challenges

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"
	"word-of-wisdom-go/pkg/services"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestRateMonitor(t *testing.T) {
	newMockDeps := func() RequestRateMonitorDeps {
		return RequestRateMonitorDeps{
			WindowDuration:              time.Duration(rand.Int64N(10000)),
			MaxUnverifiedClientRequests: 10 + rand.Int64N(20),
			MaxUnverifiedRequests:       100 + rand.Int64N(20),

			TimeProvider: services.NewMockNow(),
		}
	}

	t.Run("challengeCondition", func(t *testing.T) {
		t.Run("should allow unverified requests within the limit", func(t *testing.T) {
			deps := newMockDeps()
			monitor, _ := NewRequestRateMonitor(deps).(*requestRateMonitor)
			assert.Equal(
				t,
				RecordRequestResult{},
				monitor.challengeCondition(deps.MaxUnverifiedClientRequests-1, deps.MaxUnverifiedRequests-1),
			)
		})
		t.Run("should require client requests verification above threshold", func(t *testing.T) {
			deps := newMockDeps()
			monitor, _ := NewRequestRateMonitor(deps).(*requestRateMonitor)
			assert.Equal(
				t,
				RecordRequestResult{
					ChallengeRequired:   true,
					ChallengeComplexity: 1,
				},
				monitor.challengeCondition(deps.MaxUnverifiedClientRequests+1, deps.MaxUnverifiedRequests-1),
			)
		})
		t.Run("should grow client request verification complexity linearly", func(t *testing.T) {
			deps := newMockDeps()
			monitor, _ := NewRequestRateMonitor(deps).(*requestRateMonitor)
			wantComplexity := 1 + rand.IntN(10)
			assert.Equal(
				t,
				RecordRequestResult{
					ChallengeRequired:   true,
					ChallengeComplexity: wantComplexity,
				},
				monitor.challengeCondition(int64(wantComplexity)*deps.MaxUnverifiedClientRequests+2, deps.MaxUnverifiedRequests-1),
			)
		})
		t.Run("should require global requests verification above threshold", func(t *testing.T) {
			deps := newMockDeps()
			monitor, _ := NewRequestRateMonitor(deps).(*requestRateMonitor)
			assert.Equal(
				t,
				RecordRequestResult{
					ChallengeRequired:   true,
					ChallengeComplexity: 1,
				},
				monitor.challengeCondition(deps.MaxUnverifiedClientRequests-1, deps.MaxUnverifiedRequests+1),
			)
		})
		t.Run("should increase global requests verification if at 2x global capacity", func(t *testing.T) {
			deps := newMockDeps()
			monitor, _ := NewRequestRateMonitor(deps).(*requestRateMonitor)
			assert.Equal(
				t,
				RecordRequestResult{
					ChallengeRequired:   true,
					ChallengeComplexity: 2,
				},
				monitor.challengeCondition(deps.MaxUnverifiedClientRequests-1, deps.MaxUnverifiedRequests*2+1),
			)
		})
	})

	t.Run("RecordRequest", func(t *testing.T) {
		t.Run("should accumulate metrics within the window", func(t *testing.T) {
			deps := newMockDeps()
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			wantRes := RecordRequestResult{
				ChallengeRequired:   true,
				ChallengeComplexity: rand.IntN(100),
			}
			conditionCalls := make([][]int64, 0, len(clients)*int(deps.MaxUnverifiedClientRequests))
			deps.challengeConditionFunc = func(nextClientCount, nextGlobalCount int64) RecordRequestResult {
				conditionCalls = append(conditionCalls, []int64{nextClientCount, nextGlobalCount})
				return wantRes
			}

			monitor := NewRequestRateMonitor(deps)
			for _, clientID := range clients {
				for range deps.MaxUnverifiedClientRequests {
					res, err := monitor.RecordRequest(ctx, clientID)
					require.NoError(t, err)
					require.Equal(t, wantRes, res)
				}
			}
			for i := range clients {
				callIndexBase := int64(i) * deps.MaxUnverifiedClientRequests
				for j := range deps.MaxUnverifiedClientRequests {
					callIndex := int(callIndexBase + j)
					call := conditionCalls[callIndex]
					assert.Equal(t, []int64{j + 1, callIndexBase + j + 1}, call)
				}
			}
		})
		t.Run("should reset metrics after the window", func(t *testing.T) {
			deps := newMockDeps()
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			wantRes := RecordRequestResult{
				ChallengeRequired:   true,
				ChallengeComplexity: rand.IntN(100),
			}
			conditionCalls := make([][]int64, 0, len(clients)*int(deps.MaxUnverifiedClientRequests))
			deps.challengeConditionFunc = func(nextClientCount, nextGlobalCount int64) RecordRequestResult {
				conditionCalls = append(conditionCalls, []int64{nextClientCount, nextGlobalCount})
				return wantRes
			}

			monitor := NewRequestRateMonitor(deps)
			for _, clientID := range clients {
				for range deps.MaxUnverifiedClientRequests {
					res, err := monitor.RecordRequest(ctx, clientID)
					require.NoError(t, err)
					require.Equal(t, wantRes, res)
				}
			}
			mockNow, _ := deps.TimeProvider.(*services.MockNow)
			mockNow.Increment(deps.WindowDuration + 1*time.Millisecond)

			conditionCalls = make([][]int64, 0, len(clients)*int(deps.MaxUnverifiedClientRequests))
			for _, clientID := range clients {
				for range deps.MaxUnverifiedClientRequests {
					res, err := monitor.RecordRequest(ctx, clientID)
					require.NoError(t, err)
					require.Equal(t, wantRes, res)
				}
			}
			for i := range clients {
				callIndexBase := int64(i) * deps.MaxUnverifiedClientRequests
				for j := range deps.MaxUnverifiedClientRequests {
					callIndex := int(callIndexBase + j)
					call := conditionCalls[callIndex]
					assert.Equal(t, []int64{j + 1, callIndexBase + j + 1}, call)
				}
			}
		})
	})
}
