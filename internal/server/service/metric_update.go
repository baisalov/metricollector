package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
)

type MetricUpdateService struct {
	tm      transactionManager
	storage MetricStorage
}

func NewMetricUpdateService(storage MetricStorage, tm transactionManager) *MetricUpdateService {
	return &MetricUpdateService{
		storage: storage,
		tm:      tm,
	}
}

type transactionManager interface {
	Do(context.Context, func(context.Context) error) error
}

type MetricStorage interface {
	Get(ctx context.Context, t metric.Type, id string) (metric.Metric, error)
	Save(ctx context.Context, m metric.Metric) error
}

func (s *MetricUpdateService) Updates(ctx context.Context, metrics ...metric.Metric) error {
	return s.tm.Do(ctx, func(ctx context.Context) error {
		for _, m := range metrics {
			if _, err := s.Update(ctx, m); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *MetricUpdateService) Update(ctx context.Context, m metric.Metric) (metric.Metric, error) {
	mm, err := s.storage.Get(ctx, m.MType, m.ID)

	if err != nil {
		if !errors.Is(err, metric.ErrMetricNotFound) {
			return m, err
		}
	}

	if m.MType == metric.Counter && mm.Delta != nil {
		*m.Delta += *mm.Delta
	}

	err = s.storage.Save(ctx, m)

	if err != nil {
		return m, fmt.Errorf("can not save metric: %w", err)
	}

	return m, nil
}
