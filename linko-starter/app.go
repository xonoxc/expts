package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"boot.dev/linko/internal/store"
)

type Config struct {
	hTTPPort int
	dataDir  string
}

type App struct {
	logger *slog.Logger
	store  *store.Store
	server *server
	port   int
}

type AppContructorDeps struct {
	logger *slog.Logger
	cancel context.CancelFunc
}

func newApp(cfg Config, deps AppContructorDeps) (*App, error) {
	st, err := store.New(cfg.dataDir, deps.logger)
	if err != nil {
		return nil, err
	}

	s := newServer(*st, deps.logger, cfg.hTTPPort, deps.cancel)

	return &App{
		logger: deps.logger,
		store:  st,
		server: s,
		port:   cfg.hTTPPort,
	}, nil
}

func (a *App) run(ctx context.Context) int {
	var serverErr error
	go func() {
		serverErr = a.server.start()
	}()

	a.logger.Info("Server stated",
		"url", fmt.Sprintf("http://localhost:%d", a.port),
	)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := a.server.shutdown(shutdownCtx)
	if err != nil {
		a.logger.Error("Server shutdown failure",
			"err", err,
		)
		return 1
	}

	a.logger.Info("Linko has shut down gracefully")

	if serverErr != nil {
		a.logger.Error("Internal Server Error",
			"error:", serverErr,
		)
		return 1
	}

	return 0
}
