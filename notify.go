package chronon

import "time"

// notifiers holds a slice of channels that will receive durations,
// such as from sleepers.
type notifiers []chan<- time.Duration

// deleteAt removes the channel at the given index.  This method does no bounds checking.
func (n *notifiers) deleteAt(i int) {
	last := len(*n) - 1
	(*n)[i], (*n)[last] = (*n)[last], nil
	*n = (*n)[:last]
}

// notify dispatches the given duration to each channel in this slice.
func (n notifiers) notify(d time.Duration) {
	for _, ch := range n {
		ch <- d
	}
}

// add appends the given channel to this slice.  This method
// is idempotent: if v is already in this slice, this method
// does nothing.
func (n *notifiers) add(v chan<- time.Duration) {
	for _, ch := range *n {
		if ch == v {
			return
		}
	}

	*n = append(*n, v)
}

// remove deletes the channel from this slice.  This method
// is idempotent: if v is not in this slice, this method
// does nothing.
func (n *notifiers) remove(v chan<- time.Duration) {
	var i int
	for i < len(*n) {
		if (*n)[i] == v {
			n.deleteAt(i)
		} else {
			i++
		}
	}
}
