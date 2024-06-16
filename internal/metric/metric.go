package metric

import "strings"

type Type string

const (
	Counter Type = "counter"
	Gauge   Type = "gauge"
)

func ParseType(s string) Type {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return Type(s)
}

func (t Type) IsValid() bool {
	switch t {
	case Counter, Gauge:
		return true
	default:
		return false
	}
}

func (t Type) String() string {
	return string(t)
}

type Metric interface {
	Name() string
	Value() float64
	Type() Type
}

type metric struct {
	name  string
	value float64
}

func (m metric) Name() string {
	return m.name
}

func (m metric) Value() float64 {
	return m.value
}

func (m metric) Type() Type {
	return ""
}

type CounterMetric struct {
	metric
}

func (c *CounterMetric) Type() Type {
	return Counter
}

func (c *CounterMetric) Add(value int64) {
	c.value += float64(value)
}

func NewCounterMetric(name string, value int64) *CounterMetric {
	return &CounterMetric{metric{
		name:  name,
		value: float64(value),
	}}
}

type GaugeMetric struct {
	metric
}

func (g *GaugeMetric) Type() Type {
	return Gauge
}

func (g *GaugeMetric) Set(value float64) {
	g.value = value
}

func NewGaugeMetric(name string, value float64) *GaugeMetric {
	return &GaugeMetric{metric{
		name:  name,
		value: value,
	}}
}
