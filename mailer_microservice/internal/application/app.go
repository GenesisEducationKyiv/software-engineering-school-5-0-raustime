package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"mailer_microservice/internal/config"
	"mailer_microservice/internal/mailer_service"
	"mailer_microservice/internal/server"
)

type App struct {
	server *http.Server
}

func NewApp() *App {
	cfg := config.Load()
	if cfg == nil {
		panic("failed to load configuration")
	}

	emailSender := mailer_service.NewSMTPSender(
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPHost,
		fmt.Sprint(cfg.SMTPPort),
	)

	mailer := mailer_service.NewMailerService(emailSender, cfg.AppBaseURL)
	mailer.SetTemplateDir(cfg.TemplateDir)
	router := server.NewRouter(mailer)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{server: httpServer}
}

func (a *App) Run() error {
	log.Printf("🚀 Starting mailer service on %s", a.server.Addr)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server error: %w", err))
		}
	}()
	return nil
}

func (a *App) Close(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
