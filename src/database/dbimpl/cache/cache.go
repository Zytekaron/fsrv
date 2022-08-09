package cache

import (
	"fsrv/src/database/dbutil"
	"fsrv/src/database/entities"
	"github.com/zyedidia/generic/cache"
)

type CacheDB struct {
	db             dbutil.DBInterface
	resourceCache  cache.Cache[string, entities.Resource]
	keyCache       cache.Cache[string, entities.Resource]
	roleCache      cache.Cache[string, entities.Role]
	rateLimitCache cache.Cache[string, entities.RateLimit]
	tokenCache     cache.Cache[string, entities.Token] //todo: build out infrastructure for tokens
}

func (c CacheDB) CreateKey(key *entities.Key) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) CreateResource(resource *entities.Resource) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) CreateRole(role *entities.Role) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) CreateRateLimit(limit *entities.RateLimit) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetKeys(pageSize int, offset int) ([]*entities.Key, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetKeyIDs(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetKeyData(keyid string) (*entities.Key, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetResources(pageSize int, offset int) ([]*entities.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetResourceIDs(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetResourceData(resourceid string) (*entities.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GetRoles(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GiveRole(keyid string, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) TakeRole(keyid string, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) GrantPermission(permission *entities.Permission, role ...string) []error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) RevokePermission(permission *entities.Permission, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) SetRateLimit(key *entities.Key, limit *entities.RateLimit) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) DeleteRole(name string) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) DeleteKey(id string) error {
	//TODO implement me
	panic("implement me")
}

func (c CacheDB) DeleteResource(id string) error {
	//TODO implement me
	panic("implement me")
}
