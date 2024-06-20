package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/baisalov/metricollector/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricStorage struct {
	mock.Mock
}

func (s *MockMetricStorage) Get(ctx context.Context, t metric.Type, name string) (metric.Metric, error) {
	args := s.Called(ctx, t, name)

	var m metric.Metric

	arg := args.Get(0)
	if arg != nil {
		m = arg.(metric.Metric)
	}

	return m, args.Error(1)
}

func (s *MockMetricStorage) Save(ctx context.Context, metric metric.Metric) error {
	args := s.Called(ctx, metric)
	return args.Error(0)
}

func (s *MockMetricStorage) All(ctx context.Context) ([]metric.Metric, error) {
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

func TestMetricService_Gauge(t *testing.T) {
	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()
	name := "test_metric"
	value := 42.0
	gaugeMetric := metric.NewGaugeMetric(name, value)

	storage.On("Save", ctx, gaugeMetric).Return(nil)

	err := service.Gauge(ctx, name, value)
	assert.NoError(t, err)

	storage.AssertExpectations(t)
}

func TestMetricService_Count(t *testing.T) {
	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()
	name := "test_metric"
	value := int64(5)
	counterMetric := metric.NewCounterMetric(name, value)

	storage.On("Get", ctx, metric.Counter, name).Return(nil, metric.ErrMetricNotFound)
	storage.On("Save", ctx, counterMetric).Return(nil)

	err := service.Count(ctx, name, value)
	assert.NoError(t, err)

	storage.AssertExpectations(t)
}

func TestMetricService_Count_UpdateExisting(t *testing.T) {
	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()
	name := "test_metric"
	value := int64(5)

	existingCounterMetric := metric.NewCounterMetric(name, int64(10))

	storage.On("Get", ctx, metric.Counter, name).Return(existingCounterMetric, nil)
	storage.On("Save", ctx, existingCounterMetric).Return(nil)

	err := service.Count(ctx, name, value)
	assert.NoError(t, err)

	assert.Equal(t, float64(15), existingCounterMetric.Value())

	storage.AssertExpectations(t)
}

func TestMetricService_Count_Error(t *testing.T) {
	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()
	name := "test_metric"
	value := int64(5)

	storage.On("Get", ctx, metric.Counter, name).Return(nil, errors.New("unexpected error"))

	err := service.Count(ctx, name, value)
	assert.Error(t, err)

	storage.AssertExpectations(t)
}

func TestMetricService_Get(t *testing.T) {
	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()

	name := "test_metric_unexpected_error"

	storage.On("Get", ctx, metric.Counter, name).Return(nil, errors.New("unexpected error"))

	_, err := service.Get(ctx, metric.Counter, name)
	assert.Error(t, err)

	storage.AssertExpectations(t)

	name = "test_metric_not_found_error"

	storage.On("Get", ctx, metric.Counter, name).Return(nil, metric.ErrMetricNotFound)

	_, err = service.Get(ctx, metric.Counter, name)
	assert.Error(t, err)
	assert.ErrorIs(t, err, metric.ErrMetricNotFound)

	storage.AssertExpectations(t)

	name = "test_metric"
	value := int64(5)

	storage.On("Get", ctx, metric.Counter, name).Return(metric.NewCounterMetric(name, value), nil)

	actual, err := service.Get(ctx, metric.Counter, name)

	require.NoError(t, err)

	assert.Equal(t, metric.Counter, actual.Type())
	assert.Equal(t, name, actual.Name())
	assert.Equal(t, float64(value), actual.Value())

	storage.AssertExpectations(t)
}

func TestMetricService_All(t *testing.T) {
	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()

	m1 := metric.NewCounterMetric("test_metric_counter", 10)
	m2 := metric.NewCounterMetric("test_metric_gauge", 15)

	var expected []metric.Metric

	expected = append(expected, m1, m2)

	storage.On("All", ctx).Return(m1, m2, nil)

	actual, err := service.All(ctx)

	require.NoError(t, err)

	assert.ElementsMatch(t, expected, actual)

	storage.AssertExpectations(t)
}

func TestMetricService_All_Error(t *testing.T) {

	storage := new(MockMetricStorage)
	service := NewMetricService(storage)

	ctx := context.TODO()

	expected := errors.New("unexpected error")

	storage.On("All", ctx).Return(nil, expected)

	_, err := service.All(ctx)

	require.Error(t, err)

	assert.ErrorIs(t, err, expected)

	storage.AssertExpectations(t)
}
