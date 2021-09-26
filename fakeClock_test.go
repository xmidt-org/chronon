package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FakeClockSuite struct {
	suite.Suite

	now time.Time
}

func (suite *FakeClockSuite) requireNotDone(done <-chan struct{}) {
	select {
	case <-done:
		suite.Require().Fail("The done channel should NOT have been signalled")
	case <-time.After(100 * time.Millisecond):
	}
}

func (suite *FakeClockSuite) requireDone(done <-chan struct{}) {
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		suite.Require().Fail("The done channel should have been signalled")
	}
}

func (suite *FakeClockSuite) requireResult(result <-chan time.Time, expected time.Time) {
	select {
	case actual := <-result:
		suite.Equal(expected, actual)
	case <-time.After(100 * time.Millisecond):
		suite.Require().Fail("A result should have been sent")
	}
}

func (suite *FakeClockSuite) requireNoResult(result <-chan time.Time) {
	select {
	case <-result:
		suite.Require().Fail("A result should NOT have been sent")
	case <-time.After(100 * time.Millisecond):
	}
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
			sleepTime := <-onSleep
			suite.requireNotDone(done)

			fc.Add(sleepTime / 2)
			suite.requireNotDone(done)

			fc.Add(sleepTime / 2)
			suite.requireDone(done)
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
			suite.requireNotDone(done)

			// moving backwards shouldn't affect anything
			fc.Set(suite.now.Add(-100 * time.Second))
			suite.requireNotDone(done)

			fc.Set(suite.now.Add(sleepTime))
			suite.requireDone(done)
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

			select {
			case <-removed:
				suite.Require().Fail("The removed channel shouldn't have been signalled when adding a timer")
			default:
			}

			go func() {
				result <- (<-t.C())
			}()

			suite.requireNoResult(result)

			fc.Add(d / 2)
			suite.requireNoResult(result)

			fc.Add(d / 2)
			suite.requireResult(result, suite.now.Add(d))
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

			select {
			case <-removed:
				suite.Require().Fail("The removed channel shouldn't have been signalled when adding a timer")
			default:
			}

			go func() {
				result <- (<-t.C())
			}()

			suite.requireNoResult(result)
			fc.Set(suite.now.Add(-time.Second))
			suite.requireNoResult(result)
			fc.Set(suite.now.Add(d))
			suite.requireResult(result, suite.now.Add(d))
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

	suite.requireNoResult(result)
	fc.Add(50 * time.Millisecond)
	suite.requireNoResult(result)
	fc.Add(50 * time.Millisecond)
	suite.requireResult(result, suite.now.Add(100*time.Millisecond))
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
				case <-called:
				default:
					suite.Fail("The function should have been immediately called")
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
