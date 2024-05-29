// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FakeTickerSuite struct {
	ChrononSuite
}

func (suite *FakeTickerSuite) TestInvalidDuration() {
	for _, invalid := range []time.Duration{-TestInterval, 0} {
		suite.Run(invalid.String(), func() {
			fc := suite.newFakeClock()
			suite.Panics(func() {
				fc.NewTicker(invalid)
			})
		})
	}
}

func (suite *FakeTickerSuite) TestAdd() {
	t, fc := suite.newFakeTicker(TestInterval)

	fc.Add(TestInterval / 2)
	suite.requireNoSignal(t.C(), Immediate)

	fc.Add(TestInterval / 2)
	suite.requireSignal(t.C(), Immediate)

	fc.Set(suite.now.Add(-time.Second))
	suite.requireNoSignal(t.C(), Immediate)

	fc.Add(time.Second)
	suite.requireNoSignal(t.C(), Immediate)

	// the ticker shouldn't fire for a timestamp that it previously fired for
	fc.Add(TestInterval)
	suite.requireNoSignal(t.C(), Immediate)

	fc.Add(TestInterval)
	suite.requireSignal(t.C(), Immediate)
}

func (suite *FakeTickerSuite) TestSet() {
	t, fc := suite.newFakeTicker(TestInterval)

	fc.Set(suite.now.Add(-time.Second))
	suite.requireNoSignal(t.C(), Immediate)

	fc.Set(suite.now.Add(TestInterval / 2))
	suite.requireNoSignal(t.C(), Immediate)

	fc.Set(suite.now.Add(TestInterval))
	suite.requireSignal(t.C(), Immediate)

	fc.Set(suite.now.Add(2 * TestInterval))
	suite.requireSignal(t.C(), Immediate)

	// the ticker shouldn't fire for a timestamp that it previously fired for
	fc.Set(suite.now.Add(TestInterval))
	suite.requireNoSignal(t.C(), Immediate)
}

func (suite *FakeTickerSuite) TestStopReset() {
	t, fc := suite.newFakeTicker(TestInterval)

	t.Stop()
	fc.Add(TestInterval)
	suite.requireNoSignal(t.C(), Immediate)

	t.Reset(2 * TestInterval)
	fc.Add(TestInterval)
	suite.requireNoSignal(t.C(), Immediate)

	fc.Add(TestInterval)
	suite.requireSignal(t.C(), Immediate)

	suite.Panics(func() {
		t.Reset(-100 * time.Second)
	})

	suite.requireNoSignal(t.C(), Immediate)
}

func TestFakeTicker(t *testing.T) {
	suite.Run(t, new(FakeTickerSuite))
}
