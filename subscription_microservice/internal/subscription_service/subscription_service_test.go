package subscription_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"subscription_microservice/internal/apierrors"
	"subscription_microservice/internal/db/models"
)

// mailerServiceMock implements the mailer service interface.
type messageBrokerMock struct {
	mock.Mock
}

func (m *messageBrokerMock) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *messageBrokerMock) SendConfirmationEmail(ctx context.Context, email, token string) error {
	args := m.Called(ctx, email, token)
	return args.Error(0)
}

// subscriptionRepoMock implements the subscription repository interface.
type subscriptionRepoMock struct {
	mock.Mock
}

func (m *subscriptionRepoMock) GetByEmail(ctx context.Context, email string) (models.Subscription, error) {
	args := m.Called(ctx, email)
	if sub, ok := args.Get(0).(models.Subscription); ok {
		return sub, args.Error(1)
	}
	return models.Subscription{}, args.Error(1)
}

func (m *subscriptionRepoMock) GetByToken(ctx context.Context, token string) (models.Subscription, error) {
	args := m.Called(ctx, token)
	if sub, ok := args.Get(0).(models.Subscription); ok {
		return sub, args.Error(1)
	}
	return models.Subscription{}, args.Error(1)
}

func (m *subscriptionRepoMock) GetConfirmed(ctx context.Context, frequency string) ([]models.Subscription, error) {
	args := m.Called(ctx, frequency)
	if subs, ok := args.Get(0).([]models.Subscription); ok {
		return subs, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *subscriptionRepoMock) Create(ctx context.Context, data models.Subscription) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *subscriptionRepoMock) Update(ctx context.Context, data models.Subscription) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *subscriptionRepoMock) Delete(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func TestCreateSubscription(main *testing.T) {
	main.Run("InvalidEmailEmpty", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		err := svc.Create(context.Background(), "", "TestCity", "daily")
		require.Equal(t, apierrors.ErrInvalidEmail, err)
	})

	main.Run("InvalidEmailFormat", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		err := svc.Create(context.Background(), "invalid", "TestCity", "daily")
		require.Equal(t, apierrors.ErrInvalidEmail, err)
	})

	main.Run("InvalidCity", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		err := svc.Create(context.Background(), "user@example.com", "", "daily")
		require.Equal(t, apierrors.ErrInvalidCity, err)
	})

	main.Run("InvalidFrequency", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		err := svc.Create(context.Background(), "user@example.com", "TestCity", "weekly")
		require.Equal(t, apierrors.ErrInvalidFrequency, err)
	})

	main.Run("AlreadySubscribed", func(t *testing.T) {
		ctx := context.Background()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// Return a non-zero subscription to simulate an already subscribed user.
		existingSub := models.Subscription{
			Email: "user@example.com",
		}
		repo.On("GetByEmail", ctx, "user@example.com").Return(existingSub, nil)

		err := svc.Create(ctx, "user@example.com", "TestCity", "daily")
		require.Equal(t, apierrors.ErrAlreadySubscribed, err)
		repo.AssertCalled(t, "GetByEmail", ctx, "user@example.com")
	})

	main.Run("CreateRepoError", func(t *testing.T) {
		ctx := context.Background()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// GetByEmail returns empty subscription.
		repo.On("GetByEmail", ctx, "user@example.com").Return(models.Subscription{}, nil)
		// Repo Create fails.
		repo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		err := svc.Create(ctx, "user@example.com", "TestCity", "daily")
		require.EqualError(t, err, "db error")
		repo.AssertCalled(t, "Create", ctx, mock.Anything)
	})

	main.Run("SendConfirmationEmailError", func(t *testing.T) {
		ctx := context.Background()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		repo.On("GetByEmail", ctx, "user@example.com").Return(models.Subscription{}, nil)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		// Mailer returns error.
		broker.On("Publish", "mailer.notifications", mock.Anything).Return(errors.New("send error"))

		err := svc.Create(ctx, "user@example.com", "TestCity", "daily")
		require.Equal(t, apierrors.ErrFailedSendConfirmEmail, err)
		repo.AssertCalled(t, "Create", ctx, mock.Anything)
		broker.AssertCalled(t, "Publish", "mailer.notifications", mock.Anything)
	})

	main.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		repo.On("GetByEmail", ctx, "user@example.com").Return(models.Subscription{}, nil)
		// Capture the subscription passed to Create.
		repo.On("Create", ctx, mock.AnythingOfType("models.Subscription")).Return(nil).Run(func(args mock.Arguments) {
			sub := args.Get(1).(models.Subscription)
			// Check that token is a valid UUID and CreatedAt is set.
			_, err := uuid.Parse(sub.Token)
			require.NoError(t, err)
			require.WithinDuration(t, time.Now(), sub.CreatedAt, time.Second)
		})
		broker.On("Publish", "mailer.notifications", mock.Anything).Return(nil)

		err := svc.Create(ctx, "user@example.com", "TestCity", "daily")
		require.NoError(t, err)
		repo.AssertCalled(t, "GetByEmail", ctx, "user@example.com")
		repo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("models.Subscription"))
		broker.AssertCalled(t, "Publish", "mailer.notifications", mock.Anything)
	})
}

