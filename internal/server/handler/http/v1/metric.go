package v1

import (
	"context"
	"github.com/baisalov/metricollector/internal/metric"
	"net/http"
	"strconv"
)

type MetricHandler struct {
	service metricService
}

type metricService interface {
	Count(ctx context.Context, name string, value int64) error
	Gauge(ctx context.Context, name string, value float64) error
}

func NewMetricHandler(service metricService) *MetricHandler {
	return &MetricHandler{
		service: service,
	}
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
