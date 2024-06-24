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
	"time"
)

func main() {
	metricAgent := agent.NewMetricAgent(&provider.MemStats{}, sender.NewHTTPSender(reportAddress))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := metricAgent.Run(ctx, time.Duration(pullInterval)*time.Second, time.Duration(reportInterval)*time.Second)

	if err != nil {
		log.Printf("metric agent stop: %s\n", err.Error())
	} else {
		log.Println("metric agent stop")
	}
}
