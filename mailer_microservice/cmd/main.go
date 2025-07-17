package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mailer_microservice/internal/application"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := application.NewApp()
	if app == nil {
		log.Fatal("❌ failed to create application")
	}
	defer func() {
		if cerr := app.Close(ctx); cerr != nil {
			log.Printf("⚠️ graceful shutdown failed: %v", cerr)
		}
	}()

	// Запускаємо app.Run у окремій горутині
	errCh := make(chan error, 1)
	go func() {
		if err := app.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("📦 Received shutdown signal...")
	case err := <-errCh:
		log.Fatalf("❌ server run failed: %v", err)
	}
}
