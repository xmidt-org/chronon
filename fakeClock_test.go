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

		select {
		case <-done:
			suite.Fail("The sleeping goroutine should NOT have exited")
		default:
			// passing
		}

		fc.Add(sleepTime / 2)

		select {
		case <-done:
			suite.Fail("The sleeping goroutine should NOT have exited")
		default:
			// passing
		}

		fc.Add(sleepTime / 2)

		select {
		case <-done:
			// passing
		case <-time.After(100 * time.Millisecond):
			suite.Fail("The sleeping goroutine should have exited")
		}
	})
}

func TestFakeClock(t *testing.T) {
	suite.Run(t, new(FakeClockSuite))
}
