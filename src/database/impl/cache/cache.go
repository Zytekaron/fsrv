package cache

import (
	"fsrv/src/database"
	"fsrv/src/database/entities"
	"github.com/zyedidia/generic/cache"
)

type CacheDB struct {
	db               database.DBInterface
	resourceCache    *mutexCache[string, result[*entities.Resource]]
	keyCache         *mutexCache[string, result[*entities.Key]]
	roleCache        *mutexCache[string, result[*entities.Role]]
	rateLimitCache   *mutexCache[string, result[*entities.RateLimit]]
	rateLimitIDCache *mutexCache[string, result[string]]
	tokenCache       *mutexCache[string, result[*entities.Token]] // todo: build out infrastructure for tokens
}

func NewCache(db database.DBInterface) *CacheDB {
	return &CacheDB{
		db:               db,
		resourceCache:    newMutexCache(cache.New[string, result[*entities.Resource]](200)),
		keyCache:         newMutexCache(cache.New[string, result[*entities.Key]](1000)),
		roleCache:        newMutexCache(cache.New[string, result[*entities.Role]](25)),
		rateLimitCache:   newMutexCache(cache.New[string, result[*entities.RateLimit]](500)),
		rateLimitIDCache: newMutexCache(cache.New[string, result[string]](50)),
		tokenCache:       newMutexCache(cache.New[string, result[*entities.Token]](25)),
	}
}

func (c *CacheDB) CreateKey(key *entities.Key) error {
	err := c.db.CreateKey(key)
	if err != nil {
		return err
	}
	c.keyCache.Put(key.ID, result[*entities.Key]{key, err})
	return nil
}

func (c *CacheDB) CreateResource(resource *entities.Resource) error {
	return createData(c.resourceCache, resource, func() error {
		return c.db.CreateResource(resource)
	})
}

func (c *CacheDB) CreateRole(role *entities.Role) error {
	return c.db.CreateRole(role)
}

func (c *CacheDB) CreateRateLimit(limit *entities.RateLimit) error {
	return createData(c.rateLimitCache, limit, func() error {
		return c.db.CreateRateLimit(limit)
	})
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
	res, ok := c.keyCache.Get(keyID)
	if ok {
		return res.Val, res.Err
	}

	key, err := c.db.GetKeyData(keyID)
	// todo: check if the error should be kept before caching
	c.keyCache.Put(keyID, result[*entities.Key]{key, err})
	return key, err
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

	// todo: check if the error should be kept before caching
	c.resourceCache.Put(permission.ResourceID, result[*entities.Resource]{res, nil})
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

	// todo: check if the error should be kept before caching
	c.resourceCache.Put(permission.ResourceID, result[*entities.Resource]{res, nil})
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
