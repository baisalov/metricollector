package metric

type Type string

const (
	Counter Type = "counter"
	Gouge   Type = "gouge"
)

func (t Type) IsValid() bool {
	switch t {
	case Counter, Gouge:
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

type GougeMetric struct {
	metric
}

func (g *GougeMetric) Type() Type {
	return Gouge
}

func (g *GougeMetric) Set(value float64) {
	g.value = value
}

func NewGougeMetric(name string, value float64) *GougeMetric {
	return &GougeMetric{metric{
		name:  name,
		value: value,
	}}
}
