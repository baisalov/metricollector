package v1

import (
	"errors"
	"github.com/baisalov/metricollector/internal/metric"
	"strings"
)

func ConvertMetric(m metric.Metric) Metrics {
	res := Metrics{
		ID:    m.Name(),
		MType: m.Type().String(),
	}

	if m.Type() == metric.Counter {
		v := int64(m.Value())
		res.Delta = &v
		return res
	}

	v := m.Value()
	res.Value = &v

	return res
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m Metrics) Type() metric.Type {
	return metric.ParseType(m.MType)
}

func (m Metrics) ValidateForUpdate() error {
	t := m.Type()

	if !t.IsValid() {
		return metric.ErrIncorrectMetricType
	}

	if strings.TrimSpace(m.ID) == "" {
		return errors.New("empty metric name")
	}

	if t == metric.Gauge && m.Value == nil {
		return errors.New("incorrect gauge metric value")
	}

	if t == metric.Counter && m.Delta == nil {
		return errors.New("incorrect counter metric value")
	}

	return nil
}

func (m Metrics) ValidateForValue() error {
	t := m.Type()

	if !t.IsValid() {
		return metric.ErrIncorrectMetricType
	}

	if strings.TrimSpace(m.ID) == "" {
		return errors.New("empty metric name")
	}

	return nil
}
