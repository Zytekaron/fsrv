package syncrl

import (
	"github.com/zytekaron/gotil/v2/rl"
	"sync"
	"time"
)

type SyncBucket struct {
	bucket *rl.Bucket
	mux    sync.Mutex
}

func NewBucket(limit int, reset time.Duration) *SyncBucket {
	return &SyncBucket{
		bucket: rl.NewBucket(limit, reset),
	}
}

func (s *SyncBucket) CanDraw() bool {
	return s.CanDrawN(1)
}

func (s *SyncBucket) CanDrawN(count int) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.bucket.CanDraw(count)
}

func (s *SyncBucket) Draw() bool {
	return s.DrawN(1)
}

func (s *SyncBucket) DrawN(n int) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.bucket.Draw(n)
}

func (s *SyncBucket) RemainingTime() int64 {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.bucket.RemainingTime()
}

func (s *SyncBucket) RemainingUses() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.bucket.RemainingUses()
}
