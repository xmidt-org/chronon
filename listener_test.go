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

func (suite *ListenersSuite) TestAdd() {
	ls := new(listeners)

	mock1 := new(mockListener)
	mock1.ExpectOnAdvance(suite.now, false).Times(3)

	ls.add(mock1)
	ls.onAdvance(suite.now) // mock1

	// idempotent
	ls.add(mock1)
	ls.onAdvance(suite.now) // mock1 again

	mock2 := new(mockListener)
	mock2.ExpectOnAdvance(suite.now, false).Once()
	ls.add(mock2)
	ls.onAdvance(suite.now) // mock1 and mock2

	mock1.AssertExpectations(suite.T())
	mock2.AssertExpectations(suite.T())
}

func (suite *ListenersSuite) TestRemove() {
	ls := new(listeners)

	mock1 := new(mockListener)
	mock2 := new(mockListener)
	mock3 := new(mockListener)
	mock4 := new(mockListener)

	mock1.ExpectOnAdvance(suite.now, false).Twice()
	mock2.ExpectOnAdvance(suite.now, false).Twice()
	mock3.ExpectOnAdvance(suite.now, false).Twice()

	// mocks 1, 2, and 3 will be left in
	ls.add(mock1)
	ls.add(mock2)
	ls.remove(mock1)
	ls.remove(mock1) // idempotent

	ls.add(mock3)
	ls.add(mock4)
	ls.remove(mock4)
	ls.remove(mock4) // idempotent

	ls.add(mock1)

	ls.onAdvance(suite.now)
	ls.onAdvance(suite.now) // idempotent, since all return false

	mock1.AssertExpectations(suite.T())
	mock2.AssertExpectations(suite.T())
	mock3.AssertExpectations(suite.T())
	mock4.AssertExpectations(suite.T())
}

func (suite *ListenersSuite) TestOnAdvance() {
	ls := new(listeners)

	mock1 := new(mockListener)
	mock2 := new(mockListener)
	mock3 := new(mockListener)

	mock1.ExpectOnAdvance(suite.now, false).Once()
	mock1.ExpectOnAdvance(suite.now, true).Once()
	mock2.ExpectOnAdvance(suite.now, true).Once()
	mock3.ExpectOnAdvance(suite.now, false).Twice()
	mock3.ExpectOnAdvance(suite.now, true).Once()

	ls.add(mock1)
	ls.add(mock2)
	ls.add(mock3)

	ls.onAdvance(suite.now) // mock2 should have been removed
	ls.onAdvance(suite.now) // mock1 should have been removed
	ls.onAdvance(suite.now) // mock3 should have been removed
	ls.onAdvance(suite.now) // should be empty now

	mock1.AssertExpectations(suite.T())
	mock2.AssertExpectations(suite.T())
	mock3.AssertExpectations(suite.T())
}

func TestListeners(t *testing.T) {
	suite.Run(t, new(ListenersSuite))
}
