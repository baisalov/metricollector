package sender

import (
	"context"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strings"
	"time"
)

type HTTPSender struct {
	address string
	client  *resty.Client
}

func convertMetric(m metric.Metric) metrics {
	res := metrics{
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

type metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func NewHTTPSender(address string) *HTTPSender {
	if !strings.HasPrefix(address, "http://") || !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	client := resty.New()

	client.
		SetRetryCount(3).
		SetRetryWaitTime(10 * time.Second).
		SetRetryMaxWaitTime(30 * time.Second)

	return &HTTPSender{
		address: address,
		client:  client,
	}
}

func (s *HTTPSender) Send(ctx context.Context, m metric.Metric) error {

	addr := fmt.Sprintf("%s/update/", s.address)

	mm := convertMetric(m)

	log.Printf("sending metric: %+v\n", mm)

	res, err := s.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(mm).
		Post(addr)

	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d", res.StatusCode())
	}

	return nil
}
