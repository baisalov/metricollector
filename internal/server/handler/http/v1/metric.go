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
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type MetricHandler struct {
	provider metricProvider
	updater  metricUpdater
}

type metricUpdater interface {
	Update(ctx context.Context, m metric.Metric) (metric.Metric, error)
}

type metricProvider interface {
	Get(ctx context.Context, t metric.Type, id string) (metric.Metric, error)
	All(ctx context.Context) ([]metric.Metric, error)
}

func NewMetricHandler(provider metricProvider, updater metricUpdater) *MetricHandler {
	return &MetricHandler{
		provider: provider,
		updater:  updater,
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

	return middleware.GzipDecompress(middleware.GzipCompress(router))
}

func (h *MetricHandler) UpdateV2(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var request metric.Metric

	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			response.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		if errors.Is(err, metric.ErrIncorrectType) {
			response.Error(w, metric.ErrIncorrectType.Error(), http.StatusBadRequest)
			return
		}

		slog.Error("failed to decode request", "error", err)
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := request.Validate(); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Debug("Update", "request", request)

	res, err := h.updater.Update(r.Context(), request)
	if err != nil {
		slog.Error("failed to update metric", "error", err)
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(w, res)
}

func (h *MetricHandler) ValueV2(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var req metric.Metric

	if err := decoder.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			response.Error(w, "empty request body", http.StatusBadRequest)
			return
		}

		if errors.Is(err, metric.ErrIncorrectType) {
			response.Error(w, metric.ErrIncorrectType.Error(), http.StatusBadRequest)
			return
		}

		slog.Error("failed to decode request", "error", err)
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.TrimSpace(req.ID) == "" {
		response.Error(w, "empty metric name", http.StatusBadRequest)
	}

	slog.Debug("Value", "request", req)

	res, err := h.provider.Get(r.Context(), req.MType, req.ID)
	if err != nil {
		if errors.Is(err, metric.ErrMetricNotFound) {
			response.Error(w, "metric not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to get metric", "error", err)
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(w, res)
}

func (h *MetricHandler) AllValuesV2(w http.ResponseWriter, r *http.Request) {

	res, err := h.provider.All(r.Context())
	if err != nil {
		slog.Error("failed to get metrics", "error", err)
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(w, res)
}

func (h *MetricHandler) Update(w http.ResponseWriter, r *http.Request) {

	m := metric.Metric{
		MType: metric.ParseType(r.PathValue("type")),
		ID:    r.PathValue("name"),
	}

	metricValue := r.PathValue("value")

	if m.MType == metric.Counter {
		value, err := strconv.Atoi(metricValue)
		if err != nil {
			http.Error(w, "incorrect counter metric value", http.StatusBadRequest)
			return
		}

		delta := int64(value)

		m.Delta = &delta

	} else {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "incorrect gauge metric value", http.StatusBadRequest)
			return
		}

		m.Value = &value
	}

	if err := m.Validate(); err != nil {
		response.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.updater.Update(r.Context(), m)
	if err != nil {
		slog.Error("failed to update metric", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	m, err := h.provider.Get(r.Context(), metricType, metricName)
	if err != nil {
		if errors.Is(err, metric.ErrMetricNotFound) {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to get metric", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)

	var value string

	if m.MType == metric.Counter {
		value = strconv.FormatInt(*m.Delta, 10)
	} else {
		value = strconv.FormatFloat(*m.Value, 'f', 10, 64)
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		slog.Error("Failed to write response body", "error", err)
	}
}

func (h *MetricHandler) AllValues(w http.ResponseWriter, r *http.Request) {

	res, err := h.provider.All(r.Context())
	if err != nil {
		slog.Error("failed to get metrics", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body strings.Builder

	_, err = body.WriteString("<html><head><title>Metrics</title></head><body><ol>")
	if err != nil {
		slog.Error("failed to write content header", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var value string

	for _, m := range res {

		if m.MType == metric.Counter {
			value = strconv.FormatInt(*m.Delta, 10)
		} else {
			value = strconv.FormatFloat(*m.Value, 'f', 10, 64)
		}

		_, err = fmt.Fprintf(&body, "<li>%s: %v</li>", m.ID, value)
		if err != nil {
			slog.Error("failed to write content body", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	_, err = body.WriteString("</ol></body></html>")
	if err != nil {
		slog.Error("failed to write content bottom", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	w.WriteHeader(http.StatusOK)

	_, err = w.Write([]byte(body.String()))
	if err != nil {
		slog.Error("Failed to write response body", err)
	}
}
