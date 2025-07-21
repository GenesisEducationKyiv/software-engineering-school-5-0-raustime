package client

import (
	"net/http"

	subpb "weather_microservice/gen/go/subscription/v1/subscriptionv1connect"
)

type SubscriptionClient struct {
	Client subpb.SubscriptionServiceClient
}

func NewSubscriptionClient(baseURL string) *SubscriptionClient {
	return &SubscriptionClient{
		Client: subpb.NewSubscriptionServiceClient(
			&http.Client{},
			baseURL, 
		),
	}
}
