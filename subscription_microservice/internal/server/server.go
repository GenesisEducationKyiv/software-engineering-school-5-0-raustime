package server

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
  _ "github.com/lib/pq"

	"subscription_microservice/internal/config"
)

func Run(cfg *config.Config) error {
	httpRouter, grpcServer := build(cfg)

	if httpRouter == nil || grpcServer == nil {
	log.Fatal("failed to build server components (router or gRPC server is nil)")
	}


	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GrpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("subscription-service gRPC listening on port %s", cfg.GrpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	httpPort := cfg.HttpPort
	log.Printf("subscription-service HTTP gateway listening on port %s", httpPort)
	return http.ListenAndServe(":"+httpPort, httpRouter)
}

func initDatabase(cfg *config.Config) (*bun.DB, error) {
	sqlDB, err := sql.Open("postgres", cfg.GetDatabaseURL())
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqlDB, pgdialect.New())

	if cfg.IsBunDebugEnabled() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}
