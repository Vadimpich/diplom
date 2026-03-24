package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"diplom/internal/app"
	"diplom/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer application.Close()

	log.Printf("starting api server on %s", cfg.HTTPAddr)

	if err := application.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("run api server: %v", err)
	}
}
