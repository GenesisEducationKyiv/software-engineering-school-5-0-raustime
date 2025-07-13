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
