package memory

import (
	"context"
	"github.com/baisalov/metricollector/internal/metric"
	"sync"
)

type MetricStorage struct {
	mx      sync.RWMutex
	metrics map[string]metric.Metric
}

func NewMetricStorage() *MetricStorage {
	return &MetricStorage{
		metrics: make(map[string]metric.Metric),
	}
}

func (s *MetricStorage) key(t metric.Type, name string) string {
	return t.String() + "_" + name
}

func (s *MetricStorage) Get(_ context.Context, t metric.Type, name string) (metric.Metric, error) {
	s.mx.RLock()

	defer s.mx.RUnlock()

	m, ok := s.metrics[s.key(t, name)]
	if !ok {
		return nil, metric.ErrMetricNotFound
	}

	return m, nil
}

func (s *MetricStorage) Save(_ context.Context, m metric.Metric) error {
	s.mx.Lock()

	defer s.mx.Unlock()

	s.metrics[s.key(m.Type(), m.Name())] = m

	return nil
}

func (s *MetricStorage) All(_ context.Context) ([]metric.Metric, error) {
	s.mx.RLock()

	defer s.mx.RUnlock()

	metrics := make([]metric.Metric, 0, len(s.metrics))

	for _, m := range s.metrics {
		metrics = append(metrics, m)
	}

	return metrics, nil
}
