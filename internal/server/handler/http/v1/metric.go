package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
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

	router.Route(`/update/`, func(r chi.Router) {
		r.Post(`/{type}/{name}/{value}`, h.Update)
	})

	router.Get(`/value/{type}/{name}`, h.Value)
	router.Get(`/`, h.AllValues)

	return router
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {

	metricType := metric.ParseType(r.PathValue("type"))

	if !metricType.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := r.PathValue("name")

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metricValue := r.PathValue("value")

	switch metricType {
	case metric.Counter:
		value, err := strconv.Atoi(metricValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "incorrect counter metric value", http.StatusBadRequest)
		}

		err = h.service.Count(r.Context(), metricName, int64(value))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "incorrect gauge metric value", http.StatusBadRequest)
			return
		}

		err = h.service.Gauge(r.Context(), metricName, value)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) Value(w http.ResponseWriter, r *http.Request) {
	metricType := metric.ParseType(r.PathValue("type"))
	if !metricType.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := r.PathValue("name")

	m, err := h.service.Get(r.Context(), metricType, metricName)
	if err != nil {
		if errors.Is(err, metric.ErrMetricNotFound) {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	value := strconv.FormatFloat(m.Value(), 'g', -1, 64)

	w.Write([]byte(value))
}

func (h *MetricHandler) AllValues(w http.ResponseWriter, r *http.Request) {

	metrics, err := h.service.All(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body := "<html><head><title>Metrics</title></head><body><ol>"

	for _, m := range metrics {
		body += fmt.Sprintf("<li>%s: %v</li>", m.Name(), m.Value())
	}

	body += "</ol></body></html>"

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(body))
}
