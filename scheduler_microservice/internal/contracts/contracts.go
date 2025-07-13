package contracts

import "time"

type EmailSenderProvider interface {
	Send(to, subject, htmlBody string) error
}

// WeatherData represents weather information.
type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

type Subscription struct {
	ID          int64
	Email       string
	City        string
	Frequency   string
	Confirmed   bool
	Token       string
	CreatedAt   time.Time
	ConfirmedAt time.Time
}
