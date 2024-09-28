package main

import (
	"context"
	"log/slog"
	"os/signal"
	"time"
	"word-of-wisdom-go/pkg/api/tcp/server"
	"word-of-wisdom-go/pkg/diag"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"golang.org/x/sys/unix"
)

type runTCPServerParams struct {
	dig.In `ignore-unexported:"true"`

	RootLogger *slog.Logger

	*server.Listener

	noop bool
}

func runTCPServer(params runTCPServerParams) error {
	rootLogger := params.RootLogger
	rootCtx := context.Background()

	signalCtx, cancel := signal.NotifyContext(rootCtx, unix.SIGINT, unix.SIGTERM)
	defer cancel()

	startupErrors := make(chan error, 1)
	go func() {
		if params.noop {
			rootLogger.InfoContext(signalCtx, "NOOP: Exiting now")
			startupErrors <- nil
			return
		}

		if err := params.Listener.Start(signalCtx); err != nil {
			startupErrors <- err
			return
		}
	}()

	var startupErr error
	select {
	case startupErr = <-startupErrors:
		if startupErr != nil {
			rootLogger.ErrorContext(rootCtx, "Server error", "err", startupErr)
		} else {
			rootLogger.InfoContext(rootCtx, "Server stopped")
		}
	case <-signalCtx.Done(): // coverage-ignore
		rootLogger.InfoContext(rootCtx, "Trying to shut down gracefully")
		ts := time.Now()
		if err := params.Listener.Close(); err != nil {
			rootLogger.ErrorContext(rootCtx, "graceful shutdown failed", diag.ErrAttr(err))
		}
		rootLogger.InfoContext(rootCtx, "Service stopped",
			slog.Duration("duration", time.Since(ts)),
		)
	}
	return startupErr
}

func newTCPServerCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tcp-server",
		Short: "Command to start tcp server",
	}
	noop := false
	cmd.Flags().BoolVar(
		&noop,
		"noop",
		false,
		"Do not start. Just setup deps and exit. Useful for testing if setup is all working.",
	)
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return container.Invoke(func(params runTCPServerParams) error {
			params.noop = noop
			return runTCPServer(params)
		})
	}
	return cmd
}
