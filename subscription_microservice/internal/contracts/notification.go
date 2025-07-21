type NotificationMessage struct {
	Type    string       `json:"type"`
	To      string       `json:"to"`
	Token   string       `json:"token,omitempty"`
	City    string       `json:"city,omitempty"`
	Weather *WeatherData `json:"weather,omitempty"`
	Subject string       `json:"subject,omitempty"`
	Body    string       `json:"body,omitempty"`
}
