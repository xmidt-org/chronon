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
	WaitALittle time.Duration = 100 * time.Millisecond
)

type ChannelSuite struct {
	suite.Suite
}

func (suite *ChannelSuite) selectOn(ch interface{}, waitFor time.Duration) (int, reflect.Value, bool) {
	cases := []reflect.SelectCase{
		reflect.SelectCase{
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

func (suite *ChannelSuite) requireReceive(ch interface{}, waitFor time.Duration) interface{} {
	chosen, value, recvOK := suite.selectOn(ch, waitFor)
	suite.Require().Truef(
		chosen == 0 && recvOK,
		"Nothing received on channel [%T]",
		ch,
	)

	return value.Interface()
}

func (suite *ChannelSuite) requireReceiveEqual(ch, expected interface{}, waitFor time.Duration) {
	suite.Equal(expected, suite.requireReceive(ch, waitFor))
}

func (suite *ChannelSuite) requireSignal(ch interface{}, waitFor time.Duration) {
	chosen, _, _ := suite.selectOn(ch, waitFor)
	suite.Require().Truef(
		chosen == 0,
		"The channel [%T] should have been signalled",
		ch,
	)
}

func (suite *ChannelSuite) requireNoSignal(ch interface{}, waitFor time.Duration) {
	chosen, _, _ := suite.selectOn(ch, waitFor)
	suite.Require().Truef(
		chosen == 1,
		"The channel [%T] should NOT have been signalled",
		ch,
	)
}
