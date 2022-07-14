package syncrl

import (
	"github.com/zytekaron/gotil/v2/rl"
	"sync"
)

type SyncSuite struct {
	managers map[string]*rl.SyncBucketManager
	mutexes  map[string]*sync.RWMutex
	mux      sync.RWMutex
}

func New() *SyncSuite {
	return &SyncSuite{
		managers: make(map[string]*rl.SyncBucketManager),
		mutexes:  make(map[string]*sync.RWMutex),
	}
}

func (s *SyncSuite) Get(id string) (*rl.SyncBucketManager, bool) {
	s.mux.RLock()
	manager, ok := s.managers[id]
	s.mux.RUnlock()
	return manager, ok
}

func (s *SyncSuite) Put(id string, manager *rl.SyncBucketManager) {
	s.mux.Lock()
	s.managers[id] = manager
	s.mux.Unlock()
}

func (s *SyncSuite) Delete(id string) {
	s.mux.Lock()
	delete(s.managers, id)
	s.mux.Unlock()
}

func (s *SyncSuite) Purge() {
	s.mux.RLock()
	for _, sbm := range s.managers {
		sbm.Purge()
	}
	s.mux.RUnlock()
}
