package chronon

import "time"

type updateResult int

const (
	continueUpdates updateResult = iota
	stopUpdates
)

// listener represents anything that can respond to a clock's time being updated
type listener interface {
	// onUpdate allows this instance to respond to a clock's current time
	// being updated.  The return value from this method indicates whether
	// this listener will continue receiving updates.
	//
	// FakeClock always executes listeners under its lock, so that clock time
	// cannot change during callback execution.  Implementations of this method
	// must not attempt to reacquire a FakeClock's lock, as that would result
	// in a deadlock.
	onUpdate(time.Time) updateResult
}

// listeners is a set of listener instances that react to a containing
// fake clock's time updates.
type listeners map[listener]bool

// onUpdate dispatches an advance event to each listener and removes the
// listeners whose onUpdate method returns stopUpdates.
func (ls *listeners) onUpdate(t time.Time) {
	for l := range *ls {
		if l.onUpdate(t) == stopUpdates {
			delete(*ls, l)
		}
	}
}

// register performs initialization for the given listener by invoking
// it's onAdvance method for the first time.  If the listener returns continueUpdates,
// it will be added to this listeners so that it receives future updates.
func (ls *listeners) register(t time.Time, v listener) {
	if *ls == nil {
		*ls = make(listeners)
	}

	if v.onUpdate(t) == continueUpdates {
		(*ls)[v] = true
	}
}

// add inserts the given listener but performs no other initialization.  This method
// is appropriate for listeners that need to add themselves again, such as after a Reset.
func (ls *listeners) add(v listener) {
	if *ls == nil {
		*ls = make(listeners)
	}

	(*ls)[v] = true
}

// remove deletes the given listener.
func (ls *listeners) remove(v listener) {
	if *ls != nil {
		delete(*ls, v)
	}
}
