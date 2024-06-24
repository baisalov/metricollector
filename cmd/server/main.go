package main

import (
	"context"
	"github.com/baisalov/metricollector/internal/server/config"
	"github.com/baisalov/metricollector/internal/server/handler/http/v1"
	"github.com/baisalov/metricollector/internal/server/service"
	"github.com/baisalov/metricollector/internal/server/storage/memory"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	conf := config.MustLoad()

	log.Printf("running metric server with environments: %+v\n", conf)

	storage := memory.NewMetricStorage()

	metricService := service.NewMetricService(storage)

	h := v1.NewMetricHandler(metricService)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:    conf.Address,
		Handler: h.Handler(),
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
		log.Printf("exit reason: %s \n", err.Error())
	}
}
