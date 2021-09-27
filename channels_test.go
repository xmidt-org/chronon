package chronon

import (
	"reflect"
	"time"

	"github.com/stretchr/testify/suite"
)

type ChannelSuite struct {
	suite.Suite
}

func (suite *ChannelSuite) requireReceive(ch interface{}) interface{} {
	t := time.NewTimer(100 * time.Millisecond)
	defer t.Stop()

	chosen, value, recvOK := reflect.Select([]reflect.SelectCase{
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		},
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(t.C),
		},
	})

	suite.Require().Truef(
		chosen == 0 && recvOK,
		"Nothing received on channel [%T]",
		ch,
	)

	return value.Interface()
}

func (suite *ChannelSuite) requireReceiveEqual(ch, expected interface{}) {
	suite.Equal(expected, suite.requireReceive(ch))
}

func (suite *ChannelSuite) requireSignal(ch interface{}) {
	t := time.NewTimer(100 * time.Millisecond)
	defer t.Stop()

	chosen, _, _ := reflect.Select([]reflect.SelectCase{
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		},
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(t.C),
		},
	})

	suite.Require().Truef(
		chosen == 0,
		"The channel [%T] should have been signalled",
		ch,
	)
}

func (suite *ChannelSuite) requireNoSignal(ch interface{}) {
	t := time.NewTimer(100 * time.Millisecond)
	defer t.Stop()

	chosen, _, _ := reflect.Select([]reflect.SelectCase{
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		},
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(t.C),
		},
	})

	suite.Require().Truef(
		chosen == 1,
		"The channel [%T] should NOT have been signalled",
		ch,
	)
}
