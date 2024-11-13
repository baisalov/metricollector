package v1

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/baisalov/metricollector/internal/server/handler/http/middleware"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/transactions"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func metricMatcher(match string) func(id string) bool {
	return func(id string) bool {
		return match == id
	}
}

type metricStorageMock struct {
	mock.Mock
}

func (s *metricStorageMock) Get(ctx context.Context, t metric.Type, id string) (metric.Metric, error) {
	args := s.Called(ctx, t, id)
	return args.Get(0).(metric.Metric), args.Error(1)
}

func (s *metricStorageMock) Save(ctx context.Context, m metric.Metric) error {
	args := s.Called(ctx, m)
	return args.Error(0)
}

func (s *metricStorageMock) All(ctx context.Context) ([]metric.Metric, error) {
	args := s.Called(ctx)

	var metrics []metric.Metric

	for i := 0; i < len(args); i++ {
		arg := args.Get(i)
		if arg != nil {
			if m, ok := arg.(metric.Metric); ok {
				metrics = append(metrics, m)
			}
		}
	}

	return metrics, args.Error(len(args) - 1)
}

func setupServer(storage *metricStorageMock) *httptest.Server {
	router := chi.NewMux()

	serv := service.NewMetricUpdateService(storage, transactions.DiscardManager{})

	handler := NewMetricHandler(storage, serv)

	router.Use(middleware.GzipCompress, middleware.GzipDecompress)

	handler.Register(router)

	return httptest.NewServer(router)
}

