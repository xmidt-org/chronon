package chronon

import (
	"errors"
	"time"
)

// FakeTicker is a time.Ticker implementation driven by a containing FakeClock.
type FakeTicker struct {
	fc *FakeClock

	c    chan time.Time
	tick time.Duration
	next time.Time // the earliest time at which the next tick should fire
}

func newFakeTicker(fc *FakeClock, tick time.Duration, start time.Time) *FakeTicker {
	if tick <= 0 {
		// consistent with time.NewTicker
		panic(errors.New("non-positive interval for FakeTicker"))
	}

	return &FakeTicker{
		fc:   fc,
		c:    make(chan time.Time, 1),
		tick: tick,
		next: start.Add(tick),
	}
}

// onAdvance handles dispatching any tick events to the channel based on
// the containing FakeClock's time advancing.  This method always returns
// false, since advancing a clock never causes a ticker to expire.
func (ft *FakeTicker) onAdvance(now time.Time) bool {
	// dispatch as many ticks as are necessary
	for equalOrAfter(now, ft.next) {
		sendTime(ft.c, ft.next) // send the next instead of now, since we may send multiple
		ft.next = ft.next.Add(ft.tick)
	}

	// a ticker doesn't expire on its own.  it has to be stopped.
	return false
}

// C returns the time channel on which ticks are sent.  This channel is never closed
// and will be the same for the lifetime of this FakeTicker.
func (ft *FakeTicker) C() <-chan time.Time {
	return ft.c
}

// Reset changes the tick duration for this FakeTicker.  The next tick after
// this method is invoked will occur based on the current time of the FakeClock.
//
// If Stop had been called, this method reactivates this FakeTicker with the
// containing FakeClock.
func (ft *FakeTicker) Reset(d time.Duration) {
	if d <= 0 {
		// consistent with time.NewTicker
		panic(errors.New("non-positive interval for FakeTicker"))
	}

	ft.fc.doWith(
		func(now time.Time, ls *listeners) {
			ft.tick = d
			ft.next = now.Add(d)
			ls.add(ft) // idempotent
		},
	)
}

// Stop halts tick events.
func (ft *FakeTicker) Stop() {
	ft.fc.doWith(
		func(_ time.Time, ls *listeners) {
			ls.remove(ft)
		},
	)
}

// Fire forces this FakeTicker to fire a tick event, even if it had been stopped.
// The time used for the tick is whatever the next fire time was.
//
// Calling this method multiple times without advancing the containing FakeClock
// will result in multiple ticks with the same timestamp.  If the tick time value is
// important to testing code, advance the FakeClock or set the FakeClock's time instead
// of using this method.
func (ft *FakeTicker) Fire() {
	ft.fc.doWith(
		func(time.Time, *listeners) {
			sendTime(ft.c, ft.next)
		},
	)
}
