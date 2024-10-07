package wow

import (
	"word-of-wisdom-go/internal/di"

	"go.uber.org/dig"
)

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		NewQuery,
	)
}
