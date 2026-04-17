package appErr

import (
	"errors"
	"log/slog"
)

type errWithAttrs struct {
	error
	atters []slog.Attr
}

func (e *errWithAttrs) Unwrap() error {
	return e.error
}

func (e *errWithAttrs) Attrs() []slog.Attr {
	return e.atters
}

type attrError interface {
	Attrs() []slog.Attr
}

func Attrs(err error) []slog.Attr {
	var attrs []slog.Attr

	for err != nil {
		if ae, ok := err.(attrError); ok {
			attrs = append(attrs, ae.Attrs()...)
		}
		err = errors.Unwrap(err)
	}
	return attrs
}

func WithAttr(err error, args ...any) error {
	return &errWithAttrs{
		error:  err,
		atters: argsToAttr(args),
	}
}

func argsToAttr(args []any) []slog.Attr {
	attrs := make([]slog.Attr, len(args))

	for i := 0; i < len(args); {
		switch key := args[i].(type) {
		case slog.Attr:
			attrs = append(attrs, key)
			i++

		case string:
			if i+1 >= len(args) {
				attrs = append(
					attrs, slog.String("!BADKEY", key))
				i++
			} else {
				attrs = append(
					attrs, slog.Any(key, args[i+1]))
				i += 2
			}
		default:
			attrs = append(
				attrs, slog.Any("!BADKEY", args[i]))
			i++
		}
	}

	return attrs
}

type MultiError interface {
	error
	Unwrap() []error
}
