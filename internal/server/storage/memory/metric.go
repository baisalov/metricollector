package memory

import (
	"context"
	"github.com/baisalov/metricollector/internal/metric"
	"sync"
)

var ()

type metricStorage struct {
	mx      sync.Mutex
	metrics map[string]metric.Metric
}

func NewMetricStorage() *metricStorage {
	return &metricStorage{
		metrics: make(map[string]metric.Metric),
	}
}

func (s *metricStorage) key(t metric.Type, name string) string {
	return t.String() + "_" + name
}

func (s *metricStorage) Get(_ context.Context, t metric.Type, name string) (metric.Metric, error) {
	s.mx.Lock()

	defer s.mx.Unlock()

	m, ok := s.metrics[s.key(t, name)]
	if !ok {
		return nil, metric.ErrMetricNotFound
	}

	return m, nil
}

func (s *metricStorage) Save(_ context.Context, m metric.Metric) error {
	s.mx.Lock()

	defer s.mx.Unlock()

	s.metrics[s.key(m.Type(), m.Name())] = m

	return nil
}
