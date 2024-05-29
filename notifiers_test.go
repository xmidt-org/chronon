// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NotifiersSuite struct {
	ChrononSuite
}

func (suite *NotifiersSuite) TestAddRemove() {
	var (
		ch1 = make(chan int, 1)
		ch2 = make(chan int, 1)
		ch3 = make(chan int, 1)

		ns notifiers
	)

	ns.notify(-1)
	ns.remove(ch1) // should be a noop
	ns.notify(-2)

	ns.add(ch1)
	ns.notify(10)
	suite.requireReceiveEqual(ch1, 10, Immediate)

	ns.add(ch2)
	ns.add(ch3)
	ns.notify(20)
	suite.requireReceiveEqual(ch1, 20, Immediate)
	suite.requireReceiveEqual(ch2, 20, Immediate)
	suite.requireReceiveEqual(ch3, 20, Immediate)

	ns.remove(ch2)
	ns.notify(30)
	suite.requireReceiveEqual(ch1, 30, Immediate)
	suite.requireNoSignal(ch2, Immediate)
	suite.requireReceiveEqual(ch3, 30, Immediate)

	ns.remove(ch1)
	ns.remove(ch3)
	ns.notify(40)
	suite.requireNoSignal(ch1, Immediate)
	suite.requireNoSignal(ch2, Immediate)
	suite.requireNoSignal(ch3, Immediate)
}

func TestNotifiers(t *testing.T) {
	suite.Run(t, new(NotifiersSuite))
}
