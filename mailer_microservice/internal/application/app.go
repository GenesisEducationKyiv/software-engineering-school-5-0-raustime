package application

import (
	"log"
	"net/http"
	"fmt"
	"context"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"mailer_microservice/internal/config"
	"mailer_microservice/internal/mailer_service"
	"mailer_microservice/internal/server"
)

type App struct {
	httpServer *http.Server
}

func NewApp() *App {
	cfg := config.Load()
	if cfg == nil {
		log.Fatal("❌ failed to load configuration")
	}

	sender := mailer_service.NewSMTPSender(
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPHost,
		fmt.Sprint(cfg.SMTPPort),
	)

	mailer := mailer_service.NewMailerService(sender, cfg.AppBaseURL)
	mailer.SetTemplateDir(cfg.TemplateDir)

	router := server.NewRouter(mailer)
	h2cHandler := h2c.NewHandler(router, &http2.Server{})

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: h2cHandler,
	}

	return &App{httpServer: httpServer}
}

func (a *App) Run() error {
	log.Printf("🚀 MailerService (ConnectRPC + h2c) listening on %s", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

func (a *App) Close(_ context.Context) error {
	return a.httpServer.Close()
}
