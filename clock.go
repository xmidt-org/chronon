package chronon

import "time"

// Ticker represents a source of time events that occur at some
// interval.  This type corresponds to *time.Ticker.
type Ticker interface {
	// C returns the channel on which this ticker sends ticks.  This method
	// always returns a non-nil channel.
	//
	// As with the time package, slow clients may miss ticks.  If this channel's
	// buffer is full when it's time to send another tick, that new tick is dropped.
	C() <-chan time.Time

	// Reset reschedules or reactivates this Ticker so that it begins firing
	// ticks on the new interval.
	Reset(time.Duration)

	// Stop halts future ticks.
	Stop()
}

// Timer represents a source of an event that happens once, at a set point in time.
// This type corresponds to *time.Timer.
type Timer interface {
	// C returns the channel on which this timer sends time events.  Note that if
	// this Timer was created via AfterFunc, this method returns nil.
	C() <-chan time.Time

	// Reset either (1) reschedules this Timer to fire after the given duration from Now(), or
	// (2) schedules this Timer to fire again after the given duration from Now().
	//
	// IMPORTANT: The return value for this method is consistent with time.Timer.  All
	// the same caveats apply.  In particular, it's generally not possible to use the
	// return value from this method correctly.
	Reset(time.Duration) bool

	// Stop prevents this Timer from firing if it hasn't already.  If this timer hadn't
	// fired yet, this method returns true to indicate that this Timer was, in fact, stopped.
	// Otherwise, this method return false, indicating that this Timer had already fired.
	Stop() bool
}

// Clock represents a standard set of time operations.  Methods of this interface
// correspond to the functionality in the time package.
type Clock interface {
	// Now returns this clock's notion of the current time.
	Now() time.Time

	// Since returns the duration since this clock's current time.
	Since(t time.Time) time.Duration

	// Until returns the duration between this clock's current time and
	// the given time.
	Until(t time.Time) time.Duration

	// Sleep blocks until this Clock believes that the given
	// time duration has elapsed.
	Sleep(time.Duration)

	// After returns a timer channel with the same semantics as time.After.
	// Since the underlying Timer is not returned, the timer cannot be stopped.
	After(time.Duration) <-chan time.Time

	// AfterFunc invokes the given function after the specified duration.
	// The returned Timer can be used to halt execution.
	AfterFunc(time.Duration, func()) Timer

	// Tick returns a ticker channel with the same semantics as time.Tick.
	Tick(time.Duration) <-chan time.Time

	// NewTicker produces a Ticker which emits events at the given interval.
	NewTicker(time.Duration) Ticker

	// NewTimer produces a Timer which emits its single event at the given
	// duration in the future.
	NewTimer(time.Duration) Timer
}
