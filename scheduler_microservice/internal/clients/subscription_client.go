package clients

import (
	"context"
	"scheduler_microservice/internal/contracts"

	"net/http"
	subscriptionv1 "scheduler_microservice/gen/go/subscription/v1"
	subscriptionv1connect "scheduler_microservice/gen/go/subscription/v1/subscriptionv1connect"

	connect "connectrpc.com/connect"
)

type subscriptionClient struct {
	client subscriptionv1connect.SubscriptionServiceClient
}

func NewSubscriptionClient(httpClient *http.Client, baseURL string) *subscriptionClient {
	return &subscriptionClient{
		client: subscriptionv1connect.NewSubscriptionServiceClient(httpClient, baseURL),
	}
}

func (c *subscriptionClient) GetConfirmed(ctx context.Context, frequency string) ([]*contracts.Subscription, error) {
	resp, err := c.client.GetConfirmed(ctx, connect.NewRequest(&subscriptionv1.GetConfirmedRequest{Frequency: frequency}))
	if err != nil {
		return nil, err
	}

	var subs []*contracts.Subscription
	for _, s := range resp.Msg.Subscriptions {
		subs = append(subs, &contracts.Subscription{
			Email: s.Email,
			City:  s.City,
			Token: s.Token,
		})
	}
	return subs, nil
}
