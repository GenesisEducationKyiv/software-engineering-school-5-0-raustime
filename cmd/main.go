package main

import (
	"log"
	"weatherapi/internal"
)

func main() {
	app, err := internal.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Printf("Failed to close application: %v", err)
		}
	}()

	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
