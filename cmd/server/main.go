package main

import (
	"context"
	"github.com/baisalov/metricollector/internal/checker"
	"github.com/baisalov/metricollector/internal/closer"
	"github.com/baisalov/metricollector/internal/server/config"
	"github.com/baisalov/metricollector/internal/server/handler/http/middleware"
	"github.com/baisalov/metricollector/internal/server/handler/http/v1"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"github.com/baisalov/metricollector/internal/server/storage/postgres"
	"github.com/baisalov/metricollector/internal/transactions"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/sync/errgroup"
	"log"
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, logOpt))
	slog.SetDefault(logger)

	logger.Info("running metric server", "env", conf)

	closings := closer.NewCloser()
	defer func() {
		if err := closings.Close(); err != nil {
			slog.Error("closing error", "error", err)
		}
	}()

	check := checker.NewChecker()

	router := chi.NewMux()

	router.Use(middleware.GzipCompress, middleware.GzipDecompress)

	if conf.HashKey != "" {
		router.Use(middleware.HashCheck(conf.HashKey))
	}

	if conf.DatabaseDsn != "" {
		pool, err := pgxpool.New(context.Background(), conf.DatabaseDsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %v\n", err)
		}

		db := stdlib.OpenDBFromPool(pool)
		closings.Register("closing database connection", db)

		check.Register(checker.Wrap(db.Ping))

		storage, err := postgres.NewMetricStorage(db)
		if err != nil {
			log.Fatalf("failed to init database storage: %v\n", err)
		}

		v1.NewMetricHandler(storage, service.NewMetricUpdateService(storage, postgres.NewTransactionManager(db))).Register(router)
	} else {
		slog.Info("creating file")
		file, err := os.OpenFile(conf.StoragePath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
		if err != nil {
			log.Fatalf("failed to open file: %v\n", err)
		}

		closings.Register("closing file", file)

		slog.Info("init memory storage")
		storage, err := memory.NewMetricStorage(file, conf.StoreInterval, conf.Restore)
		if err != nil {
			log.Fatalf("failed to init storage: %v\n", err)
		}

		closings.Register("closing metric storage", storage)

		v1.NewMetricHandler(storage, service.NewMetricUpdateService(storage, transactions.DiscardManager{})).Register(router)
	}

	v1.NewHealthCheckHandler(check).Register(router)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:         conf.Address,
		Handler:      middleware.RequestLogging(router),
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
		logger.Error("server stopped", "reason", err.Error())
	}
}
