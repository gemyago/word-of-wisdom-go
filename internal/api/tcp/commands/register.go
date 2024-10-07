package commands

import (
	"word-of-wisdom-go/internal/app/challenges"
	"word-of-wisdom-go/internal/app/wow"
	"word-of-wisdom-go/internal/di"

	"go.uber.org/dig"
)

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		di.ProvideAs[*challenges.Challenges, challengesService],
		di.ProvideAs[*challenges.RequestRateMonitor, requestRateMonitor],
		di.ProvideAs[*wow.Query, wowQuery],

		NewHandler,
	)
}
