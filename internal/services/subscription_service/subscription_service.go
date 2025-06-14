package subscription_service

import (
	"context"
	"errors"
	"time"

	"weatherapi/internal/db/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var (
	ErrAlreadySubscribed    = errors.New("email already subscribed")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidToken         = errors.New("invalid token")
)

// SubscriptionService defines subscription service interface
type ISubscriptionService interface {
	CreateSubscription(ctx context.Context, email, city, frequency string) (*models.Subscription, error)
	ConfirmSubscription(ctx context.Context, token string) error
	DeleteSubscription(ctx context.Context, token string) error
	GetConfirmedSubscriptions(ctx context.Context, frequency string) ([]models.Subscription, error)
}

// subscriptionService implements SubscriptionService
type subscriptionService struct {
	db bun.IDB
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(db bun.IDB) ISubscriptionService {
	return &subscriptionService{db: db}
}

// CreateSubscription creates a new subscription
func (s *subscriptionService) CreateSubscription(ctx context.Context, email, city, frequency string) (*models.Subscription, error) {
	// Check if already subscribed
	var existing models.Subscription
	err := s.db.NewSelect().Model(&existing).Where("email = ?", email).Scan(ctx)
	if err == nil {
		return nil, ErrAlreadySubscribed
	}

	subscription := &models.Subscription{
		Email:     email,
		City:      city,
		Frequency: frequency,
		Token:     uuid.New().String(),
		CreatedAt: time.Now(),
	}

	if _, err := s.db.NewInsert().Model(subscription).Exec(ctx); err != nil {
		return nil, err
	}

	return subscription, nil
}

// ConfirmSubscription confirms a subscription
func (s *subscriptionService) ConfirmSubscription(ctx context.Context, token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return ErrInvalidToken
	}

	var subscription models.Subscription
	err := s.db.NewSelect().Model(&subscription).Where("token = ?", token).Scan(ctx)
	if err != nil {
		return ErrSubscriptionNotFound
	}

	subscription.Confirmed = true
	subscription.ConfirmedAt = time.Now()

	if _, err := s.db.NewUpdate().Model(&subscription).WherePK().Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DeleteSubscription deletes a subscription
func (s *subscriptionService) DeleteSubscription(ctx context.Context, token string) error {
	if _, err := uuid.Parse(token); err != nil {
		return ErrInvalidToken
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
		return ErrSubscriptionNotFound
	}

	return nil
}

// GetConfirmedSubscriptions retrieves confirmed subscriptions by frequency
func (s *subscriptionService) GetConfirmedSubscriptions(ctx context.Context, frequency string) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	err := s.db.NewSelect().
		Model(&subscriptions).
		Where("confirmed = TRUE").
		Where("frequency = ?", frequency).
		Scan(ctx)

	return subscriptions, err
}
