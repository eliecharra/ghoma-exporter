package main

import (
	"fmt"
	"os"

	"github.com/eliecharra/ghoma/internal/ghoma"
	"github.com/eliecharra/ghoma/internal/intrumentation/config"
)

func main() {
	conf, err := config.Get()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading config: %s", err)
		os.Exit(1)
	}

	server := ghoma.NewServer(ghoma.ServerOptions{ListenAddr: conf.ListenAddress})
	if err := server.Start(); err != nil {
		panic(err)
	}
}
