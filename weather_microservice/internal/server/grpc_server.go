package server

import (
	"context"
	weatherv1 "weather_microservice/gen/go/weather/v1"
	"weather_microservice/gen/go/weather/v1/weatherv1connect"
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
	data, err := s.service.GetWeather(ctx, r.Msg.City)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := &weatherv1.GetWeatherResponse{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}
	return connect.NewResponse(res), nil
}
