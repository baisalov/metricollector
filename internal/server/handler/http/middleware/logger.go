package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	d *responseData
}

func (w loggingResponseWriter) Write(bytes []byte) (int, error) {
	size, err := w.ResponseWriter.Write(bytes)

	w.d.size += size

	return size, err
}

func (w loggingResponseWriter) WriteHeader(statusCode int) {
	w.d.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogging(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ww := loggingResponseWriter{
			ResponseWriter: w,
			d: &responseData{
				size:   0,
				status: 0,
			},
		}

		t1 := time.Now()

		next.ServeHTTP(ww, r)

		slog.Info(r.RequestURI,
			"method", r.Method,
			"status", ww.d.status,
			"duration", time.Since(t1).Milliseconds(),
			"body", ww.d.size)
	}

	return http.HandlerFunc(fn)
}
