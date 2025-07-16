package server

import (
	"context"
	"fmt"
	"io"
	"log"

	"connectrpc.com/connect"

	mailerv1 "mailer_microservice/gen/go/mailer/v1"
	"mailer_microservice/gen/go/mailer/v1/mailerv1connect"
	"mailer_microservice/internal/contracts"
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

func (s *MailerServer) startWorker() {
	for job := range s.queue {
		go func(job mailJob) {
			req := job.req

			var err error

			if req.IsConfirmation {
				log.Printf("[MailerService] 📩 sending confirmation email to %s", req.To)
				err = s.Service.SendConfirmationEmail(job.ctx, req.To, req.Token)
			} else {
				log.Printf("[MailerService] 📩 sending weather email to %s for city %s", req.To, req.City)
				data := contracts.WeatherData{
					Temperature: float64(req.Temperature),
					Humidity:    float64(req.Humidity),
					Description: req.Description,
				}
				err = s.Service.SendWeatherEmail(job.ctx, req.To, req.City, data, req.Token)
			}

			resp := &mailerv1.EmailStatusResponse{
				RequestId: req.RequestId,
				Delivered: err == nil,
			}
			if err != nil {
				resp.Error = err.Error()
			} else {
				log.Printf("[MailerService] ✅ email delivered to %s", req.To)
			}

			select {
			case <-job.ctx.Done():
				log.Printf("⚠️ stream closed, skipping send to %s", req.To)
			default:
				if sendErr := job.stream.Send(resp); sendErr != nil {
					log.Printf("❌ error sending response to client: %v", sendErr)
				}
			}
		}(job)
	}
}
