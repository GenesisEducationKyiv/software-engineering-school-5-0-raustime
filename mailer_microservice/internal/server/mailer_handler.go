package server

import (
	"context"
	"fmt"
	"io"

	"connectrpc.com/connect"

	mailerv1 "mailer_microservice/gen/go/mailer/v1"
	"mailer_microservice/gen/go/mailer/v1/mailerv1connect"
	"mailer_microservice/internal/mailer_service"
)

type MailerServer struct {
	mailerv1connect.UnimplementedMailerServiceHandler
	Service *mailer_service.MailerService
	queue   chan mailJob
}

type mailJob struct {
	req    *mailerv1.EmailRequest
	stream *connect.BidiStream[mailerv1.EmailRequest, mailerv1.EmailStatusResponse]
	ctx    context.Context
}

func NewMailerServer(service *mailer_service.MailerService) *MailerServer {
	s := &MailerServer{
		Service: service,
		queue:   make(chan mailJob, 100),
	}
	go s.startWorker()
	return s
}

func (s *MailerServer) SendEmails(
	ctx context.Context,
	stream *connect.BidiStream[mailerv1.EmailRequest, mailerv1.EmailStatusResponse],
) error {
	for {
		req, err := stream.Receive()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return connect.NewError(connect.CodeUnknown, fmt.Errorf("receive error: %w", err))
		}
		s.queue <- mailJob{req: req, stream: stream, ctx: ctx}
	}
}
