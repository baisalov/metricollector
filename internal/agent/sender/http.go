package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type HTTPSender struct {
	address string
	hashKey string
	client  *resty.Client
}

func NewHTTPSender(address, hashKey string) *HTTPSender {
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
		hashKey: hashKey,
	}
}

func (s *HTTPSender) Send(ctx context.Context, metrics []metric.Metric) error {

	addr := fmt.Sprintf("%s/updates/", s.address)

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	err := enc.Encode(metrics)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	var hashSum []byte

	if s.hashKey != "" {
		h := hmac.New(sha256.New, []byte(s.hashKey))
		h.Write(buf.Bytes())
		hashSum = h.Sum(nil)
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

	slog.Debug("sending metric", "metric", metrics)

	res, err := s.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("HashSHA256", fmt.Sprintf("%x", hashSum)).
		SetBody(zip.Bytes()).
		Post(addr)

	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected response status: %d", res.StatusCode())
	}

	slog.Debug("metrics success send")

	return nil
}
