package chronon

import (
	"time"
)

// FakeTimer is a Timer which can be manually controlled by either
// advancing its containing FakeClock or by methods of this interface.
type FakeTimer interface {
	Timer

	// When returns the absolute time at which this timer fires its
	// event.  This value will be affected by Reset.
	When() time.Time

	// Fire forces this FakeTimer to fire its event.  Neither the containing
	// fake clock nor the value returned by When are affected by this method.
	//
	// This method returns true if this timer had been active, false if
	// the timer had already fired.
	Fire() bool
}

// fakeTimer is a Timer which can be manually controlled.  This type
// preserves the odd Reset/Stop behavior of the time package.
type fakeTimer struct {
	fc *FakeClock

	c chan time.Time
	f func(time.Time)

	when  time.Time
	fired bool // whether a time has been sent on c since creation or the last reset
}

// newfakeTimer contructs a fakeTimer with a given FakeClock container and
// the given wakeup time.
func newFakeTimer(fc *FakeClock, when time.Time) *fakeTimer {
	return &fakeTimer{
		fc:   fc,
		c:    make(chan time.Time, 1),
		when: when,
	}
}

// newAfterFunc constructs a fakeTimer appropriate for time.AfterFunc style invocation
// with a FakeClock.
func newAfterFunc(fc *FakeClock, when time.Time, f func(time.Time)) *fakeTimer {
	return &fakeTimer{
		fc:   fc,
		f:    f,
		when: when,
	}
}

// fire handles dispatching the time event appropriately.  Depending
// upon how this timer was created, this will be either sending the
// time on a channel or invoking an arbitrary function.
func (ft *fakeTimer) fire(t time.Time) {
	if ft.c != nil {
		sendTime(ft.c, t)
	} else {
		ft.f(t)
	}
}

// onUpdate processes what should happen if the current fake time is set to a new value.
// If this timer was previously triggered or if the new value should triggered it, this
// method returns true which indicates that the containing FakeClock should remove it
// from its callbacks.  Otherwise, this method returns false.
func (ft *fakeTimer) onUpdate(newNow time.Time) updateResult {
	if ft.fired {
		return stopUpdates // previously triggered
	}

	if equalOrAfter(newNow, ft.when) {
		ft.fired = true
		ft.fire(newNow)
		return stopUpdates
	}

	return continueUpdates
}

// C returns the channel on which this fakeTimer sends its time events.  This
// channel is never closed and will be the same channel instance for the life
// of this fakeTimer.
//
// If this fakeTimer was created via AfterFunc, this method returns nil.  This
// is consistent with time.AfterFunc.
func (ft *fakeTimer) C() <-chan time.Time {
	return ft.c
}

// Reset has all the same semantics as time.Timer.Reset.  This method returns true
// if this fakeTimer had been stopped or fired a timer event, false otherwise.
//
// This method is atomic with respect to the containing FakeClock.  In particular,
// this means that if the C() channel was not drained, this method can cause a deadlock.
func (ft *fakeTimer) Reset(d time.Duration) (rescheduled bool) {
	ft.fc.doWith(
		func(now time.Time, ls *listeners) {
			rescheduled = !ft.fired
			ft.when = now.Add(d)
			ft.fired = false

			if equalOrAfter(now, ft.when) {
				ls.remove(ft)
				ft.fired = true
				ft.fire(now)
			} else if !rescheduled {
				ls.add(ft)
			}
		},
	)

	return
}

// Stop cancels this timer, preserving the semantics of time.Timer.Stop.
//
// This method is atomic with respect to the containing FakeClock.
func (ft *fakeTimer) Stop() (stopped bool) {
	ft.fc.doWith(
		func(now time.Time, ls *listeners) {
			stopped = !ft.fired
			ft.fired = true
			ls.remove(ft)
		},
	)

	return
}
