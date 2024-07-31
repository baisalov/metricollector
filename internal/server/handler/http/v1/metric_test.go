package v1

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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

func setupServer(storage *memory.MetricStorage) *httptest.Server {
	serv := service.NewMetricService(storage)

	handler := NewMetricHandler(serv)

	return httptest.NewServer(handler.Handler())
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

func TestMetricHandler_UpdateV2(t *testing.T) {

	storage := memory.NewMetricStorage()

	err := storage.Save(context.Background(), metric.NewCounterMetric("IssetCounter", 10))

	require.NoError(t, err)

	err = storage.Save(context.Background(), metric.NewGaugeMetric("IssetGauge", 20))

	require.NoError(t, err)

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

func TestMetricHandler_ValueV2(t *testing.T) {

	storage := memory.NewMetricStorage()

	issetCounter := metric.NewCounterMetric("IssetCounter", 10)

	err := storage.Save(context.Background(), issetCounter)

	require.NoError(t, err)

	issetGauge := metric.NewGaugeMetric("IssetGauge", 20)

	err = storage.Save(context.Background(), issetGauge)

	require.NoError(t, err)

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

	server := setupServer(storage)
	defer server.Close()

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, server.URL+tt.request, nil)

			require.NoError(t, err)

			result, err := server.Client().Do(request)

			require.NoError(t, err)

			err = result.Body.Close()

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

	server := setupServer(storage)
	defer server.Close()

	for _, tt := range testCases {
		t.Run(tt.Name(), func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, server.URL+fmt.Sprintf("/value/%s/%s", tt.Type(), tt.Name()), nil)

			require.NoError(t, err)

			result, err := server.Client().Do(request)

			require.NoError(t, err)

			body, err := io.ReadAll(result.Body)

			require.NoError(t, err)

			err = result.Body.Close()

			require.NoError(t, err)

			require.Equal(t, http.StatusOK, result.StatusCode)

			assert.Equal(t, string(body), strconv.FormatFloat(tt.Value(), 'g', -1, 64))
		})
	}

	t.Run("bad request", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, server.URL+fmt.Sprintf("/value/%s/%s", "incorrect_type", "not_fund"), nil)

		require.NoError(t, err)

		result, err := server.Client().Do(request)

		require.NoError(t, err)

		err = result.Body.Close()

		require.NoError(t, err)

		require.Equal(t, http.StatusBadRequest, result.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, server.URL+fmt.Sprintf("/value/%s/%s", "counter", "not_fund"), nil)

		require.NoError(t, err)

		result, err := server.Client().Do(request)

		require.NoError(t, err)

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

	server := setupServer(storage)
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)

	require.NoError(t, err)

	result, err := server.Client().Do(request)

	require.NoError(t, err)

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

func TestGzipCompress(t *testing.T) {
	storage := memory.NewMetricStorage()

	server := setupServer(storage)
	defer server.Close()

	val := float64(10)

	m := Metrics{
		ID:    "test",
		MType: "gauge",
		Delta: nil,
		Value: &val,
	}

	b, err := json.Marshal(m)

	require.NoError(t, err)

	t.Run("sends_gzip", func(t *testing.T) {

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)

		_, err := zb.Write(b)

		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r, err := http.NewRequest("POST", server.URL+"/update/", buf)

		require.NoError(t, err)

		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		bb, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, string(b), string(bb))

		err = resp.Body.Close()

		require.NoError(t, err)
	})

	t.Run("accepts_gzip", func(t *testing.T) {

		buf := bytes.NewBuffer(b)

		r, err := http.NewRequest("POST", server.URL+"/update/", buf)

		require.NoError(t, err)

		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		bb, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, string(b), string(bb))

		err = resp.Body.Close()

		require.NoError(t, err)
	})
}
