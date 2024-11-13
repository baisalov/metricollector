package provider

import (
	"github.com/baisalov/metricollector/internal/metric"
	"math/rand/v2"
)

const (
	keyRandomValue = "RandomValue"
	keyPullCount   = "PollCount"
)

type Custom struct {
}

func (c Custom) Source() string {
	return "custom"
}

func (c Custom) Load() ([]metric.Metric, error) {
	return []metric.Metric{
		metric.NewGaugeMetric(keyRandomValue, rand.Float64()),
		metric.NewCounterMetric(keyPullCount, 1),
	}, nil
}
