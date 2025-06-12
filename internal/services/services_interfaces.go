// services_interfaces.go - Файл з інтерфейсами та моделями для тестування
package services

import (
	"context"
	"errors"
	"time"
)

// Інтерфейси сервісів
type WeatherService interface {
	GetWeather(ctx context.Context, city string) (*Weather, error)
}

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, email, city, frequency string) (*Subscription, error)
	ConfirmSubscription(ctx context.Context, token string) error
	DeleteSubscription(ctx context.Context, token string) error
}

type MailerService interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
}

// Моделі даних
type Weather struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

type Subscription struct {
	ID        int64     `json:"id" bun:"id,pk,autoincrement"`
	Email     string    `json:"email" bun:"email,notnull"`
	City      string    `json:"city" bun:"city,notnull"`
	Frequency string    `json:"frequency" bun:"frequency,notnull"`
	Token     string    `json:"token" bun:"token,unique,notnull"`
	Confirmed bool      `json:"confirmed" bun:"confirmed,default:false"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,default:current_timestamp"`
}

// Помилки сервісів
var (
	ErrCityNotFound          = errors.New("city not found")
	ErrAlreadySubscribed     = errors.New("email already subscribed")
	ErrSubscriptionNotFound  = errors.New("subscription not found")
	ErrInvalidToken          = errors.New("invalid token")
)