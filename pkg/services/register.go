package services

import (
	"crypto/rand"
	"time"
	"word-of-wisdom-go/pkg/di"

	"go.uber.org/dig"
)

type CryptoRandReader func(b []byte) (n int, err error)

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		NewTimeProvider,
		NewUUIDGenerator,
		di.ProvideValue(time.NewTicker),
		di.ProvideValue(CryptoRandReader(rand.Read)),
	)
}
