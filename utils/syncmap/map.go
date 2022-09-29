package syncmap

import (
	"golang.org/x/exp/constraints"
	"sync"
)

type CountMap[K comparable, V constraints.Integer | constraints.Float] struct {
	data map[K]V
	mux  sync.RWMutex
}

func New[K comparable, V constraints.Integer | constraints.Float]() *CountMap[K, V] {
	return &CountMap[K, V]{
		data: make(map[K]V),
	}
}

func (s *CountMap[K, V]) Get(id K) (V, bool) {
	s.mux.RLock()
	val, ok := s.data[id]
	s.mux.RUnlock()
	return val, ok
}

func (s *CountMap[K, V]) GetOrZero(id K) V {
	s.mux.RLock()
	val, ok := s.data[id]
	s.mux.RUnlock()
	if !ok {
		var null V
		return null
	}
	return val
}

func (s *CountMap[K, V]) Put(id K, value V) {
	s.mux.Lock()
	s.data[id] = value
	s.mux.Unlock()
}

func (s *CountMap[K, V]) Delete(id K) {
	s.mux.Lock()
	delete(s.data, id)
	s.mux.Unlock()
}

// CompareLessAndIncrement returns whether the initial value is
// less than the provided one, and increments the value if so.
func (s *CountMap[K, V]) CompareLessAndIncrement(id K, value V) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.data[id] >= value {
		return false
	}

	s.data[id]++
	return true
}

// Increment increments the value and returns the updated value
func (s *CountMap[K, V]) Increment(id K) V {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.data[id]++
	return s.data[id]
}

// Decrement decrements the value and returns the updated value
func (s *CountMap[K, V]) Decrement(id K) V {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.data[id]--
	return s.data[id]
}
