package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FakeClockSuite struct {
	ChannelSuite

	now time.Time
}

func (suite *FakeClockSuite) SetupSuite() {
	suite.now = time.Now()
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
			sleepTime := suite.requireReceive(onSleep, 100*time.Millisecond).(time.Duration)
			suite.requireNoSignal(done, Immediate)

			fc.Add(sleepTime / 2)
			suite.requireNoSignal(done, Immediate)

			fc.Add(sleepTime / 2)
			suite.requireSignal(done, 100*time.Millisecond)
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
	suite.Run("Immediate", func() {
		for _, interval := range []time.Duration{-100 * time.Millisecond, 0} {
			suite.Run(interval.String(), func() {
				var (
					fc      = NewFakeClock(suite.now)
					onTimer = make(chan time.Duration, 1)
					removed = make(chan time.Duration, 1)
				)

				fc.NotifyOnTimer(onTimer)
				fc.NotifyOnTimer(removed)
				fc.StopOnTimer(removed)
				t := fc.NewTimer(interval)

				select {
				case d := <-onTimer:
					suite.Equal(interval, d)
				default:
					suite.Require().Fail("The onTimer channel should have immediately received a value")
				}

				select {
				case <-removed:
					suite.Require().Fail("The removed channel shouldn't have been signalled when adding a timer")
				default:
				}

				select {
				case v := <-t.C():
					suite.Equal(suite.now, v)
				default:
					suite.Fail("The timer channel should have immediately received a value")
				}
			})
		}
	})

	suite.Run("Delayed", func() {
		suite.Run("Add", func() {
			var (
				fc      = NewFakeClock(suite.now)
				onTimer = make(chan time.Duration, 1)
				removed = make(chan time.Duration, 1)
				result  = make(chan time.Time)
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)
			t := fc.NewTimer(100 * time.Millisecond)
			d := <-onTimer // should have a value immediately
			suite.requireNoSignal(removed, Immediate)

			go func() {
				result <- (<-t.C())
			}()

			suite.requireNoSignal(result, WaitALittle)

			fc.Add(d / 2)
			suite.requireNoSignal(result, WaitALittle)

			fc.Add(d / 2)
			suite.requireReceiveEqual(result, suite.now.Add(d), WaitALittle)
		})

		suite.Run("Set", func() {
			var (
				fc      = NewFakeClock(suite.now)
				onTimer = make(chan time.Duration, 1)
				removed = make(chan time.Duration, 1)
				result  = make(chan time.Time)
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)
			t := fc.NewTimer(100 * time.Millisecond)
			d := <-onTimer // should have a value immediately
			suite.requireNoSignal(removed, Immediate)

			go func() {
				result <- (<-t.C())
			}()

			suite.requireNoSignal(result, WaitALittle)
			fc.Set(suite.now.Add(-time.Second))
			suite.requireNoSignal(result, WaitALittle)
			fc.Set(suite.now.Add(d))
			suite.requireReceiveEqual(result, suite.now.Add(d), WaitALittle)
		})
	})
}

func (suite *FakeClockSuite) TestAfter() {
	// we've already put NewTimer through it's paces, so this is
	// just testing the happy path

	var (
		fc     = NewFakeClock(suite.now)
		ch     = fc.After(100 * time.Millisecond)
		result = make(chan time.Time)
	)

	go func() {
		result <- (<-ch)
	}()

	suite.requireNoSignal(result, WaitALittle)
	fc.Add(50 * time.Millisecond)
	suite.requireNoSignal(result, WaitALittle)
	fc.Add(50 * time.Millisecond)
	suite.requireReceiveEqual(result, suite.now.Add(100*time.Millisecond), WaitALittle)
}

func (suite *FakeClockSuite) TestAfterFunc() {
	suite.Run("Immediate", func() {
		for _, interval := range []time.Duration{-100 * time.Millisecond, 0} {
			suite.Run(interval.String(), func() {
				var (
					fc      = NewFakeClock(suite.now)
					onTimer = make(chan time.Duration, 1)
					removed = make(chan time.Duration, 1)
					called  = make(chan struct{})
				)

				fc.NotifyOnTimer(onTimer)
				fc.NotifyOnTimer(removed)
				fc.StopOnTimer(removed)
				t := fc.AfterFunc(interval, func() {
					close(called)
				})

				suite.Nil(t.C())
				suite.requireReceiveEqual(onTimer, interval, Immediate)
				suite.requireNoSignal(removed, Immediate)
				suite.requireSignal(called, Immediate)
			})
		}
	})

	suite.Run("Delayed", func() {
		suite.Run("Add", func() {
			var (
				fc      = NewFakeClock(suite.now)
				onTimer = make(chan time.Duration, 1)
				removed = make(chan time.Duration, 1)
				called  = make(chan struct{})
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)
			t := fc.AfterFunc(100*time.Millisecond, func() {
				close(called)
			})

			suite.Nil(t.C())
			d := <-onTimer // should have a value immediately

			select {
			case <-removed:
				suite.Require().Fail("The removed channel shouldn't have been signalled when adding a timer")
			default:
			}

			select {
			case <-called:
				suite.Require().Fail("The function should NOT have been called yet")
			default:
			}

			fc.Add(d / 2)
			select {
			case <-called:
				suite.Require().Fail("The function should NOT have been called yet")
			default:
			}

			fc.Add(d / 2)
			select {
			case <-called:
			default:
				suite.Require().Fail("The function should have been called")
			}
		})

		suite.Run("Set", func() {
			var (
				fc      = NewFakeClock(suite.now)
				onTimer = make(chan time.Duration, 1)
				removed = make(chan time.Duration, 1)
				called  = make(chan struct{})
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)
			t := fc.AfterFunc(100*time.Millisecond, func() {
				close(called)
			})

			suite.Nil(t.C())
			d := <-onTimer // should have a value immediately

			select {
			case <-removed:
				suite.Require().Fail("The removed channel shouldn't have been signalled when adding a timer")
			default:
			}

			select {
			case <-called:
				suite.Require().Fail("The function should NOT have been called yet")
			default:
			}

			fc.Set(suite.now.Add(-time.Second))
			select {
			case <-called:
				suite.Require().Fail("The function should NOT have been called yet")
			default:
			}

			fc.Set(suite.now.Add(d))
			select {
			case <-called:
			default:
				suite.Require().Fail("The function should have been called")
			}
		})
	})
}

func TestFakeClock(t *testing.T) {
	suite.Run(t, new(FakeClockSuite))
}
