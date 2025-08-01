package subscription_service

import (
	"context"
	"encoding/json"
	"log"
	"net/mail"
	"time"

	"github.com/google/uuid"

	"subscription_microservice/internal/apierrors"
	"subscription_microservice/internal/contracts"
	"subscription_microservice/internal/db/models"
)

type subscriptionRepo interface {
	GetByEmail(ctx context.Context, email string) (models.Subscription, error)
	GetByToken(ctx context.Context, token string) (models.Subscription, error)
	GetConfirmed(ctx context.Context, frequency string) ([]models.Subscription, error)
	Create(ctx context.Context, data models.Subscription) error
	Update(ctx context.Context, data models.Subscription) error
	Delete(ctx context.Context, token string) error
}

type messageBroker interface {
	Publish(subject string, data []byte) error
}

type SubscriptionService struct {
	subRepo subscriptionRepo
	broker  messageBroker
}

func New(sr subscriptionRepo, broker messageBroker) SubscriptionService {
	return SubscriptionService{
		subRepo: sr,
		broker:  broker,
	}
}

func (s SubscriptionService) Create(ctx context.Context, email, city, frequency string) error {
	if email == "" {
		return apierrors.ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return apierrors.ErrInvalidEmail
	}

	// Валідація city
	if city == "" {
		return apierrors.ErrInvalidCity
	}

	// Валідація frequency
	validFreq := map[string]bool{"daily": true, "hourly": true}
	if !validFreq[frequency] {
		return apierrors.ErrInvalidFrequency
	}
	existing, err := s.subRepo.GetByEmail(ctx, email)
	if err != nil && err != apierrors.ErrSubscriptionNotFound {
		// лог будь-яких несподіваних помилок
		log.Printf("[SubscriptionService] error checking existing email: %v", err)
	}
	if existing != (models.Subscription{}) {
		return apierrors.ErrAlreadySubscribed
	}

	subscription := models.Subscription{
		Email:     email,
		City:      city,
		Frequency: frequency,
		Token:     uuid.New().String(),
		CreatedAt: time.Now(),
	}

	if err := s.subRepo.Create(ctx, subscription); err != nil {
		return err
	}

	notif := contracts.NotificationMessage{
		Type:  "confirmation",
		To:    email,
		Token: subscription.Token,
	}

	payload, err := json.Marshal(notif)
	if err != nil {
		log.Printf("[SubscriptionService] failed to marshal notification: %v", err)
		return apierrors.ErrFailedSendConfirmEmail
	}

	if err := s.broker.Publish("mailer.notifications", payload); err != nil {
		log.Printf("[SubscriptionService] failed to publish confirmation event: %v", err)
		return apierrors.ErrFailedSendConfirmEmail
	}

	return nil
}

func (s SubscriptionService) Confirm(ctx context.Context, token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return apierrors.ErrInvalidToken
	}

	var subscription models.Subscription
	subscription, err := s.subRepo.GetByToken(ctx, token)
	if err != nil {
		return apierrors.ErrSubscriptionNotFound
	}

	subscription.Confirmed = true
	subscription.ConfirmedAt = time.Now()

	return s.subRepo.Update(ctx, subscription)
}

func (s SubscriptionService) Delete(ctx context.Context, token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return apierrors.ErrInvalidToken
	}

	err := s.subRepo.Delete(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (s SubscriptionService) GetConfirmed(ctx context.Context, frequency string) ([]contracts.Subscription, error) {
	modelSubs, err := s.subRepo.GetConfirmed(ctx, frequency)
	if err != nil {
		return nil, err
	}

	// Конвертація в contracts.Subscription
	converted := make([]contracts.Subscription, len(modelSubs))
	for i, m := range modelSubs {
		converted[i] = contracts.Subscription{
			ID:          m.ID,
			Email:       m.Email,
			City:        m.City,
			Frequency:   m.Frequency,
			Confirmed:   m.Confirmed,
			Token:       m.Token,
			CreatedAt:   m.CreatedAt,
			ConfirmedAt: m.ConfirmedAt,
		}
	}

	return converted, nil
}
