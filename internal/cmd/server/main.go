package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eliecharra/ghoma/internal/ghoma"
	"github.com/eliecharra/ghoma/internal/intrumentation"
	"github.com/eliecharra/ghoma/internal/intrumentation/config"
	"go.uber.org/zap"
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

	server := ghoma.NewServer(
		ghoma.ServerOptions{ListenAddr: conf.ListenAddress},
	)
	if err := server.Start(ctx); err != nil {
		zap.L().Fatal("Unable to start server", zap.Error(err))
		os.Exit(1)
	}
}
