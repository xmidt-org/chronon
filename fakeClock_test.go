package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FakeClockSuite struct {
	ChrononSuite
}

func (suite *FakeClockSuite) TestNow() {
	fc := NewFakeClock(suite.now)
	suite.Equal(suite.now, fc.Now())
	suite.Equal(time.Second, fc.Until(suite.now.Add(time.Second)))
	suite.Equal(time.Second, fc.Since(suite.now.Add(-time.Second)))
}

func (suite *FakeClockSuite) TestAdd() {
	fc := NewFakeClock(suite.now)
	suite.Equal(suite.now, fc.Now())

	newNow := fc.Add(100 * time.Hour)
	suite.Equal(newNow, fc.Now())
	suite.Equal(suite.now.Add(100*time.Hour), fc.Now())

	nextNow := fc.Add(-650 * time.Minute)
	suite.Equal(nextNow, fc.Now())
	suite.Equal(newNow.Add(-650*time.Minute), fc.Now())
}

func (suite *FakeClockSuite) TestSleep() {
	suite.Run("NegativeDuration", func() {
		fc := NewFakeClock(suite.now)
		fc.Sleep(-1000)
	})

	suite.Run("ZeroDuration", func() {
		fc := NewFakeClock(suite.now)
		fc.Sleep(0)
	})

	suite.Run("Wakeup", func() {
		suite.Run("Add", func() {
			var (
				fc      = NewFakeClock(suite.now)
				done    = make(chan struct{})
				onSleep = make(chan time.Duration)
			)

			fc.NotifyOnSleep(onSleep)

			go func() {
				defer close(done)
				fc.Sleep(100 * time.Millisecond)
			}()

			// ensure the other goroutine is actually blocked
			sleepTime := suite.requireReceive(onSleep, WaitALittle).(time.Duration)
			suite.requireNoSignal(done, Immediate)

			fc.Add(sleepTime / 2)
			suite.requireNoSignal(done, Immediate)

			fc.Add(sleepTime / 2)
			suite.requireSignal(done, WaitALittle)
		})

		suite.Run("Set", func() {
			var (
				fc      = NewFakeClock(suite.now)
				done    = make(chan struct{})
				onSleep = make(chan time.Duration)
			)

			fc.NotifyOnSleep(onSleep)

			go func() {
				defer close(done)
				fc.Sleep(100 * time.Millisecond)
			}()

			// ensure the other goroutine is actually blocked
			sleepTime := <-onSleep
			suite.requireNoSignal(done, WaitALittle)

			// moving backwards shouldn't affect anything
			fc.Set(suite.now.Add(-100 * time.Second))
			suite.requireNoSignal(done, WaitALittle)

			fc.Set(suite.now.Add(sleepTime))
			suite.requireSignal(done, WaitALittle)
		})

		suite.Run("StopOnSleep", func() {
			var (
				fc      = NewFakeClock(suite.now)
				onSleep = make(chan time.Duration)
			)

			fc.NotifyOnSleep(onSleep)
			fc.StopOnSleep(onSleep)
			suite.Empty(fc.onSleep)
		})
	})
}

func (suite *FakeClockSuite) TestNewTimer() {
	for _, interval := range []time.Duration{-timerInterval, 0, timerInterval} {
		suite.Run(interval.String(), func() {
			var (
				fc      = suite.newFakeClock()
				onTimer = make(chan time.Duration, 1)
				removed = make(chan time.Duration, 1)
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)

			t := fc.NewTimer(interval)
			suite.Require().NotNil(t)
			suite.requireReceiveEqual(onTimer, interval, Immediate)
			suite.requireNoSignal(removed, Immediate)
		})
	}
}

func (suite *FakeClockSuite) TestAfter() {
	var (
		fc      = suite.newFakeClock()
		onTimer = make(chan time.Duration, 1)
		removed = make(chan time.Duration, 1)
	)

	fc.NotifyOnTimer(onTimer)
	fc.NotifyOnTimer(removed)
	fc.StopOnTimer(removed)

	ch := fc.After(timerInterval)
	suite.Require().NotNil(ch)
	suite.requireReceiveEqual(onTimer, timerInterval, Immediate)
	suite.requireNoSignal(removed, Immediate)
}

