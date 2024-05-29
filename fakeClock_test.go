// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
	for _, interval := range []time.Duration{-TestInterval, 0, TestInterval} {
		suite.Run(interval.String(), func() {
			var (
				fc      = suite.newFakeClock()
				onSleep = make(chan Sleeper, 1)
				removed = make(chan Sleeper, 1)
			)

			fc.NotifyOnSleep(onSleep)
			fc.NotifyOnSleep(removed)
			fc.StopOnSleep(removed)

			done := make(chan struct{})
			go func() {
				defer close(done)
				fc.Sleep(interval)
			}()

			s := suite.requireReceive(onSleep, WaitALittle).(Sleeper)
			suite.Require().NotNil(s)
			suite.requireNoSignal(removed, Immediate)

			fc.Set(s.When())
			suite.requireSignal(done, WaitALittle)
		})
	}
}

func (suite *FakeClockSuite) TestNewTimer() {
	for _, interval := range []time.Duration{-TestInterval, 0, TestInterval} {
		suite.Run(interval.String(), func() {
			var (
				fc      = suite.newFakeClock()
				onTimer = make(chan FakeTimer, 1)
				removed = make(chan FakeTimer, 1)
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)

			t := fc.NewTimer(interval)
			suite.Require().NotNil(t)
			ft := suite.requireReceive(onTimer, Immediate).(FakeTimer)
			suite.Same(t, ft)
			suite.requireNoSignal(removed, Immediate)
		})
	}
}

func (suite *FakeClockSuite) TestAfter() {
	var (
		fc      = suite.newFakeClock()
		onTimer = make(chan FakeTimer, 1)
		removed = make(chan FakeTimer, 1)
	)

	fc.NotifyOnTimer(onTimer)
	fc.NotifyOnTimer(removed)
	fc.StopOnTimer(removed)

	ch := fc.After(TestInterval)
	suite.Require().NotNil(ch)
	suite.requireNoSignal(removed, Immediate)

	t := suite.requireReceive(onTimer, Immediate).(FakeTimer)
	suite.Require().NotNil(t)
	suite.True(ch == t.C())
}

func (suite *FakeClockSuite) TestAfterFunc() {
	for _, interval := range []time.Duration{-TestInterval, 0, TestInterval} {
		suite.Run(interval.String(), func() {
			var (
				fc      = suite.newFakeClock()
				onTimer = make(chan FakeTimer, 1)
				removed = make(chan FakeTimer, 1)
			)

			fc.NotifyOnTimer(onTimer)
			fc.NotifyOnTimer(removed)
			fc.StopOnTimer(removed)

			t := fc.AfterFunc(interval, func() {})
			suite.Require().NotNil(t)
			suite.Require().Nil(t.C())

			ft := suite.requireReceive(onTimer, Immediate).(FakeTimer)
			suite.Same(t, ft)
			suite.requireNoSignal(removed, Immediate)
		})
	}
}

func (suite *FakeClockSuite) TestNewTicker() {
	var (
		fc       = suite.newFakeClock()
		onTicker = make(chan FakeTicker, 1)
		removed  = make(chan FakeTicker, 1)
	)

	fc.NotifyOnTicker(onTicker)
	fc.NotifyOnTicker(removed)
	fc.StopOnTicker(removed)

	t := fc.NewTicker(TestInterval)
	suite.Require().NotNil(t)
	suite.Require().NotNil(t.C())

	ft := suite.requireReceive(onTicker, Immediate).(FakeTicker)
	suite.Same(t, ft)
	suite.requireNoSignal(removed, Immediate)
}

func (suite *FakeClockSuite) TestTick() {
	var (
		fc       = suite.newFakeClock()
		onTicker = make(chan FakeTicker, 1)
		removed  = make(chan FakeTicker, 1)
	)

	fc.NotifyOnTicker(onTicker)
	fc.NotifyOnTicker(removed)
	fc.StopOnTicker(removed)

	ch := fc.Tick(TestInterval)
	suite.Require().NotNil(ch)

	ft := suite.requireReceive(onTicker, Immediate).(FakeTicker)
	suite.Require().NotNil(ft)
	suite.True(ch == ft.C())
	suite.requireNoSignal(removed, Immediate)
}

func TestFakeClock(t *testing.T) {
	suite.Run(t, new(FakeClockSuite))
}
