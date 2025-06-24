package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/services/subscription_service"
)

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	subscriptionService subscription_service.SubscriptionService
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionService subscription_service.SubscriptionService) SubscriptionHandler {
	return SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// Subscribe handles subscription requests
func (h SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req contracts.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validateSubscriptionRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.subscriptionService.Create(r.Context(), req.Email, req.City, req.Frequency)
	if err != nil {
		switch err {
		case apierrors.ErrAlreadySubscribed:
			http.Error(w, "Email already subscribed", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Confirm handles subscription confirmation
func (h *SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.Confirm(r.Context(), token); err != nil {
		switch err {
		case apierrors.ErrSubscriptionNotFound:
			http.Error(w, "Subscription not found", http.StatusNotFound)
		case apierrors.ErrInvalidToken:
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
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.Delete(r.Context(), token); err != nil {
		switch err {
		case apierrors.ErrSubscriptionNotFound:
			http.Error(w, "Subscription not found", http.StatusNotFound)
		case apierrors.ErrInvalidToken:
			http.Error(w, "Invalid token", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SubscriptionHandler) validateSubscriptionRequest(req contracts.SubscriptionRequest) error {
	var errs []string

	if req.Email == "" {
		errs = append(errs, apierrors.ErrInvalidEmail.Error())
	} else if _, err := mail.ParseAddress(req.Email); err != nil {

		errs = append(errs, apierrors.ErrInvalidEmail.Error())
	}
	if req.City == "" {
		errs = append(errs, apierrors.ErrInvalidCity.Error())
	}
	if req.Frequency != "hourly" && req.Frequency != "daily" {
		errs = append(errs, apierrors.ErrInvalidFrequency.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
