package dbutil

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
	GetKeyData(keyid string) (*entities.Key, error)
	GetResources(pageSize int, offset int) ([]*entities.Resource, error)
	GetResourceIDs(pageSize int, offset int) ([]string, error)
	GetResourceData(resourceid string) (*entities.Resource, error)
	GetRoles(pageSize int, offset int) ([]string, error)

	GiveRole(keyid string, role ...string) error
	TakeRole(keyid string, role ...string) error
	GrantPermission(permission *entities.Permission, role ...string) error
	RevokePermission(permission *entities.Permission, role ...string) error
	SetRateLimit(key *entities.Key, limit *entities.RateLimit) error
	GetRateLimitData(ratelimitid string) (*entities.RateLimit, error)
	GetKeyRateLimitID(keyid string) (string, error)
	UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error
	DeleteRateLimit(rateLimitID string) error

	DeleteRole(name string) error
	DeleteKey(id string) error
	DeleteResource(id string) error
}
