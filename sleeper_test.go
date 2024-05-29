// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SleeperSuite struct {
	ChrononSuite

	wakeup time.Time
	after  time.Time
}

func (suite *SleeperSuite) SetupSuite() {
	suite.ChrononSuite.SetupSuite()
	suite.wakeup = suite.now.Add(time.Second)
	suite.after = suite.wakeup.Add(time.Second)
}

func (suite *SleeperSuite) Immediate() {
	for _, interval := range []time.Duration{-TestInterval, 0} {
		suite.Run(interval.String(), func() {
			s, fc, done := suite.newSleeper(interval)
			suite.requireSignal(done, Immediate)
			suite.Equal(-interval, fc.Since(s.When()))

			suite.False(s.Wakeup())
		})
	}
}

func (suite *SleeperSuite) TestWakeup() {
	s, fc, done := suite.newSleeper(TestInterval)
	suite.requireNoSignal(done, Immediate)
	suite.Equal(TestInterval, fc.Until(s.When()))

	suite.True(s.Wakeup())
	suite.requireSignal(done, WaitALittle)

	suite.False(s.Wakeup())
}

func (suite *SleeperSuite) TestAdd() {
	s, fc, done := suite.newSleeper(TestInterval)
	suite.requireNoSignal(done, Immediate)
	suite.Equal(TestInterval, fc.Until(s.When()))

	fc.Add(TestInterval / 2)
	suite.requireNoSignal(done, WaitALittle)

	fc.Add(TestInterval / 2)
	suite.requireSignal(done, WaitALittle)

	suite.False(s.Wakeup())
}

func (suite *SleeperSuite) TestSet() {
	s, fc, done := suite.newSleeper(TestInterval)
	suite.requireNoSignal(done, Immediate)
	suite.Equal(TestInterval, fc.Until(s.When()))

	fc.Set(s.When().Add(-time.Hour))
	suite.requireNoSignal(done, WaitALittle)

	fc.Set(s.When())
	suite.requireSignal(done, WaitALittle)

	suite.False(s.Wakeup())
}

func TestSleeper(t *testing.T) {
	suite.Run(t, new(SleeperSuite))
}
