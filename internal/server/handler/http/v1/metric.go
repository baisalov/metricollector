package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/baisalov/metricollector/internal/server/handler/http/middleware"
	"github.com/baisalov/metricollector/internal/server/handler/http/response"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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

	acceptedContentType := middleware.AcceptedContentTypeJSON()

	router.Route(`/update/`, func(r chi.Router) {
		r.Post(`/{type}/{name}/{value}`, h.Update)
	})

	router.Get(`/value/{type}/{name}`, h.Value)
	router.Get(`/`, h.AllValues)

	router.Method(http.MethodPost, `/update/`, acceptedContentType(http.HandlerFunc(h.UpdateV2)))
	router.Method(http.MethodPost, `/value/`, acceptedContentType(http.HandlerFunc(h.ValueV2)))
	router.Method(http.MethodPost, `/`, acceptedContentType(http.HandlerFunc(h.AllValuesV2)))

	return router
}

func (h *MetricHandler) UpdateV2(w http.ResponseWriter, r *http.Request) {

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

func (h *MetricHandler) ValueV2(w http.ResponseWriter, r *http.Request) {

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

func (h *MetricHandler) AllValuesV2(w http.ResponseWriter, r *http.Request) {

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

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {

	metricType := metric.ParseType(r.PathValue("type"))

	metricName := r.PathValue("name")

	metricValue := r.PathValue("value")

	switch metricType {
	case metric.Counter:
		value, err := strconv.Atoi(metricValue)
		if err != nil {
			http.Error(w, "incorrect counter metric value", http.StatusBadRequest)
			return
		}

		_, err = h.service.Count(r.Context(), metricName, int64(value))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case metric.Gauge:
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
	default:
		http.Error(w, "incorrect metric type", http.StatusBadRequest)
		return
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
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	value := strconv.FormatFloat(m.Value(), 'g', -1, 64)

	_, err = w.Write([]byte(value))
	if err != nil {
		log.Println("Failed to write response body: ", err.Error())
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricHandler) AllValues(w http.ResponseWriter, r *http.Request) {

	metrics, err := h.service.All(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body strings.Builder

	_, err = body.WriteString("<html><head><title>Metrics</title></head><body><ol>")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, m := range metrics {
		_, err = fmt.Fprintf(&body, "<li>%s: %v</li>", m.Name(), m.Value())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	_, err = body.WriteString("</ol></body></html>")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	_, err = w.Write([]byte(body.String()))
	if err != nil {
		log.Println("Failed to write response body: ", err.Error())
	}

	w.WriteHeader(http.StatusOK)
}
