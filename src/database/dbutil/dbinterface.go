package dbutil

import (
	"errors"
	"fsrv/src/database/entities"
)

var (
	ErrCreateFailed  = errors.New("database creation failed")
	ErrCheckFailed   = errors.New("database integrity check failed")
	ErrDestroyFailed = errors.New("database object wipe failed")

	ErrKeyDuplicate      = errors.New("A key with the given ID already exists")
	ErrRoleDuplicate     = errors.New("A role with the given ID already exists")
	ErrResourceDuplicate = errors.New("A resource with the given ID already exists")

	ErrKeyMissing      = errors.New("the specified key does not exist")
	ErrRoleMissing     = errors.New("the specified role does not exist")
	ErrResourceMissing = errors.New("the specified resource does not exist")

	ErrRoleNameBad     = errors.New("the given role name is not allowed")
	ErrKeyNameBad      = errors.New("the given key name is not allowed")
	ErrResourceNameBad = errors.New("the given resource name is not allowed")
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
