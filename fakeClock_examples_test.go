package chronon

import (
	"context"
	"fmt"
	"time"
)

func ExampleFakeClock_Sleep() {
	fc := NewFakeClock(time.Now())

	// to coordinate with another goroutine, use NotifyOnSleep
	onSleep := make(chan Sleeper)
	fc.NotifyOnSleep(onSleep)

	// spawn our code under test in a separate goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		fc.Sleep(5 * time.Second)
		fmt.Println("code under test has awakened")
	}()

	// ensure that production code has reached the point where it's sleeping
	s := <-onSleep
	fmt.Println("code under test is now sleeping for", fc.Until(s.When()))

	// wakeup our production code
	fc.Add(5 * time.Second)

	// ensure that the goroutine finishes before we exit
	<-done

	// Output:
	// code under test is now sleeping for 5s
	// code under test has awakened
}

func ExampleFakeClock_NewTimer() {
	fc := NewFakeClock(time.Now())

	// to coordinate with another goroutine, use NotifyOnTimer
	onTimer := make(chan FakeTimer)
	fc.NotifyOnTimer(onTimer)

	// spawn our code under test in a separate goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		t := fc.NewTimer(10 * time.Minute)
		<-t.C()
		fmt.Println("code under test is no longer waiting")
	}()

	// ensure that production code has reached the point where it's
	// waiting on a timer
	ft := <-onTimer
	fmt.Println("code under test is now waiting for", fc.Until(ft.When()))

	// the clock can be advanced less than the timer interval with no effect
	fc.Add(time.Millisecond)

	// the clock can also be turned backward, again with no effect
	fc.Add(-2 * time.Hour)

	// force our code under test to stop waiting
	ft.Fire()

	// ensure that the goroutine finishes before we exit
	<-done

	// Output:
	// code under test is now waiting for 10m0s
	// code under test is no longer waiting
}

func ExampleFakeClock_NewTicker() {
	fc := NewFakeClock(time.Now())

	// to coordinate with another goroutine, use NotifyOnTicker
	onTicker := make(chan FakeTicker)
	fc.NotifyOnTicker(onTicker)

	// create a context to control termination of our code under test
	ctx, cancel := context.WithCancel(context.Background())

	// spawn our code under test in a separate goroutine
	done := make(chan struct{})
	receivedTick := make(chan struct{})
	go func() {
		defer close(done)

		t := fc.NewTicker(20 * time.Second)
		for {
			select {
			case <-t.C():
				fmt.Println("tick")
				receivedTick <- struct{}{}
			case <-ctx.Done():
				fmt.Println("code under test has been cancelled")
				return
			}
		}
	}()

	// ensure that production code has reached the point where it's
	// waiting on a ticker
	ft := <-onTicker
	fmt.Println("code under test is now waiting for ticks on", fc.Until(ft.When()))

	// force a tick by advancing the clock
	fc.Set(ft.When())
	<-receivedTick

	// can also force a tick directly
	ft.Fire()
	<-receivedTick

	// moving the clock backwards doesn't result in ticks
	fc.Add(-time.Hour)

	// all done testing
	cancel()

	// ensure that the goroutine finishes before we exit
	<-done

	// Output:
	// code under test is now waiting for ticks on 20s
	// tick
	// tick
	// code under test has been cancelled
}
