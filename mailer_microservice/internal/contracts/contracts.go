package contracts

type EmailSenderProvider interface {
	Send(to, subject, htmlBody string) error
}

// WeatherData represents weather information.
type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

type NotificationMessage struct {
	Type    string       `json:"type"`              // "confirmation", "weather", "custom"
	To      string       `json:"to"`                // Email адреса
	Token   string       `json:"token,omitempty"`   // для підтвердження/відписки
	City    string       `json:"city,omitempty"`    // для weather email
	Weather *WeatherData `json:"weather,omitempty"` // вбудований об'єкт погоди
	Subject string       `json:"subject,omitempty"` // кастомний заголовок
	Body    string       `json:"body,omitempty"`    // кастомне HTML тіло
}
