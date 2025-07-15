package main

import (
	"log"
	"os"
	"os/signal"
	"scheduler_microservice/internal/application"
	"scheduler_microservice/internal/health"
	"syscall"
)

func main() {
	app := application.NewApp()
	// Start /health endpoint on 8092
	health.StartHealthServer(app.GetConfig().Port)
	
	app.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down scheduler...")
	app.Shutdown()
}
