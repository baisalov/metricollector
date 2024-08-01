package metric

import (
	"errors"
	"strings"
)

var (
	ErrMetricNotFound = errors.New("metric not found")
	ErrEmptyID        = errors.New("empty metric id")
	ErrIncorrectValue = errors.New("incorrect metric value")
)

type Metric struct {
	ID    string   `json:"id"`
	MType Type     `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func NewCounterMetric(name string, delta int64) Metric {
	return Metric{
		MType: Counter,
		ID:    name,
		Delta: &delta,
	}
}

func NewGaugeMetric(name string, value float64) Metric {
	return Metric{
		MType: Gauge,
		ID:    name,
		Value: &value,
	}
}
func (m Metric) Validate() error {

	if !m.MType.IsValid() {
		return ErrIncorrectType
	}

	if strings.TrimSpace(m.ID) == "" {
		return ErrEmptyID
	}

	if m.MType == Gauge && m.Value == nil {
		return ErrIncorrectValue
	}

	if m.MType == Counter && m.Delta == nil {
		return ErrIncorrectValue
	}

	return nil
}
