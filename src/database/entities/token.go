package entities

import (
	"fsrv/utils/serde"
	"time"
)

// Token represents an administrator API token used to authenticate for privileged actions.
type Token struct {
	// ID is the id of the token.
	ID string `json:"id"`

	Comment string `json:"comment"`

	// ExpiresAt is the time when this token expires.
	ExpiresAt serde.Time `json:"expires_at"`
	// CreatedAt is the time when this token was created.
	CreatedAt serde.Time `json:"created_at"`
}

func (t *Token) IsExpired() bool {
	expiry := time.Time(t.ExpiresAt)
	if expiry.UnixNano() == 0 {
		return false
	}
	return time.Since(expiry).Nanoseconds() > 0
}
