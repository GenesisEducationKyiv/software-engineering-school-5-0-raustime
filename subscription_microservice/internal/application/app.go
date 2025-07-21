package application

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

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

func (a *App) Run() error {
	// gRPC
	go func() {
		listener, err := net.Listen("tcp", ":"+a.cfg.GrpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("subscription-service gRPC listening on port %s", a.cfg.GrpcPort)
		if err := a.grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// HTTP
	log.Printf("subscription-service HTTP gateway listening on port %s", a.cfg.HttpPort)
	return a.httpServer.ListenAndServe()
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
