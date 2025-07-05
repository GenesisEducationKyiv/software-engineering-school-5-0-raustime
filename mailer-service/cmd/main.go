package main

import (
	"log"
	"mailer-service/internal/config"
	"mailer-service/internal/nats"
)

func main() {
	cfg := config.Load()

	natsListener, err := nats.NewListener(cfg)
	if err != nil {
		log.Fatalf("failed to start NATS listener: %v", err)
	}

	log.Println("Mailer service is running and listening for messages...")
	if err := natsListener.Listen(); err != nil {
		log.Fatalf("NATS listener failed: %v", err)
	}
}
