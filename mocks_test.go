package chronon

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type mockListener struct {
	mock.Mock
}

func (m *mockListener) onAdvance(t time.Time) bool {
	args := m.Called(t)
	return args.Bool(0)
}

func (m *mockListener) ExpectOnAdvance(t time.Time, v bool) *mock.Call {
	return m.On("onAdvance", t).Return(v)
}
