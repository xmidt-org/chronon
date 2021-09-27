package chronon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SleeperSuite struct {
	ChannelSuite

	now    time.Time
	wakeup time.Time
	after  time.Time
}

func (suite *SleeperSuite) SetupSuite() {
	// create some convenient times for testing
	suite.now = time.Now()
	suite.wakeup = suite.now.Add(time.Second)
	suite.after = suite.wakeup.Add(time.Second)
}

func (suite *SleeperSuite) TestOnAdvance() {
	suite.Run("Exact", func() {
		s := newSleeperAt(
			suite.wakeup,
		)

		suite.Require().NotNil(s)
		suite.requireNoSignal(s.awaken)

		// idempotent
		suite.False(s.onAdvance(suite.now))
		suite.requireNoSignal(s.awaken)

		// idempotent
		suite.False(s.onAdvance(suite.now))
		suite.requireNoSignal(s.awaken)

		// wakeup using the exact time value
		suite.True(s.onAdvance(suite.wakeup))
		suite.requireSignal(s.awaken)

		// idempotent
		suite.True(s.onAdvance(suite.wakeup))
		suite.requireSignal(s.awaken)

		// idempotent
		suite.True(s.onAdvance(suite.after))
		suite.requireSignal(s.awaken)

		// idempotent
		suite.False(s.onAdvance(suite.now))
		suite.requireSignal(s.awaken)

	})

	suite.Run("After", func() {
		s := newSleeperAt(
			suite.wakeup,
		)

		suite.Require().NotNil(s)
		suite.requireNoSignal(s.awaken)

		// idempotent
		suite.False(s.onAdvance(suite.now))
		suite.requireNoSignal(s.awaken)

		// idempotent
		suite.False(s.onAdvance(suite.now))
		suite.requireNoSignal(s.awaken)

		// wakeup using a value after the time value
		suite.True(s.onAdvance(suite.after))
		suite.requireSignal(s.awaken)

		// idempotent
		suite.True(s.onAdvance(suite.wakeup))
		suite.requireSignal(s.awaken)

		// idempotent
		suite.True(s.onAdvance(suite.after))
		suite.requireSignal(s.awaken)

		// idempotent
		suite.False(s.onAdvance(suite.now))
		suite.requireSignal(s.awaken)

	})
}

func (suite *SleeperSuite) TestWait() {
	s := newSleeperAt(
		suite.wakeup,
	)

	suite.Require().NotNil(s)

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.wait()
	}()

	suite.False(s.onAdvance(suite.now))

	select {
	case <-done:
		suite.Require().Fail("The waiting goroutine should not have exited")
	default:
		// passing
	}

	suite.True(s.onAdvance(suite.wakeup))

	select {
	case <-done:
		// passing
	case <-time.After(100 * time.Millisecond):
		suite.Require().Fail("The waiting goroutine should have exited")
	}
}

func TestSleeper(t *testing.T) {
	suite.Run(t, new(SleeperSuite))
}
