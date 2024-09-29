package challenges

import (
	"context"
	cryptoRand "crypto/rand"
	"testing"
	"time"
	"word-of-wisdom-go/pkg/services"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func BenchmarkChallengesSolveChallenge(b *testing.B) {
	challenges := NewChallenges(Deps{
		TimeProvider:              services.NewTimeProvider(),
		CryptoRandReader:          cryptoRand.Read,
		MaxSolveChallengeDuration: 2 * time.Minute,
	})

	clientID := faker.UUIDHyphenated()
	challenge, err := challenges.GenerateNewChallenge(clientID)
	require.NoError(b, err)

	ctx := context.Background()

	b.Run("complexity-1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = challenges.SolveChallenge(ctx, 1, challenge)
			require.NoError(b, err)
		}
	})

	b.Run("complexity-2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = challenges.SolveChallenge(ctx, 2, challenge)
			require.NoError(b, err)
		}
	})

	b.Run("complexity-3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err = challenges.SolveChallenge(ctx, 3, challenge)
			require.NoError(b, err)
		}
	})
}
