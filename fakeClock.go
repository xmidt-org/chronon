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
	listeners listeners
	onSleeper notifiers
	onTimer   notifiers
	onTicker  notifiers
}

var _ Clock = (*FakeClock)(nil)
var _ Adder = (*FakeClock)(nil)
var _ Setter = (*FakeClock)(nil)

// NewFakeClock creates a FakeClock that uses the given time as the
// initial current time.
func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{
		now: start,
	}
}

// doWith executes a function under this clock's lock.  The supplied
// function is passed the current fake clock time and the set of listeners.
func (fc *FakeClock) doWith(f func(time.Time, *listeners)) {
	fc.lock.Lock()
	defer fc.lock.Unlock()
	f(fc.now, &fc.listeners)
}

// Add satisfies the Adder interface.  Updating this fake clock's
// time through this method is atomic with respect to all the other
// methods.
func (fc *FakeClock) Add(d time.Duration) (now time.Time) {
	fc.lock.Lock()
	now = fc.now.Add(d)
	fc.now = now
	fc.listeners.onUpdate(now)
	fc.lock.Unlock()

	return
}

// Set is similar to Advance, except that it sets an absolute time instead
// of moving this fake clock's time by a certain delta.
func (fc *FakeClock) Set(t time.Time) {
	fc.lock.Lock()
	fc.now = t
	fc.listeners.onUpdate(t)
	fc.lock.Unlock()
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
// the given duration elapses.  If d is nonpositive, this function
// immediately returns exactly as with time.Sleep.
//
// If d is positive, then any channel registered with NotifyOnSleep
// will receive d prior to blocking.
func (fc *FakeClock) Sleep(d time.Duration) {
	if d <= 0 {
		// consistent with time.Sleep
		return
	}

	fc.lock.Lock()
	sleeper := newSleeperAt(fc, fc.now.Add(d))
	fc.listeners.register(fc.now, sleeper)
	fc.onSleeper.notify(sleeper)
	fc.lock.Unlock()

	sleeper.wait()
}

// NotifyOnSleep registers a channel that receives the intervals for any goroutine
// which invokes Sleep.  Calling code must service the channel promptly, as Sleep
// does not drop events sent to this channel.
//
// A sleep channel is useful when testing concurrent code where test code
// needs to block waiting for a sleeper before modifying this FakeClock's time.
// When used for this purpose, be sure to register a sleep channel before
// invoking Sleep, usually in test setup code.
func (fc *FakeClock) NotifyOnSleep(ch chan<- Sleeper) {
	fc.lock.Lock()
	fc.onSleeper.add(ch)
	fc.lock.Unlock()
}

// StopOnSleep removes a channel from the list of channels that receive notifications
// for Sleep.  If the given channel is not present, this method does nothing.
func (fc *FakeClock) StopOnSleep(ch chan<- Sleeper) {
	fc.lock.Lock()
	fc.onSleeper.remove(ch)
	fc.lock.Unlock()
}

// NewTimer creates a Timer that fires when this FakeClock has been advanced
// by at least the given duration.  The returned timer can be stopped or reset in
// the usual fashion, which will affect what happens when the FakeClock is advanced.
//
// The Timer returned by this method can always be cast to a FakeTimer.
func (fc *FakeClock) NewTimer(d time.Duration) Timer {
	fc.lock.Lock()
	ft := newFakeTimer(fc, fc.now.Add(d))

	fc.listeners.register(fc.now, ft)
	fc.onTimer.notify(ft)

	fc.lock.Unlock()
	return ft
}

// After returns a channel which receives a time after the given duration.
func (fc *FakeClock) After(d time.Duration) <-chan time.Time {
	return fc.NewTimer(d).C()
}

// AfterFunc schedules a function to execute after this FakeClock has been advanced
// by at least the given duration.  The returned Timer can be used to cancel the
// execution, as with time.AfterFunc.  The returned Timer from this method is
// always a *FakeTimer, and its C() method always returns nil.
//
// The Timer returned by this method can always be cast to a FakeTimer.
func (fc *FakeClock) AfterFunc(d time.Duration, f func()) Timer {
	fc.lock.Lock()
	ft := newAfterFunc(fc, fc.now.Add(d), func(time.Time) { f() })

	fc.listeners.register(fc.now, ft)
	fc.onTimer.notify(ft)

	fc.lock.Unlock()
	return ft
}

// NotifyOnTimer registers a channel that receives the intervals for any timers created
// through this fake clock.  This includes implicit timers, such as with After and AfterFunc.
//
// Test code that uses this method can be notified when code under test creates timers.
func (fc *FakeClock) NotifyOnTimer(ch chan<- FakeTimer) {
	fc.lock.Lock()
	fc.onTimer.add(ch)
	fc.lock.Unlock()
}

// StopOnTimer removes a channel from the list of channels that receive notifications
// for timers.  If the given channel is not present, this method does nothing.
func (fc *FakeClock) StopOnTimer(ch chan<- FakeTimer) {
	fc.lock.Lock()
	fc.onTimer.remove(ch)
	fc.lock.Unlock()
}

// NewTicker creates a Ticker that fires when this FakeClock is advanced by
// increments of the given duration.  The returned ticker can be stopped or
// reset in the usual fashion.
//
// The Ticker returned from this method can always be cast to a FakeTicker.
func (fc *FakeClock) NewTicker(d time.Duration) Ticker {
	fc.lock.Lock()
	ft := newFakeTicker(fc, d, fc.now)

	fc.listeners.register(fc.now, ft)
	fc.onTicker.notify(ft)
	fc.lock.Unlock()
	return ft
}

// Tick returns a channel which receives time events at the given interval.
func (fc *FakeClock) Tick(d time.Duration) <-chan time.Time {
	return fc.NewTicker(d).C()
}

// NotifyOnTicker registers a channel that receives the intervals for any tickers created
// through this fake clock.  This includes implicit tickers, such as Tick.
//
// Test code that uses this method can be notified when code under test creates tickers.
func (fc *FakeClock) NotifyOnTicker(ch chan<- FakeTicker) {
	fc.lock.Lock()
	fc.onTicker.add(ch)
	fc.lock.Unlock()
}

// StopOnTicker removes a channel from the list of channels that receive notifications
// for timers.  If the given channel is not present, this method does nothing.
func (fc *FakeClock) StopOnTicker(ch chan<- FakeTicker) {
	fc.lock.Lock()
	fc.onTicker.remove(ch)
	fc.lock.Unlock()
}
