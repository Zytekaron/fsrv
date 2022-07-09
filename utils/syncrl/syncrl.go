package syncrl

import (
	"sync"
	"time"
)

type SyncSuite struct {
	managers map[string]*SyncManager
	mutexes  map[string]*sync.RWMutex
	mux      sync.RWMutex
}

func New() *SyncSuite {
	return &SyncSuite{
		managers: make(map[string]*SyncManager),
		mutexes:  make(map[string]*sync.RWMutex),
	}
}

func (s *SyncSuite) GetManager(id string) (*SyncManager, bool) {
	s.mux.RLock()
	manager, ok := s.managers[id]
	s.mux.RUnlock()
	return manager, ok
}

func (s *SyncSuite) AddManager(id string, manager *SyncManager) {
	s.mux.Lock()
	s.managers[id] = manager
	s.mux.Unlock()
}

func (s *SyncSuite) RemoveManager(id string) {
	s.mux.Lock()
	delete(s.managers, id)
	s.mux.Unlock()
}

// Purger is a daemon that purges old elements from each manager at a set interval.
func (s *SyncSuite) Purger(interval time.Duration) chan struct{} {
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

// Purge runs a single purge of old elements from each manager.
func (s *SyncSuite) Purge() {
	s.mux.RLock()
	for _, sbm := range s.managers {
		sbm.Purge()
	}
	s.mux.RUnlock()
}
