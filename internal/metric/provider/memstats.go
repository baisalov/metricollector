package provider

import (
	"github.com/baisalov/metricollector/internal/metric"
	"runtime"
)

type MemStats struct {
}

func (s MemStats) Load() []metric.Metric {
	var stats = new(runtime.MemStats)

	runtime.ReadMemStats(stats)

	return s.convert(stats)
}

func (s MemStats) convert(stats *runtime.MemStats) []metric.Metric {
	keyValues := map[string]float64{
		"Alloc":         float64(stats.Alloc),
		"BuckHashSys":   float64(stats.BuckHashSys),
		"Frees":         float64(stats.Frees),
		"GCCPUFraction": stats.GCCPUFraction,
		"GCSys":         float64(stats.GCSys),
		"HeapAlloc":     float64(stats.HeapAlloc),
		"HeapIdle":      float64(stats.HeapIdle),
		"HeapInuse":     float64(stats.HeapInuse),
		"HeapObjects":   float64(stats.HeapObjects),
		"HeapReleased":  float64(stats.HeapReleased),
		"HeapSys":       float64(stats.HeapSys),
		"LastGC":        float64(stats.LastGC),
		"Lookups":       float64(stats.Lookups),
		"MCacheInuse":   float64(stats.MCacheInuse),
		"MCacheSys":     float64(stats.MCacheSys),
		"MSpanInuse":    float64(stats.MSpanInuse),
		"MSpanSys":      float64(stats.MSpanSys),
		"Mallocs":       float64(stats.Mallocs),
		"NextGC":        float64(stats.NextGC),
		"NumForcedGC":   float64(stats.NumForcedGC),
		"NumGC":         float64(stats.NumGC),
		"OtherSys":      float64(stats.OtherSys),
		"PauseTotalNs":  float64(stats.PauseTotalNs),
		"StackInuse":    float64(stats.StackInuse),
		"StackSys":      float64(stats.StackSys),
		"Sys":           float64(stats.Sys),
		"TotalAlloc":    float64(stats.TotalAlloc),
	}

	var metrics []metric.Metric

	for k, v := range keyValues {
		metrics = append(metrics, metric.NewGaugeMetric(k, v))
	}

	return metrics
}
