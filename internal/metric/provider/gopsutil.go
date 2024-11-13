package provider

import (
	"fmt"
	"github.com/baisalov/metricollector/internal/metric"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type Gopsutil struct {
}

func (g Gopsutil) Source() string {
	return "Gopsutil"
}

func (g Gopsutil) Load() ([]metric.Metric, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to load virtual memory: %w", err)
	}

	c, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to load cpu counts: %w", err)
	}

	return []metric.Metric{
		metric.NewGaugeMetric("TotalMemory", float64(v.Total)),
		metric.NewGaugeMetric("FreeMemory", float64(v.Free)),
		metric.NewGaugeMetric("CPUutilization1", float64(c)),
	}, nil
}
