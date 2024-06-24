package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
)

type MetricService struct {
	storage MetricStorage
}

func NewMetricService(storage MetricStorage) *MetricService {
	return &MetricService{storage: storage}
}

type MetricStorage interface {
	Get(ctx context.Context, t metric.Type, name string) (metric.Metric, error)
	Save(ctx context.Context, m metric.Metric) error
	All(ctx context.Context) ([]metric.Metric, error)
}

func (s *MetricService) Gauge(ctx context.Context, name string, value float64) error {
	m := metric.NewGaugeMetric(name, value)

	err := s.storage.Save(ctx, m)

	if err != nil {
		return fmt.Errorf("can not save metric: %w", err)
	}

	return nil
}

func (s *MetricService) Count(ctx context.Context, name string, value int64) error {

	m, err := s.storage.Get(ctx, metric.Counter, name)

	if err != nil {
		if !errors.Is(err, metric.ErrMetricNotFound) {
			return err
		}
	}

	if m != nil {
		c, ok := m.(*metric.CounterMetric)
		if !ok {
			return metric.ErrIncorrectMetricType
		}

		c.Add(value)
	} else {
		m = metric.NewCounterMetric(name, value)
	}

	err = s.storage.Save(ctx, m)

	if err != nil {
		return fmt.Errorf("can not save metruc: %w", err)
	}

	return nil
}

func (s *MetricService) Get(ctx context.Context, t metric.Type, name string) (metric.Metric, error) {
	return s.storage.Get(ctx, t, name)
}

func (s *MetricService) All(ctx context.Context) ([]metric.Metric, error) {
	return s.storage.All(ctx)
}
