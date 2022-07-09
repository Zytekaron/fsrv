package syncrl

import (
	"sync"
	"time"
)

type SyncManager struct {
	limit   int
	reset   time.Duration
	buckets map[string]*SyncBucket
	mux     sync.RWMutex
}

func NewManager(limit int, reset time.Duration) *SyncManager {
	return &SyncManager{
		limit:   limit,
		reset:   reset,
		buckets: make(map[string]*SyncBucket),
	}
}

func (s *SyncManager) GetBucket(id string) *SyncBucket {
	s.mux.RLock()
	defer s.mux.RUnlock()
	bucket, ok := s.buckets[id]
	if !ok {
		bucket = NewBucket(s.limit, s.reset)
		s.buckets[id] = bucket
	}
	return bucket
}

// Purger is a daemon that purges old elements at a set interval.
func (s *SyncManager) Purger(interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.Purge()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
	return stop
}

// Purge runs a single purge of old elements.
func (s *SyncManager) Purge() {
	s.mux.RLock()
	for id, bucket := range s.buckets {
		if bucket.RemainingTime() == 0 {
			// briefly move to write lock
			s.mux.RUnlock()
			s.mux.Lock()
			delete(s.buckets, id)
			// return to read lock
			s.mux.Unlock()
			s.mux.RLock()
		}
	}
	s.mux.RUnlock()
}
