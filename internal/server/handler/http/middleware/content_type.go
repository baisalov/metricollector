package middleware

import (
	"net/http"
)

func AcceptedContentTypeJSON() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}

}
