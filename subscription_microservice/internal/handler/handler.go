package handler

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	subscriptionv1 "subscription_microservice/gen/go/subscription/v1"
	"subscription_microservice/gen/go/subscription/v1/subscriptionv1connect"
	"subscription_microservice/internal/subscription_service"
)

type SubscriptionHandler struct {
	impl *subscription_service.SubscriptionService
}

func NewHandler(svc *subscription_service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{impl: svc}
}

// Register створює HTTP handler і маршрут для підключення до HTTP-сервера.
func Register(mux *http.ServeMux, svc *subscription_service.SubscriptionService) {
	path, handler := subscriptionv1connect.NewSubscriptionServiceHandler(NewHandler(svc))
	mux.Handle(path, handler)
}

func (h *SubscriptionHandler) Create(
	ctx context.Context,
	req *connect.Request[subscriptionv1.CreateRequest],
) (*connect.Response[subscriptionv1.CreateResponse], error) {
	err := h.impl.Create(ctx, req.Msg.Email, req.Msg.City, req.Msg.Frequency)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&subscriptionv1.CreateResponse{}), nil
}

func (h *SubscriptionHandler) Confirm(
	ctx context.Context,
	req *connect.Request[subscriptionv1.ConfirmRequest],
) (*connect.Response[subscriptionv1.ConfirmResponse], error) {
	err := h.impl.Confirm(ctx, req.Msg.Token)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&subscriptionv1.ConfirmResponse{}), nil
}

func (h *SubscriptionHandler) Delete(
	ctx context.Context,
	req *connect.Request[subscriptionv1.DeleteRequest],
) (*connect.Response[subscriptionv1.DeleteResponse], error) {
	err := h.impl.Delete(ctx, req.Msg.Token)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&subscriptionv1.DeleteResponse{}), nil
}

func (h *SubscriptionHandler) GetConfirmed(
	ctx context.Context,
	req *connect.Request[subscriptionv1.GetConfirmedRequest],
) (*connect.Response[subscriptionv1.GetConfirmedResponse], error) {
	subs, err := h.impl.GetConfirmed(ctx, req.Msg.Frequency)
	if err != nil {
		return nil, err
	}

	result := make([]*subscriptionv1.Subscription, 0, len(subs))
	for _, sub := range subs {
		result = append(result, &subscriptionv1.Subscription{
			Id:          uint64(sub.ID),
			Email:       sub.Email,
			City:        sub.City,
			Frequency:   sub.Frequency,
			Token:       sub.Token,
			Confirmed:   sub.Confirmed,
			CreatedAt:   timestamppb.New(sub.CreatedAt),
			ConfirmedAt: timestamppb.New(sub.ConfirmedAt),
		})
	}

	return connect.NewResponse(&subscriptionv1.GetConfirmedResponse{
		Subscriptions: result,
	}), nil
}
