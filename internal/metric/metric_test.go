package metric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCounterMetric(t *testing.T) {
	name := "test_counter"
	value := int64(10)
	metric := NewCounterMetric(name, value)

	assert.Equal(t, name, metric.Name())
	assert.Equal(t, float64(value), metric.Value())
	assert.Equal(t, Counter, metric.Type())
}

func TestCounterMetric_Add(t *testing.T) {
	name := "test_counter"
	value := int64(10)
	m := NewCounterMetric(name, value)

	m.Add(5)
	assert.Equal(t, float64(15), m.Value())
}

func TestNewGaugeMetric(t *testing.T) {
	name := "test_gauge"
	value := 10.5
	m := NewGaugeMetric(name, value)

	assert.Equal(t, name, m.Name())
	assert.Equal(t, value, m.Value())
	assert.Equal(t, Gauge, m.Type())
}

func TestGaugeMetric_Set(t *testing.T) {
	name := "test_gauge"
	value := 10.5
	m := NewGaugeMetric(name, value)

	m.Set(20.5)
	assert.Equal(t, 20.5, m.Value())
}
