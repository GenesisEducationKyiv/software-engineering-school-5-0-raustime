package repositories

import (
	"context"

	"github.com/uptrace/bun"

	"subscription_microservice/internal/db/models"
)

type SubscriptionRepo struct {
	db *bun.DB
}

func NewSubscriptionRepo(db *bun.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) GetByEmail(ctx context.Context, email string) (models.Subscription, error) {
	var sub models.Subscription
	err := r.db.NewSelect().Model(&sub).Where("email = ?", email).Scan(ctx)
	return sub, err
}

func (r *SubscriptionRepo) GetByToken(ctx context.Context, token string) (models.Subscription, error) {
	var sub models.Subscription
	err := r.db.NewSelect().Model(&sub).Where("token = ?", token).Scan(ctx)
	return sub, err
}

func (r *SubscriptionRepo) GetConfirmed(ctx context.Context, frequency string) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.NewSelect().Model(&subs).Where("confirmed = TRUE AND frequency = ?", frequency).Scan(ctx)
	return subs, err
}

func (r *SubscriptionRepo) Create(ctx context.Context, data models.Subscription) error {
	_, err := r.db.NewInsert().Model(&data).Exec(ctx)
	return err
}

func (r *SubscriptionRepo) Update(ctx context.Context, data models.Subscription) error {
	_, err := r.db.NewUpdate().Model(&data).WherePK().Exec(ctx)
	return err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, token string) error {
	_, err := r.db.NewDelete().Model((*models.Subscription)(nil)).Where("token = ?", token).Exec(ctx)
	return err
}
