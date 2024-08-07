package main

import (
	"context"
	"github.com/baisalov/metricollector/internal/server/config"
	"github.com/baisalov/metricollector/internal/server/handler/http/middleware"
	"github.com/baisalov/metricollector/internal/server/handler/http/v1"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net"
	"net/http"
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

	log.Info("running metric server", "env", conf)

	slog.Info("creating file")
	file, err := os.OpenFile(conf.StoragePath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		log.Error("failed to open file", "error", err)
		return
	}

	defer func(file *os.File) {
		slog.Debug("closing file")
		err = file.Close()
		if err != nil {
			slog.Error("failed to close file", "error", err)
		}
	}(file)

	slog.Info("init storage")
	storage, err := memory.NewMetricStorage(file, conf.StoreInterval, conf.Restore)
	if err != nil {
		log.Error("failed to init storage", "error", err)
		return
	}

	defer func(storage *memory.MetricStorage) {
		slog.Debug("closing storage")
		err = storage.Close()
		if err != nil {
			slog.Error("failed to close storage", "error", err)
		}
	}(storage)

	metricUpdater := service.NewMetricUpdateService(storage)

	h := v1.NewMetricHandler(storage, metricUpdater)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:         conf.Address,
		Handler:      middleware.RequestLogging(h.Handler()),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,

		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, ctx := errgroup.WithContext(ctx)

	slog.Info("running server")

	g.Go(func() error {
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		<-ctx.Done()
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return httpServer.Shutdown(timeout)
	})

	if err := g.Wait(); err != nil {
		log.Error("server stopped", "reason", err.Error())
	}
}
