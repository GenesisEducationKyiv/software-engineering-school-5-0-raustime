package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
)

var (
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidCity      = errors.New("invalid city")
	ErrInvalidFrequency = errors.New("invalid frequency")
)

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	subscriptionService subscription_service.ISubscriptionService
	mailerService       mailer_service.IMailerService
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionService subscription_service.ISubscriptionService, mailerService mailer_service.IMailerService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		mailerService:       mailerService,
	}
}

// SubscriptionRequest represents subscription request
type SubscriptionRequest struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

// Subscribe handles subscription requests
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validateSubscriptionRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	subscription, err := h.subscriptionService.CreateSubscription(r.Context(), req.Email, req.City, req.Frequency)
	if err != nil {
		switch err {
		case subscription_service.ErrAlreadySubscribed:
			http.Error(w, "Email already subscribed", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := h.mailerService.SendConfirmationEmail(r.Context(), req.Email, subscription.Token); err != nil {
		http.Error(w, "Failed to send confirmation email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Confirm handles subscription confirmation
func (h *SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/api/confirm/")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.ConfirmSubscription(r.Context(), token); err != nil {
		switch err {
		case subscription_service.ErrSubscriptionNotFound:
			http.Error(w, "Subscription not found", http.StatusNotFound)
		case subscription_service.ErrInvalidToken:
			http.Error(w, "Invalid token", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Unsubscribe handles unsubscription
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/api/unsubscribe/")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.DeleteSubscription(r.Context(), token); err != nil {
		switch err {
		case subscription_service.ErrSubscriptionNotFound:
			http.Error(w, "Subscription not found", http.StatusNotFound)
		case subscription_service.ErrInvalidToken:
			http.Error(w, "Invalid token", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// validateSubscriptionRequest validates subscription request
func (h *SubscriptionHandler) validateSubscriptionRequest(req SubscriptionRequest) error {
	if req.Email == "" {
		return ErrInvalidEmail
	}
	if req.City == "" {
		return ErrInvalidCity
	}
	if req.Frequency != "hourly" && req.Frequency != "daily" {
		return ErrInvalidFrequency
	}
	return nil
}