func (suite *FakeClockSuite) TestAfterFunc() {
	for _, interval := range []time.Duration{-timerInterval, 0, timerInterval} {
		suite.Run(interval.String(), func() {
			var (
				fc      = suite.newFakeClock()
				onTimer = make(chan time.Duration, 1)
				removed = make(chan time.Duration, 1)
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)

			t := fc.AfterFunc(interval, func() {})
			suite.Require().NotNil(t)
			suite.Require().Nil(t.C())
			suite.requireReceiveEqual(onTimer, interval, Immediate)
			suite.requireNoSignal(removed, Immediate)
		})
	}
}

func (suite *FakeClockSuite) TestNewTicker() {
	const tickerInterval time.Duration = 100 * time.Millisecond

	suite.Run("Add", func() {
		var (
			fc       = NewFakeClock(suite.now)
			onTicker = make(chan time.Duration, 1)
			removed  = make(chan time.Duration, 1)
		)

		fc.NotifyOnTicker(onTicker)
		fc.NotifyOnTicker(removed)
		fc.StopOnTicker(removed)
		t := fc.NewTicker(tickerInterval)
		suite.requireReceiveEqual(onTicker, tickerInterval, Immediate)
		suite.requireNoSignal(removed, Immediate)

		suite.requireNoSignal(t.C(), Immediate)
		fc.Add(tickerInterval / 2)
		suite.requireNoSignal(t.C(), Immediate)
		fc.Add(tickerInterval / 2)
		suite.requireSignal(t.C(), Immediate)

		fc.Add(tickerInterval)
		suite.requireSignal(t.C(), Immediate)
	})

	suite.Run("Set", func() {
		var (
			fc       = NewFakeClock(suite.now)
			onTicker = make(chan time.Duration, 1)
			removed  = make(chan time.Duration, 1)
		)

		fc.NotifyOnTicker(onTicker)
		fc.NotifyOnTicker(removed)
		fc.StopOnTicker(removed)
		t := fc.NewTicker(tickerInterval)
		suite.requireReceiveEqual(onTicker, tickerInterval, Immediate)
		suite.requireNoSignal(removed, Immediate)

		suite.requireNoSignal(t.C(), Immediate)

		fc.Set(suite.now.Add(-time.Hour))
		suite.requireNoSignal(t.C(), Immediate)

		fc.Set(suite.now)
		suite.requireNoSignal(t.C(), Immediate)

		fc.Set(suite.now.Add(tickerInterval))
		suite.requireSignal(t.C(), Immediate)
	})
}

func (suite *FakeClockSuite) TestTick() {
	const tickerInterval time.Duration = 100 * time.Millisecond

	suite.Run("Add", func() {
		var (
			fc       = NewFakeClock(suite.now)
			onTicker = make(chan time.Duration, 1)
			removed  = make(chan time.Duration, 1)
		)

		fc.NotifyOnTicker(onTicker)
		fc.NotifyOnTicker(removed)
		fc.StopOnTicker(removed)
		t := fc.Tick(tickerInterval)
		suite.requireReceiveEqual(onTicker, tickerInterval, Immediate)
		suite.requireNoSignal(removed, Immediate)

		suite.requireNoSignal(t, Immediate)
		fc.Add(tickerInterval / 2)
		suite.requireNoSignal(t, Immediate)
		fc.Add(tickerInterval / 2)
		suite.requireSignal(t, Immediate)

		fc.Add(tickerInterval)
		suite.requireSignal(t, Immediate)
	})

	suite.Run("Set", func() {
		var (
			fc       = NewFakeClock(suite.now)
			onTicker = make(chan time.Duration, 1)
			removed  = make(chan time.Duration, 1)
		)

		fc.NotifyOnTicker(onTicker)
		fc.NotifyOnTicker(removed)
		fc.StopOnTicker(removed)
		t := fc.Tick(tickerInterval)
		suite.requireReceiveEqual(onTicker, tickerInterval, Immediate)
		suite.requireNoSignal(removed, Immediate)

		suite.requireNoSignal(t, Immediate)

		fc.Set(suite.now.Add(-time.Hour))
		suite.requireNoSignal(t, Immediate)

		fc.Set(suite.now)
		suite.requireNoSignal(t, Immediate)

		fc.Set(suite.now.Add(tickerInterval))
		suite.requireSignal(t, Immediate)
	})
}

func TestFakeClock(t *testing.T) {
	suite.Run(t, new(FakeClockSuite))
}
