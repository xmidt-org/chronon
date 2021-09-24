package chronon

import "time"

// systemTicker is the standard implementation of Ticker that
// is backed by a *time.Ticker.
type systemTicker struct {
	t *time.Ticker
}

func (st systemTicker) C() <-chan time.Time {
	return st.t.C
}

func (st systemTicker) Reset(d time.Duration) {
	st.t.Reset(d)
}

func (st systemTicker) Stop() {
	st.t.Stop()
}

// systemTimer is a Timer backed by a *time.Timer.
type systemTimer struct {
	t *time.Timer
}

func (st systemTimer) C() <-chan time.Time {
	return st.t.C
}

func (st systemTimer) Reset(d time.Duration) bool {
	return st.t.Reset(d)
}

func (st systemTimer) Stop() bool {
	return st.t.Stop()
}

// systemClock is a Clock that uses the time package.  All the Clock
// interface methods of this type delegate to the time package.
type systemClock struct{}

func (sc systemClock) Now() time.Time {
	return time.Now()
}

func (sc systemClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (sc systemClock) Until(t time.Time) time.Duration {
	return time.Until(t)
}

func (sc systemClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (sc systemClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (sc systemClock) AfterFunc(d time.Duration, f func()) Timer {
	return systemTimer{
		t: time.AfterFunc(d, f),
	}
}

func (sc systemClock) Tick(d time.Duration) <-chan time.Time {
	return time.Tick(d)
}

func (sc systemClock) NewTicker(d time.Duration) Ticker {
	return systemTicker{
		t: time.NewTicker(d),
	}
}

func (sc systemClock) NewTimer(d time.Duration) Timer {
	return systemTimer{
		t: time.NewTimer(d),
	}
}

// SystemClock returns a Clock implementation backed by the time package.
func SystemClock() Clock {
	return systemClock{}
}
