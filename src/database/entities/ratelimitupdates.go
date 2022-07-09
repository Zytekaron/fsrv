package entities

import "time"

type RateLimitUpdates struct {
	limit *int
	reset *time.Duration
}

func NewRateLimitUpdates() *RateLimitUpdates {
	return &RateLimitUpdates{}
}

func (u *RateLimitUpdates) WithLimit(limit int) *RateLimitUpdates {
	*u.limit = limit
	return u
}

func (u *RateLimitUpdates) WithReset(reset time.Duration) *RateLimitUpdates {
	*u.reset = reset
	return u
}
