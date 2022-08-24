package cache

import (
	"fsrv/src/database/dbutil"
	"fsrv/src/database/entities"
	"github.com/zyedidia/generic/cache"
	"log"
)

type tAndErr[T any] struct {
	V   T
	Err error
}

type CacheDB struct {
	db               dbutil.DBInterface
	resourceCache    *cache.Cache[string, tAndErr[*entities.Resource]]
	keyCache         *cache.Cache[string, tAndErr[*entities.Key]]
	roleCache        *cache.Cache[string, tAndErr[*entities.Role]]
	rateLimitCache   *cache.Cache[string, tAndErr[*entities.RateLimit]]
	rateLimitIDCache *cache.Cache[string, tAndErr[string]]
	tokenCache       *cache.Cache[string, tAndErr[*entities.Token]] //todo: build out infrastructure for tokens
}

type retrieveFunc[T any] func() (T, error)
type createFunc func() error

type taggedType interface {
	GetID() string
}

func retrieveData[T any](cache *cache.Cache[string, tAndErr[T]], key string, retrieveFn retrieveFunc[T]) (T, error) {
	data, ok := cache.Get(key)
	if ok {
		return data.V, data.Err
	}

	freshData, err := retrieveFn()
	cache.Put(key, tAndErr[T]{freshData, err})
	return freshData, err
}

func createData[T taggedType](cache *cache.Cache[string, tAndErr[T]], data T, createFn createFunc) error {
	err := createFn()
	if err != nil {
		return err
	}
	cache.Put(data.GetID(), tAndErr[T]{data, nil})
	return nil
}

func updateData[T taggedType](cache *cache.Cache[string, tAndErr[T]], createFn createFunc, retrieveFn retrieveFunc[T]) error {
	err := createFn()
	if err != nil {
		return err
	}
	data, err := retrieveFn()
	if err != nil {
		log.Fatalf("ERROR: CACHE INCONSISTENCY DETETECTED: %e", err)
		return err
	}
	cache.Put(data.GetID(), tAndErr[T]{data, nil})
	return nil
}

func NewCache(db dbutil.DBInterface) *CacheDB {
	var cacheDB CacheDB
	cacheDB.db = db
	cacheDB.resourceCache = cache.New[string, tAndErr[*entities.Resource]](100)
	cacheDB.keyCache = cache.New[string, tAndErr[*entities.Key]](100)
	cacheDB.roleCache = cache.New[string, tAndErr[*entities.Role]](20)
	cacheDB.rateLimitCache = cache.New[string, tAndErr[*entities.RateLimit]](500)
	cacheDB.rateLimitIDCache = cache.New[string, tAndErr[string]](50)
	cacheDB.tokenCache = cache.New[string, tAndErr[*entities.Token]](25)
	return &cacheDB
}

func (c *CacheDB) CreateKey(key *entities.Key) error {
	err := c.db.CreateKey(key)
	if err != nil {
		return err
	}
	c.keyCache.Put(key.ID, tAndErr[*entities.Key]{key, nil})
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
		c.keyCache.Put(keyID, tAndErr[*entities.Key]{key, err})
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

func (c *CacheDB) GiveRole(keyID string, role ...string) error {
	return updateData[*entities.Key](c.keyCache, func() error {
		return c.db.GiveRole(keyID, role...)
	}, func() (*entities.Key, error) {
		return c.db.GetKeyData(keyID)
	})
}

func (c *CacheDB) TakeRole(keyID string, role ...string) error {
	return updateData[*entities.Key](c.keyCache, func() error {
		return c.db.TakeRole(keyID, role...)
	}, func() (*entities.Key, error) {
		return c.db.GetKeyData(keyID)
	})
}

func (c *CacheDB) GrantPermission(permission *entities.Permission, roles ...string) error {
	err := c.db.GrantPermission(permission, roles...)
	if err != nil {
		return err
	}
	res, err := c.GetResourceData(permission.ResourceID)
	if err != nil {
		return err
	}

	for _, role := range roles {
		resOpAccess := entities.ResourceOperationAccess{
			ID:   role,
			Type: permission.TypeRWMD,
		}
		res.OperationNodes[resOpAccess] = permission.Status
	}

	c.resourceCache.Put(permission.ResourceID, tAndErr[*entities.Resource]{res, nil})

	return err
}

func (c *CacheDB) RevokePermission(permission *entities.Permission, roles ...string) error {
	err := c.db.RevokePermission(permission, roles...)
	if err != nil {
		return err
	}
	res, err := c.GetResourceData(permission.ResourceID)
	if err != nil {
		return err
	}
	for _, role := range roles {
		resOpAccess := entities.ResourceOperationAccess{
			ID:   role,
			Type: permission.TypeRWMD,
		}
		delete(res.OperationNodes, resOpAccess)
	}

	c.resourceCache.Put(permission.ResourceID, tAndErr[*entities.Resource]{res, nil})

	return err
}

// SetRateLimit
// NOTE: mutates underlying key to use given limitID
func (c *CacheDB) SetRateLimit(key *entities.Key, limitID string) error {
	return createData[*entities.Key](c.keyCache, key, func() error {
		err := c.db.SetRateLimit(key, limitID) //attempt to set in database
		if err != nil {
			key.RateLimitID = limitID //set in key object so that cache change is written by createData (see impl)
		}
		return err
	})
}

func (c *CacheDB) GetRateLimitData(rateLimitID string) (*entities.RateLimit, error) {
	return retrieveData[*entities.RateLimit](c.rateLimitCache, rateLimitID, func() (*entities.RateLimit, error) {
		return c.db.GetRateLimitData(rateLimitID)
	})
}

func (c *CacheDB) GetKeyRateLimitID(keyID string) (string, error) {
	return retrieveData[string](c.rateLimitIDCache, keyID, func() (string, error) {
		return c.db.GetKeyRateLimitID(keyID)
	})
}

func (c *CacheDB) UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error {
	return createData[*entities.RateLimit](c.rateLimitCache, rateLimit, func() error {
		return c.db.UpdateRateLimit(rateLimitID, rateLimit)
	})
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