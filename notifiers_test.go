package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type NotifiersSuite struct {
	ChrononSuite
}

func (suite *NotifiersSuite) TestAddRemoveNotify() {
	var (
		ch1 = make(chan time.Duration, 1)
		ch2 = make(chan time.Duration, 1)

		n notifiers
	)

	n.notify(TestInterval)

	n.add(ch1)
	n.add(ch2)
	n.notify(TestInterval)
	suite.requireReceiveEqual(ch1, TestInterval, Immediate)
	suite.requireReceiveEqual(ch2, TestInterval, Immediate)

	// idempotent
	n.add(ch2)
	n.notify(TestInterval)
	suite.requireReceiveEqual(ch1, TestInterval, Immediate)
	suite.requireReceiveEqual(ch2, TestInterval, Immediate)

	n.remove(ch1)
	n.notify(TestInterval)
	suite.requireNoSignal(ch1, Immediate)
	suite.requireReceiveEqual(ch2, TestInterval, Immediate)

	n.remove(ch2)
	n.notify(TestInterval)
	suite.requireNoSignal(ch1, Immediate)
	suite.requireNoSignal(ch2, Immediate)
}

func TestNotifiers(t *testing.T) {
	suite.Run(t, new(NotifiersSuite))
}
