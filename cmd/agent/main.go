package main

import (
	"context"
	"github.com/baisalov/metricollector/internal/agent"
	"github.com/baisalov/metricollector/internal/agent/config"
	"github.com/baisalov/metricollector/internal/agent/sender"
	"github.com/baisalov/metricollector/internal/metric/provider"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conf := config.MustLoad()

	log.Printf("running metric agent with environments: %+v\n", conf)

	metricAgent := agent.NewMetricAgent(&provider.MemStats{}, sender.NewHTTPSender(conf.ReportAddress))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := metricAgent.Run(ctx,
		time.Duration(conf.PullInterval)*time.Second,
		time.Duration(conf.ReportInterval)*time.Second)

	if err != nil {
		log.Printf("metric agent stop: %s\n", err.Error())
	} else {
		log.Println("metric agent stop")
	}
}
