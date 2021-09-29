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

func TestSleeper(t *testing.T) {
	suite.Run(t, new(SleeperSuite))
}
