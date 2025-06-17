package testserver

import (
	"weatherapi/internal/contracts"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Server struct {
	db     bun.IDB
	mailer contracts.MailerServiceProvider
	Router *echo.Echo
}

func NewServer(db bun.IDB, mailer contracts.MailerServiceProvider) *Server {
	s := &Server{
		db:     db,
		mailer: mailer,
		Router: echo.New(),
	}

	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	// наприклад:
	s.Router.POST("/api/subscribe", s.handleSubscribe)
	s.Router.GET("/api/confirm/:token", s.handleConfirm)
	s.Router.GET("/api/unsubscribe/:token", s.handleUnsubscribe)
}
