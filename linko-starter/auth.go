package main

import (
	"context"
	"errors"
	"net/http"

	pkgerr "github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserContextKey contextKey = "user"

var allowedUsers = map[string]string{
	"frodo":   "$2a$10$B6O/n6teuCzpuh66jrUAdeaJ3WvXcxRkzpN0x7H.di9G9e/NGb9Me",
	"samwise": "$2a$10$EWZpvYhUJtJcEMmm/IBOsOGIcpxUnGIVMRiDlN/nxl1RRwWGkJtty",
	// frodo: "ofTheNineFingers"
	// samwise: "theStrong"
	"saruman": "invalidFormat",
}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			httpError(
				r.Context(),
				w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		stored, exists := allowedUsers[username]
		if !exists {
			httpError(
				r.Context(),
				w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		ok, err := s.validatePassword(password, stored)
		if err != nil {
			s.logger.Debug("error validating password",
				"username:", username,
				"error:", err,
			)
			httpError(
				r.Context(),
				w, http.StatusInternalServerError, errors.New("Internal Server Error"))
			return
		}

		if !ok {
			httpError(
				r.Context(),
				w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		logCtx := r.Context().Value(logContextKey).(*LogContext)
		logCtx.Username = username

		r = r.WithContext(
			context.WithValue(r.Context(), UserContextKey, username),
		)
		next.ServeHTTP(w, r)
	})
}

func (s *server) validatePassword(password, stored string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(
		[]byte(stored),
		[]byte(password),
	)
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, pkgerr.WithStack(err)
	}
	return true, nil
}

func httpError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	if logCtx, ok := ctx.Value(logContextKey).(*LogContext); ok {
		logCtx.Error = err
	}

	switch status {
	case http.StatusUnauthorized:
		http.Error(w, "uauthorized request", status)

	case http.StatusForbidden:
		http.Error(w, "forbidden request", status)

	case http.StatusInternalServerError:
		http.Error(w, "internal server error", status)

	default:
		http.Error(w, err.Error(), status)
	}
}
