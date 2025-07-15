package server

import (
	"mailer_microservice/gen/go/mailer/v1/mailerv1connect"
	"mailer_microservice/internal/mailer_service"
	"net/http"
)

func NewRouter(service *mailer_service.MailerService) *http.ServeMux {
	mux := http.NewServeMux()

	srv := NewMailerServer(service)
	path, handler := mailerv1connect.NewMailerServiceHandler(srv)

	mux.Handle(path, handler)
	return mux
}
