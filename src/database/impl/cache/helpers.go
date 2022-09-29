package cache

import "log"

type result[V any] struct {
	Val V
	Err error
}

type retrieveFunc[T any] func() (T, error)
type createFunc func() error

type taggedType interface {
	GetID() string
}

func retrieveData[T any](cache *mutexCache[string, result[T]], id string, retrieveFn retrieveFunc[T]) (T, error) {
	data, ok := cache.Get(id)
	if ok {
		return data.Val, data.Err
	}

	freshData, err := retrieveFn()
	// todo: check if the error should be kept before caching
	cache.Put(id, result[T]{freshData, err})
	return freshData, err
}

func createData[T taggedType](cache *mutexCache[string, result[T]], data T, createFn createFunc) error {
	err := createFn()
	if err != nil {
		return err
	}

	cache.Put(data.GetID(), result[T]{data, nil})
	return nil
}

func updateData[T taggedType](cache *mutexCache[string, result[T]], createFn createFunc, retrieveFn retrieveFunc[T]) error {
	err := createFn()
	if err != nil {
		return err
	}

	data, err := retrieveFn()
	if err != nil {
		log.Fatalf("ERROR: CACHE INCONSISTENCY DETETECTED: %e", err)
		return err
	}

	cache.Put(data.GetID(), result[T]{data, nil})
	return nil
}
