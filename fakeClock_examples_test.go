package chronon

import (
	"context"
	"fmt"
	"time"
)

func ExampleFakeClock_Sleep() {
	fc := NewFakeClock(time.Now())

	// to coordinate with another goroutine, use NotifyOnSleep
	onSleep := make(chan time.Duration)
	fc.NotifyOnSleep(onSleep)

	// spawn our code under test in a separate goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		fc.Sleep(5 * time.Second)
		fmt.Println("code under test has awakened")
	}()

	// ensure that production code has reached the point where it's sleeping
	d := <-onSleep
	fmt.Println("code under test is now sleeping for", d)

	// wakeup our production code
	fc.Add(d)

	// ensure that the goroutine finishes before we exit
	<-done

	// Output:
	// code under test is now sleeping for 5s
	// code under test has awakened
}

func ExampleFakeClock_NewTimer() {
	start := time.Now()
	fc := NewFakeClock(start)

	// to coordinate with another goroutine, use NotifyOnTimer
	onTimer := make(chan time.Duration)
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
	d := <-onTimer
	fmt.Println("code under test is now waiting for", d)

	// the clock can be advanced less than the timer interval with no effect
	fc.Add(d / 3)

	// the clock can also be turned backward, again with no effect
	fc.Add(-2 * time.Hour)

	// force our code under test to stop waiting
	fc.Set(start.Add(d))

	// ensure that the goroutine finishes before we exit
	<-done

	// Output:
	// code under test is now waiting for 10m0s
	// code under test is no longer waiting
}

func ExampleFakeClock_NewTicker() {
	start := time.Now()
	fc := NewFakeClock(start)

	// to coordinate with another goroutine, use NotifyOnTicker
	onTicker := make(chan time.Duration)
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
	d := <-onTicker
	fmt.Println("code under test is now waiting for ticks on", d)

	// force a tick by advancing the clock by the tick interval
	fc.Add(d)
	<-receivedTick

	// can also force a tick by setting absolute time in the future
	fc.Set(fc.Now().Add(d))
	<-receivedTick

	// moving the clock backwards doesn't result in ticks
	fc.Add(-d)

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
