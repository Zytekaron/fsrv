package entities

import (
	"fsrv/utils/serde"
)

type RateLimit struct {
	ID    string         `json:"id" toml:"id"`
	Limit int            `json:"limit" toml:"limit"`
	Reset serde.Duration `json:"reset" toml:"reset"`
}

func (p *RateLimit) GetID() string {
	return p.ID
}
