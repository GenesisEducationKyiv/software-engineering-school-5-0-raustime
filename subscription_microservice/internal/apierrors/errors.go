package apierrors

import "errors"

var (
	ErrCityNotFound           = errors.New("city not found")
	ErrAlreadySubscribed      = errors.New("email already subscribed")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrInvalidToken           = errors.New("invalid token")
	ErrFailedSendConfirmEmail = errors.New("failed to send confirmation email")
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidCity            = errors.New("invalid city")
	ErrInvalidFrequency       = errors.New("invalid frequency")
)
