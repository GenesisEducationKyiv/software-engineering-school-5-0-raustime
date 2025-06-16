package subscription_service

import (
	"context"
	"time"

	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/db/models"
	"weatherapi/internal/services/mailer_service"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SubscriptionService defines subscription service interface
type subscriptionServiceProvider interface {
	CreateSubscription(ctx context.Context, email, city, frequency string) error
	ConfirmSubscription(ctx context.Context, token string) error
	DeleteSubscription(ctx context.Context, token string) error
	GetConfirmedSubscriptions(ctx context.Context, frequency string) ([]contracts.Subscription, error)
}

type mailSender interface {
	SendConfirmationEmail(ctx context.Context, to string, token string) error
}

// subscriptionService implements SubscriptionService
type SubscriptionService struct {
	db           bun.IDB
	mailerSender mailer_service.MailerService
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(db bun.IDB, mailer mailer_service.MailerService) SubscriptionService {
	return SubscriptionService{
		db:           db,
		mailerSender: mailer,
	}
}

// CreateSubscription creates a new subscription
func (s SubscriptionService) CreateSubscription(ctx context.Context, email, city, frequency string) error {
	return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var existing models.Subscription
		err := tx.NewSelect().Model(&existing).Where("email = ?", email).Scan(ctx)
		if err == nil {
			return apierrors.ErrAlreadySubscribed
		}

		subscription := models.Subscription{
			Email:     email,
			City:      city,
			Frequency: frequency,
			Token:     uuid.New().String(),
			CreatedAt: time.Now(),
		}

		if _, err := tx.NewInsert().Model(&subscription).Exec(ctx); err != nil {
			return err
		}

		if err := s.mailerSender.SendConfirmationEmail(ctx, email, subscription.Token); err != nil {
			return apierrors.ErrFailedSendConfirmEmail
		}

		return nil
	})
}

// ConfirmSubscription confirms a subscription
func (s SubscriptionService) ConfirmSubscription(ctx context.Context, token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return apierrors.ErrInvalidToken
	}

	var subscription models.Subscription
	err := s.db.NewSelect().Model(&subscription).Where("token = ?", token).Scan(ctx)
	if err != nil {
		return apierrors.ErrSubscriptionNotFound
	}

	subscription.Confirmed = true
	subscription.ConfirmedAt = time.Now()

	_, err = s.db.NewUpdate().Model(&subscription).WherePK().Exec(ctx)
	return err
}

// DeleteSubscription deletes a subscription
func (s SubscriptionService) DeleteSubscription(ctx context.Context, token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return apierrors.ErrInvalidToken
	}

	res, err := s.db.NewDelete().Model((*models.Subscription)(nil)).Where("token = ?", token).Exec(ctx)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return apierrors.ErrSubscriptionNotFound
	}

	return nil
}

// GetConfirmedSubscriptions retrieves confirmed subscriptions by frequency
func (s SubscriptionService) GetConfirmedSubscriptions(ctx context.Context, frequency string) ([]contracts.Subscription, error) {
	var modelSubs []models.Subscription
	err := s.db.NewSelect().
		Model(&modelSubs).
		Where("confirmed = TRUE").
		Where("frequency = ?", frequency).
		Scan(ctx)

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
