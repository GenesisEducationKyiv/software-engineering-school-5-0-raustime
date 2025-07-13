package server

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	subscriptionv1 "subscription_microservice/gen/go/subscription/v1/subscriptionv1connect"
	"subscription_microservice/internal/config"
	"subscription_microservice/internal/db/migration"
	"subscription_microservice/internal/db/repositories"
	"subscription_microservice/internal/handler"
	"subscription_microservice/internal/mailerclient"
	"subscription_microservice/internal/subscription_service"
)

func build(cfg *config.Config) (*http.ServeMux, *grpc.Server) {
	// Init DB.
	db, err := initDatabase(cfg)
	if err != nil {
		fmt.Printf("database init failed: %v\n", err)
		return nil, nil
	}

	// Run migrations.
	mr := migration.NewRunner(db, "migrations")
	if err := mr.RunMigrations(context.Background()); err != nil {
		fmt.Printf("migrations failed: %v\n", err)
		return nil, nil
	}

	mailer, err := mailerclient.New(cfg.MailerGRPCAddr)
	if err != nil {
		fmt.Printf("mailerclient creation failed: %v\n", err)
		return nil, nil
	}

	subRepo := repositories.NewSubscriptionRepo(db)
	subService := subscription_service.New(subRepo, mailer)

	// gRPC setup
	grpcServer := grpc.NewServer()

	reflection.Register(grpcServer)

	// HTTP Connect handler
	mux := http.NewServeMux()
	path, connectHandler := subscriptionv1.NewSubscriptionServiceHandler(
		handler.NewHandler(&subService),
	)
	mux.Handle(path, connectHandler)

	reflectPath, reflectHandler := grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector("subscription.v1.SubscriptionService"),
	)
	mux.Handle(reflectPath, reflectHandler)

	return mux, grpcServer
}
