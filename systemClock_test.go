package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SystemClockSuite struct {
	ChannelSuite
}

func (suite *SystemClockSuite) assertEqualOrAfter(expected, actual time.Time) {
	suite.Truef(
		expected.Equal(actual) || expected.After(actual),
		"%s was not at least %s",
		expected,
		actual,
	)
}

func (suite *SystemClockSuite) systemClock() Clock {
	clock := SystemClock()
	suite.Require().NotNil(clock)
	return clock
}

func (suite *SystemClockSuite) newTimer(d time.Duration) Timer {
	t := suite.systemClock().NewTimer(d)
	suite.Require().NotNil(t)
	return t
}

func (suite *SystemClockSuite) newTicker(d time.Duration) Ticker {
	t := suite.systemClock().NewTicker(d)
	suite.Require().NotNil(t)
	return t
}

func (suite *SystemClockSuite) TestNow() {
	clock := suite.systemClock()

	now := clock.Now()
	suite.assertEqualOrAfter(time.Now(), now)
	suite.GreaterOrEqual(int64(clock.Since(now)), int64(0))
	suite.GreaterOrEqual(int64(clock.Until(now.Add(24*time.Hour))), int64(0))
}

func (suite *SystemClockSuite) TestSleep() {
	clock := suite.systemClock()

	now := clock.Now()
	clock.Sleep(10 * time.Millisecond)
	suite.GreaterOrEqual(
		int64(clock.Since(now)),
		10*time.Millisecond,
	)
}

func (suite *SystemClockSuite) TestAfter() {
	clock := suite.systemClock()
	suite.requireSignal(
		clock.After(10*time.Millisecond),
		20*time.Millisecond,
	)
}

func (suite *SystemClockSuite) TestAfterFunc() {
	clock := suite.systemClock()
	done := make(chan struct{})

	clock.AfterFunc(10*time.Millisecond, func() { close(done) })
	suite.requireSignal(done, 20*time.Millisecond)
}

func (suite *SystemClockSuite) TestNewTimer() {
	suite.Run("Wait", func() {
		t := suite.newTimer(10 * time.Millisecond)
		suite.requireSignal(
			t.C(),
			20*time.Millisecond,
		)

		suite.False(t.Stop())
	})

	suite.Run("StopReset", func() {
		t := suite.newTimer(24 * time.Hour)
		defer t.Stop()

		suite.True(t.Stop())
		suite.False(t.Reset(10 * time.Millisecond))
		suite.requireSignal(t.C(), 20*time.Millisecond)
	})
}

func (suite *SystemClockSuite) TestTick() {
	clock := suite.systemClock()
	t := clock.Tick(10 * time.Millisecond)
	suite.requireSignal(t, 11*time.Millisecond)
	suite.requireSignal(t, 11*time.Millisecond)
}

func (suite *SystemClockSuite) TestNewTicker() {
	t := suite.newTicker(10 * time.Millisecond)
	defer t.Stop()

	suite.requireSignal(t.C(), 11*time.Millisecond)
	suite.requireSignal(t.C(), 11*time.Millisecond)

	t.Stop()
	suite.requireNoSignal(t.C(), 11*time.Millisecond)

	t.Reset(15 * time.Millisecond)
	suite.requireSignal(t.C(), 16*time.Millisecond)
	suite.requireSignal(t.C(), 16*time.Millisecond)
}

func TestSystemClock(t *testing.T) {
	suite.Run(t, new(SystemClockSuite))
}
