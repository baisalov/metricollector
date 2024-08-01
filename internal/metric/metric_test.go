package metric

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func TestNewCounterMetric(t *testing.T) {
	id := "test_counter"
	delta := int64(10)
	metric := NewCounterMetric(id, delta)

	assert.Equal(t, Counter, metric.MType)
	assert.Equal(t, id, metric.ID)
	assert.Equal(t, delta, *metric.Delta)
	assert.Nil(t, metric.Value)
	assert.NoError(t, metric.Validate())
}

func TestNewGaugeMetric(t *testing.T) {
	id := "test_gauge"
	value := 10.5
	metric := NewGaugeMetric(id, value)

	assert.Equal(t, Gauge, metric.MType)
	assert.Equal(t, id, metric.ID)
	assert.Equal(t, value, *metric.Value)
	assert.Nil(t, metric.Delta)
	assert.NoError(t, metric.Validate())
}

func TestMetric_Validate(t *testing.T) {
	metric := Metric{
		MType: Type("incorrect"),
	}

	t.Run("incorrect_type", func(t *testing.T) {
		err := metric.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, ErrIncorrectType, err)
	})

	metric.MType = Counter
	metric.ID = " "

	t.Run("empty_id", func(t *testing.T) {
		err := metric.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, ErrEmptyID, err)
	})

	metric.ID = "test"

	t.Run("incorrect_counter_value", func(t *testing.T) {
		err := metric.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, ErrIncorrectValue, err)
	})

	metric.MType = Gauge

	t.Run("incorrect_gauge_value", func(t *testing.T) {
		err := metric.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, ErrIncorrectValue, err)
	})
}

func TestMetric_ValueToString(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		m := NewCounterMetric("test", 10)

		assert.Equal(t, strconv.FormatInt(*m.Delta, 10), m.ValueToString())
	})

	t.Run("gauge", func(t *testing.T) {
		m := NewGaugeMetric("test", 10.0001000)

		assert.Equal(t, strings.TrimRight(fmt.Sprintf("%.3f", *m.Value), "0."), m.ValueToString())
	})
}
