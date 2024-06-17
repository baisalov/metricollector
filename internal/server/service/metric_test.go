package service

import (
	"context"
	"errors"
	"testing"

	"github.com/baisalov/metricollector/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricStorage struct {
	mock.Mock
}

func (m *MockMetricStorage) Get(ctx context.Context, t metric.Type, name string) (metric.Metric, error) {
	args := m.Called(ctx, t, name)

	var metr metric.Metric

	arg := args.Get(0)
	if arg != nil {
		metr = arg.(metric.Metric)
	}

	return metr, args.Error(1)
}

func (m *MockMetricStorage) Save(ctx context.Context, metric metric.Metric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
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