func encode(t *testing.T, m metric.Metric) io.Reader {
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

	storage := &metricStorageMock{}

	existCounter := metric.NewCounterMetric("ExistCounter", 23)
	existGauge := metric.NewGaugeMetric("ExistGauge", 20)

	newCounter := metric.NewCounterMetric("NewCounter", 10)
	newGouge := metric.NewGaugeMetric("NewGouge", 10)

	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(existCounter.ID))).Return(existCounter, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(existGauge.ID))).Return(existGauge, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(newCounter.ID))).Return(metric.Metric{}, metric.ErrMetricNotFound)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(newGouge.ID))).Return(metric.Metric{}, metric.ErrMetricNotFound)
	storage.On("Save", mock.Anything, mock.Anything).Return(nil)

	server := setupServer(storage)
	defer server.Close()

	t.Run("save new counter", func(t *testing.T) {

		status, res := doRequest(t, server, "/update/", encode(t, newCounter))

		require.Equal(t, http.StatusOK, status)

		var mm metric.Metric

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, newCounter, mm)

	})

	t.Run("save new gauge", func(t *testing.T) {

		status, res := doRequest(t, server, "/update/", encode(t, newGouge))

		require.Equal(t, http.StatusOK, status)

		var mm metric.Metric

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, newGouge, mm)

	})

	t.Run("update exist counter", func(t *testing.T) {

		m := metric.NewCounterMetric("ExistCounter", 10)

		status, res := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm metric.Metric

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m.ID, mm.ID)
		assert.Equal(t, m.MType, mm.MType)
		assert.Equal(t, int64(33), *mm.Delta)
	})

	t.Run("update exist gauge", func(t *testing.T) {

		m := metric.NewGaugeMetric("ExistGauge", 30)

		status, res := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm metric.Metric

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m, mm)

	})

	t.Run("save incorrect type", func(t *testing.T) {

		val := float64(10)

		m := metric.Metric{
			ID:    "IncorrectType",
			MType: "incorrect",
			Value: &val,
		}

		status, _ := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("save incorrect value 1", func(t *testing.T) {

		m := metric.Metric{
			ID:    "IncorrectValue#1",
			MType: metric.Counter,
		}

		status, _ := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("save incorrect value 2", func(t *testing.T) {

		m := metric.Metric{
			ID:    "IncorrectValue#2",
			MType: metric.Gauge,
		}

		status, _ := doRequest(t, server, "/update/", encode(t, m))

		require.Equal(t, http.StatusBadRequest, status)
	})
}

func TestMetricHandler_ValueV2(t *testing.T) {

	storage := &metricStorageMock{}

	existCounter := metric.NewCounterMetric("ExistCounter", 10)
	existGauge := metric.NewGaugeMetric("ExistGauge", 20)

	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(existCounter.ID))).Return(existCounter, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(existGauge.ID))).Return(existGauge, nil)
	storage.On("All", mock.Anything).Return(existCounter, existGauge, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher("NotFoundCounter"))).Return(metric.Metric{}, metric.ErrMetricNotFound)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher("NotFoundGauge"))).Return(metric.Metric{}, metric.ErrMetricNotFound)

	server := setupServer(storage)
	defer server.Close()

	t.Run("counter", func(t *testing.T) {

		m := metric.Metric{
			ID:    "ExistCounter",
			MType: metric.Counter,
		}

		status, res := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm metric.Metric

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m.ID, mm.ID)
		assert.Equal(t, m.MType, mm.MType)
		assert.Equal(t, *existCounter.Delta, *mm.Delta)

	})

	t.Run("gauge", func(t *testing.T) {

		m := metric.Metric{
			ID:    "ExistGauge",
			MType: metric.Gauge,
		}

		status, res := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusOK, status)

		var mm metric.Metric

		dec := json.NewDecoder(res)

		err := dec.Decode(&mm)

		require.NoError(t, err)

		assert.Equal(t, m.ID, mm.ID)
		assert.Equal(t, m.MType, mm.MType)
		assert.Equal(t, *existGauge.Value, *mm.Value)
	})

	t.Run("not found counter", func(t *testing.T) {

		m := metric.Metric{
			ID:    "NotFoundCounter",
			MType: metric.Counter,
		}

		status, _ := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("not found gauge", func(t *testing.T) {

		val := int64(10)

		m := metric.Metric{
			ID:    "NotFoundGauge",
			MType: metric.Gauge,
			Delta: &val,
		}

		status, _ := doRequest(t, server, "/value/", encode(t, m))

		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("all", func(t *testing.T) {
		m := []metric.Metric{
			existCounter,
			existGauge,
		}

		status, res := doRequest(t, server, "/", nil)

		require.Equal(t, http.StatusOK, status)

		var mm []metric.Metric

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

	storage := &metricStorageMock{}

	storage.On("Save", mock.Anything, mock.Anything).Return(nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher("testGauge"))).Return(metric.NewGaugeMetric("testGauge", 20), nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher("testCounter"))).Return(metric.NewCounterMetric("testCounter", 20), nil)

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

	storage := &metricStorageMock{}

	var testCases []metric.Metric

	m1 := metric.NewCounterMetric("test_counter_metric", 10)
	m2 := metric.NewCounterMetric("test_gauge_metric", 15)
	m3 := metric.NewGaugeMetric("test_gauge_metric_with_pointer", 1.5000000000001)

	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(m1.ID))).Return(m1, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(m2.ID))).Return(m2, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(m3.ID))).Return(m3, nil)
	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher("not_fund"))).Return(metric.Metric{}, metric.ErrMetricNotFound)

	testCases = append(testCases, m1, m2, m3)

	server := setupServer(storage)
	defer server.Close()

	for _, tt := range testCases {
		t.Run(tt.ID, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, server.URL+fmt.Sprintf("/value/%s/%s", tt.MType, tt.ID), nil)

			request.Header.Set("Accept-Encoding", "")

			require.NoError(t, err)

			result, err := server.Client().Do(request)

			require.NoError(t, err)

			body, err := io.ReadAll(result.Body)

			require.NoError(t, err)

			err = result.Body.Close()

			require.NoError(t, err)

			require.Equal(t, http.StatusOK, result.StatusCode)

			assert.Equal(t, tt.ValueToString(), string(body))
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

	m1 := metric.NewCounterMetric("test_counter_metric", 10)
	m2 := metric.NewGaugeMetric("test_gauge_metric", 15)
	m3 := metric.NewGaugeMetric("test_gauge_metric_with_pointer", 15.100000000002)

	var metrics []metric.Metric

	metrics = append(metrics, m1, m2, m3)

	storage := &metricStorageMock{}

	storage.On("All", mock.Anything).Return(m1, m2, m3, nil)

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

		if strings.Contains(html, fmt.Sprintf("<li>%s: %v</li>", m.ID, m.ValueToString())) {
			match++
		}
	}

	assert.Equal(t, match, len(metrics))
}

func TestGzipCompress(t *testing.T) {
	storage := &metricStorageMock{}

	server := setupServer(storage)
	defer server.Close()

	existMetric := metric.NewGaugeMetric("existMetric", 10)

	storage.On("Get", mock.Anything, mock.Anything, mock.MatchedBy(metricMatcher(existMetric.ID))).Return(existMetric, nil)
	storage.On("Save", mock.Anything, mock.Anything).Return(nil)

	b, err := json.Marshal(existMetric)

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
