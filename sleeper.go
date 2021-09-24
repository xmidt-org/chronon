package chronon

import (
	"sync"
	"time"
)

// sleeper represents a simple listener that wakes up at a fixed time.
type sleeper struct {
	once   sync.Once
	awaken chan struct{}
	wakeup time.Time
}

// newSleeperAt creates a sleeper that awakens at the given absolute
// time.  The wakeup value must be in the future, relative to the enclosing
// FakeClock.
func newSleeperAt(wakeup time.Time) *sleeper {
	return &sleeper{
		awaken: make(chan struct{}),
		wakeup: wakeup,
	}
}

// onAdvance tests if this sleeper should awaken.  If it should, then
// the internal channel is signaled to allow any waiters to return
// immediately.
//
// This method guards against multiple triggers, as might be the case
// when concurrent FakeClock updates occur.
func (s *sleeper) onAdvance(newNow time.Time) bool {
	if equalOrAfter(newNow, s.wakeup) {
		s.once.Do(func() {
			close(s.awaken)
		})

		return true
	}

	return false
}

// wait blocks until this sleeper awakens.
func (s *sleeper) wait() {
	<-s.awaken
}
