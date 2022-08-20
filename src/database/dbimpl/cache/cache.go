package cache

import (
	"fsrv/src/database/dbutil"
	"fsrv/src/database/entities"
	"github.com/zyedidia/generic/cache"
)

type pair[T any, U any] struct {
	V   T
	Err U
}

type CacheDB struct {
	db               dbutil.DBInterface
	resourceCache    cache.Cache[string, pair[*entities.Resource, error]]
	keyCache         cache.Cache[string, pair[*entities.Key, error]]
	roleCache        cache.Cache[string, pair[*entities.Role, error]]
	rateLimitCache   cache.Cache[string, pair[*entities.RateLimit, error]]
	rateLimitIDCache cache.Cache[string, pair[string, error]]
	tokenCache       cache.Cache[string, pair[*entities.Token, error]] //todo: build out infrastructure for tokens
}

type RetrieveFunc[T any] func() (T, error)

func retrieveData[T any](cache cache.Cache[string, pair[T, error]], key string, retrieveFn RetrieveFunc[T]) (T, error) {
	data, ok := cache.Get(key)
	if ok {
		return data.V, data.Err
	}

	freshData, err := retrieveFn()
	cache.Put(key, pair[T, error]{freshData, err})
	return freshData, err
}

type createFunc func() error

type taggedType interface {
	GetID() string
}

func createData[T taggedType](cache cache.Cache[string, pair[T, error]], data T, createFn createFunc) error {
	err := createFn()
	if err != nil {
		return err
	}
	cache.Put(data.GetID(), pair[T, error]{data, nil})
	return nil
}

func (c *CacheDB) NewCache(db dbutil.DBInterface) {
	c.db = db
}

func (c *CacheDB) CreateKey(key *entities.Key) error {
	err := c.db.CreateKey(key)
	if err != nil {
		return err
	}
	c.keyCache.Put(key.ID, pair[*entities.Key, error]{key, nil})
	return nil
}

func (c *CacheDB) CreateResource(resource *entities.Resource) error {
	return createData(c.resourceCache, resource, func() error { return c.db.CreateResource(resource) })
}

func (c *CacheDB) CreateRole(role *entities.Role) error {
	return c.db.CreateRole(role)
}

func (c *CacheDB) CreateRateLimit(limit *entities.RateLimit) error {
	return createData(c.rateLimitCache, limit, func() error { return c.db.CreateRateLimit(limit) })
}

func (c *CacheDB) GetKeys(pageSize int, offset int) ([]*entities.Key, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GetKeyIDs(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GetKeyData(keyID string) (*entities.Key, error) {
	key, ok := c.keyCache.Get(keyID)
	if ok {
		return key.V, key.Err
	} else {
		key, err := c.db.GetKeyData(keyID)
		c.keyCache.Put(keyID, pair[*entities.Key, error]{key, err})
		return key, err
	}
}

func (c *CacheDB) GetResources(pageSize int, offset int) ([]*entities.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GetResourceIDs(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GetResourceData(resourceID string) (*entities.Resource, error) {
	return retrieveData[*entities.Resource](c.resourceCache, resourceID, func() (*entities.Resource, error) {
		data, err := c.db.GetResourceData(resourceID)
		return data, err
	})
}

func (c *CacheDB) GetRoles(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GiveRole(keyid string, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) TakeRole(keyid string, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GrantPermission(permission *entities.Permission, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) RevokePermission(permission *entities.Permission, role ...string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) SetRateLimit(key *entities.Key, limit *entities.RateLimit) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GetRateLimitData(ratelimitid string) (*entities.RateLimit, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GetKeyRateLimitID(keyID string) (string, error) {
	return retrieveData[string](c.rateLimitIDCache, keyID, func() (string, error) {
		data, err := c.db.GetKeyRateLimitID(keyID)
		return data, err
	})
}

func (c *CacheDB) UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) DeleteRateLimit(rateLimitID string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) DeleteRole(name string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) DeleteKey(id string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) DeleteResource(id string) error {
	//TODO implement me
	panic("implement me")
}
