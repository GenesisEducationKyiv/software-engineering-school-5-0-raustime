package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/uptrace/bun"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"subscription_microservice/internal/broker"
	"subscription_microservice/internal/config"
	"subscription_microservice/internal/db"
	"subscription_microservice/internal/db/migration"
	"subscription_microservice/internal/db/repositories"
	"subscription_microservice/internal/handler"
	"subscription_microservice/internal/subscription_service"

	subscriptionv1 "subscription_microservice/gen/go/subscription/v1/subscriptionv1connect"

	"connectrpc.com/grpcreflect"
)

type App struct {
	cfg        *config.Config
	db         *bun.DB
	nats       *broker.NATSClient
	httpServer *http.Server
	grpcServer *grpc.Server
}

func NewApp(cfg *config.Config) (*App, error) {
	// DB
	db, err := db.Init(cfg)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	// Run migrations
	mr := migration.NewRunner(db, "migrations")
	if err := mr.RunMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// NATS
	natsClient, err := broker.NewNATSClient(cfg.NATSUrl)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	// Repositories & Services
	subRepo := repositories.NewSubscriptionRepo(db)
	subService := subscription_service.New(subRepo, natsClient)

	// Handlers
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	httpMux := http.NewServeMux()
	path, connectHandler := subscriptionv1.NewSubscriptionServiceHandler(handler.NewHandler(&subService))
	httpMux.Handle(path, connectHandler)

	reflectPath, reflectHandler := grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector("subscription.v1.SubscriptionService"),
	)
	httpMux.Handle(reflectPath, reflectHandler)

	// HTTP server
	httpServer := &http.Server{
		Addr:    ":" + cfg.HttpPort,
		Handler: httpMux,
	}

	return &App{
		cfg:        cfg,
		db:         db,
		nats:       natsClient,
		httpServer: httpServer,
		grpcServer: grpcServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	// gRPC
	go func() {
		listener, err := net.Listen("tcp", ":"+a.cfg.GrpcPort)
		if err != nil {
			errCh <- fmt.Errorf("failed to listen: %w", err)
			return
		}
		log.Printf("subscription-service gRPC listening on port %s", a.cfg.GrpcPort)
		if err := a.grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// HTTP
	go func() {
		log.Printf("subscription-service HTTP gateway listening on port %s", a.cfg.HttpPort)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("🛑 context canceled, shutting down servers...")
		// продовжуємо shutdown нижче
	case err := <-errCh:
		log.Printf("💥 server error occurred: %v", err)
		// продовжуємо shutdown нижче, після логування
	}

	// Graceful shutdown
	grpcShutdown := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(grpcShutdown)
	}()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctxShutdown); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	<-grpcShutdown
	return nil
}

func (a *App) Close(ctx context.Context) error {
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	if a.nats != nil {
		a.nats.Close()
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("db close error: %v", err)
		}
	}
	return nil
}
