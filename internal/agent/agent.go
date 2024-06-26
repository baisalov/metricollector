package agent

import (
	"context"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"golang.org/x/sync/errgroup"
	"log"
	"maps"
	"math/rand"
	"sync"
	"time"
)

const (
	keyRandomValue = "RandomValue"
	keyPullCount   = "PollCount"
)

type MetricAgent struct {
	mx       sync.RWMutex
	state    map[string]metric.Metric
	provider metricProvider
	sender   metricSender
}

type metricSender interface {
	Send(ctx context.Context, metric metric.Metric) error
}

type metricProvider interface {
	Load() []metric.Metric
}

func NewMetricAgent(provider metricProvider, sender metricSender) *MetricAgent {
	return &MetricAgent{
		mx:       sync.RWMutex{},
		state:    make(map[string]metric.Metric),
		provider: provider,
		sender:   sender,
	}
}

func (a *MetricAgent) Run(ctx context.Context, pullInterval, reportInterval time.Duration) error {

	log.Println("metric agent start")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				log.Println("start loading metrics")

				metrics := a.provider.Load()

				a.mx.Lock()

				for _, v := range metrics {
					a.state[v.Name()] = v
				}

				a.state[keyRandomValue] = metric.NewGaugeMetric(keyRandomValue, rand.Float64())

				if pullCount, ok := a.state[keyPullCount].(*metric.CounterMetric); ok {
					pullCount.Add(1)
					a.state[keyPullCount] = pullCount
				} else {
					a.state[keyPullCount] = metric.NewCounterMetric(keyPullCount, 1)
				}

				a.mx.Unlock()

			}

			time.Sleep(pullInterval)
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				log.Println("start sending metrics")

				a.mx.RLock()

				localStat := make(map[string]metric.Metric, len(a.state))
				maps.Copy(localStat, a.state)

				a.mx.RUnlock()

				for _, m := range localStat {
					err := a.sender.Send(ctx, m)
					if err != nil {
						return fmt.Errorf("cant send metric: %w", err)
					}
				}
			}

			time.Sleep(reportInterval)
		}
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("metric agent stop reason: %w", err)
	}

	return nil
}
