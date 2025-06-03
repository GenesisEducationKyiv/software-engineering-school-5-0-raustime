package jobs

import (
	"context"
	"log"
	"os"
	"time"

	"weatherapi/internal/db/models"
	"weatherapi/internal/mailer"
	openweatherapi "weatherapi/internal/openweatherapi"

	"github.com/uptrace/bun"
)

func StartWeatherNotificationLoop(db *bun.DB, sender mailer.EmailSender) {
	go func() {
		for {
			now := time.Now()

			// Надсилати щогодини (00 хвилин)
			if now.Minute() == 0 {
				sendUpdates(db, "hourly", sender)

				// О 8:00 — щоденні
				if now.Hour() == 8 {
					sendUpdates(db, "daily", sender)
				}
			}

			time.Sleep(1 * time.Minute)
		}
	}()
}
func sendUpdates(db *bun.DB, frequency string, sender mailer.EmailSender) {

	baseURL := os.Getenv("APP_BASE_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var subs []models.Subscription
	err := db.NewSelect().
		Model(&subs).
		Where("confirmed = TRUE").
		Where("frequency = ?", frequency).
		Scan(ctx)
	if err != nil {
		log.Printf("Failed to fetch %s subscriptions: %v", frequency, err)
		return
	}

	for _, sub := range subs {
		weather, err := openweatherapi.FetchWeather(sub.City)
		if err != nil {
			log.Printf("Weather fetch error for %s: %v", sub.City, err)
			continue
		}

		err = mailer.SendWeatherEmailWithSender(sender, sub.Email, sub.City, weather, baseURL, sub.Token)
		if err != nil {
			log.Printf("Failed to send weather email to %s: %v", sub.Email, err)
		} else {
			log.Printf("Sent %s weather email to %s (%s)", frequency, sub.Email, sub.City)
		}
	}
}
