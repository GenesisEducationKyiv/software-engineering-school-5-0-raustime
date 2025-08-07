package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	subpb "weather_microservice/gen/go/subscription/v1"
	"weather_microservice/internal/client"
	"weather_microservice/internal/logging"

	"connectrpc.com/connect"
)

type SubscriptionHandler struct {
	client *client.SubscriptionClient
}

func NewSubscriptionHandler(client *client.SubscriptionClient) SubscriptionHandler {
	return SubscriptionHandler{
		client: client,
	}
}

func (h SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn(ctx, "http:Subscribe", nil, http.ErrNotSupported)
		return
	}

	var reqData struct {
		Email     string `json:"email"`
		City      string `json:"city"`
		Frequency string `json:"frequency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		logger.Warn(ctx, "http:Subscribe", nil, err)
		return
	}

	req := connect.NewRequest(&subpb.CreateRequest{
		Email:     reqData.Email,
		City:      reqData.City,
		Frequency: reqData.Frequency,
	})

	_, err := h.client.Client.Create(ctx, req)
	if err != nil {
		logger.Error(ctx, "http:Subscribe", map[string]string{
			"email": reqData.Email,
			"city":  reqData.City,
		}, err)
		http.Error(w, "Failed to create subscription", http.StatusBadGateway)
		return
	}

	logger.Info(ctx, "http:Subscribe", map[string]string{
		"email": reqData.Email,
		"city":  reqData.City,
	})
	w.WriteHeader(http.StatusCreated)
}

func (h SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn(ctx, "http:Confirm", nil, http.ErrNotSupported)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/confirm/")
	req := connect.NewRequest(&subpb.ConfirmRequest{Token: token})

	_, err := h.client.Client.Confirm(ctx, req)
	if err != nil {
		logger.Error(ctx, "http:Confirm", map[string]string{"token": token}, err)
		http.Error(w, "Failed to confirm subscription", http.StatusBadGateway)
		return
	}

	logger.Info(ctx, "http:Confirm", map[string]string{"token": token})
	w.WriteHeader(http.StatusOK)
}

func (h SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn(ctx, "http:Unsubscribe", nil, http.ErrNotSupported)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/unsubscribe/")
	req := connect.NewRequest(&subpb.DeleteRequest{Token: token})

	_, err := h.client.Client.Delete(ctx, req)
	if err != nil {
		logger.Error(ctx, "http:Unsubscribe", map[string]string{"token": token}, err)
		http.Error(w, "Failed to unsubscribe", http.StatusBadGateway)
		return
	}

	logger.Info(ctx, "http:Unsubscribe", map[string]string{"token": token})
	w.WriteHeader(http.StatusOK)
}
