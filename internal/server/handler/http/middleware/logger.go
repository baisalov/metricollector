package middleware

import (
	"log"
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

func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			log.Printf("| %v | %v | %v | %v", r.RequestURI, ww.d.status, time.Since(t1), ww.d.size)
		}

		return http.HandlerFunc(fn)
	}

}
