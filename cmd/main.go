package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"weatherapi/internal/application"
)

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	defer func() {
		if cerr := app.Close(ctx); cerr != nil {
			log.Printf("Failed to close application: %v", cerr)
		}
	}()

	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
