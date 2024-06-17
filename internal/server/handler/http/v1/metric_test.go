package v1

import (
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
	name       string
	request    string
	statusCode int
}{
	{
		"update gauge metric",
		"/update/gauge/testGauge/42.42",
		http.StatusOK,
	},
	{
		"update counter metric",
		"/update/counter/testCounter/10",
		http.StatusOK,
	},
	{
		"update gauge metric without name",
		"/update/gauge",
		http.StatusNotFound,
	},
	{
		"update gauge metric without value #1",
		"/update/gauge/testGauge",
		http.StatusNotFound,
	},
	{
		"update counter metric without value #2",
		"/update/counter/testCounter",
		http.StatusNotFound,
	},
	{
		"update invalid metric type",
		"/update/invalidType/testMetric/42.42",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #1",
		"/update/gauge/testMetric/invalidValue",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #2",
		"/update/gauge/testMetric/2invalidValue",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #3",
		"/update/gauge/testMetric/2,3",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #4",
		"/update/counter/testMetric/invalidValue",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #4",
		"/update/counter/testMetric/2invalidValue",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #6",
		"/update/counter/testMetric/2,3",
		http.StatusBadRequest,
	},
	{
		"update metric with invalid value #7",
		"/update/counter/testMetric/2.3",
		http.StatusBadRequest,
	},
}

func TestMetricUpdateHandler(t *testing.T) {
	storage := memory.NewMetricStorage()

	serv := service.NewMetricService(storage)

	handler := NewMetricHandler(serv)

	mux := http.NewServeMux()

	mux.HandleFunc(`POST /update/{type}/{name}/{value}`, handler.Update)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, request)

			result := w.Result()

			err := result.Body.Close()

			require.NoError(t, err)

			assert.Equal(t, tt.statusCode, result.StatusCode)
		})
	}
}
