package retry_test

import (
	"errors"
	"fmt"
	"time"

	"aqwari.net/retry"
)

func ExampleExponential() {
	backoff := retry.Exponential(time.Hour / 2).Splay(time.Minute * 10)

	for tries := 0; tries < 10; tries++ {
		fmt.Println("Connecting to service")
		b := backoff(tries)
		fmt.Printf("Connection failed; will try again in %s", b)
		time.Sleep(b)
	}
}

func ExampleSeconds() {
	backoff := retry.Seconds(2, 4, 6, 22, 39, 18)

	for tries := 0; tries < 10; tries++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(tries))
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

	for tries := 0; tries < 8; tries++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(tries))
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

func ExampleStrategy_Add() {
	backoff := retry.Exponential(-1).Units(time.Second).Add(time.Minute)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
	}

	// Output: 1m1s
	// 1m2s
	// 1m4s
	// 1m8s
	// 1m16s
	// 1m32s
	// 2m4s
	// 3m8s
	// 5m16s
	// 9m32s
}

func ExampleStrategy_Scale() {
	// Sleep for 2ⁿ milliseconds, not seconds
	backoff := retry.Exponential(-1).Units(time.Second).Scale(1e-3)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
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
	backoff := retry.Exponential(-1).Splay(time.Second / 2)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
	}
}

func ExampleStrategy_Fixed() {
	backoff := retry.Fixed(time.Minute, time.Hour, time.Hour*2)

	for tries := 0; tries < 4; tries++ {
		fmt.Println(backoff(tries))
	}

	// Output: 1m0s
	// 1h0m0s
	// 2h0m0s
	// 2h0m0s
}

func ExampleStrategy_Unshift() {
	backoff := retry.Exponential(-1).Units(time.Second).Unshift(time.Minute)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
	}

	// Output: 1m0s
	// 1s
	// 2s
	// 4s
	// 8s
	// 16s
	// 32s
	// 1m4s
	// 2m8s
	// 4m16s
}

func ExampleStrategy_Shift() {
	backoff := retry.Seconds(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).
		Shift(2)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
	}

	// Output: 3s
	// 4s
	// 5s
	// 6s
	// 7s
	// 8s
	// 9s
	// 10s
	// 10s
	// 10s

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
	backoff := retry.Exponential(3600).
		Units(time.Minute).
		Shift(1).
		Overwrite(time.Hour)
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
