package retry_test

import (
	"errors"
	"fmt"
	"time"

	"aqwari.net/retry"
)

func ExampleExponential() {
	backoff := retry.Exponential(time.Second)

	for try := 0; try < 10; try++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(try))
	}

	// Output: Connection failed; will try again in 1s
	// Connection failed; will try again in 2s
	// Connection failed; will try again in 4s
	// Connection failed; will try again in 8s
	// Connection failed; will try again in 16s
	// Connection failed; will try again in 32s
	// Connection failed; will try again in 1m4s
	// Connection failed; will try again in 2m8s
	// Connection failed; will try again in 4m16s
	// Connection failed; will try again in 8m32s
}

func ExampleSeconds() {
	backoff := retry.Seconds(2, 4, 6, 22, 39, 18)

	for try := 0; try < 10; try++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(try))
	}
	// Output: Connection failed; will try again in 2s
	// Connection failed; will try again in 4s
	// Connection failed; will try again in 6s
	// Connection failed; will try again in 22s
	// Connection failed; will try again in 39s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
}

func ExampleMilliseconds() {
	backoff := retry.Milliseconds(2, 4, 6, 22, 39, 18)

	for try := 0; try < 8; try++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(try))
	}
	// Output: Connection failed; will try again in 2ms
	// Connection failed; will try again in 4ms
	// Connection failed; will try again in 6ms
	// Connection failed; will try again in 22ms
	// Connection failed; will try again in 39ms
	// Connection failed; will try again in 18ms
	// Connection failed; will try again in 18ms
	// Connection failed; will try again in 18ms
}

func ExampleIntervals() {
	backoff := retry.Intervals(time.Minute, time.Hour, time.Hour*2)

	for try := 0; try < 4; try++ {
		fmt.Println(backoff(try))
	}
	// Output: 1m0s
	// 1h0m0s
	// 2h0m0s
	// 2h0m0s
}

func ExampleIntervals_scaled() {
	backoff := retry.Intervals(1, 2, 3, 4, 5, 6, 7).Scale(time.Microsecond)

	for try := 0; try < 7; try++ {
		fmt.Println(backoff(try))
	}
	// Output: 1µs
	// 2µs
	// 3µs
	// 4µs
	// 5µs
	// 6µs
	// 7µs
}

func ExampleStrategy_Scale() {
	// Sleep for 2ⁿ milliseconds, not nanoseconds
	backoff := retry.Exponential(1).Scale(time.Millisecond)

	for try := 0; try < 10; try++ {
		fmt.Println(backoff(try))
	}

	// Output: 1ms
	// 2ms
	// 4ms
	// 8ms
	// 16ms
	// 32ms
	// 64ms
	// 128ms
	// 256ms
	// 512ms
}

func ExampleStrategy_Splay() {
	backoff := retry.Exponential(time.Second).Splay(time.Second / 2)

	for try := 0; try < 10; try++ {
		fmt.Println(backoff(try))
	}
}

func ExampleStrategy_Intervals() {
	backoff := retry.Intervals(time.Minute, time.Hour, time.Hour*2)

	for try := 0; try < 4; try++ {
		fmt.Println(backoff(try))
	}

	// Output: 1m0s
	// 1h0m0s
	// 2h0m0s
	// 2h0m0s
}

func ExampleStrategy_Unshift() {
	backoff := retry.Exponential(time.Minute).
		Unshift(time.Hour)

	for try := 0; try < 10; try++ {
		fmt.Printf("%d: %s\n", try, backoff(try))
	}

	// Output: 0: 1h0m0s
	// 1: 1m0s
	// 2: 2m0s
	// 3: 4m0s
	// 4: 8m0s
	// 5: 16m0s
	// 6: 32m0s
	// 7: 1h4m0s
	// 8: 2h8m0s
	// 9: 4h16m0s
}

func ExampleStrategy_Shift() {
	backoff := retry.Seconds(1, 2, 3, 4, 5, 6, 7).Shift(2)

	for try := 0; try < 10; try++ {
		fmt.Println(backoff(try))
	}

	// Output: 3s
	// 4s
	// 5s
	// 6s
	// 7s
	// 7s
	// 7s
	// 7s
	// 7s
	// 7s
}

func Example() {
	// mock for an unreliable remote service
	getdump := func(i int) error {
		if i%3 == 0 {
			return nil
		}
		return errors.New("remote call failed")
	}
	// Request a dump from a service every hour. If something goes
	// wrong, retry on lengthening intervals until we get a response,
	// then go back to per-hour dumps.
	backoff := retry.Exponential(time.Minute).Shift(2).
		Unshift(time.Hour)
	try := 0

	for i := 0; i < 7; i++ {
		if err := getdump(i); err != nil {
			try++
			fmt.Println(err)
		} else {
			try = 0
			fmt.Println("success")
		}
		fmt.Printf("sleeping %s\n", backoff(try))
	}

	// Output: success
	// sleeping 1h0m0s
	// remote call failed
	// sleeping 4m0s
	// remote call failed
	// sleeping 8m0s
	// success
	// sleeping 1h0m0s
	// remote call failed
	// sleeping 4m0s
	// remote call failed
	// sleeping 8m0s
	// success
	// sleeping 1h0m0s
}
