package main

import (
	"context"
	"github.com/baisalov/metricollector/internal/agent"
	"github.com/baisalov/metricollector/internal/agent/config"
	"github.com/baisalov/metricollector/internal/agent/sender"
	"github.com/baisalov/metricollector/internal/metric/provider"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conf := config.MustLoad()

	logOpt := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, logOpt))

	slog.SetDefault(log)

	log.Info("running metric agent", "env", conf)

	metricAgent := agent.NewMetricAgent(&provider.MemStats{}, sender.NewHTTPSender(conf.ReportAddress))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := metricAgent.Run(ctx,
		time.Duration(conf.PullInterval)*time.Second,
		time.Duration(conf.ReportInterval)*time.Second)

	if err != nil {
		log.Error("metric agent stop", "error", err)
	} else {
		log.Info("metric agent stop")
	}
}
