package challenges

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"word-of-wisdom-go/pkg/services"

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

type Challenges interface {
	GenerateNewChallenge(clientID string) (string, error)
	VerifySolution(
		complexity int,
		challenge string,
		solution string,
	) bool

	// SolveChallenge returns a nonce that is a solution of the challenge.
	// It is used by client side only and
	// in real world scenario this may sit in it's own repo
	// but keeping it simple for now.
	// Returns error if the solution was not found
	SolveChallenge(
		ctx context.Context,
		complexity int,
		challenge string,
	) (string, error)
}

type Deps struct {
	dig.In `ignore-unexported:"true"`

	// services
	services.TimeProvider
	services.CryptoRandReader

	computeHashFn func(input []byte) []byte
}

type challenges struct {
	Deps
}

func (c *challenges) generateRandomBytes(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := c.CryptoRandReader(nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

func (c *challenges) GenerateNewChallenge(clientID string) (string, error) {
	nowBytes := big.NewInt(c.Now().UnixNano()).Bytes()
	nonce, err := c.generateRandomBytes(10) // TODO: may need to make it smaller (or configurable)
	if err != nil {
		return "", err
	}
	challengeBytes := make([]byte, len(nowBytes)+len(nonce)+len(clientID))
	copy(challengeBytes, []byte(clientID))
	copy(challengeBytes[len(clientID):], nowBytes)
	copy(challengeBytes[len(clientID)+len(nowBytes):], nonce)
	return hex.EncodeToString(challengeBytes), nil
}

func (c *challenges) VerifySolution(
	complexity int,
	challenge string,
	solution string,
) bool {
	hashInputBytes := make([]byte, len(challenge)+len(solution)+1)
	copy(hashInputBytes, []byte(challenge))
	hashInputBytes[len(challenge)] = ':'
	copy(hashInputBytes[len(challenge)+1:], []byte(solution))
	actualHash := c.computeHashFn(hashInputBytes)
	leadingZerosNum := countLeadingZeros(actualHash)
	return leadingZerosNum >= complexity
}

func (c *challenges) SolveChallenge(ctx context.Context, complexity int, challenge string) (string, error) {
	challengePartEnd := len(challenge)
	hashInput := make([]byte, challengePartEnd+20) // we reserve 20 bytes for solution which should be enough
	copy(hashInput, []byte(challenge))
	hashInput[challengePartEnd] = ':'
	nonce := 0

	// TODO: If no deadline, set default deadline (configurable)
	deadline, hasDeadline := ctx.Deadline()

	for {
		nonceStr := strconv.Itoa(nonce)
		copy(hashInput[challengePartEnd+1:], []byte(nonceStr))
		hash := c.computeHashFn(hashInput[:challengePartEnd+1+len(nonceStr)])
		leadingZeros := countLeadingZeros(hash)
		if leadingZeros >= complexity {
			return nonceStr, nil
		}

		// TODO: Make sure to set the deadline on caller
		if hasDeadline && c.Deps.Now().UnixNano() >= deadline.UnixNano() {
			break
		}

		nonce++
	}
	return "", errors.New("failed to solve challenge: deadline reached")
}

func NewChallenges(deps Deps) Challenges {
	if deps.computeHashFn == nil {
		deps.computeHashFn = computeHash
	}
	return &challenges{
		Deps: deps,
	}
}
