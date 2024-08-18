package v1

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type HealthCheckHandler struct {
	ch checker
}

func NewHealthCheckHandler(ch checker) *HealthCheckHandler {
	return &HealthCheckHandler{ch: ch}
}

type checker interface {
	Check(ctx context.Context) error
}

func (h *HealthCheckHandler) Register(mux *chi.Mux) {
	mux.Get("/ping", h.Check)
}

func (h *HealthCheckHandler) Check(w http.ResponseWriter, r *http.Request) {
	if err := h.ch.Check(r.Context()); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
