package main

import (
	//"net/http"

	//"time"

	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/app"
)

var (
	GitCommit = "0.0.1"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := app.App(ctx, GitCommit); err != nil {
		panic(err)
	}
}
