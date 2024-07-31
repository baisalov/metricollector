package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipCompress(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newGzipWriter(w)

			ow = cw

			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
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
	contentType := w.ResponseWriter.Header().Get("Content-Type")

	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
		w.ResponseWriter.Header().Del("Content-Length")

		return w.writer.Write(p)
	}

	return w.ResponseWriter.Write(p)
}

func (w gzipWriter) Close() error {
	return w.writer.Close()
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
