package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func GzipDecompress(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		contentEncoding := r.Header.Get(_contentEncoding)
		sendsGzip := strings.Contains(contentEncoding, "gzip") && r.ContentLength > 0

		if sendsGzip {
			cr, err := newGzipReader(r.Body)
			if err != nil {
				slog.Error("failed to create new gzip reader", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return

			}

			r.Body = cr

			defer func() {
				if err := cr.Close(); err != nil {
					slog.Error("failed close gzip reader", "error", err)
				}
			}()

		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
