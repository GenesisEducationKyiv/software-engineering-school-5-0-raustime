package server

import (
	"log"
	mailerv1 "mailer_microservice/gen/go/mailer/v1"
	"mailer_microservice/internal/contracts"
)

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
