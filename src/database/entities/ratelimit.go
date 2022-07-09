package entities

import (
	"fsrv/utils/serde"
)

type RateLimit struct {
	ID    string         `json:"id"`
	Limit int            `json:"limit"`
	Reset serde.Duration `json:"reset"`
}
