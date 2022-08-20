package entities

import (
	"fsrv/utils/serde"
	"time"
)

// Key represents an access key used to authenticate against a Resource.
type Key struct {
	// ID is the id of the key.
	ID string `json:"id"`
	// Comment is used to note the owner or usage of a key.
	Comment string `json:"comment"`
	// Roles are the roles this token has.
	Roles []string `json:"roles"`

	// RateLimitID is the rate limit level of this token
	RateLimitID string `json:"rate_limit_id"`

	// ExpiresAt is the time when this key expires.
	ExpiresAt serde.Time `json:"expires_at"`
	// CreatedAt is the time when this key was created.
	CreatedAt serde.Time `json:"created_at"`
}

func (k *Key) IsExpired() bool {
	expiry := time.Time(k.ExpiresAt)
	if expiry.UnixNano() == 0 {
		return false
	}
	return time.Since(expiry).Nanoseconds() > 0
}

func (k *Key) GetID() string {
	return k.ID
}
