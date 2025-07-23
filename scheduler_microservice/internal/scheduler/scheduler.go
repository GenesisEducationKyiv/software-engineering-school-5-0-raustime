package scheduler

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"scheduler_microservice/internal/contracts"
)

type SubscriptionService interface {
	GetConfirmed(ctx context.Context, frequency string) ([]*contracts.Subscription, error)
}

type MailPublisher interface {
	Publish(subject string, data []byte) error
}

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (*contracts.WeatherData, error)
}

type Scheduler struct {
	subSvc     SubscriptionService
	mailPub    MailPublisher
	weatherSvc WeatherService
	stopChan   chan struct{}
}

func NewScheduler(subSvc SubscriptionService, mailPub MailPublisher, weatherSvc WeatherService) *Scheduler {
	return &Scheduler{
		subSvc:     subSvc,
		mailPub:    mailPub,
		weatherSvc: weatherSvc,
		stopChan:   make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go s.run()
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Minute() == 0 {
				s.Send("hourly")
				if now.Hour() == 8 {
					s.Send("daily")
				}
			}
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) Send(freq string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subs, err := s.subSvc.GetConfirmed(ctx, freq)
	if err != nil {
		log.Printf("Error getting subscriptions: %v", err)
		return
	}

	const maxWorkers = 10
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, sub := range subs {
		sem <- struct{}{}
		wg.Add(1)

		go func(sub *contracts.Subscription) {
			defer func() {
				<-sem
				wg.Done()
			}()
			s.processSubscription(ctx, sub, freq)
		}(sub)
	}

	wg.Wait()
}

func (s *Scheduler) processSubscription(ctx context.Context, sub *contracts.Subscription, freq string) {
	weather, err := s.weatherSvc.GetWeather(ctx, sub.City)
	if err != nil {
		log.Printf("Weather error for %s: %v", sub.City, err)
		return
	}

	msg := contracts.NotificationMessage{
		Type:    "weather",
		To:      sub.Email,
		City:    sub.City,
		Token:   sub.Token,
		Weather: weather,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal notification for %s: %v", sub.Email, err)
		return
	}

	if err := s.mailPub.Publish("mailer.notifications", payload); err != nil {
		log.Printf("Failed to publish notification to %s: %v", sub.Email, err)
		return
	}

	log.Printf("📤 Published %s weather update for %s", freq, sub.Email)
}
