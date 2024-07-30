package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/baisalov/metricollector/internal/server/handler/http/middleware"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupServer(storage *memory.MetricStorage) *httptest.Server {
	serv := service.NewMetricService(storage)

	jsonContent := middleware.AcceptedContentTypeJson()

	handler := NewMetricHandler(serv)

	return httptest.NewServer(jsonContent(handler.Handler()))
}

func encode(t *testing.T, m Metrics) io.Reader {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	err := enc.Encode(m)

	require.NoError(t, err)

	return &buf
}

func doRequest(t *testing.T, server *httptest.Server, url string, r io.Reader) (int, io.Reader) {

	request, err := http.NewRequest(http.MethodPost, server.URL+url, r)

	require.NoError(t, err)

	request.Header.Set("Content-Type", "application/json")

	result, err := server.Client().Do(request)

	require.NoError(t, err)

	var res bytes.Buffer

	_, err = io.Copy(&res, result.Body)

	require.NoError(t, err)

	err = result.Body.Close()

	require.NoError(t, err)

	require.Equal(t, "application/json", result.Header.Get("Content-Type"))

	return result.StatusCode, &res
}

func TestMetricHandler_Update(t *testing.T) {

	storage := memory.NewMetricStorage()

	storage.Save(context.Background(), metric.NewCounterMetric("IssetCounter", 10))

	storage.Save(context.Background(), metric.NewGaugeMetric("IssetGauge", 20))

	server := setupServer(storage)
	defer server.Close()

	t.Run("save counter", func(t *testing.T) {

		val := int64(10)

		m := Metrics{
			ID:    "NewCounter",
			MType: metric.Counter.String(),
			Delta: &val,
		}

		status, res := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m, mm)

	})

	t.Run("save gauge", func(t *testing.T) {

		val := float64(10)

		m := Metrics{
			ID:    "NewGouge",
			MType: metric.Gauge.String(),
			Value: &val,
		}

		status, res := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m, mm)

	})

	t.Run("save isset counter", func(t *testing.T) {

		val := int64(10)

		m := Metrics{
			ID:    "IssetCounter",
			MType: metric.Counter.String(),
			Delta: &val,
		}

		status, res := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m.ID, mm.ID)
		assert.Equal(t, m.MType, mm.MType)
		assert.Equal(t, int64(20), *mm.Delta)
	})

	t.Run("save isset gauge", func(t *testing.T) {

		val := float64(10)

		m := Metrics{
			ID:    "IssetGouge",
			MType: metric.Gauge.String(),
			Value: &val,
		}

		status, res := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m, mm)

	})

	t.Run("save incorrect type", func(t *testing.T) {

		val := float64(10)

		m := Metrics{
			ID:    "IncorrectType",
			MType: "incorrect",
			Value: &val,
		}

		status, _ := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("save incorrect value 1", func(t *testing.T) {

		m := Metrics{
			ID:    "IncorrectValue#1",
			MType: metric.Counter.String(),
		}

		status, _ := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("save incorrect value 2", func(t *testing.T) {

		m := Metrics{
			ID:    "IncorrectValue#2",
			MType: metric.Gauge.String(),
		}

		status, _ := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusBadRequest, status)
	})
}

func TestMetricHandler_Value(t *testing.T) {

	storage := memory.NewMetricStorage()

	issetCounter := metric.NewCounterMetric("IssetCounter", 10)
	storage.Save(context.Background(), issetCounter)

	issetGauge := metric.NewGaugeMetric("IssetGauge", 20)
	storage.Save(context.Background(), issetGauge)

	server := setupServer(storage)
	defer server.Close()

	t.Run("counter", func(t *testing.T) {

		m := Metrics{
			ID:    "IssetCounter",
			MType: metric.Counter.String(),
		}

		status, res := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m.ID, mm.ID)
		assert.Equal(t, m.MType, mm.MType)
		assert.Equal(t, int64(issetCounter.Value()), *mm.Delta)

	})

	t.Run("gauge", func(t *testing.T) {

		m := Metrics{
			ID:    "IssetGauge",
			MType: metric.Gauge.String(),
		}

		status, res := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m.ID, mm.ID)
		assert.Equal(t, m.MType, mm.MType)
		assert.Equal(t, issetGauge.Value(), *mm.Value)
	})

	t.Run("not found counter", func(t *testing.T) {

		m := Metrics{
			ID:    "NotFoundCounter",
			MType: metric.Counter.String(),
		}

		status, _ := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("not found gauge", func(t *testing.T) {

		val := int64(10)

		m := Metrics{
			ID:    "NotFoundGauge",
			MType: metric.Gauge.String(),
			Delta: &val,
		}

		status, _ := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("all", func(t *testing.T) {
		m := []Metrics{
			ConvertMetric(issetCounter),
			ConvertMetric(issetGauge),
		}

		status, res := doRequest(t, server, "/", nil)

		require.Equal(t, http.StatusOK, status)

		var mm []Metrics

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.ElementsMatch(t, m, mm)
	})
}
