// SPDX-FileCopyrightText: 2024 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package chronon

import "time"

// equalOrAfter returns true if a >= b as defined
// by time.Time.Equal and time.Time.After.  Essentially,
// this function tests if updating a fake clock to time a
// should trigger something waiting until time b.
func equalOrAfter(a, b time.Time) bool {
	return a.Equal(b) || a.After(b)
}

// sendTime does a nonblocking send of a given time on a time channel.
// Used by both fake timers and tickers to avoid deadlocks with slow clients.
func sendTime(c chan<- time.Time, t time.Time) {
	select {
	case c <- t:
	default:
	}
}
