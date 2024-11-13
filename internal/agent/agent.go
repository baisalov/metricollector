package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"maps"
	"sync"
	"sync/atomic"
	"time"
)

type MetricAgent struct {
	mx    sync.RWMutex
	run   *atomic.Bool
	state map[string]metric.Metric

	providers   []metricProvider
	sender      metricSender
	senderCount int
}

type metricSender interface {
	Send(ctx context.Context, metrics ...metric.Metric) error
}

type metricProvider interface {
	Source() string
	Load() ([]metric.Metric, error)
}

func NewMetricAgent(sender metricSender, senderCount int, providers ...metricProvider) *MetricAgent {
	return &MetricAgent{
		mx:          sync.RWMutex{},
		run:         &atomic.Bool{},
		state:       make(map[string]metric.Metric),
		providers:   providers,
		sender:      sender,
		senderCount: senderCount,
	}
}

func (a *MetricAgent) Run(ctx context.Context, pullInterval, reportInterval time.Duration) error {

	if !a.run.CompareAndSwap(false, true) {
		return errors.New("metric agent already started")
	}

	defer a.run.Store(false)

	g, ctx := errgroup.WithContext(ctx)

	pullTicker := time.NewTicker(pullInterval)
	defer pullTicker.Stop()

	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	ch := make(chan metric.Metric)
	defer close(ch)

	for i := 0; i < a.senderCount; i++ {
		g.Go(a.reporter(ctx, a.sender, ch))
	}

	for _, provider := range a.providers {
		g.Go(a.puller(ctx, pullTicker, provider))
	}

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				slog.Debug("report stop")
				return ctx.Err()
			case <-reportTicker.C:
				slog.Info("start sending metrics")

				a.report(ctx, ch)
			}
		}
	})

	slog.Info("metric agent start")

	if err := g.Wait(); err != nil {
		return fmt.Errorf("metric agent stop reason: %w", err)
	}

	return nil
}

func (a *MetricAgent) puller(ctx context.Context, ticker *time.Ticker, provider metricProvider) func() error {

	return func() error {
		for {
			select {
			case <-ctx.Done():
				slog.Debug("puller stop", "puller", provider.Source())
				return ctx.Err()
			case <-ticker.C:
				slog.Info("start loading metrics", "provider", provider.Source())

				metrics, err := provider.Load()
				if err != nil {
					return fmt.Errorf("%s failed to load metrics: %w", provider.Source(), err)
				}

				a.store(metrics...)
			}
		}
	}
}

func (a *MetricAgent) report(ctx context.Context, ch chan metric.Metric) {
	a.mx.RLock()

	localStat := make(map[string]metric.Metric, len(a.state))
	maps.Copy(localStat, a.state)

	a.mx.RUnlock()

	for _, m := range localStat {
		select {
		case <-ctx.Done():
			return
		case ch <- m:
		}
	}
}

func (a *MetricAgent) reporter(ctx context.Context, sender metricSender, ch chan metric.Metric) func() error {
	return func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m := <-ch:
				err := sender.Send(ctx, m)
				if err != nil {
					slog.Error("failed to report metric", "error", err)
					// return fmt.Errorf("failed to report metric: %w", err)
				}
			}

		}
	}
}

func (a *MetricAgent) store(metrics ...metric.Metric) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, v := range metrics {
		a.state[v.ID] = v
	}
}
