// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ListenersSuite struct {
	suite.Suite

	now time.Time
}

func (suite *ListenersSuite) SetupSuite() {
	suite.now = time.Now()
}

func (suite *ListenersSuite) TestAddRemove() {
	var (
		mock1 = new(mockListener)
		mock2 = new(mockListener)
		mock3 = new(mockListener)
		mock4 = new(mockListener)

		ls = new(listeners)
	)

	mock1.ExpectOnUpdate(suite.now, continueUpdates).Twice()
	mock2.ExpectOnUpdate(suite.now, continueUpdates).Twice()
	mock3.ExpectOnUpdate(suite.now, continueUpdates).Twice()

	// mocks 1, 2, and 3 will be left in
	ls.add(mock1)
	ls.add(mock1) // idempotent
	ls.add(mock2)
	ls.remove(mock1)
	ls.remove(mock1) // idempotent

	ls.add(mock3)
	ls.add(mock4)
	ls.remove(mock4)
	ls.remove(mock4) // idempotent

	ls.add(mock1)

	ls.onUpdate(suite.now)
	ls.onUpdate(suite.now) // idempotent, since all return continueUpdates

	mock1.AssertExpectations(suite.T())
	mock2.AssertExpectations(suite.T())
	mock3.AssertExpectations(suite.T())
	mock4.AssertExpectations(suite.T())
}

func (suite *ListenersSuite) TestOnUpdate() {
	ls := new(listeners)

	mock1 := new(mockListener)
	mock2 := new(mockListener)
	mock3 := new(mockListener)

	mock1.ExpectOnUpdate(suite.now, continueUpdates).Once()
	mock1.ExpectOnUpdate(suite.now, stopUpdates).Once()
	mock2.ExpectOnUpdate(suite.now, stopUpdates).Once()
	mock3.ExpectOnUpdate(suite.now, continueUpdates).Twice()
	mock3.ExpectOnUpdate(suite.now, stopUpdates).Once()

	ls.add(mock1)
	ls.add(mock2)
	ls.add(mock3)

	ls.onUpdate(suite.now) // mock2 should have been removed
	ls.onUpdate(suite.now) // mock1 should have been removed
	ls.onUpdate(suite.now) // mock3 should have been removed
	ls.onUpdate(suite.now) // should be empty now

	mock1.AssertExpectations(suite.T())
	mock2.AssertExpectations(suite.T())
	mock3.AssertExpectations(suite.T())
}

func TestListeners(t *testing.T) {
	suite.Run(t, new(ListenersSuite))
}
