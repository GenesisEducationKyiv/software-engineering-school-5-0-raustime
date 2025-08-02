// weather_microservice/cmd/main.go
package main

import (
	"context"
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
	"weather_microservice/internal/logging"
	"weather_microservice/internal/pkg/ctxkeys"
	"weather_microservice/internal/server"
)

func main() {
	cfg := config.Load()
	logger := logging.NewZapWeatherLogger(cfg.LogPath, cfg.LogLevelDefault)
	ctx := context.WithValue(context.Background(), ctxkeys.Logger, logger)

	weatherService, err := bootstrap.InitWeatherService(ctx, cfg)

	if err != nil {
		logging.Error(ctx, "Failed to initialize weather service", nil, err)
		os.Exit(1)
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
		Handler:      grpcMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	h2Server := &http2.Server{}
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logging.Error(ctx, "failed to listen on gRPC port", nil, err)
		os.Exit(1)
	}

	ctxWithSignal, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx = ctxWithSignal

	// Start HTTP API
	go func() {
		logger.Info(ctx, "main", "Starting HTTP service on: "+cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error(ctx, "HTTP server error", nil, err)
		}
	}()

	// Start gRPC (true HTTP/2 plaintext)
	go func() {
		logger.Info(ctx, "main", "Starting gRPC (HTTP/2 plaintext) service on: "+cfg.GRPCPort)
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warn(ctx, "grpc-accept", nil, err)
				continue
			}
			go h2Server.ServeConn(conn, &http2.ServeConnOpts{
				BaseConfig: grpcSrv,
			})
		}
	}()

	<-ctx.Done()
	logger.Info(ctx, "main", "Shutting down weather service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = httpSrv.Shutdown(shutdownCtx)
	_ = listener.Close() // this will terminate the accept loop for HTTP/2
}
