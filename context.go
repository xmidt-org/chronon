package chronon

import (
	"context"
	"time"
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

// AfterCtx executes a function after a given interval.  A cancellable child context is
// created and passed to this function.  The Clock used to create the Timer is extracted
// from the parent context using Get.
//
// A goroutine is spawned to execute f.  This goroutine waits for either context cancellation
// or the internal timer to fire.  In both cases, the function is called.  The ctx.Err() function can be
// used to determine if the context was cancelled.  This allows the function to do any necessary
// cleanup.
//
// This function does not return the internal Timer, since the returned CancelFunc can be used to abort
// execution.
func AfterCtx(parentCtx context.Context, d time.Duration, f func(context.Context)) context.CancelFunc {
	var (
		t           = Get(parentCtx).NewTimer(d)
		ctx, cancel = context.WithCancel(parentCtx)
	)

	go func() {
		// ensures that any abnormal exits, like panics,
		// cancel the context
		defer cancel()

		select {
		case <-t.C():
			f(ctx)

		case <-ctx.Done():
			f(ctx) // allow for cleanup
		}
	}()

	return cancel
}

// TickCtx executes a function on an interval.  A cancellable child context is created
// for all executions.  The Clock used to create the Ticker is extracted from the parent
// context using Get.
//
// A goroutine is spawned to execute f.  This goroutine can be shutdown completely by
// calling the returned CancelFunc.  Alternatively, stopping the returned ticker will pause
// executions, and resetting the ticker will resume executions.
//
// The function is passed the child context on each invocation.  This function is invoked
// both (1) when a tick occurs, and (2) when the context is cancelled.  The ctx.Err() value
// can be used to disambiguate those cases.  This allows the function to do necessary
// cleanup as part of context cancellation.
func TickCtx(parentCtx context.Context, d time.Duration, f func(context.Context)) (context.CancelFunc, Ticker) {
	var (
		t           = Get(parentCtx).NewTicker(d)
		ctx, cancel = context.WithCancel(parentCtx)
	)

	go func() {
		// ensures that any abnormal exits, like panics,
		// cancel the context
		defer cancel()

		for {
			select {
			case <-t.C():
				f(ctx)

			case <-ctx.Done():
				f(ctx) // let the function do any cleanup
				return
			}
		}
	}()

	return cancel, t
}
