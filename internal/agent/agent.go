package agent

import (
	"context"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"golang.org/x/sync/errgroup"
	"log/slog"
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

	slog.Info("metric agent start")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		time.Sleep(pullInterval)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				slog.Info("start loading metrics")

				a.pull()
			}

			time.Sleep(pullInterval)
		}
	})

	g.Go(func() error {
		time.Sleep(reportInterval)

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				slog.Info("start sending metrics")

				err := a.report(ctx)
				if err != nil {
					return err
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

func (a *MetricAgent) pull() {
	metrics := a.provider.Load()

	a.mx.Lock()

	for _, v := range metrics {
		a.state[v.ID] = v
	}

	a.state[keyRandomValue] = metric.NewGaugeMetric(keyRandomValue, rand.Float64())

	if pullCount, ok := a.state[keyPullCount]; ok {
		*pullCount.Delta = *pullCount.Delta + 1
		a.state[keyPullCount] = pullCount
	} else {
		a.state[keyPullCount] = metric.NewCounterMetric(keyPullCount, 1)
	}

	a.mx.Unlock()
}

func (a *MetricAgent) report(ctx context.Context) error {
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

	return nil
}
