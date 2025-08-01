package apierrors

import "errors"

var (
	ErrCityNotFound           = errors.New("city not found")
	ErrNoWeatherDataFound     = errors.New("no weather data found")
	ErrInvalidAPIKey          = errors.New("invalid API key")
	ErrAlreadySubscribed      = errors.New("email already subscribed")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrInvalidToken           = errors.New("invalid token")
	ErrFailedSendConfirmEmail = errors.New("failed to send confirmation email")
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidCity            = errors.New("invalid city")
	ErrInvalidFrequency       = errors.New("invalid frequency")

	// Cache-related errors.
	ErrCacheMiss        = errors.New("cache miss")
	ErrCacheConnection  = errors.New("cache connection error")
	ErrCacheTimeout     = errors.New("cache operation timeout")
	ErrCacheUnavailable = errors.New("cache unavailable")
	ErrCacheCorrupted   = errors.New("cached data corrupted")
)
