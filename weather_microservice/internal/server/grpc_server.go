package server

import (
	"context"
	weatherv1 "weather_microservice/gen/go/weather/v1"
	"weather_microservice/gen/go/weather/v1/weatherv1connect"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/weather_service"

	"connectrpc.com/connect"
)

type GRPCWeatherServer struct {
	weatherv1connect.UnimplementedWeatherServiceHandler
	service weather_service.WeatherService
}

func NewGRPCWeatherServer(service weather_service.WeatherService) *GRPCWeatherServer {
	return &GRPCWeatherServer{service: service}
}
func (s *GRPCWeatherServer) GetWeather(
	ctx context.Context,
	r *connect.Request[weatherv1.GetWeatherRequest],
) (*connect.Response[weatherv1.GetWeatherResponse], error) {
	logger := logging.FromContext(ctx)
	city := r.Msg.City

	data, err := s.service.GetWeather(ctx, city)
	if err != nil {
		logger.Error(ctx, "grpc:GetWeather", map[string]string{"city": city}, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	logger.Info(ctx, "grpc:GetWeather", map[string]interface{}{
		"city":        city,
		"temperature": data.Temperature,
		"humidity":    data.Humidity,
		"description": data.Description,
	})

	res := &weatherv1.GetWeatherResponse{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}
	return connect.NewResponse(res), nil
}
