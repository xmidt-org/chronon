package chronon

import (
	"context"
)

type contextKey struct{}

// Get obtains a Clock from a context.  This function always returns a
// valid Clock.  If no clock is associated with the given context,
// the SystemClock is returned.
func Get(ctx context.Context) Clock {
	if c, ok := ctx.Value(contextKey{}).(Clock); ok {
		return c
	}

	return SystemClock()
}

// With attaches a Clock to a context.  If c is nil, the original
// context is returned.
func With(ctx context.Context, c Clock) context.Context {
	if c == nil {
		// slight optimization:
		// Get will return the system clock if no clock is set,
		// so we can avoid creating a child context in this case
		return ctx
	}

	return context.WithValue(ctx, contextKey{}, c)
}
