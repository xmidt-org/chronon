package chronon

import (
	"sync"
	"time"
)

// Adder represents a source of time values that can be modified
// by adding a duration.  *FakeClock implements this interface.
type Adder interface {
	// Add adjusts the current time by the given delta.  The delta
	// can be negative or 0.  This method returns the new value
	// of the current time.
	Add(time.Duration) time.Time
}

// Setter represents a source of time values that can be updated
// using absolute time.  *FakeClock implements this interface.
type Setter interface {
	// Set adjusts the current time to the given value.
	Set(time.Time)
}

// FakeClock is a Clock implementation that allows control over how
// the clock advances.
type FakeClock struct {
	lock sync.RWMutex

	now       time.Time
	listeners *listeners
}

var _ Clock = (*FakeClock)(nil)
var _ Adder = (*FakeClock)(nil)
var _ Setter = (*FakeClock)(nil)

// NewFakeClock creates a FakeClock that uses the given time as the
// initial current time.
func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{
		now:       start,
		listeners: new(listeners),
	}
}

// doWith executes a function under this clock's lock.  The supplied
// function is passed the current fake clock time and the set of listeners.
func (fc *FakeClock) doWith(f func(time.Time, *listeners)) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	f(fc.now, fc.listeners)
}

// Add satisfies the Adder interface.  Updating this fake clock's
// time through this method is atomic with respect to all the other
// methods.
func (fs *FakeClock) Add(d time.Duration) (now time.Time) {
	fs.lock.Lock()
	now = fs.now.Add(d)
	fs.now = now
	fs.listeners.onAdvance(now)
	fs.lock.Unlock()

	return
}

// Set is similar to Advance, except that it sets an absolute time instead
// of moving this fake clock's time by a certain delta.
func (fs *FakeClock) Set(t time.Time) {
	fs.lock.Lock()
	fs.now = t
	fs.listeners.onAdvance(t)
	fs.lock.Unlock()
}

// Now returns the value for the current time.
func (fc *FakeClock) Now() (n time.Time) {
	fc.lock.RLock()
	n = fc.now
	fc.lock.RUnlock()
	return
}

// Since returns the duration from the given time to this FakeClock's current time.
// This method is atomic with respect to the other methods of this instance.
func (fc *FakeClock) Since(t time.Time) (d time.Duration) {
	fc.lock.RLock()
	d = fc.now.Sub(t)
	fc.lock.RUnlock()
	return
}

// Until returns the duration from this FakeClock's current time to the given time.
// This method is atomic with respect to the other methods of this instance.
func (fc *FakeClock) Until(t time.Time) (d time.Duration) {
	fc.lock.RLock()
	d = t.Sub(fc.now)
	fc.lock.RUnlock()
	return
}

// Sleep blocks until this clock is advanced sufficiently so that
// the given duration elapses.
func (fc *FakeClock) Sleep(d time.Duration) {
	if d <= 0 {
		// consistent with time.Sleep
		return
	}

	fc.lock.Lock()
	sleeper := newSleeperAt(fc.now.Add(d))
	fc.listeners.add(sleeper)
	fc.lock.Unlock()

	sleeper.wait()
}

// After returns a channel which receives a time after the given duration.
func (fc *FakeClock) After(d time.Duration) <-chan time.Time {
	return fc.NewTimer(d).C()
}

// AfterFunc schedules a function to execute after this FakeClock has been advanced
// by at least the given duration.  The returned Timer can be used to cancel the
// execution, as with time.AfterFunc.  The returned Timer from this method is
// always a *FakeTimer, and its C() method always returns nil.
func (fc *FakeClock) AfterFunc(d time.Duration, f func()) Timer {
	fc.lock.Lock()
	ft := newAfterFunc(fc, fc.now.Add(d), func(time.Time) { f() })

	// handle nonpositive durations consistently
	if !ft.onAdvance(fc.now) {
		fc.listeners.add(ft)
	}

	fc.lock.Unlock()
	return ft
}

// Tick returns a channel which receives time events at the given interval.
func (fc *FakeClock) Tick(d time.Duration) <-chan time.Time {
	return fc.NewTicker(d).C()
}

// NewTimer creates a FakeTimer which fires after the given interval.  The semantics
// of time.Timer are preserved, including the unusual Stop/Reset behavior.
//
// The returned Timer can always be safely cast to a *FakeTimer.
func (fc *FakeClock) NewTimer(d time.Duration) Timer {
	fc.lock.Lock()
	ft := newFakeTimer(fc, fc.now.Add(d))

	// handle nonpositive durations consistently with normal triggering
	if !ft.onAdvance(fc.now) {
		fc.listeners.add(ft)
	}

	fc.lock.Unlock()
	return ft
}

// NewTicker creates a FakeTicker which fires on the given interval.  The semantics
// of time.Ticker are preserved.
//
// The returned Ticker can always be safely cast to a *FakeTicker.
func (fc *FakeClock) NewTicker(d time.Duration) Ticker {
	fc.lock.Lock()
	ft := newFakeTicker(fc, d, fc.now)

	// consistent logic with NewTimer, even though onAdvance
	// always returns false (at least, right now)
	if !ft.onAdvance(fc.now) {
		fc.listeners.add(ft)
	}

	fc.lock.Unlock()
	return ft
}
