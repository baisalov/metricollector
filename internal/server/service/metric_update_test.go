package service

import (
	"context"
	"errors"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MetricStorageMock struct {
	mock.Mock
}

func (s *MetricStorageMock) Get(ctx context.Context, t metric.Type, name string) (metric.Metric, error) {
	args := s.Called(ctx, t, name)
	return args.Get(0).(metric.Metric), args.Error(1)
}

func (s *MetricStorageMock) Save(ctx context.Context, m metric.Metric) error {
	args := s.Called(ctx, m)
	return args.Error(0)
}

func TestMetricUpdateService_Update(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MetricStorageMock)
	service := NewMetricUpdateService(mockStorage)

	t.Run("Update existing counter metric", func(t *testing.T) {
		initialDelta := int64(10)
		existingDelta := int64(5)
		expectedDelta := int64(15)

		existingMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "test_count_metric",
			Delta: &existingDelta,
		}

		newMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "test_count_metric",
			Delta: &initialDelta,
		}

		mockStorage.On("Get", ctx, metric.Counter, "test_count_metric").Return(existingMetric, nil)
		mockStorage.On("Save", ctx, mock.MatchedBy(func(m metric.Metric) bool { return m.ID == "test_count_metric" })).Return(nil)

		updatedMetric, err := service.Update(ctx, newMetric)
		assert.NoError(t, err)
		assert.Equal(t, expectedDelta, *updatedMetric.Delta)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Save new counter metric", func(t *testing.T) {
		newDelta := int64(10)

		newMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "new_counter_metric",
			Delta: &newDelta,
		}

		mockStorage.On("Get", ctx, metric.Counter, "new_counter_metric").Return(metric.Metric{}, metric.ErrMetricNotFound)
		mockStorage.On("Save", ctx, mock.MatchedBy(func(m metric.Metric) bool { return m.ID == "new_counter_metric" })).Return(nil)

		updatedMetric, err := service.Update(ctx, newMetric)
		assert.NoError(t, err)
		assert.Equal(t, newDelta, *updatedMetric.Delta)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Save new gauge metric", func(t *testing.T) {
		newValue := float64(10)

		newMetric := metric.Metric{
			MType: metric.Gauge,
			ID:    "new_gauge_metric",
			Value: &newValue,
		}

		mockStorage.On("Get", ctx, metric.Gauge, "new_gauge_metric").Return(metric.Metric{}, metric.ErrMetricNotFound)
		mockStorage.On("Save", ctx, mock.MatchedBy(func(m metric.Metric) bool { return m.ID == "new_gauge_metric" })).Return(nil)

		updatedMetric, err := service.Update(ctx, newMetric)
		assert.NoError(t, err)
		assert.Equal(t, newValue, *updatedMetric.Value)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Update existing gauge metric", func(t *testing.T) {
		initialValue := float64(10)
		existingValue := float64(5)
		expectedValue := float64(10)

		existingMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "test_gauge_metric",
			Value: &existingValue,
		}

		newMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "test_gauge_metric",
			Value: &initialValue,
		}

		mockStorage.On("Get", ctx, metric.Counter, "test_gauge_metric").Return(existingMetric, nil)
		mockStorage.On("Save", ctx, mock.MatchedBy(func(m metric.Metric) bool { return m.ID == "test_gauge_metric" })).Return(nil)

		updatedMetric, err := service.Update(ctx, newMetric)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, *updatedMetric.Value)

		mockStorage.AssertExpectations(t)
	})

	t.Run("Fail to get metric due to unexpected error", func(t *testing.T) {
		newDelta := int64(10)

		newMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "error_metric",
			Delta: &newDelta,
		}

		mockStorage.On("Get", ctx, metric.Counter, "error_metric").Return(metric.Metric{}, errors.New("unexpected error"))

		_, err := service.Update(ctx, newMetric)
		assert.Error(t, err)
		assert.Equal(t, "unexpected error", err.Error())

		mockStorage.AssertExpectations(t)
	})

	t.Run("Fail to save metric", func(t *testing.T) {
		initialDelta := int64(10)
		existingDelta := int64(5)

		existingMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "fail_save_metric",
			Delta: &existingDelta,
		}

		newMetric := metric.Metric{
			MType: metric.Counter,
			ID:    "fail_save_metric",
			Delta: &initialDelta,
		}

		mockStorage.On("Get", ctx, metric.Counter, "fail_save_metric").Return(existingMetric, nil)
		mockStorage.On("Save", ctx, mock.MatchedBy(func(m metric.Metric) bool { return m.ID == "fail_save_metric" })).Return(errors.New("cannot save"))

		_, err := service.Update(ctx, newMetric)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot save")

		mockStorage.AssertExpectations(t)
	})
}
