package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
	pkgerr "github.com/pkg/errors"
	appErr "xonoxc/linko/internal"
)

type StackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

func prepareErrors(err error) []slog.Attr {
	attrs := appErr.Attrs(err)

	if multiErrr, ok := errors.AsType[appErr.MultiError](err); ok {
		for i, err := range multiErrr.Unwrap() {
			attrs = append(
				attrs, slog.Attr{
					Key:   fmt.Sprintf("error_%d", i),
					Value: slog.StringValue(err.Error()),
				})
		}
	}

	if stackErr, ok := errors.AsType[StackTracer](err); ok {
		attrs = append(attrs, slog.String(
			"stack_trace",
			fmt.Sprintf("%+v", stackErr.StackTrace()),
		))
	}

	return attrs
}

func isTerminalEnv() bool {
	fd := os.Stderr.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
