package chronon

import (
	"reflect"
)

// notifiers is a set of channels.
type notifiers map[reflect.Value]bool

// notify sends e to all channels in this set.
func (n notifiers) notify(e interface{}) {
	ev := reflect.ValueOf(e)
	for ch := range n {
		ch.Send(ev)
	}
}

// add inserts a new channel into this map.
func (n *notifiers) add(ch interface{}) {
	if *n == nil {
		*n = make(notifiers, 1)
	}

	(*n)[reflect.ValueOf(ch)] = true
}

// remove deletes a channel from this map.
func (n *notifiers) remove(ch interface{}) {
	if *n == nil {
		return
	}

	delete(*n, reflect.ValueOf(ch))
}
