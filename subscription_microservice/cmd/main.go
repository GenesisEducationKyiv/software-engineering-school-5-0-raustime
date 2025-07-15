package main

import (
	"log"

	"subscription_microservice/internal/config"
	"subscription_microservice/internal/server"
)

func main() {
	cfg := config.Load()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	if err := server.Run(cfg); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
