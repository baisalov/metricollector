package v1

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/go-chi/chi/v5"
	"io"
	"log/slog"
	"net/http"
)

type MetricHandler struct {
	service metricService
}

type metricService interface {
	Count(ctx context.Context, name string, value int64) error
	Gauge(ctx context.Context, name string, value float64) error
	Get(ctx context.Context, t metric.Type, name string) (metric.Metric, error)
	All(ctx context.Context) ([]metric.Metric, error)
}

func NewMetricHandler(service metricService) *MetricHandler {
	return &MetricHandler{
		service: service,
	}
}

func (h *MetricHandler) Handler() http.Handler {

	router := chi.NewRouter()

	router.Post(`/update/`, h.Update)
	router.Get(`/value/`, h.Value)
	router.Get(`/`, h.AllValues)

	return router
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var m Metrics

	if err := decoder.Decode(&m); err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := m.ValidateForUpdate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch m.Type() {
	case metric.Counter:

		err := h.service.Count(r.Context(), m.ID, *m.Delta)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case metric.Gauge:
		err := h.service.Gauge(r.Context(), m.ID, *m.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "incorrect metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) Value(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var request Metrics

	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := request.ValidateForValue(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := h.service.Get(r.Context(), request.Type(), request.ID)
	if err != nil {
		if errors.Is(err, metric.ErrMetricNotFound) {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConvertMetric(m)

	body, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(body)
	if err != nil {
		slog.Error("Failed to write response body", "error", err)
	}
}

func (h *MetricHandler) AllValues(w http.ResponseWriter, r *http.Request) {

	metrics, err := h.service.All(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var res []Metrics

	for _, m := range metrics {
		res = append(res, ConvertMetric(m))
	}

	body, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(body)
	if err != nil {
		slog.Error("Failed to write response body", "error", err)
	}
}
