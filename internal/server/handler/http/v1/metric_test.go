package v1

import (
	"context"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupHandler(storage *memory.MetricStorage) http.Handler {
	serv := service.NewMetricService(storage)

	handler := NewMetricHandler(serv)

	updateHandler := http.NewServeMux()
	updateHandler.HandleFunc(`POST /{type}/{name}/{value}`, handler.Update)
	updateHandler.HandleFunc(`POST /{type}`, http.NotFound)

	mux := http.NewServeMux()

	mux.Handle(`POST /update/`, http.StripPrefix("/update", updateHandler))
	mux.HandleFunc(`GET /`, handler.AllValues)
	mux.HandleFunc("POST /", http.NotFound)

	return mux
}

func TestMetricHandler_Update(t *testing.T) {

	var testCases = []struct {
		name       string
		request    string
		statusCode int
	}{
		{
			"gauge metric",
			"/update/gauge/testGauge/42.42",
			http.StatusOK,
		},
		{
			"counter metric",
			"/update/counter/testCounter/10",
			http.StatusOK,
		},
		{
			"gauge metric without name",
			"/update/gauge",
			http.StatusNotFound,
		},
		{
			"gauge metric without value #1",
			"/update/gauge/testGauge",
			http.StatusNotFound,
		},
		{
			"counter metric without value #2",
			"/update/counter/testCounter",
			http.StatusNotFound,
		},
		{
			"invalid metric type",
			"/update/invalidType/testMetric/42.42",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #1",
			"/update/gauge/testMetric/invalidValue",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #2",
			"/update/gauge/testMetric/2invalidValue",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #3",
			"/update/gauge/testMetric/2,3",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #4",
			"/update/counter/testMetric/invalidValue",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #4",
			"/update/counter/testMetric/2invalidValue",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #6",
			"/update/counter/testMetric/2,3",
			http.StatusBadRequest,
		},
		{
			"metric with invalid value #7",
			"/update/counter/testMetric/2.3",
			http.StatusBadRequest,
		},
	}

	storage := memory.NewMetricStorage()

	handler := setupHandler(storage)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			result := w.Result()

			err := result.Body.Close()

			require.NoError(t, err)

			assert.Equal(t, tt.statusCode, result.StatusCode)
		})
	}
}

func TestMetricHandler_Value(t *testing.T) {

	storage := memory.NewMetricStorage()

	ctx := context.TODO()

	var testCases []metric.Metric

	m1 := metric.NewCounterMetric("test_counter_metric", 10)

	err := storage.Save(ctx, m1)

	require.NoError(t, err)

	m2 := metric.NewCounterMetric("test_gauge_metric", 15)

	err = storage.Save(ctx, m2)

	require.NoError(t, err)

	m3 := metric.NewGaugeMetric("test_gauge_metric_with_pointer", 1.5000000000001)

	err = storage.Save(ctx, m3)

	require.NoError(t, err)

	testCases = append(testCases, m1, m2, m3)

	handler := setupHandler(storage)

	for _, tt := range testCases {
		t.Run(tt.Name(), func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/%s", tt.Type(), tt.Name()), nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			result := w.Result()

			body, err := io.ReadAll(result.Body)

			require.NoError(t, err)

			err = result.Body.Close()

			require.NoError(t, err)

			require.Equal(t, http.StatusOK, result.StatusCode)

			assert.Equal(t, string(body), strconv.FormatFloat(tt.Value(), 'g', -1, 64))
		})
	}

	t.Run("bad request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/%s", "incorrect_type", "not_fund"), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, request)

		result := w.Result()

		err = result.Body.Close()

		require.NoError(t, err)

		require.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/%s", "counter", "not_fund"), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, request)

		result := w.Result()

		err = result.Body.Close()

		require.NoError(t, err)

		require.Equal(t, http.StatusNotFound, result.StatusCode)
	})

}

func TestMetricHandler_AllValues(t *testing.T) {

	var metrics []metric.Metric

	metrics = append(metrics,
		metric.NewCounterMetric("test_counter_metric", 10),
		metric.NewGaugeMetric("test_gauge_metric", 15),
		metric.NewGaugeMetric("test_gauge_metric_with_pointer", 15.100000000002))

	storage := memory.NewMetricStorage()

	ctx := context.TODO()

	for _, m := range metrics {
		err := storage.Save(ctx, m)

		require.NoError(t, err)
	}

	handler := setupHandler(storage)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, request)

	result := w.Result()

	body, err := io.ReadAll(result.Body)

	require.NoError(t, err)

	err = result.Body.Close()

	require.NoError(t, err)

	require.Equal(t, http.StatusOK, result.StatusCode)

	require.Equal(t, result.Header.Get("Content-Type"), "text/html")

	html := string(body)

	match := 0

	for _, m := range metrics {
		if strings.Contains(html, fmt.Sprintf("<li>%s: %v</li>", m.Name(), m.Value())) {
			match++
		}
	}

	assert.Equal(t, match, len(metrics))
}
