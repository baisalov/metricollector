package main

import (
	"context"
	"github.com/baisalov/metricollector/internal/agent"
	"github.com/baisalov/metricollector/internal/agent/sender"
	"github.com/baisalov/metricollector/internal/metric/provider"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	metricAgent := agent.NewMetricAgent(&provider.MemStats{}, sender.NewHTTPSender(reportAddress))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := metricAgent.Run(ctx, pullInterval, reportInterval)

	if err != nil {
		log.Printf("metric agent stop: %v\n", err.Error())
	} else {
		log.Printf("metric agent stop")
	}
}
