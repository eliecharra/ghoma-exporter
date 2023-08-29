package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/eliecharra/ghoma/internal/ghoma"
	"github.com/eliecharra/ghoma/internal/intrumentation"
	"github.com/eliecharra/ghoma/internal/intrumentation/config"
	"github.com/eliecharra/ghoma/internal/metrics"
)

func main() {
	conf, err := config.Get()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading config: %s", err)
		os.Exit(1)
	}

	if err := intrumentation.InitLogger(conf); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unable to init logger: %s", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// TODO log interrupt
		cancel()
	}()

	metricCollector := metrics.NewCollector()
	registry := prometheus.NewRegistry()
	if err := registry.Register(metricCollector); err != nil {
		zap.L().Fatal("unable to register metrics collector", zap.Error(err))
	}

	ghomaServer := ghoma.NewServer(
		ghoma.ServerOptions{ListenAddr: conf.GhomaListenAddress},
		metricCollector,
	)
	if err := ghomaServer.Start(ctx); err != nil {
		zap.L().Fatal("Unable to start ghoma server", zap.Error(err))
	}

	servermux := http.NewServeMux()
	servermux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}))
	httpServer := &http.Server{
		Addr:    conf.ListenAddress,
		Handler: servermux,
	}
	zap.L().Info("Prometheus exporter listening", zap.String("address", conf.ListenAddress))
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				zap.L().Fatal("Unable to start exporter", zap.Error(err))
			}
		}
	}()

	<-ctx.Done()
	if err := httpServer.Shutdown(ctx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Server shutdown failed", zap.Error(err))
			os.Exit(1)
		}
	}
}
