package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FakeTimerSuite struct {
	ChrononSuite
}

func (suite *FakeTimerSuite) TestNewTimer() {
	suite.Run("Immediate", func() {
		for _, d := range []time.Duration{-100 * time.Millisecond, 0} {
			suite.Run(d.String(), func() {
				t, _ := suite.newFakeTimer(d)
				suite.False(t.Stop())

				ft := t.(*fakeTimer)
				suite.True(ft.onAdvance(suite.now))
			})
		}
	})

	suite.Run("Add", func() {
		t, fc := suite.newFakeTimer(TestInterval)

		fc.Add(-time.Second)
		suite.requireNoSignal(t.C(), Immediate)

		fc.Add(time.Second)
		suite.requireNoSignal(t.C(), Immediate)

		fc.Add(TestInterval)
		suite.requireSignal(t.C(), Immediate)

		ft := t.(*fakeTimer)
		suite.True(ft.onAdvance(suite.now))
	})

	suite.Run("Set", func() {
		t, fc := suite.newFakeTimer(TestInterval)

		fc.Set(suite.now.Add(-time.Hour))
		suite.requireNoSignal(t.C(), Immediate)

		fc.Set(suite.now.Add(TestInterval / 2))
		suite.requireNoSignal(t.C(), Immediate)

		fc.Set(suite.now.Add(TestInterval))
		suite.requireSignal(t.C(), Immediate)

		ft := t.(*fakeTimer)
		suite.True(ft.onAdvance(suite.now))
	})

	suite.Run("StopReset", func() {
		t, fc := suite.newFakeTimer(TestInterval)

		// immediate stop then reset
		suite.True(t.Stop())
		suite.False(t.Reset(2 * TestInterval))

		fc.Add(TestInterval)
		suite.requireNoSignal(t.C(), Immediate)

		fc.Add(TestInterval)
		suite.requireSignal(t.C(), Immediate)

		// stop, then reset twice
		suite.False(t.Stop())
		suite.False(t.Reset(3 * TestInterval))
		suite.True(t.Reset(TestInterval))

		fc.Add(TestInterval)
		suite.requireSignal(t.C(), Immediate)

		// reset to a negative duration
		suite.False(t.Reset(-time.Hour))
		suite.requireSignal(t.C(), Immediate)
	})
}

func (suite *FakeTimerSuite) TestAfterFunc() {
	suite.Run("Immediate", func() {
		for _, d := range []time.Duration{-100 * time.Millisecond, 0} {
			suite.Run(d.String(), func() {
				t, _, called := suite.newAfterFunc(d)
				suite.False(t.Stop())
				suite.requireSignal(called, Immediate)

				ft := t.(*fakeTimer)
				suite.True(ft.onAdvance(suite.now))
			})
		}
	})

	suite.Run("Add", func() {
		t, fc, called := suite.newAfterFunc(TestInterval)

		fc.Add(-time.Second)
		suite.requireNoSignal(called, Immediate)

		fc.Add(time.Second)
		suite.requireNoSignal(called, Immediate)

		fc.Add(TestInterval)
		suite.requireSignal(called, Immediate)

		ft := t.(*fakeTimer)
		suite.True(ft.onAdvance(suite.now))
	})

	suite.Run("Set", func() {
		t, fc, called := suite.newAfterFunc(TestInterval)

		fc.Set(suite.now.Add(-time.Hour))
		suite.requireNoSignal(called, Immediate)

		fc.Set(suite.now.Add(TestInterval / 2))
		suite.requireNoSignal(called, Immediate)

		fc.Set(suite.now.Add(TestInterval))
		suite.requireSignal(called, Immediate)

		ft := t.(*fakeTimer)
		suite.True(ft.onAdvance(suite.now))
	})

	suite.Run("StopReset", func() {
		t, fc, called := suite.newAfterFunc(TestInterval)

		// immediate stop then reset
		suite.True(t.Stop())
		suite.False(t.Reset(2 * TestInterval))

		fc.Add(TestInterval)
		suite.requireNoSignal(called, Immediate)

		fc.Add(TestInterval)
		suite.requireSignal(called, Immediate)

		// stop, then reset twice
		suite.False(t.Stop())
		suite.False(t.Reset(3 * TestInterval))
		suite.True(t.Reset(TestInterval))

		fc.Add(TestInterval)
		suite.requireSignal(called, Immediate)

		// reset to a negative duration
		suite.False(t.Reset(-time.Hour))
		suite.requireSignal(called, Immediate)
	})
}

func TestFakeTimer(t *testing.T) {
	suite.Run(t, new(FakeTimerSuite))
}
