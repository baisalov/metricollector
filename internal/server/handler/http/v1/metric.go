package v1

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/baisalov/metricollector/internal/server/handler/http/response"
	"github.com/go-chi/chi/v5"
	"io"
	"log/slog"
	"net/http"
)

type MetricHandler struct {
	service metricService
}

type metricService interface {
	Count(ctx context.Context, name string, value int64) (int64, error)
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
	router.Post(`/value/`, h.Value)
	router.Post(`/`, h.AllValues)

	return router
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var request Metrics

	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			response.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := request.ValidateForUpdate(); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Debug("Update", "request", request)

	switch request.Type() {
	case metric.Counter:
		d, err := h.service.Count(r.Context(), request.ID, *request.Delta)
		if err != nil {
			response.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		request.Delta = &d
	case metric.Gauge:
		err := h.service.Gauge(r.Context(), request.ID, *request.Value)
		if err != nil {
			response.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		response.Error(w, "incorrect metric type", http.StatusBadRequest)
		return
	}

	response.Success(w, request)
}

func (h *MetricHandler) Value(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var request Metrics

	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			response.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := request.ValidateForValue(); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Debug("Value", "request", request)

	m, err := h.service.Get(r.Context(), request.Type(), request.ID)
	if err != nil {
		if errors.Is(err, metric.ErrMetricNotFound) {
			response.Error(w, "metric not found", http.StatusNotFound)
			return
		}

		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(w, ConvertMetric(m))
}

func (h *MetricHandler) AllValues(w http.ResponseWriter, r *http.Request) {

	metrics, err := h.service.All(r.Context())
	if err != nil {
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var res []Metrics

	for _, m := range metrics {
		res = append(res, ConvertMetric(m))
	}

	response.Success(w, res)
}
