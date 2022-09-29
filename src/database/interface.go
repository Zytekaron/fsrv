package database

import (
	"fsrv/src/database/entities"
)

type DBInterface interface {
	CreateKey(key *entities.Key) error
	CreateResource(resource *entities.Resource) error
	CreateRole(role *entities.Role) error
	CreateRateLimit(limit *entities.RateLimit) error

	GetKeys(pageSize int, offset int) ([]*entities.Key, error)
	GetKeyIDs(pageSize int, offset int) ([]string, error)
	GetKeyData(keyID string) (*entities.Key, error)
	GetResources(pageSize int, offset int) ([]*entities.Resource, error)
	GetResourceIDs(pageSize int, offset int) ([]string, error)
	GetResourceData(resourceID string) (*entities.Resource, error)
	GetRoles(pageSize int, offset int) ([]string, error)

	GiveRole(keyID string, role ...string) error
	TakeRole(keyID string, role ...string) error
	GrantPermission(permission *entities.Permission, role ...string) error
	RevokePermission(permission *entities.Permission, role ...string) error
	SetRateLimit(key *entities.Key, limitID string) error
	GetRateLimitData(rateLimitID string) (*entities.RateLimit, error)
	GetKeyRateLimitID(keyID string) (string, error)
	UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error
	DeleteRateLimit(rateLimitID string) error

	DeleteRole(name string) error
	DeleteKey(id string) error
	DeleteResource(id string) error
}
