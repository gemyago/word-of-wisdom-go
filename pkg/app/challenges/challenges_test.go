package challenges

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"
	"word-of-wisdom-go/pkg/services"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallenges(t *testing.T) {
	newMockDeps := func() Deps {
		return Deps{
			TimeProvider: services.NewMockNow(),
		}
	}

	newMockHashWithLeadingZeros := func(zerosCount int, ending string) []byte {
		mockHash := make([]byte, zerosCount+len(ending))
		for i := range zerosCount {
			mockHash[i] = zeroByte
		}
		copy(mockHash[zerosCount:], []byte(ending))
		return mockHash
	}

	t.Run("GenerateNewChallenge", func(t *testing.T) {
		t.Run("should generate a new challenge and hash it", func(t *testing.T) {
			deps := newMockDeps()

			clientID := faker.UUIDHyphenated()

			mockNow, _ := deps.TimeProvider.(*services.MockNow)

			randomBytes := []byte(faker.Sentence())[:16]

			wantChallengeBytes := append(
				[]byte(clientID),
				big.NewInt(mockNow.Now().UnixNano()).Bytes()...,
			)
			wantChallengeBytes = append(wantChallengeBytes, randomBytes...)
			deps.CryptoRandReader = func(b []byte) (int, error) {
				return copy(b, randomBytes), nil
			}

			wantChallenge := hex.EncodeToString(wantChallengeBytes)

			challenges := NewChallenges(deps)
			challenge, err := challenges.GenerateNewChallenge(clientID)
			require.NoError(t, err)
			assert.Equal(t, wantChallenge, challenge)
		})
	})

	t.Run("VerifySolution", func(t *testing.T) {
		t.Run("should return true if the solution is valid", func(t *testing.T) {
			deps := newMockDeps()
			wantComplexity := 5 + rand.IntN(5)
			challenge := faker.UUIDHyphenated()

			nonce := 10 + rand.IntN(1000)
			mockHash := newMockHashWithLeadingZeros(nonce, faker.Word())

			deps.computeHashFn = func(input []byte) []byte {
				assert.Equal(t, fmt.Sprintf("%s:%d", challenge, nonce), string(input))
				return mockHash
			}

			challenges := NewChallenges(deps)
			assert.True(
				t,
				challenges.VerifySolution(wantComplexity, challenge, strconv.Itoa(nonce)),
			)
		})
		t.Run("should return false if the solution is incomplete", func(t *testing.T) {
			deps := newMockDeps()
			wantComplexity := 5 + rand.IntN(5)
			challenge := faker.UUIDHyphenated()

			nonce := 10 + rand.IntN(1000)
			mockHash := newMockHashWithLeadingZeros(wantComplexity-1, faker.Word())

			deps.computeHashFn = func(input []byte) []byte {
				assert.Equal(t, fmt.Sprintf("%s:%d", challenge, nonce), string(input))
				return mockHash
			}

			challenges := NewChallenges(deps)
			assert.False(
				t,
				challenges.VerifySolution(wantComplexity, challenge, strconv.Itoa(nonce)),
			)
		})
	})

	t.Run("SolveChallenge", func(t *testing.T) {
		t.Run("should iterate until solution is found", func(t *testing.T) {
			deps := newMockDeps()
			wantComplexity := 5 + rand.IntN(10)
			challenge := faker.UUIDHyphenated()

			wantIterations := wantComplexity * 2
			mockHashes := make([][]byte, wantIterations)
			for v := range wantIterations - 1 {
				mockHashes[v] = newMockHashWithLeadingZeros(rand.IntN(wantComplexity-1), faker.Word())
			}
			mockHashes[wantIterations-1] = newMockHashWithLeadingZeros(wantComplexity, faker.Word())

			iterationCount := 0
			deps.computeHashFn = func(input []byte) []byte {
				assert.Equal(t, fmt.Sprintf("%s:%d", challenge, iterationCount), string(input))
				nextHash := mockHashes[iterationCount]
				iterationCount++
				return nextHash
			}

			challenges := NewChallenges(deps)
			solution, err := challenges.SolveChallenge(context.Background(), wantComplexity, challenge)
			require.NoError(t, err)
			require.Equal(t, wantIterations, iterationCount)
			assert.Equal(t, strconv.Itoa(wantIterations-1), solution)
		})
		t.Run("should exit on deadline", func(t *testing.T) {
			deps := newMockDeps()
			wantComplexity := 5 + rand.IntN(10)
			challenge := faker.UUIDHyphenated()

			mockNow, _ := deps.TimeProvider.(*services.MockNow)
			wantDeadline := mockNow.Now().Add(10 * time.Second)

			iterationsCount := 0
			deps.computeHashFn = func(_ []byte) []byte {
				iterationsCount++
				if iterationsCount > 100 {
					mockNow.SetValue(wantDeadline)
				}
				return newMockHashWithLeadingZeros(rand.IntN(wantComplexity-1), faker.Word())
			}

			challenges := NewChallenges(deps)

			ctx, cancel := context.WithDeadline(context.Background(), wantDeadline)
			defer cancel()
			solution, err := challenges.SolveChallenge(ctx, wantComplexity, challenge)
			require.ErrorContains(t, err, "deadline reached")
			assert.Equal(t, "", solution)
		})
	})

	t.Run("Integration", func(t *testing.T) {
		challenges := NewChallenges(Deps{
			TimeProvider:     services.NewTimeProvider(),
			CryptoRandReader: cryptoRand.Read,
		})

		clientID := faker.UUIDHyphenated()
		challenge, err := challenges.GenerateNewChallenge(clientID)
		require.NoError(t, err)

		wantComplexity := 1 + rand.IntN(3)
		solution, err := challenges.SolveChallenge(context.Background(), wantComplexity, challenge)
		require.NoError(t, err)

		verifyResult := challenges.VerifySolution(wantComplexity, challenge, solution)
		assert.True(t, verifyResult)
	})
}
