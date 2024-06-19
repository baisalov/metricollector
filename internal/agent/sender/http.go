package sender

import (
	"context"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"log"
	"net/http"
)

type HTTPSender struct {
	address string
	client  *http.Client
}

func NewHTTPSender(address string) *HTTPSender {
	return &HTTPSender{
		address: address,
		client:  http.DefaultClient,
	}
}

func (s *HTTPSender) Send(ctx context.Context, metric metric.Metric) error {

	url := fmt.Sprintf("%s/update/%s/%s/%v", s.address, metric.Type(), metric.Name(), metric.Value())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("cant create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	client := http.DefaultClient

	res, err := client.Do(req)

	if res != nil {
		defer func() {
			if err := res.Body.Close(); err != nil {
				log.Printf("cant close response body: %v", err.Error())
			}
		}()
	}

	if err != nil {
		return fmt.Errorf("cant do request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpexted response status: %d", res.StatusCode)
	}

	return nil
}
