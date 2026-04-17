package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"xonoxc/linko/internal/build"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	logger, cf, err := mustSetupLogging()

	httpPort := flag.Int("port", 8899, "port to listen on")
	dataDir := flag.String(
		"data", "./data", "directory to store data")

	flag.Parse()

	app, err := newApp(
		Config{
			hTTPPort: *httpPort,
			dataDir:  *dataDir,
		},
		AppContructorDeps{logger, cancel},
	)
	if err != nil {
		logger.Error("App init failed",
			"error", err,
		)
		return
	}

	status := app.run(ctx)
	cancel()

	closeErr := cf()
	if closeErr != nil {
		log.Fatalf(
			"critical: failed to flush logs: %v", closeErr)
	}

	os.Exit(status)
}

type closeFunc func() error

/*
*
this is a safety net but don't rely on this
*
*/
var sensitiveKeys = []string{
	"password",
	"key",
	"apikey",
	"secret",
	"pin",
	"creditcardno",
}

func mustSetupLogging() (*slog.Logger, closeFunc, error) {
	var (
		handlers   []slog.Handler
		closeFuncs []closeFunc
	)

	replaceAttrfunc := func(groups []string, a slog.Attr) slog.Attr {
		if slices.Contains(sensitiveKeys, a.Key) {
			return slog.String(
				a.Key,
				"[Redacted]",
			)
		}

		if a.Key == "error" {
			err, exists := a.Value.Any().(error)
			if !exists {
				return a
			}

			attrs := prepareErrors(err)

			return slog.GroupAttrs(
				"error",
				append(
					[]slog.Attr{
						slog.String("message", err.Error()),
					},
					attrs...,
				)...,
			)
		}
		return a
	}

	handlers = append(handlers, tint.NewHandler(os.Stderr, &tint.Options{
		NoColor: !isTerminalEnv(),
		Level:   slog.LevelDebug,
	}))

	err := godotenv.Load()
	if err != nil {
		panic("can't load env vars")
	}

	logPath := os.Getenv("LINKO_LOG_FILE")

	file := setupLogFile(logPath)

	if file != nil {
		fileBufW := bufio.NewWriterSize(
			file,
			8192,
		)
		handlers = append(
			handlers, slog.NewJSONHandler(fileBufW, &slog.HandlerOptions{
				ReplaceAttr: replaceAttrfunc,
			}))

		closeFuncs = append(closeFuncs, func() error {
			err := fileBufW.Flush()
			if err != nil {
				return fmt.Errorf("failed to flush log file: %w", err)
			}

			err = file.Close()
			if err != nil {
				return fmt.Errorf(
					"failed to close log file: %w", err)
			}

			return nil
		})
	}

	closeManagerFunc := func() error {
		var errs []error

		for _, closer := range closeFuncs {
			errs = append(errs, closer())
		}
		return errors.Join(errs...)
	}

	env := os.Getenv("LINKO_ENV")
	hostname, _ := os.Hostname()

	logger := slog.New(slog.NewMultiHandler(handlers...)).With(
		slog.String(
			"git_sha", build.GIT_SHA),
		slog.String("env", env),
		slog.String("hostname", hostname),
		slog.String("build_time", build.BUILD_TIME),
	)

	return logger, closeManagerFunc, nil
}
