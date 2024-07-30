package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type HTTPSender struct {
	address string
	client  *http.Client
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

	return &HTTPSender{
		address: address,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *HTTPSender) Send(ctx context.Context, m metric.Metric) error {

	url := fmt.Sprintf("%s/update/", s.address)

	mm := convertMetric(m)

	log.Printf("sending metric: %+v\n", mm)

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	err := enc.Encode(&mm)
	if err != nil {
		return fmt.Errorf("failed marshal metric: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("cant create request: %w", err)
	}

	req.Close = true

	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("cant do request: %w", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("failed to close response body: %s\n", err.Error())
		}
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d", res.StatusCode)
	}

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("cant read response: %w", err)
	}

	return nil
}
