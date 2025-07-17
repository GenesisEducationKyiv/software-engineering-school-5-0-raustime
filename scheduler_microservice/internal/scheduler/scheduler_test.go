package scheduler_test

import (
	"context"
	"errors"
	"testing"
	
	"github.com/stretchr/testify/mock"

	"scheduler_microservice/internal/contracts"
	"scheduler_microservice/internal/scheduler"
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

type mockMailerSvc struct{ mock.Mock }
func (m *mockMailerSvc) SendWeatherEmail(ctx context.Context, to, city string, data *contracts.WeatherData, token string) error {
	args := m.Called(ctx, to, city, data, token)
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

	subSvc := new(mockSubSvc)
	weatherSvc := new(mockWeatherSvc)
	mailerSvc := new(mockMailerSvc)

	subSvc.On("GetConfirmed", mock.Anything, "hourly").Return([]*contracts.Subscription{sub}, nil)
	weatherSvc.On("GetWeather", mock.Anything, "Kyiv").Return(weather, nil)
	mailerSvc.On("SendWeatherEmail", mock.Anything, "test@example.com", "Kyiv", weather, "abc123").Return(nil)

	s := scheduler.NewScheduler(subSvc, mailerSvc, weatherSvc)
	s.Send("hourly")

	subSvc.AssertExpectations(t)
	weatherSvc.AssertExpectations(t)
	mailerSvc.AssertExpectations(t)
}

func TestScheduler_Send_WeatherError(t *testing.T) {
	sub := &contracts.Subscription{
		Email: "err@example.com",
		City:  "Odesa",
		Token: "fail",
	}

	subSvc := new(mockSubSvc)
	weatherSvc := new(mockWeatherSvc)
	mailerSvc := new(mockMailerSvc)

	subSvc.On("GetConfirmed", mock.Anything, "daily").Return([]*contracts.Subscription{sub}, nil)
	weatherSvc.On("GetWeather", mock.Anything, "Odesa").Return(nil, errors.New("weather error"))

	s := scheduler.NewScheduler(subSvc, mailerSvc, weatherSvc)
	s.Send("daily")

	subSvc.AssertExpectations(t)
	weatherSvc.AssertExpectations(t)
	mailerSvc.AssertNotCalled(t, "SendWeatherEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
