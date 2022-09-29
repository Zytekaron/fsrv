package cache

import (
	"fsrv/src/database"
	"fsrv/src/database/entities"
	"github.com/zyedidia/generic/cache"
	"log"
	"sync"
)

type tErr[T any] struct {
	V   T
	Err error
}

type cacheMu[K comparable, V any] struct {
	C    *cache.Cache[K, V]
	RWMu sync.RWMutex
}

type CacheDB struct {
	db               database.DBInterface
	resourceCache    cacheMu[string, tErr[*entities.Resource]]
	keyCache         cacheMu[string, tErr[*entities.Key]]
	roleCache        cacheMu[string, tErr[*entities.Role]]
	rateLimitCache   cacheMu[string, tErr[*entities.RateLimit]]
	rateLimitIDCache cacheMu[string, tErr[string]]
	tokenCache       cacheMu[string, tErr[*entities.Token]] //todo: build out infrastructure for tokens
}

type retrieveFunc[T any] func() (T, error)
type createFunc func() error

type taggedType interface {
	GetID() string
}

func retrieveData[T any](cache *cache.Cache[string, tErr[T]], key string, mu *sync.RWMutex, retrieveFn retrieveFunc[T]) (T, error) {
	mu.RLock()
	data, ok := cache.Get(key)
	if ok {
		mu.RUnlock()
		return data.V, data.Err
	}
	mu.RUnlock()
	mu.Lock()
	freshData, err := retrieveFn()
	cache.Put(key, tErr[T]{freshData, err})
	mu.Unlock()
	return freshData, err
}

func createData[T taggedType](cache *cache.Cache[string, tErr[T]], data T, mu *sync.RWMutex, createFn createFunc) error {
	err := createFn()
	if err != nil {
		return err
	}
	mu.Lock()
	cache.Put(data.GetID(), tErr[T]{data, nil})
	mu.Unlock()
	return nil
}

func updateData[T taggedType](cache *cache.Cache[string, tErr[T]], mu *sync.RWMutex, createFn createFunc, retrieveFn retrieveFunc[T]) error {
	err := createFn()
	if err != nil {
		return err
	}
	data, err := retrieveFn()
	if err != nil {
		log.Fatalf("ERROR: CACHE INCONSISTENCY DETETECTED: %e", err)
		return err
	}
	mu.Lock()
	cache.Put(data.GetID(), tErr[T]{data, nil})
	mu.Unlock()
	return nil
}

func newCacheMu[K comparable, V any](cash *cache.Cache[K, V]) cacheMu[K, V] {
	return cacheMu[K, V]{
		C: cash,
	}
}

func NewCache(db database.DBInterface) *CacheDB {
	return &CacheDB{
		db:               db,
		resourceCache:    newCacheMu(cache.New[string, tErr[*entities.Resource]](200)),
		keyCache:         newCacheMu(cache.New[string, tErr[*entities.Key]](1000)),
		roleCache:        newCacheMu(cache.New[string, tErr[*entities.Role]](25)),
		rateLimitCache:   newCacheMu(cache.New[string, tErr[*entities.RateLimit]](500)),
		rateLimitIDCache: newCacheMu(cache.New[string, tErr[string]](50)),
		tokenCache:       newCacheMu(cache.New[string, tErr[*entities.Token]](25)),
	}
}

func (c *CacheDB) CreateKey(key *entities.Key) error {
	err := c.db.CreateKey(key)
	if err != nil {
		return err
	}
	c.keyCache.C.Put(key.ID, tErr[*entities.Key]{key, err})
	return nil
}

func (c *CacheDB) CreateResource(resource *entities.Resource) error {
	return createData(c.resourceCache.C, resource, &c.resourceCache.RWMu, func() error { return c.db.CreateResource(resource) })
}

func (c *CacheDB) CreateRole(role *entities.Role) error {
	return c.db.CreateRole(role)
}

func (c *CacheDB) CreateRateLimit(limit *entities.RateLimit) error {
	return createData(c.rateLimitCache.C, limit, &c.rateLimitCache.RWMu, func() error { return c.db.CreateRateLimit(limit) })
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
	c.keyCache.RWMu.RLock()
	key, ok := c.keyCache.C.Get(keyID)
	if ok {
		c.keyCache.RWMu.RLock()
		return key.V, key.Err
	} else {
		c.keyCache.RWMu.Lock()
		key, err := c.db.GetKeyData(keyID)
		c.keyCache.C.Put(keyID, tErr[*entities.Key]{key, err})
		c.keyCache.RWMu.RLock()
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
	return retrieveData[*entities.Resource](c.resourceCache.C, resourceID, &c.resourceCache.RWMu, func() (*entities.Resource, error) {
		data, err := c.db.GetResourceData(resourceID)
		return data, err
	})
}

func (c *CacheDB) GetRoles(pageSize int, offset int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheDB) GiveRole(keyID string, role ...string) error {
	return updateData[*entities.Key](c.keyCache.C, &c.keyCache.RWMu, func() error {
		return c.db.GiveRole(keyID, role...)
	}, func() (*entities.Key, error) {
		return c.db.GetKeyData(keyID)
	})
}

func (c *CacheDB) TakeRole(keyID string, role ...string) error {
	return updateData[*entities.Key](c.keyCache.C, &c.keyCache.RWMu, func() error {
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

	c.resourceCache.C.Put(permission.ResourceID, tErr[*entities.Resource]{res, nil})

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

	c.resourceCache.C.Put(permission.ResourceID, tErr[*entities.Resource]{res, nil})

	return err
}

// SetRateLimit
// NOTE: mutates underlying key to use given limitID
func (c *CacheDB) SetRateLimit(key *entities.Key, limitID string) error {
	return createData[*entities.Key](c.keyCache.C, key, &c.keyCache.RWMu, func() error {
		err := c.db.SetRateLimit(key, limitID) //attempt to set in database
		if err != nil {
			key.RateLimitID = limitID //set in key object so that cache change is written by createData (see impl)
		}
		return err
	})
}

func (c *CacheDB) GetRateLimitData(rateLimitID string) (*entities.RateLimit, error) {
	return retrieveData[*entities.RateLimit](c.rateLimitCache.C, rateLimitID, &c.rateLimitCache.RWMu, func() (*entities.RateLimit, error) {
		return c.db.GetRateLimitData(rateLimitID)
	})
}

func (c *CacheDB) GetKeyRateLimitID(keyID string) (string, error) {
	return retrieveData[string](c.rateLimitIDCache.C, keyID, &c.rateLimitIDCache.RWMu, func() (string, error) {
		return c.db.GetKeyRateLimitID(keyID)

	})
}

func (c *CacheDB) UpdateRateLimit(rateLimitID string, rateLimit *entities.RateLimit) error {
	return createData[*entities.RateLimit](c.rateLimitCache.C, rateLimit, &c.rateLimitCache.RWMu, func() error {
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
