package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"word-of-wisdom-go/pkg/diag"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type runWOWCommandParams struct {
	dig.In `ignore-unexported:"true"`

	RootLogger *slog.Logger

	// client specific deps
	sessionDialer
	WOWCommand

	// Expected in a form host:port
	serverAddress string
	output        io.Writer
}

func runWOWCommand(ctx context.Context, params runWOWCommandParams) error {
	logger := params.RootLogger.WithGroup("client")
	logger.DebugContext(ctx, "Establishing connection", slog.String("address", params.serverAddress))

	session, cleanup, err := params.DialSession("tcp", params.serverAddress)
	if err != nil {
		return err
	}
	defer func() {
		if err = cleanup(); err != nil {
			logger.ErrorContext(ctx, "Connection cleanup failed", diag.ErrAttr(err))
		}
	}()
	result, err := params.WOWCommand.Process(ctx, session)
	if err != nil {
		return err
	}
	fmt.Fprintln(params.output, "Your Word of Wisdom for today:")
	fmt.Fprintln(params.output, result)
	return nil
}

func newClientCmd(container *dig.Container) *cobra.Command {
	serverAddress := "localhost:44221"
	cmd := &cobra.Command{
		Use:   "get-wow",
		Short: "Command to connect to the server and get word of wisdom",
	}
	cmd.Flags().StringVarP(&serverAddress, "address", "a", serverAddress, "Server address to connect to")
	noop := false
	cmd.Flags().BoolVar(
		&noop,
		"noop",
		false,
		"Do not start. Just setup deps and exit. Useful for testing if setup is all working.",
	)
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		return container.Invoke(func(params runWOWCommandParams) error {
			params.serverAddress = serverAddress
			params.output = os.Stdout
			if noop {
				params.RootLogger.InfoContext(ctx, "Establishing connection", slog.String("address", params.serverAddress))
				return nil
			}
			return runWOWCommand(ctx, params)
		})
	}
	return cmd
}
