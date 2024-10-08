package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"time"
	"word-of-wisdom-go/internal/services"

	"go.uber.org/dig"
)

const zeroByte = 48

func computeHash(input []byte) []byte {
	hash := sha256.New()
	hash.Write(input)
	return hash.Sum(nil)
}

func countLeadingZeros(hash []byte) int {
	count := 0
	for _, char := range hash {
		if char == zeroByte {
			count++
		} else {
			break
		}
	}
	return count
}

type Deps struct {
	dig.In `ignore-unexported:"true"`

	// config
	MaxSolveChallengeDuration time.Duration `name:"config.challenges.maxSolveChallengeDuration"`

	// services
	services.TimeProvider
	services.CryptoRandReader

	computeHashFn func(input []byte) []byte
}

type Challenges struct {
	deps Deps
}

func (c *Challenges) generateRandomBytes(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := c.deps.CryptoRandReader(nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func (c *Challenges) GenerateNewChallenge(clientID string) (string, error) {
	nowBytes := big.NewInt(c.deps.Now().UnixNano()).Bytes()
	nonce, err := c.generateRandomBytes(16) // TODO: may need to make it smaller (or configurable)
	if err != nil {
		return "", err
	}
	challengeBytes := make([]byte, len(nowBytes)+len(nonce)+len(clientID))
	copy(challengeBytes, []byte(clientID))
	copy(challengeBytes[len(clientID):], nowBytes)
	copy(challengeBytes[len(clientID)+len(nowBytes):], nonce)
	return hex.EncodeToString(challengeBytes), nil
}

func (c *Challenges) VerifySolution(
	complexity int,
	challenge string,
	solution string,
) bool {
	hashInputBytes := make([]byte, len(challenge)+len(solution)+1)
	copy(hashInputBytes, []byte(challenge))
	hashInputBytes[len(challenge)] = ':'
	copy(hashInputBytes[len(challenge)+1:], []byte(solution))
	actualHash := c.deps.computeHashFn(hashInputBytes)
	leadingZerosNum := countLeadingZeros(actualHash)
	return leadingZerosNum >= complexity
}

// SolveChallenge returns a nonce that is a solution of the challenge.
// It is used by client side only and
// in real world scenario this may sit in it's own repo
// but keeping it simple for now.
// Returns error if the solution was not found.
func (c *Challenges) SolveChallenge(ctx context.Context, complexity int, challenge string) (string, error) {
	challengePartEnd := len(challenge)
	hashInput := make([]byte, challengePartEnd+20) // we reserve 20 bytes for solution which should be enough
	copy(hashInput, []byte(challenge))
	hashInput[challengePartEnd] = ':'
	nonce := 0

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = c.deps.Now().Add(c.deps.MaxSolveChallengeDuration)
	}

	/*
		This can be parallelised however some research & benchmarking are required.
		Straight forward approach was attempted that didn't prove to be more performant
		than serial approach, roughly:
		- GOMAXPROCS goroutines where running the hash computation below
		- Separate goroutine is generating nonces and feeding them to the above
			worker pool via channel
		Benchmark has proven that at least up to complexity 3 the serial approach is
		faster
	*/
	for {
		nonceStr := strconv.Itoa(nonce)
		copy(hashInput[challengePartEnd+1:], []byte(nonceStr))
		hash := c.deps.computeHashFn(hashInput[:challengePartEnd+1+len(nonceStr)])
		leadingZeros := countLeadingZeros(hash)
		if leadingZeros >= complexity {
			return nonceStr, nil
		}

		if c.deps.Now().UnixNano() >= deadline.UnixNano() {
			break
		}

		nonce++
	}
	return "", errors.New("failed to solve challenge: deadline reached")
}

func NewChallenges(deps Deps) *Challenges {
	if deps.computeHashFn == nil {
		deps.computeHashFn = computeHash
	}
	return &Challenges{
		deps: deps,
	}
}
