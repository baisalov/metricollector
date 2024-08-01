package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	err := enc.Encode(m)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	var zip bytes.Buffer

	zw := gzip.NewWriter(&zip)

	_, err = zw.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	err = zw.Close()
	if err != nil {
		return fmt.Errorf("failed compress data: %w", err)
	}

	log.Printf("sending metric: %+v\n", m)

	res, err := s.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(zip.Bytes()).
		Post(addr)

	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d", res.StatusCode())
	}

	return nil
}
