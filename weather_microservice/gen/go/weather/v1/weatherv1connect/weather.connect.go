// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: weather/v1/weather.proto

package weatherv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	http "net/http"
	strings "strings"
	v1 "weather_microservice/gen/go/weather/v1"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// WeatherServiceName is the fully-qualified name of the WeatherService service.
	WeatherServiceName = "weather.WeatherService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// WeatherServiceGetWeatherProcedure is the fully-qualified name of the WeatherService's GetWeather
	// RPC.
	WeatherServiceGetWeatherProcedure = "/weather.WeatherService/GetWeather"
)

// WeatherServiceClient is a client for the weather.WeatherService service.
type WeatherServiceClient interface {
	GetWeather(context.Context, *connect.Request[v1.GetWeatherRequest]) (*connect.Response[v1.GetWeatherResponse], error)
}

// NewWeatherServiceClient constructs a client for the weather.WeatherService service. By default,
// it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and
// sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC()
// or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewWeatherServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) WeatherServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	weatherServiceMethods := v1.File_weather_v1_weather_proto.Services().ByName("WeatherService").Methods()
	return &weatherServiceClient{
		getWeather: connect.NewClient[v1.GetWeatherRequest, v1.GetWeatherResponse](
			httpClient,
			baseURL+WeatherServiceGetWeatherProcedure,
			connect.WithSchema(weatherServiceMethods.ByName("GetWeather")),
			connect.WithClientOptions(opts...),
		),
	}
}

// weatherServiceClient implements WeatherServiceClient.
type weatherServiceClient struct {
	getWeather *connect.Client[v1.GetWeatherRequest, v1.GetWeatherResponse]
}

// GetWeather calls weather.WeatherService.GetWeather.
func (c *weatherServiceClient) GetWeather(ctx context.Context, req *connect.Request[v1.GetWeatherRequest]) (*connect.Response[v1.GetWeatherResponse], error) {
	return c.getWeather.CallUnary(ctx, req)
}

// WeatherServiceHandler is an implementation of the weather.WeatherService service.
type WeatherServiceHandler interface {
	GetWeather(context.Context, *connect.Request[v1.GetWeatherRequest]) (*connect.Response[v1.GetWeatherResponse], error)
}

// NewWeatherServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewWeatherServiceHandler(svc WeatherServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	weatherServiceMethods := v1.File_weather_v1_weather_proto.Services().ByName("WeatherService").Methods()
	weatherServiceGetWeatherHandler := connect.NewUnaryHandler(
		WeatherServiceGetWeatherProcedure,
		svc.GetWeather,
		connect.WithSchema(weatherServiceMethods.ByName("GetWeather")),
		connect.WithHandlerOptions(opts...),
	)
	return "/weather.WeatherService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case WeatherServiceGetWeatherProcedure:
			weatherServiceGetWeatherHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedWeatherServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedWeatherServiceHandler struct{}

func (UnimplementedWeatherServiceHandler) GetWeather(context.Context, *connect.Request[v1.GetWeatherRequest]) (*connect.Response[v1.GetWeatherResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("weather.WeatherService.GetWeather is not implemented"))
}
