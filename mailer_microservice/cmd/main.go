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

	if err := app.Run(); err != nil {
		log.Fatalf("❌ server run failed: %v", err)
	}

	// ⏸️ Блокуємо main, поки не прийде сигнал
	<-ctx.Done()

	log.Println("📦 Server is shutting down...")
}
