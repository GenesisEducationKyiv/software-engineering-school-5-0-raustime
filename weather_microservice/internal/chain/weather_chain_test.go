package chain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"weather_microservice/internal/chain"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/pkg/ctxkeys"
)

// --- Mock WeatherAPIProvider ---
type mockWeatherAPI struct {
	mock.Mock
}

func (m *mockWeatherAPI) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(contracts.WeatherData), args.Error(1)
}

// --- Mock Logger ---
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(ctx context.Context, source string, data interface{}) {
	m.Called(ctx, source, data)
}

func (m *mockLogger) Error(ctx context.Context, source string, data interface{}, err error) {
	m.Called(ctx, source, data, err)
}

// --- Mock Next Handler ---
type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) SetNext(h chain.WeatherHandler) chain.WeatherHandler {
	args := m.Called(h)
	return args.Get(0).(chain.WeatherHandler)
}

func (m *mockHandler) Handle(ctx context.Context, city string) (contracts.WeatherData, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(contracts.WeatherData), args.Error(1)
}

func (m *mockHandler) GetProviderName() string {
	args := m.Called()
	return args.String(0)
}

func TestBaseWeatherHandler_Success(t *testing.T) {
	api := new(mockWeatherAPI)
	logger := new(mockLogger)

	handler := chain.NewBaseWeatherHandler(api, "weatherapi")

	expected := contracts.WeatherData{Temperature: 20.5, Humidity: 50, Description: "Clear"}
	api.On("FetchWeather", mock.Anything, "Kyiv").Return(expected, nil)
	logger.On("Info", mock.Anything, "weatherapi", expected).Once()

	ctx := context.WithValue(context.Background(), ctxkeys.Logger, logger)

	data, err := handler.Handle(ctx, "weatherapi-Kyiv")
	assert.NoError(t, err)
	assert.Equal(t, expected, data)

	api.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestBaseWeatherHandler_ErrorAndNextFallback(t *testing.T) {
	api := new(mockWeatherAPI)
	logger := new(mockLogger)
	next := new(mockHandler)

	handler := chain.NewBaseWeatherHandler(api, "weatherapi")
	handler.SetNext(next)

	api.On("FetchWeather", mock.Anything, "Lviv").Return(contracts.WeatherData{}, errors.New("api failed"))
	logger.On("Error", mock.Anything, "weatherapi", nil, mock.Anything).Once()
	next.On("Handle", mock.Anything, "weatherapi-Lviv").Return(contracts.WeatherData{
		Temperature: 15.2,
		Humidity:    70,
		Description: "Partly Cloudy",
	}, nil)

	ctx := context.WithValue(context.Background(), ctxkeys.Logger, logger)

	data, err := handler.Handle(ctx, "weatherapi-Lviv")
	assert.NoError(t, err)
	assert.Equal(t, 15.2, data.Temperature)
	assert.Equal(t, "Partly Cloudy", data.Description)

	api.AssertExpectations(t)
	logger.AssertExpectations(t)
	next.AssertExpectations(t)
}

func TestBaseWeatherHandler_FinalFailure(t *testing.T) {
	api := new(mockWeatherAPI)
	logger := new(mockLogger)

	handler := chain.NewBaseWeatherHandler(api, "weatherapi")

	api.On("FetchWeather", mock.Anything, "CityX").Return(contracts.WeatherData{}, errors.New("network error"))
	logger.On("Error", mock.Anything, "weatherapi", nil, mock.Anything).Once()

	ctx := context.WithValue(context.Background(), ctxkeys.Logger, logger)

	data, err := handler.Handle(ctx, "weatherapi-CityX")
	assert.Error(t, err)
	assert.Empty(t, data.Description)
	assert.Contains(t, err.Error(), "all weather providers failed")

	api.AssertExpectations(t)
	logger.AssertExpectations(t)
}
