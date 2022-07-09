package entities

import (
	"fsrv/utils/serde"
	"time"
)

// Key represents an access key used to authenticate against a Permission.
type Key struct {
	// ID is the id of the key.
	ID string `json:"id"`
	// Comment is used to note the owner or usage of a token.
	Comment string `json:"comment"`
	// RateLimit is the rate limit level of this token
	RateLimit string `json:"rate_limit"`
	// Roles are the roles this token has.
	Roles []string `json:"roles"`
	// ExpiresAt is the time when this token expires.
	ExpiresAt serde.Time `json:"expires_at"`
	// CreatedAt is the time when this token was created.
	CreatedAt serde.Time `json:"created_at"`
}

func (k *Key) IsExpired() bool {
	expiry := time.Time(k.ExpiresAt)
	if expiry.UnixNano() == 0 {
		return false
	}
	return time.Since(expiry).Nanoseconds() > 0
}
