package chronon

import "time"

// listener represents anything that can respond to a clock being advanced
type listener interface {
	// onAdvance allows this instance to respond to a clock's current time
	// being updated.  If this instance wants to continue being advanced, this
	// method should return true.  If this method returns false, this instance
	// should expect additional updates.
	//
	// FakeClock always executes listeners under its lock, so that clock time
	// cannot change during callback execution.  Implementations of this method
	// must not attempt to reacquire a FakeClock's lock, as that would result
	// in a deadlock.
	onAdvance(time.Time) bool
}

// listeners is a mutable list of objects which can respond to a clock's time change.
type listeners []listener

// deleteAt removes the listener at the given index.  This method does no bounds checking.
func (ls *listeners) deleteAt(i int) {
	last := len(*ls) - 1
	(*ls)[i], (*ls)[last] = (*ls)[last], nil
	*ls = (*ls)[:last]
}

// onAdvance dispatches an advance event to each listener and removes the
// listeners whose onAdvance method returns true.
func (ls *listeners) onAdvance(t time.Time) {
	var i int
	for i < len(*ls) {
		if (*ls)[i].onAdvance(t) {
			// this callback is finished
			ls.deleteAt(i)
		} else {
			i++
		}
	}
}

// add appends the given listener.  This method is idempotent:  if
// v is already a listener, it will not be added.
func (ls *listeners) add(v listener) {
	// scan for v first.  we're optimized around fast traversal of a
	// small number of listeners, so a simple linear search will suffice.
	for _, l := range *ls {
		if v == l {
			return
		}
	}

	*ls = append(*ls, v)
}

// remove deletes the given listener.  This method is idempotent:
// if v is not present, this method does nothing.  This method also
// handles the edge case where v is present more than once in the slice,
// though in general that should never happen since add() is also
// idempotent.
func (ls *listeners) remove(v listener) {
	var i int
	for i < len(*ls) {
		if (*ls)[i] == v {
			ls.deleteAt(i)
		} else {
			i++
		}
	}
}
