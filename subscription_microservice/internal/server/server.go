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

	"subscription_microservice/internal/config"
)

func Run(cfg *config.Config) error {
	httpRouter, grpcServer := build(cfg)

	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.Port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("subscription-service gRPC listening on port %s", cfg.Port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	httpPort := cfg.Port
	log.Printf("subscription-service HTTP gateway listening on port %s", httpPort)
	return http.ListenAndServe(":"+httpPort, httpRouter)
}

func initDatabase(cfg *config.Config) (*bun.DB, error) {
	sqlDB, err := sql.Open("pg", cfg.GetDatabaseURL())
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
