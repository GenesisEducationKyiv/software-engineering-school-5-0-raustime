package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"log"

	subpb "weather_microservice/gen/go/subscription/v1"
	"weather_microservice/internal/client"

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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqData struct {
		Email     string `json:"email"`
		City      string `json:"city"`
		Frequency string `json:"frequency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	req := connect.NewRequest(&subpb.CreateRequest{
		Email:     reqData.Email,
		City:      reqData.City,
		Frequency: reqData.Frequency,
	})

	_, err := h.client.Client.Create(r.Context(), req)
	if err != nil {
		log.Printf("↪ RPC Create -> %s", req.Spec().Procedure)
		log.Printf("[SubscriptionHandler] failed to create subscription: %v", err)
		http.Error(w, "Failed to create subscription", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/confirm/")
	req := connect.NewRequest(&subpb.ConfirmRequest{Token: token})

	_, err := h.client.Client.Confirm(r.Context(), req)
	if err != nil {
		http.Error(w, "Failed to confirm subscription", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := strings.TrimPrefix(r.URL.Path, "/api/unsubscribe/")
	req := connect.NewRequest(&subpb.DeleteRequest{Token: token})

	_, err := h.client.Client.Delete(r.Context(), req)
	if err != nil {
		http.Error(w, "Failed to unsubscribe", http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}
