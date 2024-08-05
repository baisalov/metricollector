package middleware

import (
	"compress/gzip"
	"log/slog"
	"net/http"
	"strings"
)

const (
	_contentEncoding = "Content-Encoding"
	_acceptEncoding  = "Accept-Encoding"
)

func GzipCompress(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ow := w

		acceptEncoding := r.Header.Get(_acceptEncoding)
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newGzipWriter(w)

			ow = cw

			defer func() {
				if err := cw.Close(); err != nil {
					slog.Error("failed close gzip writer", "error", err)
				}
			}()
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(fn)
}

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		writer:         gzip.NewWriter(w),
	}
}

func (w gzipWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w gzipWriter) Write(p []byte) (int, error) {
	contentEncoding := w.ResponseWriter.Header().Get(_contentEncoding)
	sendsGzip := strings.Contains(contentEncoding, "gzip")

	if sendsGzip {
		return w.writer.Write(p)
	}

	return w.ResponseWriter.Write(p)
}

func (w gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		contentType := w.ResponseWriter.Header().Get("Content-Type")

		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			w.ResponseWriter.Header().Set(_contentEncoding, "gzip")
			w.ResponseWriter.Header().Add("Vary", _acceptEncoding)
			w.ResponseWriter.Header().Del("Content-Length")
		}
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w gzipWriter) Close() error {
	return w.writer.Close()
}
