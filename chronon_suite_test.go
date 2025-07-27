// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"reflect"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	// Immediate is a constant that indicates a receive operation should not spawn a timer.
	Immediate time.Duration = 0

	// WaitALittle is a convenient value for doing a channel select and waiting a small amount of time.
	// This value is a tad longer than TestInterval to allow for jitter when waiting for concurrent events.
	WaitALittle time.Duration = 110 * time.Millisecond

	// TestInterval is a positive interval convenient for testing
	TestInterval time.Duration = 100 * time.Millisecond
)

// ChrononSuite is an embeddable test suite that extends stretchr's suite
// with some useful assertions and utilities.
type ChrononSuite struct {
	suite.Suite

	now time.Time
}

func (suite *ChrononSuite) SetupSuite() {
	suite.now = time.Now()
}

// newFakeClock creates a *FakeClock under test that has standard
// assertions applied to it.
func (suite *ChrononSuite) newFakeClock() *FakeClock {
	suite.T().Helper()
	fc := NewFakeClock(suite.now)
	suite.Require().NotNil(fc)
	suite.Require().True(suite.now.Equal(fc.Now()))
	return fc
}

// newSleeper spawns a goroutine that is blocked in a FakeClock.Sleep(d) call.  The Sleeper,
// containing FakeClock, and a done channel are returned for tests to manipulate.  The done
// channel is closed when Sleep returns.
func (suite *ChrononSuite) newSleeper(d time.Duration) (Sleeper, *FakeClock, <-chan struct{}) {
	suite.T().Helper()

	var (
		fc      = suite.newFakeClock()
		onSleep = make(chan Sleeper)
		done    = make(chan struct{})
	)

	fc.NotifyOnSleep(onSleep)

	go func() {
		defer close(done)
		fc.Sleep(d)
	}()

	s := suite.requireReceive(onSleep, WaitALittle).(Sleeper)
	fc.StopOnSleep(onSleep)

	return s, fc, done
}

// newFakeTimer creates a fake timer and a *FakeClock to control it.
// Standard assertions are run against both the clock and the timer.
func (suite *ChrononSuite) newFakeTimer(d time.Duration) (FakeTimer, *FakeClock) {
	suite.T().Helper()
	fc := suite.newFakeClock()

	t := fc.NewTimer(d)
	suite.Require().NotNil(t)
	suite.Require().NotNil(t.C())

	if d > 0 {
		suite.requireNoSignal(t.C(), Immediate)
	} else {
		suite.requireSignal(t.C(), Immediate)
	}

	return t.(FakeTimer), fc
}

// newAfterFunc creates a delayed function using AfterFunc and runs standard assertions.
// The returned channel is signaled when the function is called.
func (suite *ChrononSuite) newAfterFunc(d time.Duration) (FakeTimer, *FakeClock, <-chan struct{}) {
	suite.T().Helper()
	fc := suite.newFakeClock()

	called := make(chan struct{}, 1)
	t := fc.AfterFunc(d, func() { called <- struct{}{} })
	suite.Require().NotNil(t)
	suite.Require().Nil(t.C())

	return t.(FakeTimer), fc, called
}

// newFakeTicker creates a fake ticker and a fake clock to control it with.
// Standard assertions are run on both objects.
func (suite *ChrononSuite) newFakeTicker(d time.Duration) (Ticker, *FakeClock) {
	suite.T().Helper()
	fc := suite.newFakeClock()

	t := fc.NewTicker(d)
	suite.Require().NotNil(t)
	suite.Require().NotNil(t.C())

	// a valid ticker's channel should never be signaled when first created
	suite.requireNoSignal(t.C(), Immediate)

	return t, fc
}

func (suite *ChrononSuite) selectOn(ch interface{}, waitFor time.Duration) (int, reflect.Value, bool) {
	suite.T().Helper()
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		},
	}

	if waitFor > 0 {
		t := time.NewTimer(100 * time.Millisecond)
		defer t.Stop()

		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(t.C),
		})
	} else {
		cases = append(cases, reflect.SelectCase{
			Dir: reflect.SelectDefault,
		})
	}

	return reflect.Select(cases)
}

func (suite *ChrononSuite) requireReceive(ch interface{}, waitFor time.Duration) interface{} {
	suite.T().Helper()
	chosen, value, recvOK := suite.selectOn(ch, waitFor)
	suite.Require().Truef(
		chosen == 0 && recvOK,
		"Nothing received on channel [%T]",
		ch,
	)

	return value.Interface()
}

func (suite *ChrononSuite) requireReceiveEqual(ch, expected interface{}, waitFor time.Duration) {
	suite.T().Helper()
	suite.Equal(expected, suite.requireReceive(ch, waitFor))
}

func (suite *ChrononSuite) requireSignal(ch interface{}, waitFor time.Duration) {
	suite.T().Helper()
	chosen, _, _ := suite.selectOn(ch, waitFor)
	suite.Require().Truef(
		chosen == 0,
		"The channel [%T] should have been signaled",
		ch,
	)
}

func (suite *ChrononSuite) requireNoSignal(ch interface{}, waitFor time.Duration) {
	suite.T().Helper()
	chosen, _, _ := suite.selectOn(ch, waitFor)
	suite.Require().Truef(
		chosen == 1,
		"The channel [%T] should NOT have been signaled",
		ch,
	)
}
