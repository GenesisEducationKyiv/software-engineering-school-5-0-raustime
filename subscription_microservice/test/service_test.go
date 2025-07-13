package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"subscription_microservice/internal/apierrors"
	"subscription_microservice/internal/db/models"
	"subscription_microservice/internal/subscription_service"
)

type mailerMock struct {
	mock.Mock
}

func (m *mailerMock) SendConfirmationEmail(ctx context.Context, email, token string) error {
	args := m.Called(ctx, email, token)
	return args.Error(0)
}

type repoMock struct {
	mock.Mock
}

func (r *repoMock) GetByEmail(ctx context.Context, email string) (models.Subscription, error) {
	args := r.Called(ctx, email)
	return args.Get(0).(models.Subscription), args.Error(1)
}

func (r *repoMock) GetByToken(ctx context.Context, token string) (models.Subscription, error) {
	args := r.Called(ctx, token)
	return args.Get(0).(models.Subscription), args.Error(1)
}

func (r *repoMock) GetConfirmed(ctx context.Context, frequency string) ([]models.Subscription, error) {
	args := r.Called(ctx, frequency)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (r *repoMock) Create(ctx context.Context, data models.Subscription) error {
	args := r.Called(ctx, data)
	return args.Error(0)
}

func (r *repoMock) Update(ctx context.Context, data models.Subscription) error {
	args := r.Called(ctx, data)
	return args.Error(0)
}

func (r *repoMock) Delete(ctx context.Context, token string) error {
	args := r.Called(ctx, token)
	return args.Error(0)
}

func TestCreateSubscription(t *testing.T) {
	repo := &repoMock{}
	mailer := &mailerMock{}
	svc := subscription_service.New(repo, mailer)

	ctx := context.Background()
	email := "test@example.com"
	city := "Kyiv"
	freq := "daily"

	t.Run("invalid email", func(t *testing.T) {
		err := svc.Create(ctx, "", city, freq)
		require.ErrorIs(t, err, apierrors.ErrInvalidEmail)
	})

	t.Run("already subscribed", func(t *testing.T) {
		repo.On("GetByEmail", ctx, email).Return(models.Subscription{Email: email}, nil)
		err := svc.Create(ctx, email, city, freq)
		require.ErrorIs(t, err, apierrors.ErrAlreadySubscribed)
	})

	t.Run("success", func(t *testing.T) {
		repo.ExpectedCalls = nil // reset
		mailer.ExpectedCalls = nil

		repo.On("GetByEmail", ctx, email).Return(models.Subscription{}, nil)
		repo.On("Create", ctx, mock.MatchedBy(func(m models.Subscription) bool {
			return m.Email == email && m.City == city && m.Frequency == freq
		})).Return(nil)
		mailer.On("SendConfirmationEmail", ctx, email, mock.Anything).Return(nil)

		err := svc.Create(ctx, email, city, freq)
		require.NoError(t, err)
	})
}

func TestConfirmSubscription(t *testing.T) {
	repo := &repoMock{}
	mailer := &mailerMock{}
	svc := subscription_service.New(repo, mailer)

	ctx := context.Background()
	token := uuid.New().String()

	t.Run("invalid token", func(t *testing.T) {
		err := svc.Confirm(ctx, "invalid-token")
		require.ErrorIs(t, err, apierrors.ErrInvalidToken)
	})

	t.Run("not found", func(t *testing.T) {
		repo.On("GetByToken", ctx, token).Return(models.Subscription{}, errors.New("not found"))
		err := svc.Confirm(ctx, token)
		require.ErrorIs(t, err, apierrors.ErrSubscriptionNotFound)
	})

	t.Run("success", func(t *testing.T) {
		sub := models.Subscription{Token: token}
		repo.On("GetByToken", ctx, token).Return(sub, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)
		err := svc.Confirm(ctx, token)
		require.NoError(t, err)
	})
}

func TestDeleteSubscription(t *testing.T) {
	repo := &repoMock{}
	mailer := &mailerMock{}
	svc := subscription_service.New(repo, mailer)

	ctx := context.Background()
	valid := uuid.New().String()

	t.Run("invalid token", func(t *testing.T) {
		err := svc.Delete(ctx, "invalid-token")
		require.ErrorIs(t, err, apierrors.ErrInvalidToken)
	})

	t.Run("success", func(t *testing.T) {
		repo.On("Delete", ctx, valid).Return(nil)
		err := svc.Delete(ctx, valid)
		require.NoError(t, err)
	})
}
