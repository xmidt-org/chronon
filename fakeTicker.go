package chronon

import (
	"errors"
	"time"
)

// FakeTicker represents a Ticker which can be manually controlled, either by
// advancing its containing fake clock or through methods of this interface.
type FakeTicker interface {
	Ticker

	// When returns the next time at which a tick will fire.  This value will
	// change after each tick or if this ticker is reset.
	When() time.Time

	// Fire forcibly sends a tick, unless this ticker is not active.  This
	// method returns true if the tick was sent, false if this ticker had been stopped.
	//
	// This method does not update the fake clock's current time or the
	// return value from When.  If the actual time for each tick is important
	// to production code, force ticks to fire by using FakeClock.Set and passing
	// the value returned by When.
	Fire() bool
}

// fakeTicker is a time.Ticker implementation driven by a containing FakeClock.
type fakeTicker struct {
	fc *FakeClock

	c    chan time.Time
	tick time.Duration
	next time.Time // the earliest time at which the next tick should fire
}

func newFakeTicker(fc *FakeClock, tick time.Duration, start time.Time) *fakeTicker {
	if tick <= 0 {
		// consistent with time.NewTicker
		panic(errors.New("non-positive interval for fakeTicker"))
	}

	return &fakeTicker{
		fc:   fc,
		c:    make(chan time.Time, 1),
		tick: tick,
		next: start.Add(tick),
	}
}

func (ft *fakeTicker) When() (t time.Time) {
	ft.fc.doWith(
		func(time.Time, *listeners) {
			t = ft.next
		},
	)

	return
}

func (ft *fakeTicker) Fire() (fired bool) {
	ft.fc.doWith(
		func(_ time.Time, ls *listeners) {
			fired = ls.active(ft)
			if fired {
				sendTime(ft.c, ft.next)
			}
		},
	)

	return
}

// onUpdate handles dispatching any tick events to the channel based on
// the containing FakeClock's time advancing.  This method always returns
// false, since advancing a clock never causes a ticker to expire.
func (ft *fakeTicker) onUpdate(now time.Time) updateResult {
	// dispatch as many ticks as are necessary
	for equalOrAfter(now, ft.next) {
		sendTime(ft.c, ft.next) // send the next instead of now, since we may send multiple
		ft.next = ft.next.Add(ft.tick)
	}

	// a ticker doesn't expire on its own.  it has to be stopped.
	return continueUpdates
}

// C returns the time channel on which ticks are sent.  This channel is never closed
// and will be the same for the lifetime of this fakeTicker.
func (ft *fakeTicker) C() <-chan time.Time {
	return ft.c
}

// Reset changes the tick duration for this fakeTicker.  The next tick after
// this method is invoked will occur based on the current time of the FakeClock.
//
// If Stop had been called, this method reactivates this fakeTicker with the
// containing FakeClock.
func (ft *fakeTicker) Reset(d time.Duration) {
	if d <= 0 {
		// consistent with time.NewTicker
		panic(errors.New("non-positive interval for fakeTicker"))
	}

	ft.fc.doWith(
		func(now time.Time, ls *listeners) {
			ft.tick = d
			ft.next = now.Add(d)
			ls.add(ft)
		},
	)
}

// Stop halts tick events.
func (ft *fakeTicker) Stop() {
	ft.fc.doWith(
		func(_ time.Time, ls *listeners) {
			ls.remove(ft)
		},
	)
}
