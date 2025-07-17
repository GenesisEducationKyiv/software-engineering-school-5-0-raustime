// weather_microservice/cmd/main.go
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"

	"weather_microservice/gen/go/weather/v1/weatherv1connect"
	"weather_microservice/internal/bootstrap"
	"weather_microservice/internal/config"
	"weather_microservice/internal/server"
)

func main() {
	cfg := config.Load()
	weatherService, err := bootstrap.InitWeatherService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize weather service: %v", err)
	}

	// HTTP API
	httpRouter := server.NewRouter(cfg, weatherService)
	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      httpRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// gRPC (ConnectRPC) API over HTTP/2 prior knowledge (no TLS)
	path, handler := weatherv1connect.NewWeatherServiceHandler(
		server.NewGRPCWeatherServer(weatherService),
	)
	grpcMux := http.NewServeMux()
	grpcMux.Handle(path, handler)
	grpcSrv := &http.Server{
		Handler: grpcMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	h2Server := &http2.Server{}
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}


	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start HTTP API
	go func() {
		log.Println("Starting HTTP service on:", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC (true HTTP/2 plaintext)
	go func() {
		log.Println("Starting gRPC (HTTP/2 plaintext) service on:", cfg.GRPCPort)
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Accept error: %v", err)
				continue
			}
			go h2Server.ServeConn(conn, &http2.ServeConnOpts{
				BaseConfig: grpcSrv,
			})
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down weather service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = httpSrv.Shutdown(shutdownCtx)
	_ = listener.Close() // this will terminate the accept loop for HTTP/2
}
