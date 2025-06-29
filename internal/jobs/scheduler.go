package jobs

import (
	"context"
	"log"
	"time"

	"weatherapi/internal/contracts"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"
)

// Interfaces for dependency injection and testing.

// subscriptionService defines required methods from SubscriptionService.
type subscriptionService interface {
	GetConfirmed(ctx context.Context, frequency string) ([]contracts.Subscription, error)
}

// mailerService defines required methods from MailerService.
type mailerService interface {
	SendWeatherEmail(ctx context.Context, to, city string, data contracts.WeatherData, token string) error
}

// weatherService defines required methods from WeatherService.
type weatherService interface {
	GetWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

type Scheduler struct {
	subscriptionService subscriptionService
	mailerService       mailerService
	weatherService      weatherService
	stopChan            chan struct{}
}

// NewScheduler creates a new job scheduler.
func NewScheduler(
	subscriptionService subscription_service.SubscriptionService,
	mailerService mailer_service.MailerService,
	weatherService weather_service.WeatherService,
) Scheduler {
	return Scheduler{
		subscriptionService: subscriptionService,
		mailerService:       mailerService,
		weatherService:      weatherService,
		stopChan:            make(chan struct{}),
	}
}

// Start starts the job scheduler.
func (s Scheduler) Start() {
	go s.weatherNotificationLoop()
}

// Stop stops the job scheduler.
func (s Scheduler) Stop() {
	close(s.stopChan)
}

// weatherNotificationLoop runs the weather notification loop.
func (s Scheduler) weatherNotificationLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()

			// Send hourly updates (00 minutes).
			if now.Minute() == 0 {
				s.sendWeatherUpdates("hourly")

				// Send daily updates at 8:00 AM.
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

// sendWeatherUpdates sends weather updates for specified frequency.
func (s Scheduler) sendWeatherUpdates(frequency string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subscriptions, err := s.subscriptionService.GetConfirmed(ctx, frequency)
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
		weatherData := contracts.WeatherData{
			Temperature: weather.Temperature,
			Humidity:    weather.Humidity,
			Description: weather.Description,
		}

		err = s.mailerService.SendWeatherEmail(ctx, subscription.Email, subscription.City, weatherData, subscription.Token)
		if err != nil {
			log.Printf("Failed to send weather email to %s: %v", subscription.Email, err)
		} else {
			log.Printf("Sent %s weather email to %s (%s)", frequency, subscription.Email, subscription.City)
		}
	}
}
