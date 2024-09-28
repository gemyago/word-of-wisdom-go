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

	t.Run("RecordRequest", func(t *testing.T) {
		t.Run("should allow unverified requests within a window", func(t *testing.T) {
			deps := newMockDeps()
			monitor := NewRequestRateMonitor(deps)
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			for _, clientID := range clients {
				for range deps.MaxUnverifiedClientRequests {
					res, err := monitor.RecordRequest(ctx, clientID)
					require.NoError(t, err)
					require.Equal(t, RecordRequestResult{}, res)
				}
			}
		})

		t.Run("should require client requests verification above threshold", func(t *testing.T) {
			deps := newMockDeps()
			monitor := NewRequestRateMonitor(deps)
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			for _, clientID := range clients {
				for range deps.MaxUnverifiedClientRequests {
					_, err := monitor.RecordRequest(ctx, clientID)
					require.NoError(t, err)
				}
				res, err := monitor.RecordRequest(ctx, clientID)
				require.NoError(t, err)
				assert.Equal(t, RecordRequestResult{
					ChallengeRequired:   true,
					ChallengeComplexity: 1,
				}, res)
			}
		})

		t.Run("should reset per client metrics after window elapses", func(t *testing.T) {
			deps := newMockDeps()
			monitor := NewRequestRateMonitor(deps)
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			for _, clientID := range clients {
				for range deps.MaxUnverifiedClientRequests {
					_, err := monitor.RecordRequest(ctx, clientID)
					require.NoError(t, err)
				}
			}

			mockNow, _ := deps.TimeProvider.(*services.MockNow)
			mockNow.Increment(deps.WindowDuration + 1*time.Millisecond)

			for _, clientID := range clients {
				res, err := monitor.RecordRequest(ctx, clientID)
				require.NoError(t, err)
				assert.Equal(t, RecordRequestResult{}, res)
			}
		})

		t.Run("should require verification if global threshold reached", func(t *testing.T) {
			deps := newMockDeps()
			monitor := NewRequestRateMonitor(deps)
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			for range deps.MaxUnverifiedRequests {
				_, err := monitor.RecordRequest(ctx, faker.UUIDHyphenated())
				require.NoError(t, err)
			}
			for _, clientID := range clients {
				res, err := monitor.RecordRequest(ctx, clientID)
				require.NoError(t, err)
				require.Equal(t, RecordRequestResult{
					ChallengeRequired:   true,
					ChallengeComplexity: 1,
				}, res)
			}
		})

		t.Run("should reset global verification metrics after window elapses", func(t *testing.T) {
			deps := newMockDeps()
			monitor := NewRequestRateMonitor(deps)
			ctx := context.Background()

			clients := []string{
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
				faker.UUIDHyphenated(),
			}

			for range deps.MaxUnverifiedRequests {
				_, err := monitor.RecordRequest(ctx, faker.UUIDHyphenated())
				require.NoError(t, err)
			}

			mockNow, _ := deps.TimeProvider.(*services.MockNow)
			mockNow.Increment(deps.WindowDuration + (1 * time.Millisecond))

			for _, clientID := range clients {
				res, err := monitor.RecordRequest(ctx, clientID)
				require.NoError(t, err)
				require.Equal(t, RecordRequestResult{}, res)
			}
		})
	})
}
