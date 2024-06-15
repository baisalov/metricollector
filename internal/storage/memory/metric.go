package memory

import (
	"errors"
	"github.com/baisalov/metricollector/internal/metric"
	"sync"
)

var (
	mx      sync.Mutex
	storage = make(map[string]any)
)

type MetricStorage struct {
}

func (s MetricStorage) key(t metric.Type, name string) string {
	return t.String() + "_" + name
}

func (s MetricStorage) Get(t metric.Type, name string) (metric.Metric, error) {
	mx.Lock()

	defer mx.Unlock()

	m, ok := storage[s.key(t, name)]
	if !ok {
		return nil, metric.ErrMetricNotFound
	}

	switch t {
	case metric.Gouge:
		g, ok := m.(*metric.GougeMetric)
		if !ok {
			return nil, errors.New("incorrect type cast")
		}

		return g, nil
	case metric.Counter:
		c, ok := m.(*metric.GougeMetric)
		if !ok {
			return nil, errors.New("incorrect type cast")
		}

		return c, nil
	default:
		return nil, errors.New("incorrect type cast")
	}
}

func (s MetricStorage) Save(m metric.Metric) error {
	mx.Lock()

	defer mx.Unlock()

	storage[s.key(m.Type(), m.Name())] = m.Value()

	return nil
}
