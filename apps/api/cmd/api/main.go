package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ginkit/api/internal/app"
	"ginkit/api/internal/env"
	"ginkit/api/internal/logger"
	"ginkit/api/internal/server"
	"ginkit/api/internal/version"
)

func main() {
	if err := run(); err != nil {
		slog.Error("exiting", slog.Any("error", err))
		os.Exit(1)
	}
}

// run wires the application together. Start here to add middleware, readiness
// checks, or anything that needs the application's lifetime context.
func run() error {
	logger.Setup()
	slog.Info("starting", slog.String("version", version.Version))
	ctx := shutdownContext()

	// Readiness checks go here; see handlers.Check.
	router, _, err := app.New(ctx, app.ConfigFromEnv())
	if err != nil {
		return err
	}

	return server.Run(ctx, server.Config{
		Addr:       fmt.Sprintf(":%d", env.Port()),
		Handler:    router,
		DrainDelay: env.ShutdownDelay(),
	})
}

// shutdownContext ends on the first SIGINT or SIGTERM, which starts the
// graceful shutdown. A second signal exits at once instead of waiting for it.
func shutdownContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		cancel()
		sig := <-signals
		slog.Warn("second signal received, exiting without draining", slog.String("signal", sig.String()))
		os.Exit(1)
	}()

	return ctx
}
