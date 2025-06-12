package jobs

import (
	"context"
	"log"
	"time"

	"weatherapi/internal/services"
)

// Scheduler manages background jobs
type Scheduler struct {
	subscriptionService services.SubscriptionService
	mailerService       services.MailerService
	weatherService      services.WeatherService
	stopChan           chan struct{}
}

// NewScheduler creates a new job scheduler
func NewScheduler(
	subscriptionService services.SubscriptionService,
	mailerService services.MailerService,
	weatherService services.WeatherService,
) *Scheduler {
	return &Scheduler{
		subscriptionService: subscriptionService,
		mailerService:       mailerService,
		weatherService:      weatherService,
		stopChan:           make(chan struct{}),
	}
}

// Start starts the job scheduler
func (s *Scheduler) Start() {
	go s.weatherNotificationLoop()
}

// Stop stops the job scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// weatherNotificationLoop runs the weather notification loop
func (s *Scheduler) weatherNotificationLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			
			// Send hourly updates (00 minutes)
			if now.Minute() == 0 {
				s.sendWeatherUpdates("hourly")
				
				// Send daily updates at 8:00 AM
				if now.Hour() == 8 {
					s.sendWeatherUpdates("daily")
				}
			}
		case <-s.stopChan:
			log.Println("Stopping weather notification loop")
			return
		}
	}
}

// sendWeatherUpdates sends weather updates for specified frequency
func (s *Scheduler) sendWeatherUpdates(frequency string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subscriptions, err := s.subscriptionService.GetConfirmedSubscriptions(ctx, frequency)
	if err != nil {
		log.Printf("Failed to fetch %s subscriptions: %v", frequency, err)
		return
	}

	for _, subscription := range subscriptions {
		weather, err := s.weatherService.GetWeather(ctx, subscription.City)
		if err != nil {
			log.Printf("Weather fetch error for %s: %v", subscription.City, err)
			continue
		}

		err = s.mailerService.SendWeatherEmail(ctx, subscription.Email, subscription.City, weather, subscription.Token)
		if err != nil {
			log.Printf("Failed to send weather email to %s: %v", subscription.Email, err)
		} else {
			log.Printf("Sent %s weather email to %s (%s)", frequency, subscription.Email, subscription.City)
		}
	}
}