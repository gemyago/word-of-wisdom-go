package server

import (
	"word-of-wisdom-go/internal/api/tcp/commands"
	"word-of-wisdom-go/internal/di"

	"go.uber.org/dig"
)

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		di.ProvideAs[*commands.CommandHandler, commandHandler],

		NewListener,
	)
}
