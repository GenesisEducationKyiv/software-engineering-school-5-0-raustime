package application

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"github.com/nats-io/nats.go"

	"mailer_microservice/internal/broker"
	"mailer_microservice/internal/config"
	"mailer_microservice/internal/mailer_service"
	"mailer_microservice/internal/notification"
	"mailer_microservice/internal/server"
)

type App struct {
	httpServer *http.Server
	natsConn   *nats.Conn
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

	// Connect to NATS
	nc, err := nats.Connect(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("❌ failed to connect to NATS: %v", err)
	}

	jsClient, err := broker.NewJetStreamClient(nc)
	if err != nil {
		log.Fatalf("❌ failed to get JetStream context: %v", err)
	}

	notifConsumer := notification.NewNotificationConsumer(mailer)

	// JetStream subscription with manual ack and retry
	err = jsClient.Subscribe("mailer.notifications", "mailer-consumer", func(msg *nats.Msg) {
	go func(m *nats.Msg) {
		err := notifConsumer.HandleMessage(context.Background(), m.Data)
		if err != nil {
			_ = m.Nak()
			return
			}
		_ = m.Ack()
		}(msg)
	})
	
	if err != nil {
		log.Fatalf("❌ failed to subscribe to JetStream: %v", err)
	}

	log.Println("📬 Subscribed to JetStream topic 'mailer.notifications'")

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
	if a.natsConn != nil {
		a.natsConn.Close()
	}
	return a.httpServer.Close()
}
