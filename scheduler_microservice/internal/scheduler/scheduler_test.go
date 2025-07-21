package scheduler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"scheduler_microservice/internal/contracts"
	"scheduler_microservice/internal/scheduler"

	"github.com/stretchr/testify/mock"
)

type mockSubSvc struct{ mock.Mock }

func (m *mockSubSvc) GetConfirmed(ctx context.Context, frequency string) ([]*contracts.Subscription, error) {
	args := m.Called(ctx, frequency)
	return args.Get(0).([]*contracts.Subscription), args.Error(1)
}

type mockWeatherSvc struct{ mock.Mock }

func (m *mockWeatherSvc) GetWeather(ctx context.Context, city string) (*contracts.WeatherData, error) {
	args := m.Called(ctx, city)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contracts.WeatherData), args.Error(1)
}

type mockPublisher struct{ mock.Mock }

func (m *mockPublisher) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func TestScheduler_Send_Success(t *testing.T) {
	sub := &contracts.Subscription{
		Email: "test@example.com",
		City:  "Kyiv",
		Token: "abc123",
	}
	weather := &contracts.WeatherData{
		Temperature: 21.5,
		Humidity:    50,
		Description: "Clear sky",
	}

	expectedMsg := contracts.NotificationMessage{
		Type:    "weather",
		To:      "test@example.com",
		City:    "Kyiv",
		Token:   "abc123",
		Weather: weather,
	}
	expectedPayload, _ := json.Marshal(expectedMsg)

	subSvc := new(mockSubSvc)
	weatherSvc := new(mockWeatherSvc)
	mailPub := new(mockPublisher)

	subSvc.On("GetConfirmed", mock.Anything, "hourly").Return([]*contracts.Subscription{sub}, nil)
	weatherSvc.On("GetWeather", mock.Anything, "Kyiv").Return(weather, nil)
	mailPub.On("Publish", "mailer.notifications", expectedPayload).Return(nil)

	s := scheduler.NewScheduler(subSvc, mailPub, weatherSvc)
	s.Send("hourly")

	subSvc.AssertExpectations(t)
	weatherSvc.AssertExpectations(t)
	mailPub.AssertExpectations(t)
}

func TestScheduler_Send_WeatherError(t *testing.T) {
	sub := &contracts.Subscription{
		Email: "fail@example.com",
		City:  "Odesa",
		Token: "fail",
	}

	subSvc := new(mockSubSvc)
	weatherSvc := new(mockWeatherSvc)
	mailPub := new(mockPublisher)

	subSvc.On("GetConfirmed", mock.Anything, "daily").Return([]*contracts.Subscription{sub}, nil)
	weatherSvc.On("GetWeather", mock.Anything, "Odesa").Return(nil, errors.New("weather error"))

	s := scheduler.NewScheduler(subSvc, mailPub, weatherSvc)
	s.Send("daily")

	subSvc.AssertExpectations(t)
	weatherSvc.AssertExpectations(t)
	mailPub.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}
