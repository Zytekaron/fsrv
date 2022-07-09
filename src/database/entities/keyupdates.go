package entities

import (
	"fsrv/utils/serde"
)

type KeyUpdates struct {
	comment   *string
	addRoles  *[]string
	delRoles  *[]string
	expiresAt *serde.Time
}

func NewKeyUpdates() *KeyUpdates {
	return &KeyUpdates{}
}

func (u *KeyUpdates) WithComment(comment string) *KeyUpdates {
	*u.comment = comment
	return u
}

func (u *KeyUpdates) AddRoles(roles []string) *KeyUpdates {
	*u.addRoles = roles
	return u
}

func (u *KeyUpdates) RemoveRoles(roles []string) *KeyUpdates {
	*u.delRoles = roles
	return u
}

func (u *KeyUpdates) WithExpiresAt(expiresAt serde.Time) *KeyUpdates {
	*u.expiresAt = expiresAt
	return u
}
