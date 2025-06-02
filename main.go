package main

import (
	"context"
	"log"
	"os"
	"time"

	"weatherapi/internal/api"
	"weatherapi/internal/db"
	"weatherapi/internal/jobs"
	"weatherapi/internal/mailer"

	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	_ = godotenv.Load() // ігноруємо помилку для Docker

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("DB_URL is empty")
	}

	sqldb, err := sql.Open("pg", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	dbconn := bun.NewDB(sqldb, pgdialect.New())

	if err := dbconn.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}

	// 🔁 Виконуємо автоматичні міграції
	if err := db.RunMigrations(context.Background(), dbconn); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	mailerSender := &mailer.SMTPSender{}

	jobs.StartWeatherNotificationLoop(dbconn, mailerSender)

	r := api.SetupRouter(dbconn, mailerSender)
	r.SetTrustedProxies([]string{})
	// Додаємо CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	log.Printf("Starting server on :8080")
	err = r.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Server failed: %s", err)
	}

}
