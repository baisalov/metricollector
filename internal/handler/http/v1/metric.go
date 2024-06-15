package v1

import (
	"github.com/baisalov/metricollector/internal/server/service"
	"net/http"
	"strconv"
	"strings"
)

type MetricHandler struct {
	service *service.MetricService
}

func NewMetricHandler(metricService *service.MetricService) *MetricHandler {
	return &MetricHandler{service: metricService}
}

func (h *MetricHandler) GougeHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	name = strings.TrimSpace(name)
	if name == "" {
		http.Error(w, "incorrect metric name", http.StatusBadRequest)
		return
	}

	value, err := strconv.ParseFloat(r.PathValue("value"), 64)
	if err != nil {
		http.Error(w, "incorrect metric value", http.StatusBadRequest)
		return
	}

	err = h.service.Gouge(name, value)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) CounterHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	name = strings.TrimSpace(name)
	if name == "" {
		http.Error(w, "incorrect metric name", http.StatusBadRequest)
		return
	}

	value, err := strconv.Atoi(r.PathValue("value"))
	if err != nil {
		http.Error(w, "incorrect metric value", http.StatusBadRequest)
		return
	}

	err = h.service.Count(name, int64(value))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
