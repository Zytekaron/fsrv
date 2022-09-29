package syncrl

import (
	"github.com/zytekaron/gorl"
	"sync"
)

type SyncSuite struct {
	managers map[string]*gorl.BucketManager
	mutexes  map[string]*sync.RWMutex
	mux      sync.RWMutex
}

func New() *SyncSuite {
	return &SyncSuite{
		managers: make(map[string]*gorl.BucketManager),
		mutexes:  make(map[string]*sync.RWMutex),
	}
}

func (s *SyncSuite) Get(id string) (*gorl.BucketManager, bool) {
	s.mux.RLock()
	manager, ok := s.managers[id]
	s.mux.RUnlock()
	return manager, ok
}

func (s *SyncSuite) Put(id string, manager *gorl.BucketManager) {
	s.mux.Lock()
	s.managers[id] = manager
	s.mux.Unlock()
}

func (s *SyncSuite) Delete(id string) {
	s.mux.Lock()
	delete(s.managers, id)
	s.mux.Unlock()
}

func (s *SyncSuite) PurgeAll() {
	s.mux.RLock()
	for _, sbm := range s.managers {
		sbm.Purge()
	}
	s.mux.RUnlock()
}
