package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const logContextKey contextKey = "log_context"

type LogContext struct {
	Username string
	Error    error
}

type HandlerReturnFunc = func(http.Handler) http.Handler

func RequestIdMiddleWare() HandlerReturnFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqId := r.Header.Get("X-Request-ID")
			if strings.TrimSpace(reqId) == "" {
				reqId = uuid.NewString()
			}

			ctxedReq := r.WithContext(context.WithValue(r.Context(), "req_id", reqId))

			w.Header().Set("X-Request-ID", reqId)

			next.ServeHTTP(w, ctxedReq)
		},
		)
	}
}

func requestLogger(logger *slog.Logger) HandlerReturnFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				spyReader := &SpyReadCloser{
					ReadCloser: r.Body,
				}
				spyWriter := &spyResponseWriter{ResponseWriter: w}

				r.Body = spyReader

				logCtx := &LogContext{}
				ctxedReq := r.WithContext(
					context.WithValue(r.Context(), logContextKey, logCtx))

				startTime := time.Now()

				next.ServeHTTP(spyWriter, ctxedReq)

				username := "anonymous"
				if logCtx.Username != "" {
					username = logCtx.Username
				}

				reqId := ctxedReq.Context().Value("req_id").(string)

				logAttrs := []slog.Attr{
					slog.String("method", ctxedReq.Method),
					slog.String("url", ctxedReq.URL.Path),
					slog.String("client_ip",
						redactIP(
							ctxedReq.RemoteAddr,
						),
					),
					slog.Duration("duration", time.Since(startTime)),
					slog.Int("request_body_bytes", spyReader.bytesRead),
					slog.String("request_id", reqId),
					slog.String("username", username),
					slog.Int("response_status", spyWriter.statusCode),
					slog.Int("response_body_bytes", spyWriter.bytesWritten),
				}

				if logCtx.Error != nil {
					logAttrs = append(logAttrs, slog.Any("http_error", logCtx.Error))
				}

				logger.LogAttrs(
					r.Context(), slog.LevelInfo, "Served request:", logAttrs...)
			})
	}
}

type SpyReadCloser struct {
	io.ReadCloser
	bytesRead int
}

func (r *SpyReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	r.bytesRead += n
	return n, err
}

type spyResponseWriter struct {
	http.ResponseWriter
	bytesWritten int
	statusCode   int
}

func (w *spyResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	n, err := w.ResponseWriter.Write(p)
	w.bytesWritten += n

	return n, err
}

func (w *spyResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
