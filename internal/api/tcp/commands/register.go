package commands

import (
	"word-of-wisdom-go/internal/app"
	"word-of-wisdom-go/internal/di"

	"go.uber.org/dig"
)

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		di.ProvideAs[*app.Challenges, challengesService],
		di.ProvideAs[*app.RequestRateMonitor, requestRateMonitor],
		di.ProvideAs[*app.WowQuery, wowQuery],

		NewHandler,
	)
}
