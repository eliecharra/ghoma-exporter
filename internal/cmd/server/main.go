package main

import (
	"github.com/eliecharra/ghoma/internal/ghoma"
)

func main() {
	server := ghoma.NewServer()
	if err := server.Start(); err != nil {
		panic(err)
	}
}
