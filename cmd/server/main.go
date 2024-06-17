package main

import (
	"context"
	"fmt"
	"github.com/baisalov/metricollector/internal/server/handler/http/v1"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	mux := http.NewServeMux()

	storage := memory.MetricStorage{}

	metricService := service.NewMetricService(storage)

	h := v1.NewMetricHandler(metricService)

	mux.HandleFunc(`POST /update/`, h.Update)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return httpServer.ListenAndServe()
	})

	g.Go(func() error {
		<-ctx.Done()
		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
}
