package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
)

func HashCheck(hashKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("HashSHA256")

			if header != "" {

				slog.Debug("hash header", "sum", header)

				sum, err := hex.DecodeString(header)
				if err != nil {
					slog.Error("failed to decode hex string", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				var buf bytes.Buffer

				tr := io.TeeReader(r.Body, &buf)

				data, err := io.ReadAll(tr)
				if err != nil {
					slog.Error("failed to read request body", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(&buf)

				h := hmac.New(sha256.New, []byte(hashKey))
				h.Write(data)
				sign := h.Sum(nil)

				if !bytes.EqualFold(sign, sum[:]) {
					slog.Error("invalid body sign")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
