package contracts

type IEmailSender interface {
	Send(to, subject, htmlBody string) error
}

type WeatherData struct {
	Temperature float64
	Humidity    float64
	Description string
}