func TestConfirmSubscription(main *testing.T) {
	main.Run("InvalidToken", func(t *testing.T) {
		ctx := context.Background()
		// Create dummies since mailer is unused here.
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		err := svc.Confirm(ctx, "invalid-token")
		require.Equal(t, apierrors.ErrInvalidToken, err)
	})

	main.Run("SubscriptionNotFound", func(t *testing.T) {
		ctx := context.Background()
		token := uuid.NewString()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// Simulate error when get by token is called.
		repo.On("GetByToken", ctx, token).Return(models.Subscription{}, errors.New("not found"))

		err := svc.Confirm(ctx, token)
		require.Equal(t, apierrors.ErrSubscriptionNotFound, err)
		repo.AssertCalled(t, "GetByToken", ctx, token)
	})

	main.Run("UpdateError", func(t *testing.T) {
		ctx := context.Background()
		token := uuid.NewString()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// Simulate successful get by token.
		subscription := models.Subscription{
			Token: token,
		}
		repo.On("GetByToken", ctx, token).Return(subscription, nil)
		// Simulate update failure.
		repo.On("Update", ctx, mock.AnythingOfType("models.Subscription")).Return(errors.New("update error"))

		err := svc.Confirm(ctx, token)
		require.EqualError(t, err, "update error")
		repo.AssertCalled(t, "GetByToken", ctx, token)
		repo.AssertCalled(t, "Update", ctx, mock.AnythingOfType("models.Subscription"))
	})

	main.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		token := uuid.NewString()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// Simulate successful get by token.
		subscription := models.Subscription{
			Token: token,
		}
		repo.On("GetByToken", ctx, token).Return(subscription, nil)
		// Capture the subscription passed to Update to check Confirm settings.
		repo.On("Update", ctx, mock.AnythingOfType("models.Subscription")).Return(nil).Run(func(args mock.Arguments) {
			updatedSub := args.Get(1).(models.Subscription)
			require.True(t, updatedSub.Confirmed)
			require.WithinDuration(t, time.Now(), updatedSub.ConfirmedAt, time.Second)
		})

		err := svc.Confirm(ctx, token)
		require.NoError(t, err)
		repo.AssertCalled(t, "GetByToken", ctx, token)
		repo.AssertCalled(t, "Update", ctx, mock.AnythingOfType("models.Subscription"))
	})
}

func TestDeleteSubscription(t *testing.T) {
	ctx := context.Background()
	validToken := uuid.NewString()
	invalidToken := "invalid-token"

	t.Run("InvalidToken", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		err := svc.Delete(ctx, invalidToken)
		require.Equal(t, apierrors.ErrInvalidToken, err)
	})

	t.Run("DeleteError", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)
		repo.On("Delete", ctx, validToken).Return(errors.New("delete error"))

		err := svc.Delete(ctx, validToken)
		require.EqualError(t, err, "delete error")
		repo.AssertCalled(t, "Delete", ctx, validToken)
	})

	t.Run("OK", func(t *testing.T) {
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)
		repo.On("Delete", ctx, validToken).Return(nil)

		err := svc.Delete(ctx, validToken)
		require.NoError(t, err)
		repo.AssertCalled(t, "Delete", ctx, validToken)
	})
}

func TestGetConfirmedSubscription(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ctx := context.Background()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// Simulate repository error
		repo.On("GetConfirmed", ctx, "daily").Return(nil, errors.New("db error"))

		subs, err := svc.GetConfirmed(ctx, "daily")
		require.Error(t, err)
		require.Nil(t, subs)
		repo.AssertCalled(t, "GetConfirmed", ctx, "daily")
	})

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		repo := &subscriptionRepoMock{}
		broker := &messageBrokerMock{}
		svc := New(repo, broker)

		// Prepare models.Subscription slice to be returned by repo.
		modelSubs := []models.Subscription{
			{
				ID:          1,
				Email:       "user1@example.com",
				City:        "TestCity",
				Frequency:   "daily",
				Confirmed:   true,
				Token:       "token-1",
				CreatedAt:   time.Now().Add(-time.Hour),
				ConfirmedAt: time.Now().Add(-30 * time.Minute),
			},
			{
				ID:          2,
				Email:       "user2@example.com",
				City:        "TestCity",
				Frequency:   "daily",
				Confirmed:   true,
				Token:       "token-2",
				CreatedAt:   time.Now().Add(-2 * time.Hour),
				ConfirmedAt: time.Now().Add(-90 * time.Minute),
			},
		}
		repo.On("GetConfirmed", ctx, "daily").Return(modelSubs, nil)

		contractsSubs, err := svc.GetConfirmed(ctx, "daily")
		require.NoError(t, err)
		require.Len(t, contractsSubs, len(modelSubs))

		// Verify conversion from models.Subscription to contracts.Subscription.
		for i, sub := range modelSubs {
			converted := contractsSubs[i]
			require.Equal(t, sub.ID, converted.ID)
			require.Equal(t, sub.Email, converted.Email)
			require.Equal(t, sub.City, converted.City)
			require.Equal(t, sub.Frequency, converted.Frequency)
			require.Equal(t, sub.Confirmed, converted.Confirmed)
			require.Equal(t, sub.Token, converted.Token)
			require.WithinDuration(t, sub.CreatedAt, converted.CreatedAt, time.Second)
			require.WithinDuration(t, sub.ConfirmedAt, converted.ConfirmedAt, time.Second)
		}
		repo.AssertCalled(t, "GetConfirmed", ctx, "daily")
	})
}
