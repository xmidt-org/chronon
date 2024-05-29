// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type mockListener struct {
	mock.Mock
}

func (m *mockListener) onUpdate(t time.Time) updateResult {
	args := m.Called(t)
	return args.Get(0).(updateResult)
}

func (m *mockListener) ExpectOnUpdate(t time.Time, r updateResult) *mock.Call {
	return m.On("onUpdate", t).Return(r)
}
