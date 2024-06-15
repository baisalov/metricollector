package service

import (
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
	Get(p metric.Type, name string) (metric.Metric, error)
	Save(m metric.Metric) error
}

func (s *MetricService) Gouge(name string, value float64) error {
	m := metric.NewGougeMetric(name, value)

	err := s.storage.Save(m)

	if err != nil {
		return fmt.Errorf("can not save metric: %w", err)
	}

	return nil
}

func (s *MetricService) Count(name string, value int64) error {

	m, err := s.storage.Get(metric.Counter, name)

	if err != nil {
		if !errors.Is(err, metric.ErrMetricNotFound) {
			return nil
		}
	}

	var c = new(metric.CounterMetric)

	if m != nil {
		c, ok := m.(*metric.CounterMetric)
		if !ok {
			return metric.ErrIncorrectMetricType
		}

		c.Add(value)
	} else {
		c = metric.NewCounterMetric(name, value)
	}

	err = s.storage.Save(c)

	if err != nil {
		return fmt.Errorf("can not save metruc: %w", err)
	}

	return nil
}
