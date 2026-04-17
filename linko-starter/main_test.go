package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_requestLogger(t *testing.T) {
	logBuffer := &bytes.Buffer{}

	logger := slog.New(
		slog.NewTextHandler(
			logBuffer, &slog.HandlerOptions{
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Time(
							slog.TimeKey, time.Date(
								2023, 10, 1, 12, 34, 57, 0, time.UTC),
						)
					}
					return a
				},
			}))

	requestLoggerMiddleware := requestLogger(logger)
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	loggedHandler := requestLoggerMiddleware(dummyHandler)

	req := httptest.NewRequest("GET", "http://lin.ko/api/stats", nil)
	rr := httptest.NewRecorder()
	loggedHandler.ServeHTTP(rr, req)

	const expectedStatusCode = http.StatusOK

	actualLogBuffer := logBuffer.String()

	if !strings.Contains(actualLogBuffer, "Served request") {
		t.Error("missing message")
	}

	if !strings.Contains(actualLogBuffer, "method=GET") {
		t.Error("missing method")
	}

	if !strings.Contains(actualLogBuffer, "client_ip") {
		t.Error("missing method")
	}

	if expectedStatusCode != rr.Code {
		t.Errorf("expected %d got: %d", expectedStatusCode, rr.Code)
	}
}
