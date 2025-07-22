package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"subscription_microservice/internal/application"
	"subscription_microservice/internal/config"
)

func main() {
	cfg := config.Load()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	app, err := application.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Run(ctx); err != nil {
			log.Fatalf("Application error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("🔌 Shutting down gracefully...")
	if err := app.Close(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
