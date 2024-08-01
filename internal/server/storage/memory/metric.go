package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"io"
	"log/slog"
	"maps"
	"sync"
	"time"
)

type MetricStorage struct {
	mx          sync.RWMutex
	metrics     map[string]metric.Metric
	archiver    io.ReadWriteSeeker
	syncArchive bool
	stopArchive chan int
}

func NewMetricStorage(archiver io.ReadWriteSeeker, archiveInterval int64, restore bool) (*MetricStorage, error) {
	storage := &MetricStorage{
		metrics:  make(map[string]metric.Metric),
		archiver: archiver,
	}

	if restore {
		err := storage.restore()
		if err != nil {
			return nil, fmt.Errorf("failed to restore storage: %w", err)
		}
	}

	if archiveInterval < 1 {
		storage.syncArchive = true
	} else {

		storage.stopArchive = make(chan int)

		go func() {
			for {
				time.Sleep(time.Duration(archiveInterval) * time.Second)
				select {
				case <-storage.stopArchive:
					close(storage.stopArchive)
					return
				default:
					if err := storage.archive(); err != nil {
						slog.Error("failed to archive metrics by timer", "error", err)
					}
				}
			}
		}()
	}

	return storage, nil
}

func (s *MetricStorage) key(t metric.Type, id string) string {
	return t.String() + "_" + id
}

func (s *MetricStorage) Get(_ context.Context, t metric.Type, id string) (metric.Metric, error) {
	s.mx.RLock()

	defer s.mx.RUnlock()

	m, ok := s.metrics[s.key(t, id)]
	if !ok {
		return metric.Metric{}, metric.ErrMetricNotFound
	}

	return m, nil
}

func (s *MetricStorage) Save(_ context.Context, m metric.Metric) error {
	s.mx.Lock()

	s.metrics[s.key(m.MType, m.ID)] = m

	s.mx.Unlock()

	if s.syncArchive {
		return s.archive()
	}

	return nil
}

func (s *MetricStorage) All(_ context.Context) ([]metric.Metric, error) {
	s.mx.RLock()

	defer s.mx.RUnlock()

	metrics := make([]metric.Metric, 0, len(s.metrics))

	for _, m := range s.metrics {
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (s *MetricStorage) restore() error {
	var metrics map[string]metric.Metric

	bytes, err := io.ReadAll(s.archiver)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &metrics)
		if err != nil {
			return fmt.Errorf("failed to deserialize storage: %w", err)
		}

		s.metrics = metrics
	}

	return nil
}

func (s *MetricStorage) archive() error {

	s.mx.RLock()

	metrics := make(map[string]metric.Metric, len(s.metrics))
	maps.Copy(metrics, s.metrics)

	s.mx.RUnlock()

	data, err := json.Marshal(&metrics)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %w", err)
	}

	_, err = s.archiver.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to reset file: %w", err)
	}

	_, err = s.archiver.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil

}

func (s *MetricStorage) Close() error {
	if !s.syncArchive {
		s.stopArchive <- 1
	}

	return s.archive()
}
