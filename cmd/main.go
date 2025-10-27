package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/app"
)

var (
	GitCommit = "0.0.0-dev"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	if err := app.App(ctx, GitCommit); err != nil {
		cancel()
		log.Fatal(err)
	}

	cancel()
}
