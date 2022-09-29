package entities

import (
	"fsrv/utils/serde"
)

type RateLimit struct {
	ID     string         `json:"id" toml:"id"`
	Limit  int64          `json:"limit" toml:"limit"`
	Burst  int64          `json:"burst" toml:"burst"`
	Refill serde.Duration `json:"refill" toml:"refill"`
}

func (p *RateLimit) GetID() string {
	return p.ID
}
