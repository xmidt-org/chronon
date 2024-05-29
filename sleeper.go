// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"sync"
	"time"
)

// Sleeper represents a handle to a goroutine that is blocked in a fake
// clock's Sleep method.  Test code can get access to a Sleeper by calling
// NotifyOnSleep.
type Sleeper interface {
	// When returns the point in time at which the sleeping goroutine will
	// wake up.  This value will be the duration passed to Sleep added to
	// the containing FakeClock's Now().
	When() time.Time

	// Wakeup forces the sleeping goroutine to awaken.  Neither the fake clock's
	// notion of the current time nor this Sleeper's When time are affected
	// by this method.  This method is idempotent:  invoking it more than once
	// has no further effect.
	//
	// If the fake clock's time is important to update as a result of sleeping,
	// use FakeClock.Set with the value of When.
	Wakeup() bool
}

// sleeper is the internal Sleeper implementation.
type sleeper struct {
	fc *FakeClock

	once   sync.Once
	awaken chan struct{}
	when   time.Time
}

func (s *sleeper) When() time.Time {
	return s.when
}

func (s *sleeper) Wakeup() (awakened bool) {
	// keep the same order of acquiring locks
	// as in onUpdate
	s.fc.doWith(func(_ time.Time, ls *listeners) {
		s.once.Do(func() {
			awakened = true
			close(s.awaken)
		})

		ls.remove(s)
	})

	return
}

// newSleeperAt creates a sleeper that awakens at the given absolute
// time.  The wakeup value must be in the future, relative to the enclosing
// FakeClock.
func newSleeperAt(fc *FakeClock, when time.Time) *sleeper {
	return &sleeper{
		fc:     fc,
		awaken: make(chan struct{}),
		when:   when,
	}
}

// onUpdate tests if this sleeper should awaken.  If it should, then
// the internal channel is signaled to allow any waiters to return
// immediately.
//
// This method guards against multiple triggers, as might be the case
// when concurrent FakeClock updates occur.
func (s *sleeper) onUpdate(newNow time.Time) updateResult {
	if equalOrAfter(newNow, s.when) {
		s.once.Do(func() {
			close(s.awaken)
		})

		return stopUpdates
	}

	return continueUpdates
}

// wait blocks until this sleeper awakens.
func (s *sleeper) wait() {
	<-s.awaken
}
