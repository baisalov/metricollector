package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type hashStatusReWriter struct {
	http.ResponseWriter
}

func (w hashStatusReWriter) WriteHeader(statusCode int) {
	if statusCode < 500 {
		w.ResponseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func HashCheck(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Header.Get("HashSHA256") != "" && r.Method == http.MethodPost {

				isCorrect, err := checkHash(r, key)
				if err != nil {
					slog.Error(err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if !isCorrect {
					slog.Warn("invalid body sign")
					w = hashStatusReWriter{ResponseWriter: w}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func checkHash(r *http.Request, key string) (bool, error) {
	sum, err := hex.DecodeString(r.Header.Get("HashSHA256"))
	if err != nil {
		return false, fmt.Errorf("failed to decode hex string: %w", err)
	}

	var buf bytes.Buffer

	tr := io.TeeReader(r.Body, &buf)

	data, err := io.ReadAll(tr)
	if err != nil {
		return false, fmt.Errorf("failed to read request body: %w", err)
	}

	r.Body = io.NopCloser(&buf)

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	sign := h.Sum(nil)

	return bytes.EqualFold(sign, sum[:]), nil
}
