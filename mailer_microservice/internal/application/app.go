package application

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"mailer_microservice/internal/broker"
	"mailer_microservice/internal/config"
	"mailer_microservice/internal/mailer_service"
	"mailer_microservice/internal/notification"
	"mailer_microservice/internal/server"
)

type App struct {
	httpServer *http.Server
	natsClient *broker.NATSClient
}

func NewApp() *App {
	cfg := config.Load()
	if cfg == nil {
		log.Fatal("❌ failed to load configuration")
	}

	// SMTP sender
	sender := mailer_service.NewSMTPSender(
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPHost,
		fmt.Sprint(cfg.SMTPPort),
	)

	// Mailer service
	mailer := mailer_service.NewMailerService(sender, cfg.AppBaseURL)
	mailer.SetTemplateDir(cfg.TemplateDir)

	// Start NATS consumer
	natsClient, err := broker.NewNATSClient(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("❌ failed to connect to NATS: %v", err)
	}

	notifConsumer := notification.NewNotificationConsumer(mailer)

	_, err = natsClient.Subscribe("mailer.notifications", func(msg *broker.Message) {
		go notifConsumer.HandleMessage(context.Background(), msg.Data)
	})
	if err != nil {
		log.Fatalf("❌ failed to subscribe to mailer.notifications: %v", err)
	}
	log.Println("📬 Subscribed to 'mailer.notifications'")

	// HTTP + h2c server
	router := server.NewRouter(mailer)
	h2cHandler := h2c.NewHandler(router, &http2.Server{})

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: h2cHandler,
	}

	return &App{
		httpServer: httpServer,
		natsClient: natsClient,
	}
}

func (a *App) Run() error {
	log.Printf("🚀 MailerService (ConnectRPC + h2c) listening on %s", a.httpServer.Addr)
	return a.httpServer.ListenAndServe()
}

func (a *App) Close(_ context.Context) error {
	return a.httpServer.Close()
}
